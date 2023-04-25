package definitions

import (
	"fmt"

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

// AnnotateObject annotates an sdk.Object with additional metadata.
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
	if meta["project"] == nil && ma.Project != "" {
		meta["project"] = ma.Project
	}
	return object, nil
}
