package testutils

import (
	"runtime"
	"strings"
)

// RemoveCR removes carriage return which is part of the character sequence
// signifying new line on Windows.
func RemoveCR(s string) string {
	if runtime.GOOS == "windows" {
		return strings.ReplaceAll(s, "\r", "")
	}
	return s
}
