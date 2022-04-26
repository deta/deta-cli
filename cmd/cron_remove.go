package cmd

import (
	"github.com/deta/deta-cli/cmd/logic"
	"github.com/spf13/cobra"
)

var (
	cronRemoveCmd = &cobra.Command{
		Use:   "remove [path]",
		Short: "Remove a schedule from a deta micro",
		Args:  cobra.MaximumNArgs(1),
		RunE:  removeCron,
	}
)

func init() {
	cronCmd.AddCommand(cronRemoveCmd)
}

func removeCron(cmd *cobra.Command, args []string) error {
	return logic.RemoveCron(client, args)
}
