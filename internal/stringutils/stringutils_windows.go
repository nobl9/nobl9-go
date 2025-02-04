//go:build windows

package stringutils

import (
	"strings"
)

// RemoveCR removes carriage return which is part of the character sequence
// signifying new line on Windows.
func RemoveCR(s string) string {
	return strings.ReplaceAll(s, "\r", "")
}
