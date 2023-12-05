package alertpolicy

import (
	"encoding/json"
	"strconv"

	"github.com/goccy/go-yaml"

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
	Conditions       []AlertCondition  `json:"conditions"`
	AlertMethods     []AlertMethodsRef `json:"alertMethods"`
}

func (spec Spec) GetAlertMethods() []AlertMethodsRef {
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

type AlertMethodsRef struct {
	Metadata AlertMethodsRefMetadata `json:"metadata"`
}

type AlertMethodsRefMetadata struct {
	Name    string `json:"name" validate:"required,objectName"`
	Project string `json:"project,omitempty" validate:"objectName"`
}

// UnmarshalYAML Using json unmarshal allows us to correctly receive float64 value in Value field
// https://nobl9.atlassian.net/browse/PC-11300
func (d *AlertCondition) UnmarshalYAMLALT(bytes []byte) error {
	jsonByte, err := yaml.YAMLToJSON(bytes)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(jsonByte, &d); err != nil {
		return err
	}

	return nil
}

// Unmarshal TODO handle correct Value parsing https://nobl9.atlassian.net/browse/PC-11300
func (d *AlertCondition) UnmarshalYAML(bytes []byte) error {
	var tempCondition struct {
		Measurement      string `json:"measurement"`
		Value            string `json:"value"`
		AlertingWindow   string `json:"alertingWindow,omitempty"`
		LastsForDuration string `json:"lastsFor,omitempty"`
		Operator         string `json:"op,omitempty"`
	}
	if err := yaml.Unmarshal(bytes, &tempCondition); err != nil {
		return err
	}
	d.Measurement = tempCondition.Measurement
	d.AlertingWindow = tempCondition.AlertingWindow
	d.LastsForDuration = tempCondition.LastsForDuration
	d.Operator = tempCondition.Operator

	if tempCondition.Measurement == v1alpha.MeasurementAverageBurnRate.String() ||
		tempCondition.Measurement == v1alpha.MeasurementBurnedBudget.String() {
		val, err := strconv.ParseFloat(tempCondition.Value, 64)
		if err != nil {
			return err
		}
		d.Value = val

		return nil
	}

	d.Value = tempCondition.Value

	return nil
}
