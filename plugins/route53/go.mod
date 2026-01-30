module github.com/jghiloni/esxi-acme-mgmt/plugins/route53

go 1.25.6

require (
	github.com/alecthomas/kong v1.13.0
	github.com/caddyserver/certmagic v0.25.1
	github.com/jghiloni/esxi-acme-mgmt/plugins/common v0.0.1
	github.com/libdns/route53 v1.6.0
	github.com/mholt/acmez/v3 v3.1.4
)

require (
	github.com/aws/aws-sdk-go-v2 v1.39.1 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.31.10 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.18.14 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.8 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.8 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.8 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/route53 v1.58.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.29.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.38.5 // indirect
	github.com/aws/smithy-go v1.23.0 // indirect
	github.com/caddyserver/zerossl v0.1.4 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/libdns/libdns v1.1.1 // indirect
	github.com/miekg/dns v1.1.69 // indirect
	github.com/stretchr/testify v1.8.3 // indirect
	github.com/zeebo/blake3 v0.2.4 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	go.uber.org/zap/exp v0.3.0 // indirect
	golang.org/x/crypto v0.46.0 // indirect
	golang.org/x/mod v0.30.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	golang.org/x/tools v0.39.0 // indirect
)

replace github.com/jghiloni/esxi-acme-mgmt/plugins/common => ../common
