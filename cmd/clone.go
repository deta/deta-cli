package cmd

import (
	"github.com/deta/deta-cli/cmd/logic"
	"github.com/spf13/cobra"
)

var (
	cloneCmd = &cobra.Command{
		Use:     "clone [path]",
		Short:   "Clone a deta micro",
		RunE:    clone,
		Example: cloneExamples(),
		Args:    cobra.MaximumNArgs(1),
	}
)

func init() {
	cloneCmd.Flags().StringVar(&progName, "name", "", "deta micro name")
	cloneCmd.Flags().StringVar(&projectName, "project", "", "project to clone the micro from")
	cloneCmd.MarkFlagRequired("name")

	rootCmd.AddCommand(cloneCmd)
}

func clone(cmd *cobra.Command, args []string) error {
	return logic.Clone(client, progName, projectName, args)
}

func cloneExamples() string {
	return `
1. deta clone --name my-micro

Clone latest deployment of micro 'my-micro' from 'default' project to directory './my-micro'.

2. deta clone --name my-micro --project my-project micros/my-micro-dir

Clone latest deployment of micro 'my-micro' from project 'my-project' to directory './micros/my-micro-dir'.
'./micros/my-micro-dir' must be an empty directory if it already exists. `
}
