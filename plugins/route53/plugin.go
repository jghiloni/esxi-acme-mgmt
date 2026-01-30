package main

import (
	"context"

	"github.com/alecthomas/kong"
	"github.com/caddyserver/certmagic"
	"github.com/jghiloni/esxi-acme-mgmt/plugins/common"
	"github.com/libdns/route53"
	"github.com/mholt/acmez/v3"
	"github.com/mholt/acmez/v3/acme"
)

const providerName = "route53"

type providerArgs struct {
	AccessKeyID     string `env:"AWS_ACCESS_KEY_ID"`
	SecretAccessKey string `env:"AWS_SECRET_ACCESS_KEY"`
	Profile         string `env:"AWS_PROFILE"`
	SessionToken    string `env:"AWS_SESSION_TOKEN"`
	HostedZoneID    string `env:"AWS_HOSTED_ZONE_ID"`
	Region          string `env:"AWS_REGION" default:"us-east-1"`
}

type route53PluginProvider struct {
	delegate acmez.Solver
}

func (r *route53PluginProvider) Present(ctx context.Context, challenge acme.Challenge) error {
	return r.delegate.Present(ctx, challenge)
}

func (r *route53PluginProvider) CleanUp(ctx context.Context, challenge acme.Challenge) error {
	return r.delegate.CleanUp(ctx, challenge)
}

func (*route53PluginProvider) Name() string {
	return providerName
}

func (r *route53PluginProvider) WithArgs(args []string) {
	var parsedArgs providerArgs
	k := kong.Must(&parsedArgs)
	k.Parse(args)

	r.delegate = &certmagic.DNS01Solver{
		DNSManager: certmagic.DNSManager{
			DNSProvider: &route53.Provider{
				Region:                  parsedArgs.Region,
				Profile:                 parsedArgs.Profile,
				AccessKeyId:             parsedArgs.AccessKeyID,
				SecretAccessKey:         parsedArgs.SecretAccessKey,
				SessionToken:            parsedArgs.SessionToken,
				WaitForRoute53Sync:      false,
				SkipRoute53SyncOnDelete: true,
				HostedZoneID:            parsedArgs.HostedZoneID,
			},
		},
	}
}

var DNSProvider common.Provider = new(route53PluginProvider)

func main() {
	// no op for a plugin
}
