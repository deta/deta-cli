package cmd

import (
	"github.com/deta/deta-cli/cmd/logic"
	"github.com/spf13/cobra"
)

var (
	visorOpenCmd = &cobra.Command{
		Use:   "open [path]",
		Short: "Open deta visor page for a deta micro",
		Args:  cobra.MaximumNArgs(1),
		RunE:  visorOpen,
	}
)

func init() {
	visorCmd.AddCommand(visorOpenCmd)
}

func visorOpen(cmd *cobra.Command, args []string) error {
	return logic.VisorOpen(client, args)
}