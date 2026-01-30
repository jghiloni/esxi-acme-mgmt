package main

import (
	"context"

	"github.com/alecthomas/kong"
	"github.com/caddyserver/certmagic"
	"github.com/jghiloni/esxi-acme-mgmt/plugins/common"
	"github.com/libdns/cloudflare"
	"github.com/mholt/acmez/v3"
	"github.com/mholt/acmez/v3/acme"
)

const providerName = "cloudflare"

type providerArgs struct {
	APIToken  string `env:"CLOUDFLARE_API_TOKEN"`
	ZoneToken string `env:"CLOUDFLARE_ZONE_TOKEN"`
}

type cloudflarePluginProvider struct {
	delegate acmez.Solver
}

func (r *cloudflarePluginProvider) Present(ctx context.Context, challenge acme.Challenge) error {
	return r.delegate.Present(ctx, challenge)
}

func (r *cloudflarePluginProvider) CleanUp(ctx context.Context, challenge acme.Challenge) error {
	return r.delegate.CleanUp(ctx, challenge)
}

func (*cloudflarePluginProvider) Name() string {
	return providerName
}

func (r *cloudflarePluginProvider) WithArgs(args []string) {
	var parsedArgs providerArgs
	k := kong.Must(&parsedArgs)
	k.Parse(args)

	r.delegate = &certmagic.DNS01Solver{
		DNSManager: certmagic.DNSManager{
			DNSProvider: &cloudflare.Provider{
				APIToken:  parsedArgs.APIToken,
				ZoneToken: parsedArgs.ZoneToken,
			},
		},
	}
}

var DNSProvider common.Provider = new(cloudflarePluginProvider)

func main() {
	// no op for a plugin
}
