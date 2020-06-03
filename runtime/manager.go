package runtime

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	entryPoints = map[string]string{
		"main.py":  "python",
		"index.js": "node",
	}
	depFiles = map[string]string{
		"python": "requirements.txt",
		"node":   "package.json",
	}
	detaDir    = ".deta"
	configFile = "config"
	stateFile  = "state"
)

// Manager runtime manager handles files management and other services
type Manager struct {
	rootDir    string // working directory for the program
	detaPath   string // dir for assisting in runtime management
	configPath string // configuration path
	statePath  string // state path
}

// NewManager returns a new runtime manager for the root dir of the program
func NewManager(rootDir string) (*Manager, error) {
	detaPath := filepath.Join(rootDir, detaDir)
	err := os.MkdirAll(detaPath, 0760)
	if err != nil {
		return nil, err
	}
	return &Manager{
		rootDir:    rootDir,
		detaPath:   detaPath,
		configPath: filepath.Join(detaPath, configFile),
		statePath:  filepath.Join(detaPath, stateFile),
	}, nil
}

// GetRuntime figures out the runtime of the program from entrypoint file if present in the root dir
func (m *Manager) GetRuntime() (string, error) {
	var runtime string
	var found bool

	err := filepath.Walk(m.rootDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		_, filename := filepath.Split(path)
		if r, ok := entryPoints[filename]; ok {
			if !found {
				found = true
				runtime = r
			} else {
				return errors.New("Conflicting entrypoint files found")
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if !found {
		return "", fmt.Errorf("No supported runtime found in %s", m.rootDir)
	}
	return runtime, nil
}

// if a file or dir is hidden
func (m *Manager) isHidden(path string) (bool, error) {
	_, filename := filepath.Split(path)
	switch runtime.GOOS {
	case "windows":
		// TODO: implement for windows
		return false, fmt.Errorf("Not implemented")
	default:
		return strings.HasPrefix(filename, "."), nil
	}
}

// stores hashes of the current state of all files in the root directory
func (m *Manager) storeState() error {
	checksumMap := make(map[string]string)
	err := filepath.Walk(m.rootDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		contents, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}
		hashSum := fmt.Sprintf("%x", sha256.Sum256(contents))
		checksumMap[path] = hashSum
		return nil
	})
	if err != nil {
		return err
	}

	marshalled, err := json.Marshal(checksumMap)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(m.statePath, marshalled, 0760)
	if err != nil {
		return err
	}
	return nil
}
