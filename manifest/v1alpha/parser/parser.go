package parser

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/dataexport"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/rolebinding"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/usergroup"
)

type unmarshalFunc func(v interface{}) error

// UseGenericObjects is a global flag instructing ParseObject to decode
// raw object into GenericObject instead of a concrete representation.
var UseGenericObjects = false

// UseStrictDecodingMode is a global flag instructing ParseObject to
// disallow unknown fields from decoded object definitions.
var UseStrictDecodingMode = false

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
		return genericParseObject[service.Service](unmarshal)
	case manifest.KindSLO:
		return genericParseObject[slo.SLO](unmarshal)
	case manifest.KindProject:
		return genericParseObject[project.Project](unmarshal)
	case manifest.KindAgent:
		return genericParseObject[v1alpha.Agent](unmarshal)
	case manifest.KindDirect:
		return genericParseObject[v1alpha.Direct](unmarshal)
	case manifest.KindAlert:
		return genericParseObject[v1alpha.Alert](unmarshal)
	case manifest.KindAlertMethod:
		return genericParseObject[v1alpha.AlertMethod](unmarshal)
	case manifest.KindAlertPolicy:
		return genericParseObject[v1alpha.AlertPolicy](unmarshal)
	case manifest.KindAlertSilence:
		return genericParseObject[v1alpha.AlertSilence](unmarshal)
	case manifest.KindRoleBinding:
		return genericParseObject[rolebinding.RoleBinding](unmarshal)
	case manifest.KindDataExport:
		return genericParseObject[dataexport.DataExport](unmarshal)
	case manifest.KindAnnotation:
		return genericParseObject[v1alpha.Annotation](unmarshal)
	case manifest.KindUserGroup:
		return genericParseObject[usergroup.UserGroup](unmarshal)
	default:
		return nil, fmt.Errorf("%s is %w", kind, manifest.ErrInvalidKind)
	}
}

func parseGenericObject(unmarshal unmarshalFunc) (manifest.Object, error) {
	return genericParseObject[v1alpha.GenericObject](unmarshal)
}

func getUnmarshalFunc(data []byte, format manifest.ObjectFormat) (unmarshalFunc, error) {
	var unmarshal unmarshalFunc
	switch format {
	case manifest.ObjectFormatJSON:
		unmarshal = func(v interface{}) error {
			dec := json.NewDecoder(bytes.NewReader(data))
			if UseStrictDecodingMode {
				dec.DisallowUnknownFields()
			}
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
		var opts []yaml.DecodeOption
		if UseStrictDecodingMode {
			opts = append(opts, yaml.Strict())
		}
		unmarshal = func(v interface{}) error {
			return yaml.UnmarshalWithOptions(data, v, opts...)
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
