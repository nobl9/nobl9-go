package v1alpha

import (
	"encoding/json"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/manifest"
)

type unmarshalFunc func(data []byte, v interface{}) error

func ParseObject(data []byte, kind manifest.Kind, format manifest.ObjectFormat) (manifest.Object, error) {
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
	//exhaustive:enforce
	switch kind {
	case manifest.KindService:
		object, err = genericParseObject[Service](data, unmarshal)
	case manifest.KindSLO:
		object, err = genericParseObject[SLO](data, unmarshal)
	case manifest.KindProject:
		object, err = genericParseObject[Project](data, unmarshal)
	case manifest.KindAgent:
		object, err = genericParseObject[Agent](data, unmarshal)
	case manifest.KindDirect:
		object, err = genericParseObject[Direct](data, unmarshal)
	case manifest.KindAlert:
		object, err = genericParseObject[Alert](data, unmarshal)
	case manifest.KindAlertMethod:
		object, err = genericParseObject[AlertMethod](data, unmarshal)
	case manifest.KindAlertPolicy:
		object, err = genericParseObject[AlertPolicy](data, unmarshal)
	case manifest.KindAlertSilence:
		object, err = genericParseObject[AlertSilence](data, unmarshal)
	case manifest.KindRoleBinding:
		object, err = genericParseObject[RoleBinding](data, unmarshal)
	case manifest.KindDataExport:
		object, err = genericParseObject[DataExport](data, unmarshal)
	case manifest.KindAnnotation:
		object, err = genericParseObject[Annotation](data, unmarshal)
	case manifest.KindUserGroup:
		object, err = genericParseObject[UserGroup](data, unmarshal)
	default:
		return nil, fmt.Errorf("%s is %w", kind, manifest.ErrInvalidKind)
	}
	if err != nil {
		return nil, err
	}
	return object, nil
}

// postParser allows objects to implement their own logic for setting defaults and correcting parsing results.
type postParser[T manifest.Object] interface {
	postParse() T
}

func genericParseObject[T manifest.Object](data []byte, unmarshal unmarshalFunc) (T, error) {
	var object T
	if err := unmarshal(data, &object); err != nil {
		return object, err
	}
	if v, ok := any(object).(postParser[T]); ok {
		object = v.postParse()
	}
	return object, nil
}
