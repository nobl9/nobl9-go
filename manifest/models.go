// Package manifest provides
package manifest

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// StringInterpolationPlaceholder common symbol to use in strings for interpolation e.g. "My amazing {} Service"
const StringInterpolationPlaceholder = "{}"

// ObjectInternal represents part of object which is only for internal usage,
// not exposed to the client, for internal usage
type ObjectInternal struct {
	Organization string `json:"organization,omitempty" example:"nobl9-dev"`
	ManifestSrc  string `json:",omitempty" example:"x.yml"`
}

type ObjectInternalI interface {
	GetOrganization() string
	GetManifestSource() string
	SetOrganization(org string)
	SetManifestSource(src string)
}

type LabelKey = string
type LabelValue = string
type Labels map[LabelKey][]LabelValue

// AlertSilenceMetadata defines only basic metadata fields - name and project which uniquely identifies
// object on project level.
type AlertSilenceMetadata struct {
	Name    string `json:"name" validate:"required,objectName" example:"name"`
	Project string `json:"project,omitempty" validate:"objectName" example:"default"`
}

// Metadata represents part of object which is common for all available Objects, for internal usage
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
type MetadataHolder struct {
	Metadata Metadata `json:"metadata"`
}

type ProjectMetadata struct {
	Name        string `json:"name" validate:"required,objectName" example:"name"`
	DisplayName string `json:"displayName,omitempty" validate:"omitempty,min=0,max=63" example:"Shopping App"`
	Labels      Labels `json:"labels,omitempty" validate:"omitempty,labels"`
}

type RoleBindingMetadata struct {
	Name string `json:"name" validate:"required,objectName" example:"name"`
}

// ObjectHeader represents Header which is common for all available Objects
type ObjectHeader struct {
	APIVersion string `json:"apiVersion" validate:"required" example:"n9/v1alpha"`
	Kind       Kind   `json:"kind" validate:"required" example:"kind"`
	MetadataHolder
	ObjectInternal
}

// ObjectGeneric represents struct to which every Object is parsable
// Specific types of Object have different structures as Spec
type ObjectGeneric struct {
	ObjectHeader
	Spec json.RawMessage `json:"spec"`
}

// JSONToGenericObjects parse JSON Array of Objects into generic objects
func JSONToGenericObjects(jsonPayload []byte) ([]ObjectGeneric, error) {
	var objects []ObjectGeneric
	if err := json.Unmarshal(jsonPayload, &objects); err != nil {
		if stxErr, ok := err.(*json.SyntaxError); ok {
			return nil, fmt.Errorf("malformed JSON payload, syntax error: %s offset: %d", stxErr.Error(), stxErr.Offset)
		}
		return nil, errors.New("malformed JSON payload - pass single list of valid JSON objects")
	}
	return objects, nil
}

// UnsupportedKindErr returns appropriate error for missing value in field kind
// for not empty field kind returns always that is not supported for this apiVersion
// so have to be validated before
func UnsupportedKindErr(o ObjectGeneric) error {
	if strings.TrimSpace(o.Kind.String()) == "" {
		return EnhanceError(o, errors.New("missing or empty field kind for an Object"))
	}
	return EnhanceError(o, fmt.Errorf("invalid Object kind: %s for apiVersion: %s", o.Kind, o.APIVersion))
}

// UnsupportedAPIVersionErr returns appropriate error for missing value in field apiVersion
// for not empty field apiVersion returns always that this version is not supported so have to be
// validated before
func UnsupportedAPIVersionErr(o ObjectGeneric) error {
	if strings.TrimSpace(o.APIVersion) == "" {
		return EnhanceError(o, errors.New("missing or empty field apiVersion for an Object"))
	}
	return EnhanceError(o, fmt.Errorf("not supported apiVersion: %s", o.APIVersion))
}

// EnhanceError annotates error with path of manifest source, if it exists
// if not returns the same error as passed as argument
func EnhanceError(o ObjectGeneric, err error) error {
	if err != nil && o.ManifestSrc != "" {
		err = fmt.Errorf("%s: %w", o.ManifestSrc, err)
	}
	return err
}

// StringInterpolation for arguments ("{}-my-{}-string-{}", "xd") returns string xd-my-xd-string-xd
func StringInterpolation(withPlaceholder, replacer string) string {
	return strings.ReplaceAll(withPlaceholder, StringInterpolationPlaceholder, replacer)
}
