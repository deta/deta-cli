module github.com/deta/deta-cli

go 1.13

require (
	github.com/aws/aws-sdk-go v1.31.7
	github.com/deta/deta-cli/auth v0.0.0-00010101000000-000000000000
	github.com/deta/deta-cli/cmd v0.0.0-00010101000000-000000000000
	github.com/spf13/cobra v1.0.0
	golang.org/x/crypto v0.0.0-20200510223506-06a226fb4e37
)

replace github.com/deta/deta-cli/cmd => ./cmd

replace github.com/deta/deta-cli/auth => ./auth
