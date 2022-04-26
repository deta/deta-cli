package logic

import (
	"fmt"

	"github.com/deta/deta-cli/api"
	"github.com/deta/deta-cli/auth"
	"github.com/deta/deta-cli/runtime"
)


func Login(client *api.DetaClient, authManager *auth.Manager, args []string) error {
	fmt.Println("Please, log in from the web page. Waiting...")
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
