// +build !windows

package runtime

// mock function so that the compiler does not complain
// when compiling for linux platforms
func (m *Manager) isHiddenWindows(path string) (bool, error) {
	return false, nil
}
