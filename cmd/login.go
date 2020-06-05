package cmd

import (
	"fmt"

	"github.com/deta/deta-cli/runtime"
	"github.com/spf13/cobra"
)

var (
	loginCmd = &cobra.Command{
		Use:   "login",
		Short: "login to deta",
		RunE:  login,
	}
)

func init() {
	rootCmd.AddCommand(loginCmd)
}

func login(cmd *cobra.Command, args []string) error {
	if err := authManager.Login(); err != nil {
		return err
	}
	resp, err := client.ListSpaces()
	if err != nil {
		return err
	}

	runtimeManager, err := runtime.NewManager(nil)
	if err != nil {
		return err
	}

	u := &runtime.UserInfo{
		DefaultSpace: resp[0].SpaceID,
	}

	err = runtimeManager.StoreUserInfo(u)
	if err != nil {
		return err
	}
	fmt.Println("Logged in successfully.")
	return nil
}
