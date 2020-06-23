package cmd

import (
	"os"

	"github.com/deta/deta-cli/api"
	"github.com/deta/deta-cli/auth"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "deta",
		Short: "Deta CLI for mananging deta micros",
		Long: `Deta command line interface for managing deta micros. 
Complete documentation available at https://docs.deta.sh`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}

	// deta client
	client = api.NewDetaClient()

	// auth manager
	authManager = auth.NewManager()
)

// Execute xx
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}
}
