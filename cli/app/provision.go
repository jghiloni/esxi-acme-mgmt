package app

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jghiloni/esxi-acme-mgmt/plugins/common"
	"github.com/mholt/acmez/v3"
	"github.com/mholt/acmez/v3/acme"
)

const (
	privateKeyFile = "rui.key"
	certFile       = "rui.crt"
	castore        = "castore.pem"
)

type ProvisionCommand struct {
	configDir         string
	certsDir          string
	pluginDir         string
	outputDir         string
	runDir            string
	dnsProviderName   string
	acmeURL           string
	accountEmail      string
	createAccount     bool
	accountPrivateKey crypto.Signer
	certPrivateKey    crypto.Signer
}

func (s *ProvisionCommand) BeforeApply(opts RunOptions) error {
	s.configDir = filepath.Join(opts.BaseDir, ".config")
	s.certsDir = filepath.Join(opts.BaseDir, "certs")
	s.runDir = filepath.Join(opts.BaseDir, "run")
	s.pluginDir = opts.PluginsDir
	s.outputDir = opts.TargetDirectory
	s.dnsProviderName = opts.Provider
	s.acmeURL = opts.ACMEDirectoryURL
	s.accountEmail = opts.AccountEmail

	err := os.MkdirAll(s.configDir, 0o700)
	if err != nil {
		return fmt.Errorf("could not ensure config directory exists: %w", err)
	}

	if err = os.MkdirAll(s.certsDir, 0o755); err != nil {
		return fmt.Errorf("could not ensure certs directory exists: %w", err)
	}

	if err = os.MkdirAll(s.runDir, 0o700); err != nil {
		return fmt.Errorf("could not ensure run directory exists: %w", err)
	}

	s.accountPrivateKey, s.createAccount, err = s.readOrCreatePrivateKey(filepath.Join(s.configDir, "acme.apk"), elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("get account private key: %w", err)
	}

	s.certPrivateKey, _, err = s.readOrCreatePrivateKey(filepath.Join(s.configDir, "acme.cpk"), elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("get cert private key: %w", err)
	}

	return nil
}

func (s *ProvisionCommand) Run(ctx context.Context, providerArgs []string) error {
	pidFile, err := os.OpenFile(filepath.Join(s.runDir, "pid"), os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, fs.ErrExist) {
		return errors.New("a current start command is currently in progress")
	}
	defer func() {
		os.Remove(pidFile.Name())
	}()
	fmt.Fprintf(pidFile, "%d", os.Getpid())
	pidFile.Close()

	needsRenewal, err := s.checkIfCertNeedsRenewal(ctx)
	if err != nil {
		return err
	}

	if !needsRenewal {
		slog.Info("no certs currently need renewal")
		return nil
	}

	account := acme.Account{
		Contact:              []string{fmt.Sprintf("mailto:%s", s.accountEmail)},
		TermsOfServiceAgreed: true,
		PrivateKey:           s.accountPrivateKey,
	}

	solver, err := common.LoadProvider(s.pluginDir, s.dnsProviderName, providerArgs)
	if err != nil {
		return fmt.Errorf("could not load DNS solver plugin for provider %s: %w", s.dnsProviderName, err)
	}

	client := acmez.Client{
		Client: &acme.Client{
			Directory:   s.acmeURL,
			Logger:      slog.Default(),
			UserAgent:   Name,
			PollTimeout: time.Minute,
		},
		ChallengeSolvers: map[string]acmez.Solver{
			acme.ChallengeTypeDNS01: solver,
		},
	}

	if s.createAccount {
		account, err = client.NewAccount(ctx, account)
		if err != nil {
			return fmt.Errorf("could not create ACME account: %w", err)
		}
	}

	fqdn, err := s.getLocalFQDN()
	if err != nil {
		return fmt.Errorf("could not get local FQDN: %w", err)
	}

	certs, err := client.ObtainCertificateForSANs(ctx, account, s.certPrivateKey, []string{fqdn})
	if err != nil {
		return fmt.Errorf("could not get certs from ACME server: %w", err)
	}

	return s.replaceActiveKey(certs)
}

func (*ProvisionCommand) readOrCreatePrivateKey(apkPath string, curve elliptic.Curve, r io.Reader) (crypto.Signer, bool, error) {
	apkBytes, err := os.ReadFile(apkPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, false, fmt.Errorf("error loading private key: %w", err)
	}

	var pk *ecdsa.PrivateKey
	switch apkBytes {
	case nil:
		pk, err = ecdsa.GenerateKey(curve, r)
		if err != nil {
			return nil, false, fmt.Errorf("could not generate new private key for account: %w", err)
		}

		x509Encoded, _ := x509.MarshalECPrivateKey(pk)
		pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: x509Encoded})

		return pk, true, os.WriteFile(apkPath, pemEncoded, 0o400)
	default:
		block, _ := pem.Decode(apkBytes)
		x509Encoded := block.Bytes
		pk, err = x509.ParseECPrivateKey(x509Encoded)
		return pk, false, err
	}
}

type acmeCertWithPath struct {
	acme.Certificate
	PEMPath string `json:"pemPath"`
}

func (s *ProvisionCommand) checkIfCertNeedsRenewal(context.Context) (bool, error) {
	// if the original cert is not a symlink, then it's still the original cert
	// and we need to replace it
	certPath := filepath.Join(s.outputDir, certFile)
	fi, err := os.Lstat(certPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return false, err
	}

	if fi.Mode()&os.ModeSymlink == 0 {
		return true, nil
	}

	linkPath, err := os.Readlink(certPath)
	if err != nil {
		return false, err
	}

	jsonPath := strings.TrimSuffix(linkPath, ".pem") + ".json"
	fp, err := os.Open(jsonPath)
	if err != nil {
		return false, err
	}
	defer fp.Close()

	var c acmeCertWithPath
	if err = json.NewDecoder(fp).Decode(&c); err != nil {
		return false, err
	}

	if c.PEMPath != linkPath {
		return false, fmt.Errorf("expected cert path %s does not match stored path %s", linkPath, c.PEMPath)
	}

	if c.RenewalInfo == nil {
		return false, errors.New("cert info missing renewal info")
	}

	windowStart := c.RenewalInfo.SuggestedWindow.Start.Add(-time.Nanosecond)
	windowEnd := c.RenewalInfo.SuggestedWindow.End.Add(time.Hour)

	now := time.Now()
	return now.After(windowStart) && now.Before(windowEnd), nil
}

func (s *ProvisionCommand) replaceActiveKey(certs []acme.Certificate) error {
	// there should only be one here, but ¯\_(ツ)_/¯
	chainPEM := &bytes.Buffer{}
	augmentedCerts := make([]acmeCertWithPath, 0, len(certs))
	for _, cert := range certs {
		var certsDER [][]byte

		sum := sha256.Sum256(cert.ChainPEM)
		filename := hex.EncodeToString(sum[:])

		chainBytes := cert.ChainPEM
		for {
			var block *pem.Block
			block, chainBytes = pem.Decode(chainBytes)
			if block != nil && block.Type == "CERTIFICATE" {
				_, err := x509.ParseCertificate(block.Bytes)
				if err != nil {
					return err
				}

				certsDER = append(certsDER, block.Bytes)
				continue
			}

			break
		}

		if len(certsDER) > 0 {
			certFP, err := os.Create(filepath.Join(s.certsDir, filename+".pem"))
			if err != nil {
				return err
			}
			b := &pem.Block{Bytes: certsDER[0], Type: "CERTIFICATE"}

			err = pem.Encode(certFP, b)
			certFP.Close()
			if err != nil {
				return err
			}

			for _, der := range certsDER[1:] {
				block := &pem.Block{Bytes: der, Type: "CERTIFICATE"}
				if err = pem.Encode(chainPEM, block); err != nil {
					return err
				}
				fmt.Fprintln(chainPEM)
			}
		}

		augmentedCert := acmeCertWithPath{
			Certificate: cert,
			PEMPath:     filepath.Join(s.certsDir, filename+".pem"),
		}

		jsonFilename := filepath.Join(s.certsDir, filename+".json")
		jfp, err := os.Create(jsonFilename)
		if err != nil {
			return err
		}
		err = json.NewEncoder(jfp).Encode(augmentedCert)
		jfp.Close()
		if err != nil {
			return err
		}

		augmentedCerts = append(augmentedCerts, augmentedCert)
	}

	if len(augmentedCerts) == 0 {
		return errors.New("no certificates were generated")
	}

	return s.backupAndResetActiveTLSFiles(augmentedCerts[0], chainPEM.Bytes())
}

func (*ProvisionCommand) getLocalFQDN() (string, error) {
	cmd := exec.Command("/bin/hostname", "-f")

	out := &strings.Builder{}
	cmd.Stdout = out

	err := cmd.Run()
	return strings.TrimSpace(out.String()), err
}

func (s *ProvisionCommand) backupAndResetActiveTLSFiles(cert acmeCertWithPath, caChain []byte) error {
	activeKeyFile := filepath.Join(s.outputDir, privateKeyFile)
	activeCertFile := filepath.Join(s.outputDir, certFile)
	castoreFile := filepath.Join(s.outputDir, castore)

	if err := s.backupAndReplace(activeKeyFile, filepath.Join(s.configDir, "acme.cpk")); err != nil {
		return err
	}

	if err := s.backupAndReplace(activeCertFile, cert.PEMPath); err != nil {
		return err
	}

	cas, err := os.OpenFile(castoreFile, os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer cas.Close()

	fmt.Fprintln(cas)
	cas.Write(caChain)
	return nil
}

func (s *ProvisionCommand) backupAndReplace(origFile string, newFile string) error {
	fi, err := os.Lstat(origFile)
	if err != nil {
		return err
	}

	if fi.IsDir() {
		return fmt.Errorf("%s is a dir", origFile)
	}

	backupFilePath := fmt.Sprintf("%s.%d.bak", origFile, time.Now().UnixNano())
	if fi.Mode()&os.ModeSymlink != os.ModeSymlink {
		src, berr := os.Open(origFile)
		if berr != nil {
			return berr
		}
		defer src.Close()

		dst, berr := os.OpenFile(backupFilePath, os.O_CREATE|os.O_EXCL, fi.Mode()&os.ModePerm)
		if berr != nil {
			return berr
		}
		defer dst.Close()
		_, berr = io.Copy(dst, src)
		if berr != nil {
			return berr
		}
	}

	if err = os.Remove(origFile); err != nil {
		return err
	}

	if err = os.Symlink(newFile, origFile); err != nil {
		s.restoreBackup(origFile, backupFilePath)
		return err
	}

	return nil
}

func (s *ProvisionCommand) restoreBackup(origFile, backupFile string) {
	src, e := os.Open(backupFile)
	if e != nil {
		return
	}
	defer src.Close()
	dst, e := os.Create(origFile)
	if e != nil {
		return
	}
	defer dst.Close()

	io.Copy(dst, src)
}
