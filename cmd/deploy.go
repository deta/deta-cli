package cmd

import (
	"github.com/deta/deta-cli/cmd/logic"
	"github.com/spf13/cobra"
)

var (
	deployCmd = &cobra.Command{
		Use:     "deploy [path]",
		Short:   "Deploy a deta micro",
		Args:    cobra.MaximumNArgs(1),
		Example: deployExamples(),
		RunE:    deploy,
	}
)

func init() {
	rootCmd.AddCommand(deployCmd)
}

func deploy(cmd *cobra.Command, args []string) error {
	return logic.Deploy(client, args)
}

func deployExamples() string {
	return `
1. deta deploy

Deploy a deta micro rooted in the current directory.

2. deta deploy micros/my-micro-1

Deploy a deta micro rooted in 'micros/my-micro-1' directory.`
}
