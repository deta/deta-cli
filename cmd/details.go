package cmd

import (
	"github.com/deta/deta-cli/cmd/logic"
	"github.com/spf13/cobra"
)

var (
	detailsCmd = &cobra.Command{
		Use:   "details [flags] [path]",
		Short: "Details about a deta micro",
		RunE:  details,
		Args:  cobra.MaximumNArgs(1),
	}
)

func init() {
	rootCmd.AddCommand(detailsCmd)
}

func details(cmd *cobra.Command, args []string) error {
	return logic.Details(client, args)
}
