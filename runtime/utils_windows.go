// +build windows

package runtime

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

func (m *Manager) isHiddenWindows(path string) (bool, error) {
	_, filename := filepath.Split(path)
	// consider paths starting with "." also hidden in windows
	if strings.HasPrefix(filename, ".") && filename != "." {
		return true, nil
	}

	if !filepath.IsAbs(path){
		path = filepath.Join(m.rootDir, path)
	}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist){
		// files that don't exist are taken as hidden
		return true, nil
	}

	filePtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return false, err
	}
	attrs, err := windows.GetFileAttributes(filePtr)
	if err != nil {
		return false, err
	}
	return attrs&windows.FILE_ATTRIBUTE_HIDDEN != 0, nil
}
