package v1alpha

import (
	"encoding/json"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/manifest"
)

type unmarshalFunc func(data []byte, v interface{}) error

func ParseObject(data []byte, kind manifest.Kind, format manifest.RawObjectFormat) (manifest.Object, error) {
	var unmarshal unmarshalFunc
	switch format {
	case manifest.RawObjectFormatJSON:
		unmarshal = json.Unmarshal
	case manifest.RawObjectFormatYAML:
		unmarshal = yaml.Unmarshal
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
	case manifest.KindSLO:
		object, err = genericParseObject[SLO](data, unmarshal)
	default:
		return nil, fmt.Errorf("%s is %w", kind, manifest.ErrInvalidKind)
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
