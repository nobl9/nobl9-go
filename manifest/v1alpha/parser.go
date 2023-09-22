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

// UseGenericObjects is a global flag instructing ParseObject to decode
// raw object into GenericObject instead of a concrete representation.
var UseGenericObjects = false

func ParseObject(data []byte, kind manifest.Kind, format manifest.ObjectFormat) (manifest.Object, error) {
	unmarshal, err := getUnmarshalFunc(data, format)
	if err != nil {
		return nil, err
	}

	var object manifest.Object
	if UseGenericObjects {
		object, err = parseGenericObject(unmarshal)
	} else {
		object, err = parseObject(kind, unmarshal)
	}
	if err != nil {
		return nil, err
	}
	return object, nil
}

func parseObject(kind manifest.Kind, unmarshal unmarshalFunc) (manifest.Object, error) {
	//exhaustive:enforce
	switch kind {
	case manifest.KindService:
		return genericParseObject[Service](unmarshal)
	case manifest.KindSLO:
		return genericParseObject[SLO](unmarshal)
	case manifest.KindProject:
		return genericParseObject[Project](unmarshal)
	case manifest.KindAgent:
		return genericParseObject[Agent](unmarshal)
	case manifest.KindDirect:
		return genericParseObject[Direct](unmarshal)
	case manifest.KindAlert:
		return genericParseObject[Alert](unmarshal)
	case manifest.KindAlertMethod:
		return genericParseObject[AlertMethod](unmarshal)
	case manifest.KindAlertPolicy:
		return genericParseObject[AlertPolicy](unmarshal)
	case manifest.KindAlertSilence:
		return genericParseObject[AlertSilence](unmarshal)
	case manifest.KindRoleBinding:
		return genericParseObject[RoleBinding](unmarshal)
	case manifest.KindDataExport:
		return genericParseObject[DataExport](unmarshal)
	case manifest.KindAnnotation:
		return genericParseObject[Annotation](unmarshal)
	case manifest.KindUserGroup:
		return genericParseObject[UserGroup](unmarshal)
	default:
		return nil, fmt.Errorf("%s is %w", kind, manifest.ErrInvalidKind)
	}
}

func parseGenericObject(unmarshal unmarshalFunc) (manifest.Object, error) {
	return genericParseObject[GenericObject](unmarshal)
}

func getUnmarshalFunc(data []byte, format manifest.ObjectFormat) (unmarshalFunc, error) {
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
	return unmarshal, nil
}

func genericParseObject[T manifest.Object](unmarshal unmarshalFunc) (T, error) {
	var object T
	if err := unmarshal(&object); err != nil {
		return object, err
	}
	return object, nil
}
