package cmd

import "github.com/spf13/cobra"

var (
	visorDisableCmd = &cobra.Command{
		Use:   "disable [value]",
		Short: "Disable visor for a deta program",
		Args:  cobra.MaximumNArgs(1),
		RunE:  disableVisor,
	}
)

func init() {
	visorCmd.AddCommand(visorDisableCmd)
}

func disableVisor(cmd *cobra.Command, args []string) error {
	return updateVisor("off", args)
}
