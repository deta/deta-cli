package cmd

import (
	"github.com/deta/deta-cli/cmd/logic"
	"github.com/spf13/cobra"
)

var (
	visorCmd = &cobra.Command{
		Use:   "visor [command]",
		Short: "Change visor settings for a deta micro",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}
)

func init() {
	rootCmd.AddCommand(visorCmd)
}

func updateVisor(mode string, args []string) error {
	return logic.UpdateVisor(client, mode, args)
}
