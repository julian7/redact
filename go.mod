module github.com/julian7/redact

go 1.19

require (
	github.com/ProtonMail/go-crypto v0.0.0-20220930113650-c6815a8c17ad
	github.com/hectane/go-acl v0.0.0-20190604041725-da78bae5fc95
	github.com/julian7/tester v0.0.0-20190708141839-fd2332449f51
	github.com/spf13/afero v1.9.2
	// there is an issue with v2.17.1 which breaks rendering subcommands'
	// options. Staying with v2.16.3 until it's fixed
	// Ref: https://github.com/urfave/cli/issues/1505
	github.com/urfave/cli/v2 v2.16.3
	golang.org/x/crypto v0.0.0-20220926161630-eccd6366d1be
)

require (
	github.com/cloudflare/circl v1.2.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	golang.org/x/sys v0.0.0-20220928140112-f11e5e49a4ec // indirect
	golang.org/x/text v0.3.7 // indirect
)
