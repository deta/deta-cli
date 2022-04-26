package cmd

import (
	"github.com/deta/deta-cli/cmd/logic"
	"github.com/spf13/cobra"
)

var (
	followFlag bool

	logsCmd = &cobra.Command{
		Use:   "logs [flags]",
		Short: "Get logs from a micro",
		Long: `Get logs from a visor disabled micro of the last 30 mins. 
Use command with the --follow flag to follow logs.
Using --follow automatically disables visor and renables it when the command exits.`,
		Args: cobra.NoArgs,
		RunE: logs,
	}
)

func init() {
	logsCmd.Flags().BoolVarP(&followFlag, "follow", "f", false, "follow logs")
	rootCmd.AddCommand(logsCmd)
}

func logs(cmd *cobra.Command, args []string) error {
	return logic.Logs(client, followFlag, args)
}
