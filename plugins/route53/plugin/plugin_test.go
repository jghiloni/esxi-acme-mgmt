package plugin_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/jghiloni/esxi-acme-mgmt/plugins/route53/plugin"
	"github.com/mholt/acmez/v3"
	"github.com/mholt/acmez/v3/acme"
	. "github.com/onsi/gomega"
)

func TestCertificateProvision(t *testing.T) {
	RegisterTestingT(t)

	var missing []string
	for _, e := range []string{
		"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_REGION",
		"AWS_HOSTED_ZONE_ID", "ESXI_ACME_MGMT_TEST_DOMAIN"} {
		if _, set := os.LookupEnv(e); !set {
			missing = append(missing, e)
		}
	}

	if len(missing) > 0 {
		t.Skipf("skipping TestCertificateProvision because the following env vars are missing: %v", missing)
	}

	accountPEM, err := os.ReadFile("./testdata/test_account.pem")
	Expect(err).NotTo(HaveOccurred())

	block, _ := pem.Decode(accountPEM)

	signer, err := x509.ParseECPrivateKey(block.Bytes)
	Expect(err).NotTo(HaveOccurred())

	accountJSON, err := os.ReadFile("./testdata/test_account.json")
	Expect(err).NotTo(HaveOccurred())

	var account acme.Account
	Expect(json.Unmarshal(accountJSON, &account)).NotTo(HaveOccurred())

	account.PrivateKey = signer

	solver := plugin.NewRoute53Plugin()
	solver.WithArgs([]string{})

	client := acmez.Client{
		Client: &acme.Client{
			Directory:   "https://acme-staging-v02.api.letsencrypt.org/directory",
			Logger:      slog.Default(),
			UserAgent:   "esxi-acme-mgmt/r53-plugin-test",
			PollTimeout: time.Minute,
		},
		ChallengeSolvers: map[string]acmez.Solver{
			acme.ChallengeTypeDNS01: solver,
		},
	}

	account, err = client.NewAccount(context.Background(), account)
	Expect(err).NotTo(HaveOccurred())

	certPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	Expect(err).NotTo(HaveOccurred())

	domain := os.Getenv("ESXI_ACME_MGMT_TEST_DOMAIN")
	certs, err := client.ObtainCertificateForSANs(context.Background(), account, certPrivateKey, []string{domain})
	Expect(err).NotTo(HaveOccurred())
	Expect(certs).To(HaveLen(2))
}
