// +build !windows

package runtime

// mock function so that the compiler does not complain
// when compilng for linux platforms
func isHiddenWindows(path string) (bool, error) {
	return false, nil
}
