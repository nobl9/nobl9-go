package v1alpha

import "github.com/nobl9/nobl9-go/manifest"

//go:generate go run ../../scripts/generate-object-impl.go Alert

// Alert represents triggered alert
type Alert struct {
	APIVersion string        `json:"apiVersion"`
	Kind       manifest.Kind `json:"kind"`
	Metadata   AlertMetadata `json:"metadata"`
	Spec       AlertSpec     `json:"spec"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
}

type AlertMetadata struct {
	Name    string `json:"name" validate:"required,objectName"`
	Project string `json:"project,omitempty" validate:"objectName"`
}

// AlertSpec represents content of Alert's Spec
type AlertSpec struct {
	AlertPolicy         Metadata         `json:"alertPolicy"`
	SLO                 Metadata         `json:"slo"`
	Service             Metadata         `json:"service"`
	Objective           AlertObjective   `json:"objective"`
	Severity            string           `json:"severity" validate:"required,severity" example:"High"`
	Status              string           `json:"status" example:"Resolved"`
	TriggeredMetricTime string           `json:"triggeredMetricTime"`
	TriggeredClockTime  string           `json:"triggeredClockTime"`
	ResolvedClockTime   *string          `json:"resolvedClockTime,omitempty"`
	ResolvedMetricTime  *string          `json:"resolvedMetricTime,omitempty"`
	CoolDown            string           `json:"coolDown"`
	Conditions          []AlertCondition `json:"conditions"`
}

type AlertObjective struct {
	Value       float64 `json:"value" example:"100"`
	Name        string  `json:"name" validate:"omitempty"`
	DisplayName string  `json:"displayName" validate:"omitempty"`
}
