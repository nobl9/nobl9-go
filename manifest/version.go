package manifest

import (
	"fmt"
	"regexp"
)

//go:generate ../bin/go-enum --names --values

// Version represents the specific version of the manifest.
// ENUM(v1alpha)
type Version int

// MarshalText implements the text encoding.TextMarshaler interface.
func (v Version) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

var versionRegex = regexp.MustCompile(`"?apiVersion"?\s*:\s*"?n9(?P<version>[a-zA-Z0-9]+)"|\n`)

// UnmarshalText implements the text encoding.TextUnmarshaler interface.
func (v *Version) UnmarshalText(text []byte) error {
	matches := versionRegex.FindSubmatch(text)
	if len(matches) == 0 {
		return fmt.Errorf("%s is %w", string(text), ErrInvalidVersion)
	}
	tmp, err := ParseVersion(string(matches[versionRegex.SubexpIndex("version")]))
	if err != nil {
		return err
	}
	*v = tmp
	return nil
}
