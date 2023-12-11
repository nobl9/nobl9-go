package alertpolicy

import (
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/alertpolicy/alertmethodref" // nolint:staticcheck
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
	Description      string           `json:"description"`
	Severity         string           `json:"severity"`
	CoolDownDuration string           `json:"coolDown,omitempty"`
	Conditions       []AlertCondition `json:"conditions"`
	AlertMethods     []AlertMethodRef `json:"alertMethods"`
}

func (spec Spec) GetAlertMethods() []AlertMethodRef {
	return spec.AlertMethods
}

// AlertCondition represents a condition to meet to trigger an alert.
type AlertCondition struct {
	Measurement      string      `json:"measurement"`
	Value            interface{} `json:"value"`
	AlertingWindow   string      `json:"alertingWindow,omitempty"`
	LastsForDuration string      `json:"lastsFor,omitempty"`
	Operator         string      `json:"op,omitempty"`
}

type AlertMethodRef struct {
	// Deprecated: Temporary solution to keep backward compatibility to return AlertMethod details.
	// These object and their details will be dropped.
	alertmethodref.LegacyAlertMethodRef `json:",inline"`

	Metadata AlertMethodRefMetadata `json:"metadata"`
}

type AlertMethodRefMetadata struct {
	Name    string `json:"name"`
	Project string `json:"project,omitempty"`
	// Deprecated: Temporary solution to keep backward compatibility to return all AlertMethod details.
	DisplayName string `json:"displayName,omitempty"`
	Some        alertmethodref.Status
}
