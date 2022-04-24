package cmd

import (
	"github.com/deta/deta-cli/cmd/logic"

	"github.com/spf13/cobra"
)

var (
	showLogs bool
	runCmd   = &cobra.Command{
		Use:     "run [flags] [action] [-- <input args>]",
		Short:   "Run a deta micro",
		Example: runExamples(),
		RunE:    run,
	}
)

func init() {
	runCmd.Flags().BoolVarP(&showLogs, "logs", "l", false, "show micro logs")
	rootCmd.AddCommand(runCmd)
}

func run(cmd *cobra.Command, args []string) error {
	return logic.Run(client, showLogs, args)
}

func runExamples() string {
	return `
1. deta run -- --name Jimmy --age 33 -active

Run deta micro with the following input:
{
	"name": "Jimmy",
	"age": "33",
	"active": true
}

2. deta run --logs test -- --username admin

Run deta micro and show micro logs with action 'test' and the following input:
{
	"username": "admin"
}

3. deta run delete -- --emails jimmy@deta.sh --emails joe@deta.sh

Run deta micro with action 'delete' and the following input:
{
	"emails": ["jimmy@deta.sh", "joe@deta.sh"]
}  

See https://docs.deta.sh for more examples and details. 
`
}
