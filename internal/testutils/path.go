package testutils

import (
	"os"
	"path/filepath"
)

// FindModuleRoot returns the absolute path to the modules root.
func FindModuleRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	dir = filepath.Clean(dir)
	for {
		if fi, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && !fi.IsDir() {
			return dir
		}
		d := filepath.Dir(dir)
		if d == dir {
			break
		}
		dir = d
	}
	return ""
}
