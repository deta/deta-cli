package cmd

import (
	"github.com/deta/deta-cli/cmd/logic"
	"github.com/spf13/cobra"
)

var (
	loginCmd = &cobra.Command{
		Use:   "login",
		Short: "Login to deta",
		RunE:  login,
	}
)

func init() {
	rootCmd.AddCommand(loginCmd)
}

func login(cmd *cobra.Command, args []string) error {
	return logic.Login(client, authManager, args)
}
