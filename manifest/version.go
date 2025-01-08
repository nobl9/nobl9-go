package manifest

import "strings"

//go:generate ../bin/go-enum --names --values --marshal

// Version represents the specific version of the manifest.
// ENUM(v1alpha = n9/v1alpha)
type Version string

// VersionString returns the second element of the [Version].
// For example, given "n9/v1alpha", it returns "v1alpha".
func (v Version) VersionString() string {
	return strings.TrimPrefix(v.String(), "n9/")
}
