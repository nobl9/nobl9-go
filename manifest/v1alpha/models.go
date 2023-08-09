package v1alpha

import (
	"fmt"
	"strings"

	"github.com/nobl9/nobl9-go/manifest"
)

// StringInterpolationPlaceholder common symbol to use in strings for interpolation e.g. "My amazing {} Service"
const StringInterpolationPlaceholder = "{}"

// ObjectInternal represents part of object which is only for internal usage,
// not exposed to the client
// Deprecated
type ObjectInternal struct {
	Organization string `json:"organization,omitempty" example:"nobl9-dev"`
	ManifestSrc  string `json:",omitempty" example:"x.yml"`
}

// Metadata represents part of object which is common for all available Objects, for internal usage
// Deprecated
type Metadata struct {
	Name        string `json:"name" validate:"required,objectName" example:"name"`
	DisplayName string `json:"displayName,omitempty" validate:"omitempty,min=0,max=63" example:"Prometheus Source"`
	Project     string `json:"project,omitempty" validate:"objectName" example:"default"`
	Labels      Labels `json:"labels,omitempty" validate:"omitempty,labels"`
}

// FullName returns full name of an object as `{name}.{project}`
func (m Metadata) FullName() string {
	return fmt.Sprintf("%s.%s", m.Name, m.Project)
}

// MetadataHolder is an intermediate structure that can provides metadata related
// field to other structures
// Deprecated
type MetadataHolder struct {
	Metadata Metadata `json:"metadata"`
}

// ObjectHeader represents Header which is common for all available Objects
// Deprecated
type ObjectHeader struct {
	APIVersion     string        `json:"apiVersion" validate:"required" example:"n9/v1alpha"`
	Kind           manifest.Kind `json:"kind" validate:"required" example:"kind"`
	MetadataHolder `json:",inline"`
	ObjectInternal `json:",inline"`
}

// StringInterpolation for arguments ("{}-my-{}-string-{}", "xd") returns string xd-my-xd-string-xd
func StringInterpolation(withPlaceholder, replacer string) string {
	return strings.ReplaceAll(withPlaceholder, StringInterpolationPlaceholder, replacer)
}
