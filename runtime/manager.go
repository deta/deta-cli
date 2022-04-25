package runtime

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	SPACE    = " "
	NEGATION = '!'

	PythonSkipPattern = `(.*\.pyc)|(.*\.rst)|(__pycache__)|(.*~$)|(.*\.deta)|(^.*env$)|(^.*venv$)`

	NodeSkipPattern = `(node_modules)|(.*~$)|(.*\.deta)`

	Python = "python"
	Node   = "node"

	// DefaultProject default project slug
	DefaultProject = "default"

	// drwxrw----
	dirPermMode = 0760
	// -rw-rw---
	filePermMode = 0660
)

type Pattern struct {
	Value *regexp.Regexp
	Skip  bool
}

var (
	// supported runtimes Note: index 0 is the default runtime
	runtimes = map[string][]string{
		Python: {"python3.9", "python3.8", "python3.7"},
		Node:   {"nodejs14.x", "nodejs12.x"},
	}

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

	// maps lib entry files to runtimes
	libEntryFiles = map[string][]string{
		Python: []string{"_entry.py"},
		Node:   []string{"_entry.js"},
	}

	// skipPaths maps runtimes to paths that should be skipped
	skipPaths = map[string][]Pattern{
		Python: {
			Pattern{
				Value: regexp.MustCompilePOSIX(PythonSkipPattern),
				Skip:  true,
			},
		},
		Node: {
			Pattern{
				Value: regexp.MustCompilePOSIX(NodeSkipPattern),
				Skip:  true,
			},
		},
	}

	// local paths to store information
	detaDir      = ".deta"
	userInfoFile = "user_info"
	progInfoFile = "prog_info"
	stateFile    = "state"
	ignoreFile   = ".detaignore"

	// DepCommands maps runtimes to the dependency managers
	DepCommands = map[string]string{
		Python: "pip",
		Node:   "npm",
	}

	// ErrNoEntrypoint noe entrypoint file present
	ErrNoEntrypoint = errors.New("no entrypoint file present")
	// ErrEntrypointConflict conflicting entrypoint files
	ErrEntrypointConflict = errors.New("conflicting entrypoint files present")
)

// Manager runtime manager handles files management and other services
type Manager struct {
	rootDir      string               // working directory for the program
	detaPath     string               // dir for storing program info and state
	userInfoPath string               // path to info file about the user
	progInfoPath string               // path to info file about the program
	statePath    string               // path to state file about the program
	ignorePath   string               // path to .detaignore file
	skipPaths    map[string][]Pattern // files that will be skipped
}

// Runtime holds name and version of current runtime used
type Runtime struct {
	Name    string
	Version string
}

// NewManager returns a new runtime manager for the root dir of the program
// if initDirs is true, it creates dirs under root
func NewManager(root *string, initDirs bool) (*Manager, error) {
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

	if initDirs {
		err := os.MkdirAll(detaPath, dirPermMode)
		if err != nil {
			return nil, err
		}
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

	ignorePath := filepath.Join(rootDir, ignoreFile)

	manager := &Manager{
		rootDir:      rootDir,
		detaPath:     detaPath,
		userInfoPath: userInfoPath,
		progInfoPath: filepath.Join(detaPath, progInfoFile),
		statePath:    filepath.Join(detaPath, stateFile),
		skipPaths:    skipPaths,
		ignorePath:   ignorePath,
	}

	// not handling error as we don't want cli to crash if .detaignore is not found
	manager.handleIgnoreFile()

	return manager, nil
}

func (m *Manager) handleIgnoreFile() error {
	runtime, err := m.GetRuntime()
	if err != nil {
		return err
	}

	contents, err := m.readFile(m.ignorePath)
	if err != nil {
		return err
	}

	lines, err := readLines(contents)
	if err != nil {
		return err
	}

	for _, line := range lines {
		line = strings.Trim(line, SPACE)
		if len(line) != 0 {
			value, err := regexp.CompilePOSIX(line)
			if err != nil {
				continue // ignore current line and continue
			}

			pattern := Pattern{
				Value: value,
				Skip:  true,
			}

			firstChar := line[0]
			if firstChar == NEGATION {
				remainingChar := line[1:]
				if len(remainingChar) == 0 {
					continue // ignore current line and continue
				}

				value, err := regexp.CompilePOSIX(remainingChar)
				if err != nil {
					continue // ignore current line and continue
				}

				pattern = Pattern{
					Value: value,
					Skip:  false,
				}
			}

			m.skipPaths[runtime.Name] = append([]Pattern{pattern}, m.skipPaths[runtime.Name]...)
		}
	}

	return nil
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

	progInfo, err := progInfoFromBytes(contents)
	if err != nil {
		return nil, err
	}

	runtime, err := CheckRuntime(progInfo.Runtime)
	if err != nil {
		return nil, err
	}

	progInfo.RuntimeName = runtime.Name
	progInfo.Runtime = runtime.Version

	return progInfo, nil
}

// StoreUserInfo stores the user info
func (m *Manager) StoreUserInfo(u *UserInfo) error {
	marshalled, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(m.userInfoPath, marshalled, filePermMode)
}

// GetUserInfo gets the user info
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
	if err != nil {
		if err == io.EOF {
			return true, nil
		}
		return false, err
	}
	for _, n := range names {
		isHidden, err := m.isHidden(filepath.Join(checkDir, n))
		if err != nil {
			return false, err
		}
		if !isHidden {
			return false, nil
		}
	}
	return true, nil
}

// CheckRuntime checks if given runtime is supported
func CheckRuntime(runtime string) (*Runtime, error) {
	for k, v := range runtimes {
		if contains(v, runtime) {
			return &Runtime{Name: k, Version: runtime}, nil
		}
	}
	return nil, fmt.Errorf("unsupported runtime")
}

// GetDefaultRuntimeVersion returns default runtime version
func GetDefaultRuntimeVersion(name string) string {
	return runtimes[name][0] // index 0 is the default runtime
}

// GetRuntime gets runtime from proginfo or figures out the runtime of the program from entrypoint file if present in the root dir
func (m *Manager) GetRuntime() (*Runtime, error) {
	progInfo, _ := m.GetProgInfo()
	if progInfo != nil {
		return &Runtime{
			Name:    progInfo.RuntimeName,
			Version: progInfo.Runtime,
		}, nil
	}

	var runtime *Runtime
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
				runtime = &Runtime{
					Name:    r,
					Version: GetDefaultRuntimeVersion(r),
				}
			} else {
				return errors.New("conflicting entrypoint files found")
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrNoEntrypoint
	}
	return runtime, nil
}

// if a file or dir is hidden
func (m *Manager) isHidden(path string) (bool, error) {
	_, filename := filepath.Split(path)
	return strings.HasPrefix(filename, ".") && filename != ".", nil
}

// should skip if the file or dir should be skipped
func (m *Manager) shouldSkip(path string, runtime string) (bool, error) {
	// do not skip .detaignore file
	if regexp.MustCompile(ignoreFile).MatchString(path) {
		return false, nil
	}

	for _, re := range m.skipPaths[runtime] {
		if re.Value.MatchString(filepath.ToSlash(path)) {
			return re.Skip, nil
		}
	}

	hidden, err := m.isHidden(path)
	if err != nil {
		return false, err
	}

	return hidden, nil
}

// reads the contents of a file, returns contents
func (m *Manager) readFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

// reads the contents of a file, returns contents and if file is binary or not
func (m *Manager) readFileIsBinary(path string) ([]byte, bool, error) {
	contents, err := m.readFile(path)
	if err != nil {
		return nil, false, err
	}
	return contents, isBinary(contents), nil
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

		shouldSkip, err := m.shouldSkip(path, r.Name)
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

		hashSum, err := m.calcChecksum(filepath.Join(m.rootDir, path))
		if err != nil {
			return err
		}
		sm[filepath.ToSlash(path)] = hashSum
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
		Changes:     make(map[string]string),
		BinaryFiles: make(map[string]string),
	}

	err = filepath.Walk(m.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		path, err = filepath.Rel(m.rootDir, path)
		if err != nil {
			return err
		}

		shouldSkip, err := m.shouldSkip(path, r.Name)
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

		contents, isBinary, err := m.readFileIsBinary(filepath.Join(m.rootDir, path))
		if err != nil {
			return err
		}

		if isBinary {
			sc.BinaryFiles[filepath.ToSlash(path)] = base64.StdEncoding.EncodeToString(contents)
		} else {
			sc.Changes[filepath.ToSlash(path)] = string(contents)
		}
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
		Changes:     make(map[string]string),
		BinaryFiles: make(map[string]string),
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
		shouldSkip, err := m.shouldSkip(path, r.Name)
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
		if _, ok := deletions[filepath.ToSlash(path)]; ok {
			delete(deletions, filepath.ToSlash(path))
		}

		checksum, err := m.calcChecksum(filepath.Join(m.rootDir, path))
		if err != nil {
			return err
		}

		if storedState[filepath.ToSlash(path)] != checksum {
			contents, isBinary, err := m.readFileIsBinary(filepath.Join(m.rootDir, path))
			if err != nil {
				return err
			}
			if isBinary {
				sc.BinaryFiles[filepath.ToSlash(path)] = base64.StdEncoding.EncodeToString(contents)
			} else {
				sc.Changes[filepath.ToSlash(path)] = string(contents)
			}
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

	if len(sc.Changes) == 0 && len(sc.Deletions) == 0 && len(sc.BinaryFiles) == 0 {
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
		return nil, fmt.Errorf("unsupported runtime '%s'", runtime)
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
		var deps []string

		lines, err := readLines(contents)
		if err != nil {
			return nil, err
		}

		for _, l := range lines {
			l = strings.ReplaceAll(l, " ", "")
			// skip empty lines and commentes #
			if l != "" && !strings.HasPrefix(l, "#") {
				deps = append(deps, l)
			}
		}
		return deps, nil
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
	if progInfo == nil || err != nil {
		return nil, fmt.Errorf("no program information found")
	}

	if progInfo.Runtime == "" {
		rtime, err := m.GetRuntime()
		if err != nil {
			return nil, err
		}
		progInfo.RuntimeName = rtime.Name
		progInfo.Runtime = rtime.Version
	}
	deps, err := m.readDeps(progInfo.RuntimeName)
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
	envs := make(map[string]string)

	lines, err := readLines(contents)
	if err != nil {
		return nil, err
	}

	for n, l := range lines {
		// skip empty lines and commentes #
		if l != "" && !strings.HasPrefix(l, "#") {
			sepIndex := strings.Index(l, "=")
			// expect format KEY=VALUE
			if sepIndex <= 0 {
				return nil, fmt.Errorf("unexpected format in line %d, expected KEY=VALUE", n+1)
			}
			key := l[0:sepIndex]
			value := l[sepIndex+1:]
			envs[key] = strings.Trim(value, "\"")
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

// WriteProgramFiles writes program files to target dir, target dir is relative to root dir if relative is true
func (m *Manager) WriteProgramFiles(zipFile []byte, targetDir *string, relative bool, runtimeVersion string) error {
	var writeDir string
	if relative {
		writeDir = m.rootDir
		// use root dir as dir to store if targetDir is not provided
		if targetDir != nil && *targetDir != writeDir {
			writeDir = filepath.Join(m.rootDir, *targetDir)
			err := os.MkdirAll(writeDir, dirPermMode)
			if err != nil {
				return err
			}
		}
	} else {
		if targetDir == nil {
			return fmt.Errorf("target dir not provided")
		}
		writeDir = *targetDir
	}

	runtime, err := CheckRuntime(runtimeVersion)
	if err != nil {
		return err
	}

	// unzip zip file into wrtie dir skipping lib entry file for runtime
	return unzip(zipFile, writeDir, libEntryFiles[runtime.Name])
}

// Clean removes `.deta` folder created by the runtime manager if it's empty
func (m *Manager) Clean() error {
	isEmpty, err := isDirEmpty(m.detaPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	// only delete `m.detaPath` if it's empty
	if isEmpty {
		return os.RemoveAll(m.detaPath)
	}
	return nil
}
