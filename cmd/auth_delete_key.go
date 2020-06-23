package cmd

import (
	"fmt"
	"os"

	"github.com/deta/deta-cli/api"
	"github.com/deta/deta-cli/runtime"
	"github.com/spf13/cobra"
)

var (
	deleteAPIKeyCmd = &cobra.Command{
		Use:   "delete-api-key",
		Short: "Delete api keys for a deta micro",
		Args:  cobra.MaximumNArgs(1),
		RunE:  deleteAPIKey,
	}
)

func init() {
	deleteAPIKeyCmd.Flags().StringVarP(&apiKeyName, "name", "n", "", "api-key name")
	deleteAPIKeyCmd.Flags().StringVarP(&apiKeyDesc, "desc", "d", "", "api-key description")
	deleteAPIKeyCmd.MarkFlagRequired("name")

	authCmd.AddCommand(deleteAPIKeyCmd)
}

func deleteAPIKey(cmd *cobra.Command, args []string) error {
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
		return fmt.Errorf("No deta micro initialized in '%s'", wd)
	}

	progInfo, err := runtimeManager.GetProgInfo()
	if err != nil {
		return err
	}

	err = client.DeleteAPIKey(&api.DeleteAPIKeyRequest{
		ProgramID: progInfo.ID,
		Name:      apiKeyName,
	})
	if err != nil {
		return err
	}
	fmt.Println(fmt.Sprintf("Succesfully deleted api key '%s'", apiKeyName))
	return nil
}
