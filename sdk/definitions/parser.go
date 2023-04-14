package definitions

import (
	"bytes"
	"fmt"
	"io"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/nobl9/nobl9-go/sdk"
)

var (
	errNoDefinitionsInInput = errors.New("no definitions in input")
	errMalformedInput       = errors.New("malformed input")
)

// AnnotateObject injects to objects additional fields with values passed as map in parameter
// If objects does not contain project - default value is added.
func AnnotateObject(
	object sdk.AnyJSONObj,
	annotations map[string]string,
	project string,
	isProjectOverwritten bool,
) (sdk.AnyJSONObj, error) {
	for k, v := range annotations {
		object[k] = v
	}
	m, ok := object["metadata"].(map[string]interface{})

	switch {
	case !ok:
		return nil, fmt.Errorf("cannot retrieve metadata section")
	// If project in YAML is empty - fill project
	case m["project"] == nil:
		m["project"] = project
		object["metadata"] = m
	// If value in YAML is not empty but is different from --project flag value.
	case m["project"] != nil && m["project"] != project && isProjectOverwritten:
		return nil, fmt.Errorf(
			"the project from the provided object %s does not match "+
				"the project %s. You must pass '--project=%s' to perform this operation",
			m["project"],
			project,
			m["project"])
	}
	return object, nil
}

// processRawDefinitionsToJSONArray function converts raw definitions to JSON array.
func processRawDefinitionsToJSONArray(a Annotations, rds rawDefinitions) ([]sdk.AnyJSONObj, error) {
	jsonArray := make([]sdk.AnyJSONObj, 0, len(rds))
	for _, rd := range rds {
		defsInJSON, err := decodeYAMLToJSON(rd.Definition)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", rd.ResolvedSource, err)
		}
		for _, defInJSON := range defsInJSON {
			annotations := map[string]string{
				"manifestSrc":  rd.ResolvedSource,
				"organization": a.Organization,
			}
			// FIXME: sdk.Annotate adds Project to all objects, no matter the Kind, it should not do that
			// for Project or RoleBinding (project agnostic objects), it should be fixed in the sdk.
			annotated, err := AnnotateObject(defInJSON, annotations, a.Project, a.ProjectOverwritesCfgFile)
			if err != nil {
				return nil, err
			}
			jsonArray = append(jsonArray, annotated)
		}
	}
	return jsonArray, nil
}

func decodeYAMLToJSON(content []byte) ([]sdk.AnyJSONObj, error) {
	s := yaml.NewYAMLToJSONDecoder(bytes.NewReader(content))
	var jsonArray []sdk.AnyJSONObj
	for {
		var rawData interface{}
		if err := s.Decode(&rawData); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		switch obj := rawData.(type) {
		// Single N9 object.
		case map[string]interface{}:
			if len(obj) > 0 {
				jsonArray = append(jsonArray, obj)
			}
		// Multiple N9 objects.
		case []interface{}:
			// Try parsing each to a single N9App object.
			for _, def := range obj {
				switch o := def.(type) {
				case sdk.AnyJSONObj:
					if len(o) > 0 {
						jsonArray = append(jsonArray, o)
					}
				default:
					return nil, errMalformedInput
				}
			}
		// Empty object.
		case nil:
		// Something unexpected.
		default:
			return nil, errMalformedInput
		}
	}
	if len(jsonArray) == 0 {
		return nil, errNoDefinitionsInInput
	}
	return jsonArray, nil
}
