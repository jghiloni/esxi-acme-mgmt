module github.com/jghiloni/esxi-acme-mgmt/cli

go 1.25.6

require (
	github.com/alecthomas/kong v1.13.0
	github.com/mholt/acmez/v3 v3.1.4
	github.com/samber/slog-syslog/v2 v2.5.2
	github.com/jghiloni/esxi-acme-mgmt/plugins/common v0.0.1
)

require (
	github.com/samber/lo v1.52.0 // indirect
	github.com/samber/slog-common v0.19.0 // indirect
	golang.org/x/crypto v0.47.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/text v0.33.0 // indirect
)

replace github.com/jghiloni/esxi-acme-mgmt/plugins/common => ../plugins/common
