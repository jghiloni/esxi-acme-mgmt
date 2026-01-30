package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
)

const (
	Name       = "esxi-acme-mgmt"
	versionFmt = `
Version: %s
Build:   %s
`
)

var (
	version = "0.0.0"
	build   = "dev"

	defaultAppOptions = &StartOptions{
		Version:   version,
		Build:     build,
		Stdout:    os.Stdout,
		Stderr:    os.Stderr,
		LogWriter: os.Stderr,
		Args:      os.Args[1:],
	}

	ErrNilContext = errors.New("context cannot be nil")
)

type StartOptions struct {
	Version   string
	Build     string
	Stdout    io.Writer
	Stderr    io.Writer
	LogWriter io.WriteCloser
	Args      []string
}

type RunOptions struct {
	Provision        *ProvisionCommand `cmd:"" help:"start the process of getting a new certificate"`
	Stop             *StopCommand      `cmd:"" help:"stop a running provision command"`
	AccountEmail     string            `required:"true" env:"LE_ESXI_ACCOUNT_EMAIL" help:"The email addressed associated with your letsencrypt address"`
	BaseDir          string            `default:"${basedir}" type:"existingdir" hidden:"true"`
	TargetDirectory  string            `default:"/etc/vmware/ssl" type:"existingdir" help:"The directory where generated certs should be output"`
	PluginsDir       string            `default:"${basedir}/plugins" env:"LE_ESXI_PLUGINS_DIR" type:"existingdir" help:"The directory where the plugin .so files are"`
	Provider         string            `required:"true" env:"LE_ESXI_DNS_PROVIDER" help:"The name of the provider that should be loaded via the plugins"`
	ACMEDirectoryURL string            `default:"https://acme-v02.api.letsencrypt.org/directory" env:"LE_ESXI_ACME_DIR_URL" help:"The ACME Directory URL for challenges"`
	ProviderArgs     []string          `optional:"true" arg:"" env:"LE_ESXI_PROVIDER_ARGS" passthrough:"true" help:"Arguments that will be passed to the provider separated by commas"` // TODO fix when we know where args go
}

type commandlineArgs struct {
	RunOptions
	Version kong.VersionFlag `short:"V" help:"Show the version info and quit"`
	Verbose logLevel         `short:"v" default:"2" help:"Increase verbosity. Default is --verbose=2 (-v 2), can be between 0 and 4. 0 is silent, 4 is debug level"`
}

func Run(ctx context.Context, opts *StartOptions) {
	if ctx == nil {
		log.Fatal(ErrNilContext)
	}

	if opts == nil {
		opts = defaultAppOptions
	}

	if opts != defaultAppOptions {
		if strings.TrimSpace(opts.Version) == "" {
			opts.Version = defaultAppOptions.Version
		}

		if strings.TrimSpace(opts.Build) == "" {
			opts.Build = defaultAppOptions.Build
		}

		if opts.Stdout == nil {
			opts.Stdout = defaultAppOptions.Stdout
		}

		if opts.Stderr == nil {
			opts.Stderr = defaultAppOptions.Stderr
		}

		if opts.LogWriter == nil {
			opts.LogWriter = defaultAppOptions.LogWriter
		}

		if opts.Args == nil {
			opts.Args = defaultAppOptions.Args
		}
	}

	exe, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	options := []kong.Option{
		kong.Name(Name),
		kong.Writers(opts.Stdout, opts.Stderr),
		kong.Vars{
			"basedir": filepath.Dir(filepath.Dir(exe)),
			"version": fmt.Sprintf(versionFmt, opts.Version, opts.Build),
		},
		kong.BindTo(opts.LogWriter, (*io.WriteCloser)(nil)),
		kong.BindTo(ctx, (*context.Context)(nil)),
	}

	var args commandlineArgs
	k, err := kong.New(&args, options...)
	if err != nil {
		log.Fatal(err)
	}

	kctx, err := k.Parse(opts.Args)
	if err != nil {
		log.Fatal(err)
	}

	kctx.FatalIfErrorf(kctx.Run(args.RunOptions, args.ProviderArgs))
}
