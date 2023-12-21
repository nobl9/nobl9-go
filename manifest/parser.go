package manifest

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"sync"

	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"
)

var ErrNoDefinitionsFound = errors.New("no definitions in input")

// DecodeObjects reads objects from the provided bytes slice.
// It detects if the input is in JSON (manifest.RawObjectFormatJSON) or YAML (manifest.RawObjectFormatYAML format.
func DecodeObjects(data []byte) ([]Object, error) {
	if IsJSONBuffer(data) {
		return decodeJSON(data)
	}
	return decodeYAML(data)
}

// DecodeObject returns a single, concrete object implementing manifest.Object.
// It expects exactly one object in the decoded byte slice.
func DecodeObject[T Object](data []byte) (object T, err error) {
	objects, err := DecodeObjects(data)
	if err != nil {
		return object, err
	}
	if len(objects) != 1 {
		return object, fmt.Errorf("unexpected number of objects: %d, expected exactly one", len(objects))
	}
	var isOfType bool
	object, isOfType = objects[0].(T)
	if !isOfType {
		return object, fmt.Errorf("object of type %T is not of type %T", objects[0], *new(T))
	}
	return object, nil
}

// processRawDefinitions function converts raw definitions to a slice of manifest.Object.
func processRawDefinitions(rds rawDefinitions) ([]Object, error) {
	result := make([]Object, 0, len(rds))
	for _, rd := range rds {
		objects, err := DecodeObjects(rd.Definition)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", rd.ResolvedSource, err)
		}
		for _, obj := range objects {
			if obj == nil {
				continue
			}
			result = append(result, annotateWithManifestSource(obj, rd.ResolvedSource))
		}
	}
	return result, nil
}

// annotateWithManifestSource annotates manifest.Object with the manifest definition source.
func annotateWithManifestSource(object Object, source string) Object {
	if object.GetManifestSource() == "" && source != "" {
		object = object.SetManifestSource(source)
	}
	return object
}

func decodeJSON(data []byte) ([]Object, error) {
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
		return nil, ErrNoDefinitionsFound
	}
	objects := make([]Object, 0, len(res))
	for i := range res {
		objects = append(objects, res[i].Object)
	}
	return objects, nil
}

func decodeYAML(data []byte) ([]Object, error) {
	scanner := bufio.NewScanner(bytes.NewBuffer(data))
	// Documents can have any size, at most it will be the whole data.
	// This means sometimes we might exceed the limit imposed by bufio.Scanner.
	maxTokenSize := len(data) + 1
	scanner.Buffer(make([]byte, 0, len(data)), maxTokenSize)
	scanner.Split(splitYAMLDocument)
	var res []genericObject
	for scanner.Scan() {
		doc := scanner.Bytes()
		if len(bytes.TrimSpace(doc)) == 0 {
			continue
		}
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
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, ErrNoDefinitionsFound
	}
	objects := make([]Object, 0, len(res))
	for i := range res {
		objects = append(objects, res[i].Object)
	}
	return objects, nil
}

// genericObject is a container for manifest.Object which helps in decoding process.
type genericObject struct {
	Object Object
}

// UnmarshalJSON implements json.Unmarshaler.
func (o *genericObject) UnmarshalJSON(data []byte) error {
	return o.unmarshalGeneric(data, ObjectFormatJSON)
}

// UnmarshalYAML implements yaml.BytesUnmarshaler.
func (o *genericObject) UnmarshalYAML(data []byte) error {
	return o.unmarshalGeneric(data, ObjectFormatYAML)
}

// unmarshalGeneric decodes a single raw manifest.Object representation into respective manifest.ObjectFormat.
// It uses an intermediate decoding step to extract manifest.Version and manifest.Kind from the object.
// Decoding is then delegated to the parser for specific manifest.Version.
func (o *genericObject) unmarshalGeneric(data []byte, format ObjectFormat) error {
	var object struct {
		ApiVersion Version `json:"apiVersion" yaml:"apiVersion"`
		Kind       Kind    `json:"kind" yaml:"kind"`
	}
	var unmarshal func(data []byte, v interface{}) error
	//exhaustive: enforce
	switch format {
	case ObjectFormatJSON:
		unmarshal = json.Unmarshal
	case ObjectFormatYAML:
		unmarshal = yaml.Unmarshal
	}
	if err := unmarshal(data, &object); err != nil {
		return err
	}
	parser, found := parsers[object.ApiVersion]
	if !found {
		return ErrInvalidVersion
	}
	var err error
	o.Object, err = parser(data, object.Kind, format)
	if err != nil {
		return err
	}
	return nil
}

var jsonBufferRegex = regexp.MustCompile(`^\s*\[?\s*{`)

// IsJSONBuffer scans the provided buffer, looking for an open brace indicating this is JSON.
// While a simple list like ["a", "b", "c"] is still a valid JSON,
// it does not really concern us when processing complex objects.
func IsJSONBuffer(buf []byte) bool {
	return jsonBufferRegex.Match(buf)
}

type ident uint8

const (
	identArray = iota + 1
	identObject
)

var jsonArrayIdentRegex = regexp.MustCompile(`^\s*\[`)

func getJsonIdent(data []byte) ident {
	if jsonArrayIdentRegex.Match(data) {
		return identArray
	}
	return identObject
}

var yamlArrayIdentRegex = regexp.MustCompile(`(?m)^- `)

func getYamlIdent(data []byte) ident {
	// If we encounter square brackets array syntax, well... let's still recognize it's a valid array
	// but obviously it cannot be a complex object as this syntax won't allow it.
	if yamlArrayIdentRegex.Match(data) || jsonArrayIdentRegex.Match(data) {
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

var (
	mu      sync.Mutex
	parsers = make(map[Version]parserFunc)
)

type parserFunc = func(data []byte, kind Kind, format ObjectFormat) (Object, error)

// RegisterParser registers a new parserFunc for a given manifest.Version.
// It's intended exclusively for internal usage.
func RegisterParser(version Version, parser parserFunc) {
	mu.Lock()
	defer mu.Unlock()
	if _, found := parsers[version]; found {
		panic(fmt.Sprintf("parser for version %s already registered", version))
	}
	parsers[version] = parser
}
