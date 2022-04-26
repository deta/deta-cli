package cmd

import (
	"github.com/deta/deta-cli/cmd/logic"
	"github.com/spf13/cobra"
)

var (
	forcePull bool

	pullCmd = &cobra.Command{
		Use:     "pull [flags]",
		Short:   "Pull the lastest deployed code of a deta micro",
		RunE:    pull,
		Example: pullExamples(),
		Args:    cobra.MaximumNArgs(1),
	}
)

func init() {
	pullCmd.Flags().BoolVarP(&forcePull, "force", "f", false, "force overwrite of existing files")
	rootCmd.AddCommand(pullCmd)
}

func pull(cmd *cobra.Command, args []string) error {
	return logic.Pull(client, forcePull, args)
}

func pullExamples() string {
	return `
1. deta pull

Pull latest changes of deta micro present in the current directory. 
Asks for approval before overwriting the files in the current directory.

2. deta pull --force

Force pull latest changes of deta micro present in the current directory.
Overwrites the files present in the current directory.`
}
