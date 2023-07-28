package definitions

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"unicode"

	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"

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

// genericObject is a container for manifest.Object which helps in decoding process.
type genericObject struct {
	Object manifest.Object
}

// UnmarshalJSON implements json.Unmarshaler.
func (o *genericObject) UnmarshalJSON(data []byte) error {
	return o.unmarshalGeneric(data, manifest.RawObjectFormatJSON)
}

// UnmarshalYAML implements yaml.BytesUnmarshaler.
func (o *genericObject) UnmarshalYAML(data []byte) error {
	return o.unmarshalGeneric(data, manifest.RawObjectFormatYAML)
}

// unmarshalGeneric decodes a single raw manifest.Object representation into respective manifest.RawObjectFormat.
// It uses an intermediate decoding step to extract manifest.Version and manifest.Kind from the object.
// Decoding is then delegated to the parser for specific manifest.Version.
func (o *genericObject) unmarshalGeneric(data []byte, format manifest.RawObjectFormat) error {
	var object struct {
		ApiVersion manifest.Version `json:"apiVersion" yaml:"apiVersion"`
		Kind       manifest.Kind    `json:"kind" yaml:"kind"`
	}
	var unmarshal func(data []byte, v interface{}) error
	switch format {
	case manifest.RawObjectFormatJSON:
		unmarshal = json.Unmarshal
	case manifest.RawObjectFormatYAML:
		unmarshal = yaml.Unmarshal
	}
	if err := unmarshal(data, &object); err != nil {
		return err
	}
	switch object.ApiVersion {
	case manifest.VersionV1alpha:
		parsed, err := v1alpha.ParseObject(data, object.Kind, format)
		if err != nil {
			return err
		}
		o.Object = parsed
	default:
		return manifest.ErrInvalidVersion
	}
	return nil
}

func decodePrototype(data []byte) ([]manifest.Object, error) {
	if isJSONBuffer(data) {
		return decodePrototypeJSON(data)
	}
	return decodePrototypeYAML(data)
}

func decodePrototypeJSON(data []byte) ([]manifest.Object, error) {
	var res []genericObject
	switch getJsonIdent(data) {
	case identArray:
		if err := json.Unmarshal(data, &res); err != nil {
			return nil, err
		}
	case identObject:
		var object genericObject
		if err := json.Unmarshal(data, &object); err != nil {
			return nil, err
		}
		res = append(res, object)
	}
	if len(res) == 0 {
		return nil, errNoDefinitionsInInput
	}
	objects := make([]manifest.Object, 0, len(res))
	for i := range res {
		objects = append(objects, res[i].Object)
	}
	return objects, nil
}

func decodePrototypeYAML(data []byte) ([]manifest.Object, error) {
	scanner := bufio.NewScanner(bytes.NewBuffer(data))
	scanner.Split(splitYAMLDocument)
	var res []genericObject
	for scanner.Scan() {
		doc := scanner.Bytes()
		switch getYamlIdent(doc) {
		case identArray:
			var a []genericObject
			if err := yaml.Unmarshal(doc, &a); err != nil {
				return nil, err
			}
			res = append(res, a...)
		case identObject:
			var object genericObject
			if err := yaml.Unmarshal(doc, &object); err != nil {
				return nil, err
			}
			res = append(res, object)
		}
	}
	if len(res) == 0 {
		return nil, errNoDefinitionsInInput
	}
	objects := make([]manifest.Object, 0, len(res))
	for i := range res {
		objects = append(objects, res[i].Object)
	}
	return objects, nil
}

func decodeYAMLToJSON(data []byte) ([]sdk.AnyJSONObj, error) {
	return nil, nil
	//dec := yaml.NewYAMLToJSONDecoder(bytes.NewReader(data))
	//var jsonArray []sdk.AnyJSONObj
	//for {
	//	var rawData interface{}
	//	if err := dec.Decode(&rawData); err != nil {
	//		if err == io.EOF {
	//			break
	//		}
	//		return nil, err
	//	}
	//	switch obj := rawData.(type) {
	//	case map[string]interface{}:
	//		if len(obj) > 0 {
	//			jsonArray = append(jsonArray, obj)
	//		}
	//	case []interface{}:
	//		for _, def := range obj {
	//			switch o := def.(type) {
	//			case sdk.AnyJSONObj:
	//				if len(o) > 0 {
	//					jsonArray = append(jsonArray, o)
	//				}
	//			default:
	//				return nil, errMalformedInput
	//			}
	//		}
	//	case nil:
	//	default:
	//		return nil, errMalformedInput
	//	}
	//}
	//if len(jsonArray) == 0 {
	//	return nil, errNoDefinitionsInInput
	//}
	//return jsonArray, nil
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
)

func getJsonIdent(data []byte) ident {
	if len(data) > 0 && data[0] == '[' {
		return identArray
	}
	return identObject
}

var yamlArrayIdentRegex = regexp.MustCompile(`(?m)^\s*[\[-]\s`)

func getYamlIdent(data []byte) ident {
	if len(data) > 0 && (yamlArrayIdentRegex.Match(data) || data[0] == '[') {
		return identArray
	}
	return identObject
}

// yamlDocSep includes a prefixed newline character as we do now want to split on the first
// document separator located at the beginning of the file.
const yamlDocSep = "\n---"

// splitYAMLDocument is a bufio.SplitFunc for splitting YAML streams into individual documents.
func splitYAMLDocument(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	// We have a potential document terminator.
	if i := bytes.Index(data, []byte(yamlDocSep)); i >= 0 {
		sep := len(yamlDocSep)
		i += sep
		after := data[i:]
		if len(after) == 0 {
			if atEOF {
				return len(data), data[:len(data)-sep], nil
			}
			return 0, nil, nil
		}
		if j := bytes.IndexByte(after, '\n'); j >= 0 {
			return i + j + 1, data[0 : i-sep], nil
		}
		return 0, nil, nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
