package cmd

import "github.com/spf13/cobra"

var (
	visorEnableCmd = &cobra.Command{
		Use:   "enable [value]",
		Short: "Enable visor for a deta program",
		Args:  cobra.MaximumNArgs(1),
		RunE:  enableVisor,
	}
)

func init() {
	visorCmd.AddCommand(visorEnableCmd)
}

func enableVisor(cmd *cobra.Command, args []string) error {
	return updateVisor("debug", args)
}
