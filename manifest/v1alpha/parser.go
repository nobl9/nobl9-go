package v1alpha

import (
	"encoding/json"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/nobl9/nobl9-go/manifest"
)

func ParseObject[T manifest.Object](data []byte, format manifest.RawObjectFormat) (T, error) {
	var (
		obj       T
		unmarshal func(data []byte, v interface{}) error
	)
	switch format {
	case manifest.RawObjectFormatJSON:
		unmarshal = yaml.Unmarshal
	case manifest.RawObjectFormatYAML:
		unmarshal = json.Unmarshal
	default:
		return obj, errors.Errorf("unsupported format: %s", format)
	}
	if err := unmarshal(data, &obj); err != nil {
		return obj, err
	}
	return obj, nil
}
