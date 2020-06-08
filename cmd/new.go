package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/deta/deta-cli/api"

	"github.com/deta/deta-cli/runtime"
	"github.com/spf13/cobra"
)

var (
	nodeFlag    bool
	pythonFlag  bool
	newProgName string

	newCmd = &cobra.Command{
		Use:   "new",
		Short: "Create a new program",
		RunE:  new,
		Args:  cobra.MaximumNArgs(1),
	}
)

func init() {
	newCmd.Flags().BoolVar(&nodeFlag, "node", false, "create a program with node runtime")
	newCmd.Flags().BoolVar(&pythonFlag, "python", false, "create a program with python runtime")
	newCmd.Flags().StringVarP(&newProgName, "name", "n", "", "name of the new program")
}

func new(cmd *cobra.Command, args []string) error {
	if nodeFlag && pythonFlag {
		os.Stderr.WriteString("Can not set both node and python flags")
	}

	var wd string
	if len(args) == 0 {
		// if path not provided as args
		// get current working directory
		cd, err := os.Getwd()
		if err != nil {
			return err
		}
		wd = cd
	} else {
		wd = args[0]
	}

	if newProgName == "" {
		// use current working dir as the default name of the program
		// replace spaces with underscore from the dir name if present
		newProgName = strings.ReplaceAll(filepath.Base(wd), " ", "_")
	}

	runtimeManager, err := runtime.NewManager(&wd)
	if err != nil {
		return err
	}

	// checks if a program is already present in the working directory
	isInitialized, err := runtimeManager.IsInitialized()
	if err != nil {
		return err
	}
	if isInitialized {
		os.Stderr.WriteString(fmt.Sprintf("A deta program already present in '%s'", wd))
		return nil
	}

	// get user information
	userInfo, err := runtimeManager.GetUserInfo()
	if err != nil {
		os.Stderr.WriteString("No user details found. Please, log in with 'deta login'.")
		return err
	}

	req := &api.NewProgramRequest{
		Space: userInfo.DefaultSpace,
		Name:  newProgName,
	}

	// check for runtime
	if nodeFlag {
		req.Runtime = runtime.Node
	} else if pythonFlag {
		req.Runtime = runtime.Python
	} else {
		os.Stderr.WriteString("Missing runtime. Please, choose a runtime with 'deta new -node' or 'deta new -python'")
		return nil
	}

	// send new program request
	res, err := client.NewProgram(req)
	if err != nil {
		return err
	}

	// save new program info
	newProgInfo := &runtime.ProgInfo{
		ID:      res.ID,
		Space:   res.Space,
		Runtime: res.Runtime,
		Name:    res.Name,
		Path:    res.Path,
		Project: res.Project,
		Account: res.Account,
		Region:  res.Region,
		Deps:    res.Deps,
		Envs:    res.Envs,
		Public:  res.Public,
	}
	err = runtimeManager.StoreProgInfo(newProgInfo)
	if err != nil {
		return err
	}

	// download created program
	o, err := client.DownloadProgram(&api.DownloadProgramRequest{
		ProgramID: res.ID,
		Runtime:   res.Runtime,
		Account:   res.Account,
		Region:    res.Region,
	})
	if err != nil {
		return err
	}

	// write downloaded files to dir
	err = runtimeManager.WriteProgramFiles(o.Files, &wd)
	if err != nil {
		return err
	}

	// store the program state
	go runtimeManager.StoreState()
	return nil
}
