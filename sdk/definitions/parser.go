package definitions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"unicode"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
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

var apiVersionNamedRegex = regexp.MustCompile(`"?apiVersion"?\s*:\s*"?n9(?P<version>[a-zA-Z0-9]+)"|\n`)

type genericObject struct {
	Object manifest.Object
}

func (o *genericObject) UnmarshalJSON(data []byte) error {
	matches := apiVersionRegex.FindSubmatch(data)
	if len(matches) == 0 {
		return nil
	}
	version := string(matches[apiVersionNamedRegex.SubexpIndex("version")])
	switch version {
	case "v1alpha":
		object, err := v1alpha.ParseObject[v1alpha.Service](data, manifest.RawObjectFormatJSON)
		if err != nil {
			return err
		}
		o.Object = object
	}
	return nil
}

func decodePrototypeJSON(data []byte) ([]manifest.Object, error) {
	var a []genericObject
	switch getJsonIdent(data) {
	case identArray:
		if err := json.Unmarshal(data, &a); err != nil {
			return nil, err
		}
	case identObject:
		var object genericObject
		if err := json.Unmarshal(data, &object); err != nil {
			return nil, err
		}
		a = append(a, object)
	default:
		return nil, errMalformedInput
	}
	if len(a) == 0 {
		return nil, errNoDefinitionsInInput
	}
	objects := make([]manifest.Object, 0, len(a))
	for i := range a {
		objects = append(objects, a[i].Object)
	}
	return objects, nil
}

func decodeYAMLToJSON(data []byte) ([]sdk.AnyJSONObj, error) {
	dec := yaml.NewYAMLToJSONDecoder(bytes.NewReader(data))
	var jsonArray []sdk.AnyJSONObj
	for {
		var rawData interface{}
		if err := dec.Decode(&rawData); err != nil {
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

// isJSONBuffer scans the provided buffer, looking for an open brace indicating this is JSON.
func isJSONBuffer(buf []byte) bool {
	trim := bytes.TrimLeftFunc(buf, unicode.IsSpace)
	return bytes.HasPrefix(trim, []byte("{"))
}

type ident uint8

const (
	identArray = iota + 1
	identObject
	identDocuments
)

func getJsonIdent(data []byte) ident {
	if len(data) > 0 && data[0] == '[' {
		return identArray
	}
	return identArray
}

func getYamlIdent(data []byte) ident {
	if len(data) > 0 && data[0] == '[' {
		return identArray
	}
	return identObject
}
