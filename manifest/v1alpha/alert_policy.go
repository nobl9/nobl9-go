package v1alpha

import "github.com/nobl9/nobl9-go/manifest"

//go:generate go run ../../scripts/generate-object-impl.go AlertPolicy

// AlertPolicy represents a set of conditions that can trigger an alert.
type AlertPolicy struct {
	APIVersion string              `json:"apiVersion"`
	Kind       manifest.Kind       `json:"kind"`
	Metadata   AlertPolicyMetadata `json:"metadata"`
	Spec       AlertPolicySpec     `json:"spec"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
}

type AlertPolicyMetadata struct {
	Name        string `json:"name" validate:"required,objectName"`
	DisplayName string `json:"displayName,omitempty" validate:"omitempty,min=0,max=63"`
	Project     string `json:"project,omitempty" validate:"objectName"`
	Labels      Labels `json:"labels,omitempty" validate:"omitempty,labels"`
}

// AlertPolicySpec represents content of AlertPolicy's Spec.
type AlertPolicySpec struct {
	Description      string              `json:"description" validate:"description" example:"Message budget is at risk"`
	Severity         string              `json:"severity" validate:"required,severity" example:"High"`
	CoolDownDuration string              `json:"coolDown,omitempty" validate:"omitempty,validDuration,nonNegativeDuration,durationAtLeast=5m" example:"5m"` //nolint:lll
	Conditions       []AlertCondition    `json:"conditions" validate:"required,min=1,dive"`
	AlertMethods     []PublicAlertMethod `json:"alertMethods"`
}

func (spec AlertPolicySpec) GetAlertMethods() []PublicAlertMethod {
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

// AlertPolicyWithSLOs struct which mapped one to one with kind: alert policy and slo yaml definition
type AlertPolicyWithSLOs struct {
	AlertPolicy AlertPolicy `json:"alertPolicy"`
	SLOs        []SLO       `json:"slos"`
}
