package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/deta/deta-cli/api"
	"github.com/deta/deta-cli/runtime"
	"github.com/spf13/cobra"
)

var (
	cloneCmd = &cobra.Command{
		Use:   "clone [flags] [path]",
		Short: "Clone a deta micro",
		RunE:  clone,
		Args:  cobra.MaximumNArgs(1),
	}
)

func init() {
	cloneCmd.Flags().StringVar(&progName, "name", "", "deta micro name")
	cloneCmd.Flags().StringVar(&projectName, "project", "", "deta project")
	cloneCmd.MarkFlagRequired("name")

	rootCmd.AddCommand(cloneCmd)
}

func clone(cmd *cobra.Command, args []string) error {
	var newDirCreated bool

	// clean up if a new dir was created
	cleanup := func(dir string) {
		if newDirCreated {
			err := os.RemoveAll(dir)
			if err != nil {
				os.Stderr.WriteString(fmt.Sprintf("failed to remove dir `%s`", dir))
			}
		}
	}

	cd, err := os.Getwd()
	if err != nil {
		return err
	}
	wd := filepath.Join(cd, progName)
	if len(args) > 0 {
		wd = filepath.Join(cd, args[0])
	}

	if i, err := os.Stat(wd); err == nil {
		if !i.IsDir() {
			return fmt.Errorf("'%s' is not a directory", wd)
		}
		f, err := os.Open(wd)
		if err != nil {
			return err
		}
		_, err = f.Readdirnames(1)
		if err == nil {
			f.Close()
			return fmt.Errorf("'%s' already exists and is not empty", wd)
		} else if err != io.EOF {
			f.Close()
			return err
		}
	} else if os.IsNotExist(err) {
		err = os.MkdirAll(wd, 0760)
		if err != nil {
			return err
		}
		newDirCreated = true
	} else {
		return err
	}

	runtimeManager, err := runtime.NewManager(&wd, true)
	if err != nil {
		cleanup(wd)
		return err
	}

	u, err := runtimeManager.GetUserInfo()
	if err != nil {
		cleanup(wd)
		return err
	}

	if u == nil {
		fmt.Println("login required, log in with `deta login`")
		cleanup(wd)
		return nil
	}

	if projectName == "" {
		projectName = u.DefaultProject
	}

	progDetails, err := client.GetProgDetails(&api.GetProgDetailsRequest{
		Program: progName,
		Project: projectName,
		Space:   u.DefaultSpace,
	})
	if err != nil {
		cleanup(wd)
		return err
	}

	progInfo := &runtime.ProgInfo{
		ID:      progDetails.ID,
		Space:   progDetails.Space,
		Runtime: progDetails.Runtime,
		Name:    progDetails.Name,
		Path:    progDetails.Path,
		Project: progDetails.Project,
		Account: progDetails.Account,
		Region:  progDetails.Region,
		Deps:    progDetails.Deps,
		Envs:    progDetails.Envs,
		Public:  progDetails.Public,
	}

	fmt.Println("Cloning...")
	o, err := client.DownloadProgram(&api.DownloadProgramRequest{
		ProgramID: progInfo.ID,
		Runtime:   progInfo.Runtime,
		Account:   progInfo.Account,
		Region:    progInfo.Region,
	})

	if err != nil {
		cleanup(wd)
		return err
	}

	err = runtimeManager.WriteProgramFiles(o.Files, &wd, false)
	if err != nil {
		cleanup(wd)
		runtimeManager.Clean()
		return err
	}
	err = runtimeManager.StoreProgInfo(progInfo)
	if err != nil {
		cleanup(wd)
		return err
	}
	fmt.Println(fmt.Sprintf("Successfully cloned deta micro to '%s'", wd))
	return nil
}
