package cmd

import (
	"github.com/spf13/cobra"
)

var (
	cronCmd = &cobra.Command{
		Use:   "cron",
		Short: "Change cron settings for a deta micro",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}
)

func init() {
	rootCmd.AddCommand(cronCmd)
}
