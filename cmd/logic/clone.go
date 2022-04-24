package logic

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/deta/deta-cli/api"
	"github.com/deta/deta-cli/runtime"
)

func Clone(client *api.DetaClient, progName string, projectName string, args []string) error {
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

	u, err := getUserInfo(runtimeManager, client)
	if err != nil {
		return err
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

	var cronExpression string
	if progDetails.ScheduleID > 0 {
		schedule, err := client.GetSchedule(&api.GetScheduleRequest{
			ProgramID: progDetails.ID,
		})
		if err != nil {
			cleanup(wd)
			return err
		}
		cronExpression = schedule.Expression
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
		Visor:   progDetails.Visor,
		Cron:    cronExpression,
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

	err = runtimeManager.WriteProgramFiles(o.ZipFile, &wd, false, progInfo.Runtime)
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
	err = runtimeManager.StoreState()
	if err != nil {
		cleanup(wd)
		return err
	}
	fmt.Printf("Successfully cloned deta micro to '%s'\n", wd)
	return nil
}