package logic

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/deta/deta-cli/api"

	"github.com/deta/deta-cli/runtime"
)



func NewProgram(client *api.DetaClient, progName string, projectName string, runtimeName string, pythonFlag bool, nodeFlag bool, args []string,) error {
	if nodeFlag && pythonFlag {
		return fmt.Errorf("can not set both node and python flags")
	}

	if (nodeFlag || pythonFlag) && len(runtimeName) != 0 {
		return fmt.Errorf("can not set both node/python flags and runtime flag")
	}

	var wd string
	if len(args) == 0 {
		cd, err := os.Getwd()
		if err != nil {
			return err
		}
		wd = cd
	} else {
		wd = args[0]
	}

	runtimeManager, err := runtime.NewManager(&wd, true)
	if err != nil {
		return err
	}

	// check if program root dir is empty
	isEmpty, err := runtimeManager.IsProgDirEmpty()
	if err != nil {
		return err
	}

	progRuntime, err := runtimeManager.GetRuntime()
	if err != nil {
		if errors.Is(err, runtime.ErrNoEntrypoint) && !isEmpty {
			if progName == "" {
				os.Stderr.WriteString(fmt.Sprintf("No entrypoint file found in '%s'. Please, provide a name or path to create a new micro elsewhere. See `deta new --help`.'\n", wd))
				return nil
			}
			runtimeManager.Clean()
			wd = filepath.Join(wd, progName)
			err := os.MkdirAll(wd, 0760)
			if err != nil {
				return err
			}
			runtimeManager, err = runtime.NewManager(&wd, true)
			if err != nil {
				return err
			}
		}
	}

	if progName == "" {
		// use current working dir as the default name of the program
		// replace spaces with underscore from the dir name if present
		progName = strings.ReplaceAll(filepath.Base(wd), " ", "_")
	}

	// checks if a program is already present in the working directory
	isInitialized, err := runtimeManager.IsInitialized()
	if err != nil {
		return err
	}
	if isInitialized {
		return fmt.Errorf("a deta micro already present in '%s'", wd)
	}

	isEmpty, err = runtimeManager.IsProgDirEmpty()
	if err != nil {
		return err
	}
	if !isEmpty && len(runtimeName) == 0 {
		progRuntime, err = runtimeManager.GetRuntime()
		if err != nil {
			return err
		}
		if nodeFlag && progRuntime.Name != runtime.Node {
			return fmt.Errorf("'%s' does not contain node entrypoint file", wd)
		} else if pythonFlag && progRuntime.Name != runtime.Python {
			return fmt.Errorf("'%s' does not contain python entrypoint file", wd)
		}
	} else {
		if nodeFlag {
			progRuntime = &runtime.Runtime{
				Name:    runtime.Node,
				Version: runtime.GetDefaultRuntimeVersion(runtime.Node),
			}
		} else if pythonFlag {
			progRuntime = &runtime.Runtime{
				Name:    runtime.Python,
				Version: runtime.GetDefaultRuntimeVersion(runtime.Python),
			}
		} else if len(runtimeName) != 0 {
			progRuntime, err = parseRuntime(runtimeName)
			if err != nil {
				return err
			}
		} else {
			os.Stderr.WriteString("Missing runtime. Please, choose a runtime with 'deta new --node' or 'deta new --python'\n")
			return nil
		}
	}

	// get user information
	userInfo, err := getUserInfo(runtimeManager, client)
	if err != nil {
		return err
	}

	project := userInfo.DefaultProject
	if projectName != "" {
		project = projectName
	}

	req := &api.NewProgramRequest{
		Space:   userInfo.DefaultSpace,
		Project: project,
		Name:    progName,
		Runtime: progRuntime.Version,
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
		Visor:   res.Visor,
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

	// wait for permissions to propagate before viewing/updating program
	time.Sleep(1 * time.Second)

	// dowload template files if dir is empty
	if isEmpty {
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
		err = runtimeManager.WriteProgramFiles(o.ZipFile, nil, true, res.Runtime)
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

	if c != nil {
		_, err = client.Deploy(&api.DeployRequest{
			ProgramID:   res.ID,
			Changes:     c.Changes,
			BinaryFiles: c.BinaryFiles,
			Deletions:   c.Deletions,
			Account:     res.Account,
			Region:      res.Region,
		})
		if err != nil {
			return err
		}
	}

	dc, err := runtimeManager.GetDepChanges()
	if err != nil {
		return err
	}
	runtimeManager.StoreState()

	if dc != nil {
		fmt.Println("Adding dependencies...")
		command := runtime.DepCommands[progRuntime.Name]
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
				Program: newProgInfo.ID,
				Space:   userInfo.DefaultSpace,
				Project: project,
			})
			if err != nil {
				return err
			}
			newProgInfo.Deps = progDetails.Deps
			runtimeManager.StoreProgInfo(newProgInfo)
		}
	}
	return nil
}
