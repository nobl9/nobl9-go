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

// processRawDefinitionsToJSONArray function converts raw definitions to JSON array.
func processRawDefinitionsToJSONArray(a MetadataAnnotations, rds rawDefinitions) ([]sdk.AnyJSONObj, error) {
	jsonArray := make([]sdk.AnyJSONObj, 0, len(rds))
	for _, rd := range rds {
		a.ManifestSource = rd.ResolvedSource
		defsInJSON, err := decodeYAMLToJSON(rd.Definition)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", rd.ResolvedSource, err)
		}
		for _, defInJSON := range defsInJSON {
			annotated, err := a.AnnotateObject(defInJSON)
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
		case map[string]interface{}:
			if len(obj) > 0 {
				jsonArray = append(jsonArray, obj)
			}
		case []interface{}:
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
		case nil:
		default:
			return nil, errMalformedInput
		}
	}
	if len(jsonArray) == 0 {
		return nil, errNoDefinitionsInInput
	}
	return jsonArray, nil
}
