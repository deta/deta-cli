package cmd

import (
	"github.com/deta/deta-cli/cmd/logic"
	"github.com/spf13/cobra"
)

var (
	authDisableCmd = &cobra.Command{
		Use:   "disable",
		Short: "Disable http auth for a deta micro",
		Args:  cobra.MaximumNArgs(1),
		RunE:  disableAuth,
	}
)

func init() {
	authCmd.AddCommand(authDisableCmd)
}

func disableAuth(cmd *cobra.Command, args []string) error {
	return logic.UpdateAuth(client, false, args)
}
