package definitions

import (
	"fmt"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/sdk"
)

// MetadataAnnotations defines a set of annotations appended to applied objects definitions.
// These annotations are only set if the resource definition does not contain them already.
type MetadataAnnotations struct {
	Organization string
	Project      string
	// When using Read or ReadSources this field is set by these functions,
	// anything here will provided here will be overwritten.
	ManifestSource string
}

// AnnotateObjectPrototype annotates an sdk.Kind with additional metadata.
// If objects does not contain project - default value is added.
// If value 'metadata.project' in the definition is different from
// the Project provided in MetadataAnnotations, an error is returned.
func (ma MetadataAnnotations) AnnotateObjectPrototype(object manifest.Object) (sdk.AnyJSONObj, error) {
	switch object.GetAPIVersion() {
	case "n9/v1alpha":
		
	}
	if object["organization"] == nil && ma.Organization != "" {
		object["organization"] = ma.Organization
	}
	if object["manifestSrc"] == nil && ma.ManifestSource != "" {
		object["manifestSrc"] = ma.ManifestSource
	}
	meta, ok := object["metadata"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot retrieve metadata section")
	}
	kindStr, ok := object["kind"].(string)
	if !ok {
		return nil, fmt.Errorf("cannot retrieve object kind")
	}
	kind, err := manifest.ParseKind(kindStr)
	if err != nil {
		return nil, err
	}
	switch kind {
	case manifest.KindProject, manifest.KindRoleBinding, manifest.KindUserGroup:
		// Do not append the project name.
	default:
		if meta["project"] == nil && ma.Project != "" {
			meta["project"] = ma.Project
		}
	}
	return object, nil
}

// AnnotateObject annotates an sdk.Kind with additional metadata.
// If objects does not contain project - default value is added.
// If value 'metadata.project' in the definition is different from
// the Project provided in MetadataAnnotations, an error is returned.
func (ma MetadataAnnotations) AnnotateObject(object sdk.AnyJSONObj) (sdk.AnyJSONObj, error) {
	if object["organization"] == nil && ma.Organization != "" {
		object["organization"] = ma.Organization
	}
	if object["manifestSrc"] == nil && ma.ManifestSource != "" {
		object["manifestSrc"] = ma.ManifestSource
	}
	meta, ok := object["metadata"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot retrieve metadata section")
	}
	kindStr, ok := object["kind"].(string)
	if !ok {
		return nil, fmt.Errorf("cannot retrieve object kind")
	}
	kind, err := manifest.ParseKind(kindStr)
	if err != nil {
		return nil, err
	}
	switch kind {
	case manifest.KindProject, manifest.KindRoleBinding, manifest.KindUserGroup:
		// Do not append the project name.
	default:
		if meta["project"] == nil && ma.Project != "" {
			meta["project"] = ma.Project
		}
	}
	return object, nil
}
