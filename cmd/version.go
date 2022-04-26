package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// set with Makefile during compilation
	detaVersion string
	platform    string

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print deta version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s %s %s\n", rootCmd.Use, detaVersion, platform)
		},
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}
