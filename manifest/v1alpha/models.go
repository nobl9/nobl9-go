// Package manifest provides
package v1alpha

import (
	"fmt"
	"strings"

	"github.com/nobl9/nobl9-go/manifest"
)

// StringInterpolationPlaceholder common symbol to use in strings for interpolation e.g. "My amazing {} Service"
const StringInterpolationPlaceholder = "{}"

// ObjectInternal represents part of object which is only for internal usage,
// not exposed to the client, for internal usage
type ObjectInternal struct {
	Organization string `json:"organization,omitempty" example:"nobl9-dev"`
	ManifestSrc  string `json:",omitempty" example:"x.yml"`
}

type LabelKey = string
type LabelValue = string
type Labels map[LabelKey][]LabelValue

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

// ObjectHeader represents Header which is common for all available Objects
type ObjectHeader struct {
	APIVersion string        `json:"apiVersion" validate:"required" example:"n9/v1alpha"`
	Kind       manifest.Kind `json:"kind" validate:"required" example:"kind"`
	MetadataHolder
	ObjectInternal
}

//// ObjectGeneric represents struct to which every Object is parsable
//// Specific types of Object have different structures as Spec
//type ObjectGeneric struct {
//	ObjectHeader
//	Spec json.RawMessage `json:"spec"`
//}
//
//// JSONToGenericObjects parse JSON Array of Objects into generic objects
//func JSONToGenericObjects(jsonPayload []byte) ([]ObjectGeneric, error) {
//	var objects []ObjectGeneric
//	if err := json.Unmarshal(jsonPayload, &objects); err != nil {
//		if stxErr, ok := err.(*json.SyntaxError); ok {
//			return nil, fmt.Errorf("malformed JSON payload, syntax error: %s offset: %d", stxErr.Error(), stxErr.Offset)
//		}
//		return nil, errors.New("malformed JSON payload - pass single list of valid JSON objects")
//	}
//	return objects, nil
//}
//
//// UnsupportedKindErr returns appropriate error for missing value in field kind
//// for not empty field kind returns always that is not supported for this apiVersion
//// so have to be validated before
//func UnsupportedKindErr(o ObjectGeneric) error {
//	if strings.TrimSpace(o.Kind.String()) == "" {
//		return EnhanceError(o, errors.New("missing or empty field kind for an Object"))
//	}
//	return EnhanceError(o, fmt.Errorf("invalid Object kind: %s for apiVersion: %s", o.Kind, o.APIVersion))
//}

// StringInterpolation for arguments ("{}-my-{}-string-{}", "xd") returns string xd-my-xd-string-xd
func StringInterpolation(withPlaceholder, replacer string) string {
	return strings.ReplaceAll(withPlaceholder, StringInterpolationPlaceholder, replacer)
}
