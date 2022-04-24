package cmd

import (
	"github.com/deta/deta-cli/cmd/logic"
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
	return logic.ListProjects(client, args)
}
