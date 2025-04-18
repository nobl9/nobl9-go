package parser

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/agent"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/alert"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/alertmethod"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/alertpolicy"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/alertsilence"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/annotation"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/budgetadjustment"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/dataexport"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/direct"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/report"
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

// UseJSONNumber is a global flag instructing ParseObject to decode
// JSON numbers into [json.Number] instead of float64 as is [encoding/json]
// pkg default.
var UseJSONNumber = false

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
		return genericParseObject[agent.Agent](unmarshal)
	case manifest.KindDirect:
		return genericParseObject[direct.Direct](unmarshal)
	case manifest.KindAlert:
		return genericParseObject[alert.Alert](unmarshal)
	case manifest.KindAlertMethod:
		return genericParseObject[alertmethod.AlertMethod](unmarshal)
	case manifest.KindAlertPolicy:
		return genericParseObject[alertpolicy.AlertPolicy](unmarshal)
	case manifest.KindAlertSilence:
		return genericParseObject[alertsilence.AlertSilence](unmarshal)
	case manifest.KindRoleBinding:
		return genericParseObject[rolebinding.RoleBinding](unmarshal)
	case manifest.KindDataExport:
		return genericParseObject[dataexport.DataExport](unmarshal)
	case manifest.KindAnnotation:
		return genericParseObject[annotation.Annotation](unmarshal)
	case manifest.KindUserGroup:
		return genericParseObject[usergroup.UserGroup](unmarshal)
	case manifest.KindBudgetAdjustment:
		return genericParseObject[budgetadjustment.BudgetAdjustment](unmarshal)
	case manifest.KindReport:
		return genericParseObject[report.Report](unmarshal)
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
			if UseJSONNumber {
				dec.UseNumber()
			}
			if UseStrictDecodingMode {
				dec.DisallowUnknownFields()
			}
			return dec.Decode(v)
		}
	case manifest.ObjectFormatYAML:
		unmarshal = func(v interface{}) error {
			var opts []yaml.DecodeOption
			if UseStrictDecodingMode {
				opts = append(opts, yaml.Strict())
			}
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
