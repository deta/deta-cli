package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/deta/deta-cli/api"

	"github.com/deta/deta-cli/runtime"
	"github.com/spf13/cobra"
)

var (
	nodeFlag    bool
	pythonFlag  bool
	newProgName string
	projectName string

	newCmd = &cobra.Command{
		Use:   "new [flags] [path]",
		Short: "Create a new micro",
		RunE:  new,
		Args:  cobra.MaximumNArgs(1),
	}
)

func init() {
	// flags
	newCmd.Flags().BoolVarP(&nodeFlag, "node", "n", false, "create a micro with node runtime")
	newCmd.Flags().BoolVarP(&pythonFlag, "python", "p", false, "create a micro with python runtime")
	newCmd.Flags().StringVar(&newProgName, "name", "", "name of the new micro")
	newCmd.Flags().StringVar(&projectName, "project", "", "project to create the micro under")

	rootCmd.AddCommand(newCmd)
}

func new(cmd *cobra.Command, args []string) error {
	if nodeFlag && pythonFlag {
		return fmt.Errorf("can not set both node and python flags")
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
		return fmt.Errorf("a deta micro already present in '%s'", wd)
	}

	// check if program root dir is empty
	isEmpty, err := runtimeManager.IsProgDirEmpty()
	if err != nil {
		return err
	}

	var progRuntime string
	if !isEmpty {
		progRuntime, err = runtimeManager.GetRuntime()
		if err != nil {
			return err
		}
		if nodeFlag && progRuntime != runtime.Node {
			return fmt.Errorf("'%s' does not contain node entrypoint file", wd)
		} else if pythonFlag && progRuntime != runtime.Python {
			return fmt.Errorf("'%s' does not contain python entrypoint file", wd)
		}
	} else {
		if nodeFlag {
			progRuntime = runtime.Node
		} else if pythonFlag {
			progRuntime = runtime.Python
		} else {
			os.Stderr.WriteString("Missing runtime. Please, choose a runtime with 'deta new --node' or 'deta new --python'\n")
			return nil
		}
	}

	// get user information
	userInfo, err := runtimeManager.GetUserInfo()
	if err != nil {
		return err
	}

	if userInfo == nil {
		return fmt.Errorf("login required, log in with 'deta login'")
	}

	project := userInfo.DefaultProject
	if projectName != "" {
		project = projectName
	}

	req := &api.NewProgramRequest{
		Space:   userInfo.DefaultSpace,
		Project: project,
		Name:    newProgName,
		Runtime: progRuntime,
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

	msg := "Successfully created a new micro"
	fmt.Println(msg)
	output, err := progInfoToOutput(newProgInfo)
	if err != nil {
		os.Stderr.WriteString("Micro created but failed to show details\n")
	}
	fmt.Println(output)

	// dowload template files if dir is empty
	if isEmpty {
		// wait for permissions to propagate before viewing program
		time.Sleep(1 * time.Second)
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
		err = runtimeManager.WriteProgramFiles(o.Files, nil)
		if err != nil {
			return err
		}
		// store the program state
		// ignore error here as it's okay
		// if state is not stored for new program
		runtimeManager.StoreState()
		return nil
	}

	c, err := runtimeManager.GetChanges()
	if err != nil {
		return err
	}

	_, err = client.Deploy(&api.DeployRequest{
		ProgramID: res.ID,
		Changes:   c.Changes,
		Deletions: c.Deletions,
		Account:   res.Account,
		Region:    res.Region,
	})
	if err != nil {
		return err
	}

	dc, err := runtimeManager.GetDepChanges()
	if err != nil {
		return err
	}
	runtimeManager.StoreState()

	if dc != nil {
		msg = fmt.Sprintf("%s%s", msg, "Adding dependencies...")
		fmt.Println(msg)
		command := runtime.DepCommands[res.Runtime]
		if len(dc.Added) > 0 {
			installCmd := fmt.Sprintf("%s install", command)
			for _, a := range dc.Added {
				installCmd = fmt.Sprintf("%s %s", installCmd, a)
			}
			o, err := client.UpdateProgDeps(&api.UpdateProgDepsRequest{
				ProgramID: res.ID,
				Command:   installCmd,
			})
			if err != nil {
				return err
			}
			fmt.Println(o.Output)
			if o.HasError {
				fmt.Println()
				return fmt.Errorf("failed to update dependecies: error on one or more dependencies, no dependencies were added, see output for details")
			}
			// store updated program info
			progDetails, err := client.GetProgDetails(&api.GetProgDetailsRequest{
				ProgramID: newProgInfo.ID,
			})
			if err != nil {
				newProgInfo.ReloadDeps = true
			}
			newProgInfo.Deps = progDetails.Deps
			runtimeManager.StoreProgInfo(newProgInfo)
		}
	}
	return nil
}
