package cmd

import (
	"fmt"

	"github.com/deta/deta-cli/runtime"
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
	fmt.Println("Please, log in from the web page. Waiting..")
	if err := authManager.Login(); err != nil {
		return err
	}

	u, err := client.GetUserInfo()
	if err != nil {
		return err
	}

	runtimeManager, err := runtime.NewManager(nil, false)
	if err != nil {
		return err
	}

	runtimeManager.StoreUserInfo(&runtime.UserInfo{
		DefaultSpace:     u.DefaultSpace,
		DefaultSpaceName: u.DefaultSpaceName,
		DefaultProject:   u.DefaultProject,
	})
	fmt.Println("Logged in successfully.")
	return nil
}
