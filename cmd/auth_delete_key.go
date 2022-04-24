package cmd

import (
	"github.com/deta/deta-cli/cmd/logic"
	"github.com/spf13/cobra"
)

var (
	deleteAPIKeyCmd = &cobra.Command{
		Use:     "delete-api-key",
		Short:   "Delete api key for a deta micro",
		Args:    cobra.MaximumNArgs(1),
		Example: authDeleteKeyExamples(),
		RunE:    deleteAPIKey,
	}
)

func init() {
	deleteAPIKeyCmd.Flags().StringVarP(&apiKeyName, "name", "n", "", "api-key name")
	deleteAPIKeyCmd.MarkFlagRequired("name")

	authCmd.AddCommand(deleteAPIKeyCmd)
}

func deleteAPIKey(cmd *cobra.Command, args []string) error {
	return logic.DeleteAPIKey(client, apiKeyName, args)
}

func authDeleteKeyExamples() string {
	return `
1. deta auth delete-api-key --name agent1

Delete api key with name 'agent1'`
}
