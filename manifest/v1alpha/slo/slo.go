package slo

import (
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

//go:generate go run ../../../internal/cmd/objectimpl SLO

// New creates a new SLO based on provided Metadata nad Spec.
func New(metadata Metadata, spec Spec) SLO {
	return SLO{
		APIVersion: manifest.VersionV1alpha,
		Kind:       manifest.KindSLO,
		Metadata:   metadata,
		Spec:       spec,
	}
}

// SLO struct which mapped one to one with kind: slo yaml definition, external usage.
type SLO struct {
	APIVersion manifest.Version `json:"apiVersion"`
	Kind       manifest.Kind    `json:"kind"`
	Metadata   Metadata         `json:"metadata"`
	Spec       Spec             `json:"spec"`
	Status     *Status          `json:"status,omitempty"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
}

// Metadata provides identity information for SLO.
type Metadata struct {
	Name        string                      `json:"name"`
	DisplayName string                      `json:"displayName,omitempty"`
	Project     string                      `json:"project,omitempty"`
	Labels      v1alpha.Labels              `json:"labels,omitempty"`
	Annotations v1alpha.MetadataAnnotations `json:"annotations,omitempty"`
}

// Spec holds detailed information specific to SLO.
type Spec struct {
	Description     string       `json:"description"`
	Indicator       *Indicator   `json:"indicator,omitempty"`
	BudgetingMethod string       `json:"budgetingMethod"`
	Objectives      []Objective  `json:"objectives"`
	Service         string       `json:"service"`
	TimeWindows     []TimeWindow `json:"timeWindows"`
	AlertPolicies   []string     `json:"alertPolicies"`
	Attachments     []Attachment `json:"attachments,omitempty"`
	// CreatedAt is the date of the [SLO] creation in RFC3339 format.
	// Read-only field.
	CreatedAt string `json:"createdAt,omitempty"`
	// CreatedBy is the id of the user who first created the SLO.
	// Read-only field.
	CreatedBy string `json:"createdBy,omitempty"`
	// Deprecated: this implementation of Composite will be removed and replaced with SLO.Spec.Objectives.Composite.
	Composite     *Composite     `json:"composite,omitempty"`
	AnomalyConfig *AnomalyConfig `json:"anomalyConfig,omitempty"`
}

// Attachment represents user defined URL attached to SLO.
type Attachment struct {
	URL         string  `json:"url"`
	DisplayName *string `json:"displayName,omitempty"`
}

// ObjectiveBase base structure representing an objective.
type ObjectiveBase struct {
	DisplayName string   `json:"displayName"`
	Value       *float64 `json:"value"`
	Name        string   `json:"name"`
	NameChanged bool     `json:"-"`
}

func (o ObjectiveBase) GetValue() float64 {
	var v float64
	if o.Value != nil {
		v = *o.Value
	}
	return v
}

// Objective represents single objective for SLO, for internal usage.
type Objective struct {
	ObjectiveBase `json:",inline"`
	// <!-- Go struct field and type names renaming budgetTarget to target has been postponed after GA as requested
	// in PC-1240. -->
	BudgetTarget    *float64          `json:"target"`
	TimeSliceTarget *float64          `json:"timeSliceTarget,omitempty"`
	CountMetrics    *CountMetricsSpec `json:"countMetrics,omitempty"`
	RawMetric       *RawMetricSpec    `json:"rawMetric,omitempty"`
	Composite       *CompositeSpec    `json:"composite,omitempty"`
	Operator        *string           `json:"op,omitempty"`
	// Primary is used to highlight the main (primary) objective of the [SLO].
	Primary *bool `json:"primary,omitempty"`
}

func (o Objective) GetBudgetTarget() float64 {
	var v float64
	if o.BudgetTarget != nil {
		v = *o.BudgetTarget
	}
	return v
}

func (o Objective) IsComposite() bool {
	return o.Composite != nil
}

// Indicator represents integration with metric source can be. e.g. Prometheus, Datadog, for internal usage.
type Indicator struct {
	MetricSource MetricSourceSpec `json:"metricSource"`
	RawMetric    *MetricSpec      `json:"rawMetric,omitempty"`
}

type MetricSourceSpec struct {
	Name    string        `json:"name"`
	Project string        `json:"project,omitempty"`
	Kind    manifest.Kind `json:"kind,omitempty"`
}

// Composite represents configuration for Composite SLO.
// Deprecated: this implementation of Composite will be removed and replaced with SLO.Spec.Objectives.Composite.
type Composite struct {
	BudgetTarget      *float64                    `json:"target"`
	BurnRateCondition *CompositeBurnRateCondition `json:"burnRateCondition,omitempty"`
}

// CompositeVersion represents composite version history stored for restoring process.
type CompositeVersion struct {
	Version      int32
	Created      string
	Dependencies []string
}

// CompositeBurnRateCondition represents configuration for Composite SLO  with occurrences budgeting method.
type CompositeBurnRateCondition struct {
	Value    float64 `json:"value"`
	Operator string  `json:"op"`
}

// AnomalyConfig represents relationship between anomaly type and selected notification methods.
// This will be removed (moved into Anomaly Policy) in PC-8502
type AnomalyConfig struct {
	NoData *AnomalyConfigNoData `json:"noData"`
}

// AnomalyConfigNoData contains alertMethods used for No Data anomaly type.
type AnomalyConfigNoData struct {
	AlertMethods []AnomalyConfigAlertMethod `json:"alertMethods"`
}

// AnomalyConfigAlertMethod represents a single alert method used in AnomalyConfig
// defined by name and project.
type AnomalyConfigAlertMethod struct {
	Name    string `json:"name"`
	Project string `json:"project,omitempty"`
}

// Status holds dynamic fields returned when the Service is fetched from Nobl9 platform.
// Status is not part of the static object definition.
type Status struct {
	UpdatedAt    string        `json:"updatedAt,omitempty"`
	ReplayStatus *ReplayStatus `json:"timeTravel,omitempty"`
}

type ReplayStatus struct {
	Status    string `json:"status"`
	Unit      string `json:"unit"`
	Value     int    `json:"value"`
	StartTime string `json:"startTime,omitempty"`
}
