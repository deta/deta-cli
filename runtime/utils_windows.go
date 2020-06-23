// +build windows

package runtime

import (
	"path/filepath"

	"golang.org/x/sys/windows"
)

func isHiddenWindows(path string) (bool, error) {
	_, filename := filepath.Split(path)
	filePtr, err := windows.UTF16PtrFromString(filename)
	if err != nil {
		return false, err
	}
	attrs, err := windows.GetFileAttributes(filePtr)
	if err != nil {
		return false, err
	}
	return attrs&windows.FILE_ATTRIBUTE_HIDDEN != 0, nil
}
