package alertpolicy

import (
	"encoding/json"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

//go:generate go run ../../../internal/cmd/objectimpl AlertPolicy

func New(metadata Metadata, spec Spec) AlertPolicy {
	return AlertPolicy{
		APIVersion: manifest.VersionV1alpha,
		Kind:       manifest.KindAlertPolicy,
		Metadata:   metadata,
		Spec:       spec,
	}
}

// AlertPolicy represents a set of conditions that can trigger an alert.
type AlertPolicy struct {
	APIVersion manifest.Version `json:"apiVersion"`
	Kind       manifest.Kind    `json:"kind"`
	Metadata   Metadata         `json:"metadata"`
	Spec       Spec             `json:"spec"`

	Organization   string `json:"organization,omitempty" nobl9:"computed"`
	ManifestSource string `json:"manifestSrc,omitempty" nobl9:"computed"`
}

type Metadata struct {
	Name        string                      `json:"name"`
	DisplayName string                      `json:"displayName,omitempty"`
	Project     string                      `json:"project,omitempty"`
	Labels      v1alpha.Labels              `json:"labels,omitempty"`
	Annotations v1alpha.MetadataAnnotations `json:"annotations,omitempty"`
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
	Metadata AlertMethodRefMetadata `json:"metadata"`

	// Deprecated: Temporary solution to keep backward compatibility to return AlertMethod details.
	// These object and their details will be dropped.
	legacyAlertMethodRef interface{}
}

type AlertMethodRefMetadata struct {
	Name    string `json:"name"`
	Project string `json:"project,omitempty"`
}

// EmbedAlertMethodRef sets AlertMethodRef to an arbitrary value.
// Deprecated: Temporary solution to keep backward compatibility to return AlertMethod details.
// These objects and their details will be dropped.
func (a *AlertMethodRef) EmbedAlertMethodRef(ref interface{}) {
	a.legacyAlertMethodRef = ref
}

func (a *AlertMethodRef) MarshalJSON() ([]byte, error) {
	if a.legacyAlertMethodRef != nil {
		return json.Marshal(a.legacyAlertMethodRef)
	}
	return json.Marshal(struct {
		Metadata AlertMethodRefMetadata `json:"metadata"`
	}{
		Metadata: a.Metadata,
	})
}
