module github.com/jghiloni/esxi-acme-mgmt/plugins/cloudflare

go 1.25.6

require (
	github.com/alecthomas/kong v1.13.0
	github.com/caddyserver/certmagic v0.25.1
	github.com/jghiloni/esxi-acme-mgmt/plugins/common v0.0.0-20260130030503-5c263b4c4ee2
	github.com/libdns/cloudflare v0.2.2
	github.com/mholt/acmez/v3 v3.1.4
)

require (
	github.com/caddyserver/zerossl v0.1.4 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/libdns/libdns v1.1.1 // indirect
	github.com/miekg/dns v1.1.72 // indirect
	github.com/zeebo/blake3 v0.2.4 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	go.uber.org/zap/exp v0.3.0 // indirect
	golang.org/x/crypto v0.47.0 // indirect
	golang.org/x/mod v0.32.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	golang.org/x/tools v0.41.0 // indirect
)

replace github.com/jghiloni/esxi-acme-mgmt/plugins/common => ../common
