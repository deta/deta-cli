// +build windows

package runtime

import (
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

const NewLine = "\r\n"

// other binary extensions
var otherBinaryExts = map[string]struct{}{
	".mo": {},
}

func isHiddenWindows(path string) (bool, error) {
	_, filename := filepath.Split(path)
	// consider paths starting with "." also hidden in windows
	if strings.HasPrefix(filename, ".") && filename != "." {
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
