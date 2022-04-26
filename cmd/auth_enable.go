package cmd

import (
	"github.com/deta/deta-cli/cmd/logic"
	"github.com/spf13/cobra"
)

var (
	authEnableCmd = &cobra.Command{
		Use:   "enable",
		Short: "Enable http auth for a deta micro",
		Args:  cobra.MaximumNArgs(1),
		RunE:  enableAuth,
	}
)

func init() {
	authCmd.AddCommand(authEnableCmd)
}

func enableAuth(cmd *cobra.Command, args []string) error {
	return logic.UpdateAuth(client, true, args)
}
