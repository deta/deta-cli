package cmd

import (
	"github.com/deta/deta-cli/cmd/logic"
	"github.com/spf13/cobra"
)

var (
	purgeDepsCmd = &cobra.Command{
		Use:     "purge-dependencies [path]",
		Short:   "Remove all installed dependencies for a deta micro",
		Args:    cobra.MaximumNArgs(1),
		Example: purgeDepsExamples(),
		RunE:    purgeDeps,
	}
)

func init() {
	deployCmd.AddCommand(purgeDepsCmd)
}

func purgeDeps(cmd *cobra.Command, args []string) error {
	return logic.PurgeDeps(client, args)
}

func purgeDepsExamples() string {
	return `
1. deta deploy purge-dependencies	

Remove all dependencies installed for the micro in the current directory

2. deta deploy purge-dependencies ./my-micro

Remove all dependencies installed for the micro in './my-micro' directory`
}
