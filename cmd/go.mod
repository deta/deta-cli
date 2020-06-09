module github.com/deta/deta-cli/cmd

go 1.13

replace github.com/deta/deta-cli/auth => ../auth

replace github.com/deta/deta-cli/api => ../api

replace github.com/deta/deta-cli/runtime => ../runtime

require (
	github.com/deta/deta-cli/api v0.0.0-00010101000000-000000000000
	github.com/deta/deta-cli/auth v0.0.0-00010101000000-000000000000
	github.com/deta/deta-cli/runtime v0.0.0-00010101000000-000000000000
	github.com/spf13/cobra v1.0.0
)
