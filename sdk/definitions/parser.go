package definitions

import (
	"bytes"
	"fmt"
	"io"

	"github.com/nobl9/nobl9-go/sdk"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/yaml"
)

var (
	errNoDefinitionsInInput = errors.New("no definitions in input")
	errMalformedInput       = errors.New("malformed input")
)

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
			annotated, err := sdk.Annotate(defInJSON, annotations, a.Project, a.ProjectOverwritesCfgFile)
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
