package sdk

//go:generate stringer -trimprefix=Kind -type=Kind

import (
	"strings"

	"github.com/pkg/errors"
)

var ErrInvalidKind = errors.New("invalid object Kind")

// Kind represents available objects in API to perform operations.
type Kind int

func (k Kind) ToLower() string {
	return strings.ToLower(k.String())
}

// List of available Kind in API.
const (
	KindSLO Kind = iota + 1
	KindService
	KindAgent
	KindAlertPolicy
	KindAlertSilence
	KindAlert
	KindProject
	KindAlertMethod
	KindMetricSource
	KindDirect
	KindDataExport
	KindUsageSummary
	KindRoleBinding
	KindSLOErrorBudgetStatus
	KindAnnotation
	KindGroup
)

var stringToKind = map[string]Kind{
	"slo":          KindSLO,
	"service":      KindService,
	"agent":        KindAgent,
	"alertpolicy":  KindAlertPolicy,
	"alertsilence": KindAlertSilence,
	"alert":        KindAlert,
	"project":      KindProject,
	"alertmethod":  KindAlertMethod,
	"direct":       KindDirect,
	"dataexport":   KindDataExport,
	"rolebinding":  KindRoleBinding,
	"annotation":   KindAnnotation,
	"group":        KindGroup,
}

func KindFromString(s string) (Kind, error) {
	kind, valid := stringToKind[strings.ToLower(s)]
	if !valid {
		return 0, ErrInvalidKind
	}
	return kind, nil
}
