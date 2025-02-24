//go:build !windows

package stringutils

// RemoveCR by default does nothing.
// It only modifies the string if the binary is built for Windows.
// See stringutils_windows.go for more details.
func RemoveCR(s string) string { return s }
