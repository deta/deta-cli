// +build !windows

package runtime

// NewLine new line in unix
const NewLine = "\n"

// mock function so that the compiler does not complain
// when compilng for linux platforms
func isHiddenWindows(path string) (bool, error) {
	return false, nil
}
