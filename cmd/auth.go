package cmd

import (
	"github.com/deta/deta-cli/cmd/logic"
	"github.com/spf13/cobra"
)

var (
	authCmd = &cobra.Command{
		Use:   "auth",
		Short: "Change auth settings for a deta micro",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}
)

func init() {
	rootCmd.AddCommand(authCmd)
}

func updateAuth(value bool, args []string) error {
	return logic.UpdateAuth(client, value, args)
}
