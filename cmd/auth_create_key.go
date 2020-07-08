package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/deta/deta-cli/api"
	"github.com/deta/deta-cli/runtime"
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
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	if len(args) != 0 {
		wd = args[0]
	}
	runtimeManager, err := runtime.NewManager(&wd, false)
	if err != nil {
		return nil
	}

	isInitialized, err := runtimeManager.IsInitialized()
	if err != nil {
		return err
	}

	if !isInitialized {
		return fmt.Errorf("no deta micro initialized in '%s'", wd)
	}

	progInfo, err := runtimeManager.GetProgInfo()
	if err != nil {
		return err
	}

	o, err := client.CreateAPIKey(&api.CreateAPIKeyRequest{
		ProgramID:   progInfo.ID,
		Name:        apiKeyName,
		Description: apiKeyDesc,
	})
	if err != nil {
		return err
	}

	fmt.Println("THE FOLLOWING API-KEY WILL ONLY BE SHOWN ONCE.")
	if outfile == "" {
		fmt.Println("Please, copy it and keep it in a safe place.")
	} else {
		fmt.Println("Please, keep the file in a safe place. Creating file...")
	}

	prettyOutput, err := prettyPrint(o)
	if err != nil {
		return fmt.Errorf("failed to print the key")
	}
	fmt.Println(prettyOutput)

	if outfile != "" {
		outfilepath := filepath.Join(wd, outfile)
		err := ioutil.WriteFile(outfilepath, []byte(prettyOutput), 0660)
		if err != nil {
			return fmt.Errorf("failed to save to file '%s'", outfilepath)
		}
		fmt.Println(fmt.Sprintf("Saved to file '%s'", outfilepath))
	}
	return nil
}

func authCreateKeyExamples() string {
	return `
1. deta auth create-api-key --name agent1 --desc "api key for agent 1"

Create an api key with name 'agent1' and description 'api key for agent 1'

2. deta auth create-api-key --name agent1 --outfile agent_1_key.txt

Create an api key with name 'agent1' and save it to file 'agent_1_key.txt'`
}
