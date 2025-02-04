//go:build !windows

package stringutils

// RemoveCR by default does nothing, it only modifies the string if built for Windows.
func RemoveCR(s string) string { return s }
