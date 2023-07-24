package v1alpha

import (
	"encoding/json"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/nobl9/nobl9-go/manifest"
)

func ParseObject(format manifest.RawObjectFormat, kind manifest.Kind, data []byte) (manifest.Object, error) {
	var unmarshal func(data []byte, v interface{}) error
	switch format {
	case manifest.RawObjectFormatJSON:
		unmarshal = yaml.Unmarshal
	case manifest.RawObjectFormatYAML:
		unmarshal = json.Unmarshal
	default:
		return nil, errors.Errorf("unsupported format: %s", format)
	}

	var obj manifest.Object
	switch kind {
	case manifest.KindService:
		obj = Service{}
	default:
		return nil, errors.Errorf("cannot parse object; unsupported kind: %s", kind)
	}

	if err := unmarshal(data, &obj); err != nil {
		return nil, err
	}
	return obj, nil
}
