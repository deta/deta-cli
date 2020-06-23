package cmd

import (
	"fmt"

	"github.com/deta/deta-cli/api"
	"github.com/deta/deta-cli/runtime"
	"github.com/spf13/cobra"
)

var (
	projectsCmd = &cobra.Command{
		Use:   "projects",
		Short: "List deta projects",
		RunE:  listProjects,
		Args:  cobra.NoArgs,
	}
)

func init() {
	rootCmd.AddCommand(projectsCmd)
}

func listProjects(cmd *cobra.Command, args []string) error {
	runtimeManager, err := runtime.NewManager(nil, false)
	if err != nil {
		return err
	}

	u, err := runtimeManager.GetUserInfo()
	if err != nil {
		return err
	}

	if u == nil {
		return fmt.Errorf("login required, login in `deta login`")
	}

	res, err := client.GetProjects(&api.GetProjectsRequest{
		SpaceID: u.DefaultSpace,
	})

	if err != nil {
		return err
	}

	output, err := prettyPrint(res.Projects)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}
