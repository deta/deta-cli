package cmd

import (
	"fmt"
	"os"

	"github.com/deta/deta-cli/api"
	"github.com/deta/deta-cli/runtime"
	"github.com/spf13/cobra"
)

var (
	visorCmd = &cobra.Command{
		Use:   "visor [command]",
		Short: "Change visor settings for a deta micro",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}
)

func init() {
	rootCmd.AddCommand(visorCmd)
}

func updateVisor(mode string, args []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	if len(args) != 0 {
		wd = args[0]
	}
	runtimeManager, err := runtime.NewManager(&wd)
	if err != nil {
		return err
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

	err = client.UpdateVisorMode(&api.UpdateVisorModeRequest{
		ProgramID: progInfo.ID,
		Mode:      mode,
	})
	if err != nil {
		return err
	}
	msg := fmt.Sprintf("Successfully disabled visor mode")
	if mode == "debug" {
		msg = fmt.Sprintf("Successfully enabled visor mode")
	}
	fmt.Println(msg)
	return nil
}
