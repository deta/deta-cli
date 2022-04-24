package cmd

import (
	"github.com/deta/deta-cli/cmd/logic"
	"github.com/spf13/cobra"
)

var (
	watchCmd = &cobra.Command{
		Use:     "watch [path]",
		Short:   "Deploy changes in real time",
		RunE:    watch,
		Example: watchExamples(),
		Args:    cobra.MaximumNArgs(1),
	}
)

func init() {
	rootCmd.AddCommand(watchCmd)
}

func watch(cmd *cobra.Command, args []string) error {
	return logic.Watch(client, args)
}

func watchExamples() string {
	return `
1. deta watch

Watch for changes in the current directory and deploy changes in real time.

2. deta watch my-micro

Watch for changes in './my-micro' directory and deploy changes in real time.`
}
