package alertpolicy

import (
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

//go:generate go run ../../../scripts/generate-object-impl.go AlertPolicy

func New(metadata Metadata, spec Spec) AlertPolicy {
	return AlertPolicy{
		APIVersion: v1alpha.APIVersion,
		Kind:       manifest.KindAlertPolicy,
		Metadata:   metadata,
		Spec:       spec,
	}
}

// AlertPolicy represents a set of conditions that can trigger an alert.
type AlertPolicy struct {
	APIVersion string        `json:"apiVersion"`
	Kind       manifest.Kind `json:"kind"`
	Metadata   Metadata      `json:"metadata"`
	Spec       Spec          `json:"spec"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
}

type Metadata struct {
	Name        string         `json:"name"`
	DisplayName string         `json:"displayName,omitempty"`
	Project     string         `json:"project,omitempty"`
	Labels      v1alpha.Labels `json:"labels,omitempty"`
}

// Spec represents content of AlertPolicy's Spec.
type Spec struct {
	Description      string            `json:"description"`
	Severity         string            `json:"severity"`
	CoolDownDuration string            `json:"coolDown,omitempty"`
	Conditions       []AlertCondition  `json:"conditions" validate:"required,min=1,dive"`
	AlertMethods     []AlertMethodsRef `json:"alertMethods"`
}

func (spec Spec) GetAlertMethods() []AlertMethodsRef {
	return spec.AlertMethods
}

// AlertCondition represents a condition to meet to trigger an alert.
type AlertCondition struct {
	Measurement      string      `json:"measurement" validate:"required,alertPolicyMeasurement" example:"BurnedBudget"`
	Value            interface{} `json:"value" validate:"required" example:"0.97"`
	AlertingWindow   string      `json:"alertingWindow,omitempty" validate:"omitempty,validDuration,nonNegativeDuration,durationMinutePrecision" example:"30m"` //nolint:lll
	LastsForDuration string      `json:"lastsFor,omitempty" validate:"omitempty,validDuration,nonNegativeDuration" example:"15m"`                               //nolint:lll
	Operator         string      `json:"op,omitempty" validate:"omitempty,operator" example:"lt"`
}

type AlertMethodsRef struct {
	Metadata AlertMethodsRefMetadata `json:"metadata"`
}

type AlertMethodsRefMetadata struct {
	Name    string `json:"name" validate:"required,objectName"`
	Project string `json:"project,omitempty" validate:"objectName"`
}
