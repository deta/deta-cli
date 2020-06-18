package runtime

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	// Python runtime
	Python = "python3.7"
	// Node runtime
	Node = "nodejs12.x"

	// DefaultProject default project slug
	DefaultProject = "default"

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
	// maps runtimes to dep files
	depFiles = map[string]string{
		Python: "requirements.txt",
		Node:   "package.json",
	}
	// skipDirs maps runtimes to dirs that should be skipped
	skipPaths = map[string][]*regexp.Regexp{
		Python: []*regexp.Regexp{
			regexp.MustCompile(".*\\.pyc"),
			regexp.MustCompile(".*\\.rst"),
			regexp.MustCompile("__pycache__"),
		},
		Node: []*regexp.Regexp{
			regexp.MustCompile("node_modules"),
		},
	}

	// local paths to store information
	detaDir      = ".deta"
	userInfoFile = "user_info"
	progInfoFile = "prog_info"
	stateFile    = "state"

	// DepCommands maps runtimes to the dependency managers
	DepCommands = map[string]string{
		Python: "pip",
		Node:   "npm",
	}
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

// GetRuntime gets runtime from proginfo or figures out the runtime of the program from entrypoint file if present in the root dir
func (m *Manager) GetRuntime() (string, error) {
	progInfo, _ := m.GetProgInfo()
	if progInfo != nil {
		return progInfo.Runtime, nil
	}

	var runtime string
	var found bool
	err := filepath.Walk(m.rootDir, func(path string, info os.FileInfo, err error) error {
		if path == m.rootDir {
			return nil
		}
		if info.IsDir() {
			return filepath.SkipDir
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
		return strings.HasPrefix(filename, ".") && filename != ".", nil
	}
}

// should skip if the file or dir should be skipped
func (m *Manager) shouldSkip(path string, runtime string) (bool, error) {
	hidden, err := m.isHidden(path)
	if err != nil {
		return false, err
	}
	if hidden {
		return true, nil
	}
	for _, re := range skipPaths[runtime] {
		if match := re.MatchString(path); match {
			return true, nil
		}
	}
	return false, nil
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
	r, err := m.GetRuntime()
	if err != nil {
		return err
	}

	sm := make(stateMap)
	err = filepath.Walk(m.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		path, err = filepath.Rel(m.rootDir, path)
		if err != nil {
			return err
		}

		shouldSkip, err := m.shouldSkip(path, r)
		if err != nil {
			return err
		}

		if info.IsDir() {
			if shouldSkip {
				return filepath.SkipDir
			}
			return nil
		}
		if shouldSkip {
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
	r, err := m.GetRuntime()
	if err != nil {
		return nil, err
	}

	sc := &StateChanges{
		Changes: make(map[string]string),
	}

	err = filepath.Walk(m.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		path, err = filepath.Rel(m.rootDir, path)
		if err != nil {
			return err
		}

		shouldSkip, err := m.shouldSkip(path, r)
		if err != nil {
			return err
		}
		if info.IsDir() {
			if shouldSkip {
				return filepath.SkipDir
			}
			return nil
		}
		if shouldSkip {
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
	r, err := m.GetRuntime()
	if err != nil {
		return nil, err
	}

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
		path, err = filepath.Rel(m.rootDir, path)
		if err != nil {
			return err
		}
		shouldSkip, err := m.shouldSkip(path, r)
		if err != nil {
			return err
		}
		if info.IsDir() {
			if shouldSkip {
				return filepath.SkipDir
			}
			return nil
		}
		if shouldSkip {
			return nil
		}

		// update deletions
		if _, ok := deletions[path]; ok {
			delete(deletions, path)
		}

		checksum, err := m.calcChecksum(filepath.Join(m.rootDir, path))
		if err != nil {
			return err
		}

		if storedState[path] != checksum {
			contents, err := m.readFile(filepath.Join(m.rootDir, path))
			if err != nil {
				return err
			}
			sc.Changes[path] = string(contents)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	sc.Deletions = make([]string, len(deletions))
	i := 0
	for k := range deletions {
		sc.Deletions[i] = k
		i++
	}

	if len(sc.Changes) == 0 && len(sc.Deletions) == 0 {
		return nil, nil
	}
	return sc, nil
}

type pkgJSON struct {
	Deps map[string]string `json:"dependencies"`
}

// readDeps from the dependecy files based on runtime
func (m *Manager) readDeps(runtime string) ([]string, error) {
	depFile, ok := depFiles[runtime]
	if !ok {
		return nil, fmt.Errorf("unsupported runtime %s", runtime)
	}
	contents, err := m.readFile(filepath.Join(m.rootDir, depFile))
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
		var pj pkgJSON
		err = json.Unmarshal(contents, &pj)
		if err != nil {
			return nil, err
		}
		if len(pj.Deps) == 0 {
			return nil, nil
		}
		for k, v := range pj.Deps {
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
		return nil, fmt.Errorf("no program information found")
	}

	if progInfo.Runtime == "" {
		progInfo.Runtime, err = m.GetRuntime()
		if err != nil {
			return nil, err
		}
	}
	deps, err := m.readDeps(progInfo.Runtime)
	if err != nil {
		return nil, err
	}

	// no previous deps so return all new local deps as added
	if len(progInfo.Deps) == 0 {
		if len(deps) == 0 {
			return nil, nil
		}
		return &DepChanges{
			Added: deps,
		}, nil
	}

	var dc DepChanges

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

	if len(dc.Added) == 0 && len(dc.Removed) == 0 {
		return nil, nil
	}

	return &dc, nil
}

// readEnvs read env variables from the env file
func (m *Manager) readEnvs(envFile string) (map[string]string, error) {
	contents, err := m.readFile(filepath.Join(m.rootDir, envFile))
	if err != nil {
		return nil, err
	}
	if len(contents) == 0 {
		return nil, nil
	}
	lines := strings.Split(string(contents), "\n")
	// expect format KEY=VALUE
	envs := make(map[string]string)
	for _, l := range lines {
		l = strings.ReplaceAll(l, " ", "")
		// skip empty lines and commentes #
		if l != "" && !strings.HasPrefix(l, "#") {
			vars := strings.Split(l, "=")
			if len(vars) != 2 {
				return nil, fmt.Errorf("Env file has invalid format")
			}
			envs[vars[0]] = vars[1]
		}
	}
	return envs, nil
}

// GetEnvChanges gets changes in stored env keys and keys of the envFile
func (m *Manager) GetEnvChanges(envFile string) (*EnvChanges, error) {
	vars, err := m.readEnvs(envFile)
	if err != nil {
		return nil, err
	}

	progInfo, err := m.GetProgInfo()
	if progInfo == nil {
		return &EnvChanges{
			Vars: vars,
		}, nil
	}

	if len(progInfo.Envs) == 0 {
		return &EnvChanges{
			Vars: vars,
		}, nil
	}

	ec := EnvChanges{
		Vars: make(map[string]string),
	}

	// mark all stored envs as removed
	// mark them as unremoved later if seen them in the deps file
	removedEnvs := make(map[string]struct{}, len(progInfo.Envs))
	for _, e := range progInfo.Envs {
		removedEnvs[e] = struct{}{}
	}

	for k, v := range vars {
		if _, ok := removedEnvs[k]; ok {
			// delete from removed if seen
			delete(removedEnvs, k)
		}
		ec.Vars[k] = v
	}

	for e := range removedEnvs {
		ec.Removed = append(ec.Removed, e)
	}

	if len(ec.Vars) == 0 && len(ec.Removed) == 0 {
		return nil, nil
	}

	return &ec, nil
}

// WriteProgramFiles writes program files to target dir
func (m *Manager) WriteProgramFiles(progFiles map[string]string, targetDir *string) error {
	writeDir := m.rootDir
	// use root dir as dir to store if targetDir is not provided
	if targetDir != nil && *targetDir != writeDir {
		writeDir = filepath.Join(m.rootDir, *targetDir)
		err := os.MkdirAll(writeDir, dirPermMode)
		if err != nil {
			return err
		}
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
		_, f := filepath.Split(file)
		if f != "" {
			file = filepath.Join(writeDir, file)
			err := ioutil.WriteFile(file, []byte(content), filePermMode)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
