package cmd

import (
	"github.com/deta/deta-cli/cmd/logic"
	"github.com/spf13/cobra"
)

var (
	outfile    string
	apiKeyName string
	apiKeyDesc string

	createAPIKeyCmd = &cobra.Command{
		Use:     "create-api-key [path]",
		Short:   "Create api keys for a deta micro",
		Args:    cobra.MaximumNArgs(1),
		Example: authCreateKeyExamples(),
		RunE:    createAPIKey,
	}
)

func init() {
	createAPIKeyCmd.Flags().StringVarP(&outfile, "outfile", "o", "", "file to save the api-key")
	createAPIKeyCmd.Flags().StringVarP(&apiKeyName, "name", "n", "", "api-key name")
	createAPIKeyCmd.Flags().StringVarP(&apiKeyDesc, "desc", "d", "", "api-key description")
	createAPIKeyCmd.MarkFlagRequired("name")

	authCmd.AddCommand(createAPIKeyCmd)
}

func createAPIKey(cmd *cobra.Command, args []string) error {
	return logic.CreateAPIKey(client, outfile, apiKeyName, apiKeyDesc, args)
}

func authCreateKeyExamples() string {
	return `
1. deta auth create-api-key --name agent1 --desc "api key for agent 1"

Create an api key with name 'agent1' and description 'api key for agent 1'

2. deta auth create-api-key --name agent1 --outfile agent_1_key.txt

Create an api key with name 'agent1' and save it to file 'agent_1_key.txt'`
}
