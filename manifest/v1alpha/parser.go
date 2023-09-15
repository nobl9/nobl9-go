package v1alpha

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/manifest"
)

type unmarshalFunc func(v interface{}) error

func ParseObject(data []byte, kind manifest.Kind, format manifest.ObjectFormat) (manifest.Object, error) {
	var unmarshal unmarshalFunc
	switch format {
	case manifest.ObjectFormatJSON:
		unmarshal = func(v interface{}) error {
			dec := json.NewDecoder(bytes.NewReader(data))
			dec.DisallowUnknownFields()
			return dec.Decode(v)
		}
	case manifest.ObjectFormatYAML:
		// Workaround for https://github.com/goccy/go-yaml/issues/313.
		// If the library changes its interpretation of empty pointer fields,
		// we should switch to native yaml.Unmarshal instead.
		var err error
		data, err = yaml.YAMLToJSON(data)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert YAML to JSON")
		}
		unmarshal = func(v interface{}) error {
			return yaml.UnmarshalWithOptions(data, v, yaml.Strict())
		}
	default:
		return nil, errors.Errorf("unsupported format: %s", format)
	}

	var (
		object manifest.Object
		err    error
	)
	//exhaustive:enforce
	switch kind {
	case manifest.KindService:
		object, err = genericParseObject[Service](unmarshal)
	case manifest.KindSLO:
		object, err = genericParseObject[SLO](unmarshal)
	case manifest.KindProject:
		object, err = genericParseObject[Project](unmarshal)
	case manifest.KindAgent:
		object, err = genericParseObject[Agent](unmarshal)
	case manifest.KindDirect:
		object, err = genericParseObject[Direct](unmarshal)
	case manifest.KindAlert:
		object, err = genericParseObject[Alert](unmarshal)
	case manifest.KindAlertMethod:
		object, err = genericParseObject[AlertMethod](unmarshal)
	case manifest.KindAlertPolicy:
		object, err = genericParseObject[AlertPolicy](unmarshal)
	case manifest.KindAlertSilence:
		object, err = genericParseObject[AlertSilence](unmarshal)
	case manifest.KindRoleBinding:
		object, err = genericParseObject[RoleBinding](unmarshal)
	case manifest.KindDataExport:
		object, err = genericParseObject[DataExport](unmarshal)
	case manifest.KindAnnotation:
		object, err = genericParseObject[Annotation](unmarshal)
	case manifest.KindUserGroup:
		object, err = genericParseObject[UserGroup](unmarshal)
	default:
		return nil, fmt.Errorf("%s is %w", kind, manifest.ErrInvalidKind)
	}
	if err != nil {
		return nil, err
	}
	return object, nil
}

func genericParseObject[T manifest.Object](unmarshal unmarshalFunc) (T, error) {
	var object T
	if err := unmarshal(&object); err != nil {
		return object, err
	}
	return object, nil
}
