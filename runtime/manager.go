package runtime

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

const (
	// Python runtime
	Python = "python"
	// Node runtime
	Node = "node"

	// drwxrw----
	dirPermMode = 0760
	// -rw-rw---
	filePermMode = 0660
)

var (
	// maps entrypoint files to runtimes
	entryPoints = map[string]string{
		"main.py":  Python,
		"index.js": Node,
	}
	depFiles = map[string]string{
		Python: "requirements.txt",
		Node:   "package.json",
	}
	detaDir      = ".deta"
	userInfoFile = "user_info"
	progInfoFile = "prog_info"
	stateFile    = "state"
)

// Manager runtime manager handles files management and other services
type Manager struct {
	rootDir      string // working directory for the program
	detaPath     string // dir for storing program info and state
	userInfoPath string // path to info file about the user
	progInfoPath string // path to info file about the program
	statePath    string // path to state file about the program
}

// NewManager returns a new runtime manager for the root dir of the program
func NewManager(root *string) (*Manager, error) {
	var rootDir string
	if root != nil {
		rootDir = *root
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		rootDir = wd
	}

	detaPath := filepath.Join(rootDir, detaDir)
	err := os.MkdirAll(detaPath, dirPermMode)
	if err != nil {
		return nil, err
	}

	// user info is stored in ~/.deta/userInfo as it's global
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(filepath.Join(home, detaDir), dirPermMode)
	if err != nil {
		return nil, err
	}
	userInfoPath := filepath.Join(home, detaDir, userInfoFile)

	return &Manager{
		rootDir:      rootDir,
		detaPath:     detaPath,
		userInfoPath: userInfoPath,
		progInfoPath: filepath.Join(detaPath, progInfoFile),
		statePath:    filepath.Join(detaPath, stateFile),
	}, nil
}

// StoreProgInfo stores program info to disk
func (m *Manager) StoreProgInfo(p *ProgInfo) error {
	marshalled, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(m.progInfoPath, marshalled, filePermMode)
}

// GetProgInfo gets the program info stored
func (m *Manager) GetProgInfo() (*ProgInfo, error) {
	contents, err := m.readFile(m.progInfoPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return progInfoFromBytes(contents)
}

// StoreUserInfo stores the user info
func (m *Manager) StoreUserInfo(u *UserInfo) error {
	marshalled, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(m.userInfoPath, marshalled, filePermMode)
}

// GetUserInfo gets the program info stored
func (m *Manager) GetUserInfo() (*UserInfo, error) {
	contents, err := m.readFile(m.userInfoPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return userInfoFromBytes(contents)
}

// IsInitialized checks if the root directory is initialized as a deta program
func (m *Manager) IsInitialized() (bool, error) {
	_, err := os.Stat(m.progInfoPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// IsProgDirEmpty checks if dir contains any files/folders which are not hidden
// if dir is nil, it sets the root dir
func (m *Manager) IsProgDirEmpty() (bool, error) {
	checkDir := m.rootDir
	f, err := os.Open(checkDir)
	if err != nil {
		return false, err
	}
	defer f.Close()
	names, err := f.Readdirnames(-1)
	if err == io.EOF {
		return true, nil
	}
	for _, n := range names {
		isHidden, err := m.isHidden(n)
		if err != nil {
			return false, err
		}
		if !isHidden {
			return false, nil
		}
	}
	return true, nil
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

// reads the contents of a file
func (m *Manager) readFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	contents, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return contents, nil
}

// calculates the sha256 sum of contents of file in path
func (m *Manager) calcChecksum(path string) (string, error) {
	contents, err := m.readFile(path)
	if err != nil {
		return "", err
	}
	hashSum := fmt.Sprintf("%x", sha256.Sum256(contents))
	return hashSum, nil
}

// StoreState stores hashes of the current state of all files(not hidden) in the root program directory
func (m *Manager) StoreState() error {
	sm := make(stateMap)
	err := filepath.Walk(m.rootDir, func(path string, info os.FileInfo, err error) error {
		hidden, err := m.isHidden(path)
		if err != nil {
			return err
		}

		if info.IsDir() {
			// skip hidden directories
			if hidden {
				return filepath.SkipDir
			}
			return nil
		}
		// skip hidden files
		if hidden {
			return nil
		}

		hashSum, err := m.calcChecksum(path)
		if err != nil {
			return err
		}
		sm[path] = hashSum
		return nil
	})
	if err != nil {
		return err
	}

	marshalled, err := json.Marshal(sm)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(m.statePath, marshalled, filePermMode)
	if err != nil {
		return err
	}
	return nil
}

// gets the current stored state
func (m *Manager) getStoredState() (stateMap, error) {
	contents, err := m.readFile(m.statePath)
	if err != nil {
		return nil, err
	}
	s, err := stateMapFromBytes(contents)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// readAll reads all the files and returns the contents as stateChanges
func (m *Manager) readAll() (*StateChanges, error) {
	sc := &StateChanges{
		Changes: make(map[string]string),
	}

	err := filepath.Walk(m.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		hidden, err := m.isHidden(path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			if hidden {
				return filepath.SkipDir
			}
			return nil
		}
		if hidden {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		contents, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}
		sc.Changes[path] = string(contents)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return sc, nil
}

// GetChanges checks if the state has changed in the root directory
func (m *Manager) GetChanges() (*StateChanges, error) {
	sc := &StateChanges{
		Changes: make(map[string]string),
	}

	storedState, err := m.getStoredState()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return m.readAll()
		}
		return nil, err
	}

	// mark all paths in current state as deleted
	// if seen later on walk, remove from deletions
	deletions := make(map[string]struct{}, len(storedState))
	for k := range storedState {
		deletions[k] = struct{}{}
	}

	err = filepath.Walk(m.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		hidden, err := m.isHidden(path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			if hidden {
				return filepath.SkipDir
			}
			return nil
		}
		if hidden {
			return nil
		}

		// update deletions
		if _, ok := deletions[path]; ok {
			delete(deletions, path)
		}

		checksum, err := m.calcChecksum(path)
		if err != nil {
			return err
		}

		if storedState[path] != checksum {
			contents, err := m.readFile(path)
			if err != nil {
				return err
			}
			sc.Changes[path] = string(contents)
		}
		return nil
	})

	sc.Deletions = make([]string, len(deletions))
	i := 0
	for k := range deletions {
		sc.Deletions[i] = k
		i++
	}
	return sc, nil
}

// readDeps from the dependecy files based on runtime
func (m *Manager) readDeps(runtime string) ([]string, error) {
	contents, err := m.readFile(depFiles[runtime])
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	switch runtime {
	case Python:
		return strings.Split(string(contents), "\n"), nil
	case Node:
		var nodeDeps []string
		var pkgJSON map[string]interface{}
		err = json.Unmarshal(contents, &pkgJSON)
		if err != nil {
			return nil, err
		}
		deps, ok := pkgJSON["dependencies"]
		if !ok {
			return nil, nil
		}
		if reflect.TypeOf(deps).String() != "map[string]string" {
			return nil, fmt.Errorf("'package.json' is of unexpected format")
		}

		for k, v := range deps.(map[string]string) {
			nodeDeps = append(nodeDeps, fmt.Sprintf("%s@%s", k, v))
		}
		return nodeDeps, nil
	default:
		return nil, fmt.Errorf("unsupported runtime '%s'", runtime)
	}
}

// GetDepChanges gets dependencies from program
func (m *Manager) GetDepChanges() (*DepChanges, error) {
	progInfo, err := m.GetProgInfo()
	if progInfo == nil {
		runtime, err := m.GetRuntime()
		if err != nil {
			return nil, err
		}
		deps, err := m.readDeps(runtime)
		if err != nil {
			return nil, err
		}
		return &DepChanges{
			Added: deps,
		}, nil
	}

	if len(progInfo.Deps) == 0 {
		if progInfo.Runtime == "" {
			progInfo.Runtime, err = m.GetRuntime()
		}
		deps, err := m.readDeps(progInfo.Runtime)
		if err != nil {
			return nil, err
		}
		return &DepChanges{
			Added: deps,
		}, nil
	}

	var dc DepChanges

	deps, err := m.readDeps(progInfo.Runtime)
	if err != nil {
		return nil, err
	}

	// mark all stored deps as removed deps
	// mark them as unremoved later if seen them in the deps file
	removedDeps := make(map[string]struct{}, len(progInfo.Deps))
	for _, d := range progInfo.Deps {
		removedDeps[d] = struct{}{}
	}

	for _, d := range deps {
		if _, ok := removedDeps[d]; ok {
			// remove from deleted if seen
			delete(removedDeps, d)
		} else {
			// add as new dep if not seen
			dc.Added = append(dc.Added, d)
		}
	}

	for d := range removedDeps {
		dc.Removed = append(dc.Removed, d)
	}
	return &dc, nil
}

// WriteProgramFiles writes program files to target dir
func (m *Manager) WriteProgramFiles(progFiles map[string]string, targetDir *string) error {
	writeDir := m.rootDir
	// use root dir as dir to store if targetDir is not provided
	if targetDir != nil && *targetDir != writeDir {
		writeDir = filepath.Join(m.rootDir, *targetDir)
	}

	// need to create dirs first before writing the files
	for f := range progFiles {
		dir, _ := filepath.Split(f)
		if dir != "" {
			dir = filepath.Join(writeDir, dir)
			err := os.MkdirAll(dir, dirPermMode)
			if err != nil {
				return err
			}
		}
	}

	// write the files
	for file, content := range progFiles {
		file = filepath.Join(writeDir, file)
		err := ioutil.WriteFile(file, []byte(content), filePermMode)
		if err != nil {
			return err
		}
	}
	return nil
}
