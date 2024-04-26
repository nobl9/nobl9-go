package alert

import (
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

//go:generate go run ../../../scripts/generate-object-impl.go Alert

// New creates a new Alert based on provided Metadata nad Spec.
func New(metadata Metadata, spec Spec) Alert {
	return Alert{
		APIVersion: manifest.VersionV1alpha,
		Kind:       manifest.KindAlert,
		Metadata:   metadata,
		Spec:       spec,
	}
}

// Alert represents triggered alert
type Alert struct {
	APIVersion manifest.Version `json:"apiVersion"`
	Kind       manifest.Kind    `json:"kind"`
	Metadata   Metadata         `json:"metadata"`
	Spec       Spec             `json:"spec"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
}

type Metadata struct {
	Name    string `json:"name" validate:"required,objectName"`
	Project string `json:"project,omitempty" validate:"objectName"`
}

// Spec represents content of Alert's Spec
type Spec struct {
	AlertPolicy                 ObjectMetadata `json:"alertPolicy"`
	SLO                         ObjectMetadata `json:"slo"`
	Service                     ObjectMetadata `json:"service"`
	Objective                   Objective      `json:"objective"`
	Severity                    string         `json:"severity"`
	Status                      string         `json:"status"`
	TriggeredMetricTime         string         `json:"triggeredMetricTime"`
	TriggeredClockTime          string         `json:"triggeredClockTime"`
	ResolvedClockTime           *string        `json:"resolvedClockTime,omitempty"`
	ResolvedMetricTime          *string        `json:"resolvedMetricTime,omitempty"`
	CoolDown                    string         `json:"coolDown"`
	Conditions                  []Condition    `json:"conditions"`
	CoolDownStartedAtMetricTime *string        `json:"coolDownStartedAtMetricTime,omitempty"`
	ResolutionReason            *string        `json:"resolutionReason,omitempty"`
}

type Objective struct {
	Value       float64 `json:"value"`
	Name        string  `json:"name"`
	DisplayName string  `json:"displayName"`
}

type ObjectMetadata struct {
	Name        string         `json:"name"`
	DisplayName string         `json:"displayName,omitempty"`
	Project     string         `json:"project,omitempty"`
	Labels      v1alpha.Labels `json:"labels,omitempty"`
}

type Condition struct {
	Measurement      string           `json:"measurement"`
	Value            interface{}      `json:"value"`
	AlertingWindow   string           `json:"alertingWindow,omitempty"`
	LastsForDuration string           `json:"lastsFor,omitempty"`
	Operator         string           `json:"op,omitempty"`
	Status           *ConditionStatus `json:"status,omitempty"`
}

type ConditionStatus struct {
	FirstMetMetricTime   string  `json:"firstMetMetricTime,omitempty"`
	LastMetMetricTime    *string `json:"lastMetMetricTime,omitempty"`
	LastForMetMetricTime *string `json:"lastsForMetMetricTime,omitempty"`
}

var validator = validation.New[Alert]()

func validate(_ Alert) *v1alpha.ObjectError { return nil }
