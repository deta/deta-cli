package runtime

import (
	"archive/zip"
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// other binary extensions
var otherBinaryExts = map[string]struct{}{
	".mo": {},
}

func readLines(data []byte) ([]string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

// contains checks if the given string exists on given array
func contains(arr []string, str string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}

	return false
}

// checks if dir is empty
func isDirEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	names, err := f.Readdirnames(-1)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return true, nil
		}
		return false, err
	}
	return len(names) == 0, nil
}

// unzip unzips a zip file into dest skipping skipFileNames
func unzip(zipFile []byte, dest string, skipFileNames []string) error {
	// map for faster lookup later
	skipFilesMap := make(map[string]struct{})
	for _, name := range skipFileNames {
		skipFilesMap[name] = struct{}{}
	}

	r, err := zip.NewReader(bytes.NewReader(zipFile), int64(len(zipFile)))
	if err != nil {
		return err
	}

	for _, f := range r.File {
		// skip if in skipFilesMap
		if _, ok := skipFilesMap[strings.TrimPrefix(f.Name, string(os.PathSeparator))]; ok {
			continue
		}

		fpath := filepath.Join(dest, f.Name)
		// check for zipslip
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return errors.New("zipslip in zip file")
		}

		// make folder if is a folder
		if f.FileInfo().IsDir() {
			err = os.MkdirAll(fpath, os.ModePerm)
			if err != nil {
				return err
			}
			continue
		}

		// make and copy file
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		copyDest, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		srcFile, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(copyDest, srcFile)
		if err != nil {
			return err
		}

		// close files without defer to close before next iteration of loop
		copyDest.Close()
		srcFile.Close()

	}
	return nil
}
