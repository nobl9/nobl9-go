package v1alpha

import (
	"encoding/json"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/nobl9/nobl9-go/manifest"
)

type unmarshalFunc func(data []byte, v interface{}) error

func ParseObject(data []byte, kind manifest.Kind, format manifest.RawObjectFormat) (manifest.Object, error) {
	var unmarshal unmarshalFunc
	switch format {
	case manifest.RawObjectFormatJSON:
		unmarshal = yaml.Unmarshal
	case manifest.RawObjectFormatYAML:
		unmarshal = json.Unmarshal
	default:
		return nil, errors.Errorf("unsupported format: %s", format)
	}

	var (
		object manifest.Object
		err    error
	)
	switch kind {
	case manifest.KindService:
		object, err = genericParseObject[Service](data, unmarshal)
	default:
		return nil, manifest.ErrInvalidKind
	}
	if err != nil {
		return nil, err
	}
	return object, nil
}

func genericParseObject[T manifest.Object](data []byte, unmarshal unmarshalFunc) (T, error) {
	var object T
	if err := unmarshal(data, &object); err != nil {
		return object, err
	}
	return object, nil
}
