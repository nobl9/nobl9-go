package sdk

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/goccy/go-yaml"

	"github.com/nobl9/nobl9-go/manifest"
)

// EncodeObjects writes objects to the given [io.Writer] in the specified [manifest.ObjectFormat].
func EncodeObjects(objects []manifest.Object, out io.Writer, format manifest.ObjectFormat) error {
	return encodeObjects(objects, out, format)
}

// EncodeObject writes a single object to the given [io.Writer] in the specified [manifest.ObjectFormat].
func EncodeObject(object manifest.Object, out io.Writer, format manifest.ObjectFormat) error {
	return encodeObjects(object, out, format)
}

// PrintObjects prints objects to the given [io.Writer] in the specified [manifest.ObjectFormat].
// Deprecated: Use EncodeObjects instead.
func PrintObjects(objects []manifest.Object, out io.Writer, format manifest.ObjectFormat) error {
	return encodeObjects(objects, out, format)
}

// PrintObject prints a single object to the given [io.Writer] in the specified [manifest.ObjectFormat].
// Deprecated: Use EncodeObject instead.
func PrintObject(object manifest.Object, out io.Writer, format manifest.ObjectFormat) error {
	return encodeObjects(object, out, format)
}

func encodeObjects(objects any, out io.Writer, format manifest.ObjectFormat) error {
	switch format {
	case manifest.ObjectFormatJSON:
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(objects)
	case manifest.ObjectFormatYAML:
		enc := yaml.NewEncoder(out, yaml.CustomMarshaler(yamlNumberMarshaler))
		return enc.Encode(objects)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// yamlNumberMarshaler is a custom marshaler for [json.Number].
// It is used to avoid converting int to float64 when converting JSON to YAML for generic
// [manifest.Object] representations, like [v1alpha.GenericObject].
func yamlNumberMarshaler(number json.Number) ([]byte, error) {
	return []byte(number.String()), nil
}
