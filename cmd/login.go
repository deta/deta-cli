package cmd

import (
	"github.com/deta/deta-cli/auth"
	"github.com/spf13/cobra"
)

var (
	loginCmd = &cobra.Command{
		Use:   "login",
		Short: "login to deta",
		RunE:  login,
	}
	authManager = auth.NewManager()
)

func init() {
	rootCmd.AddCommand(loginCmd)
}

func login(cmd *cobra.Command, args []string) error {
	if err := authManager.Login(); err != nil {
		return err
	}
	return nil
}
