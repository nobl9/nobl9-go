package slo

import (
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/twindow"
)

//go:generate go run ../../scripts/generate-object-impl.go SLO

// New creates a new Service based on provided Metadata nad Spec.
func New(metadata Metadata, spec Spec) SLO {
	return SLO{
		APIVersion: manifest.VersionV1alpha.String(),
		Kind:       manifest.KindSLO,
		Metadata:   metadata,
		Spec:       spec,
	}
}

// SLO struct which mapped one to one with kind: slo yaml definition, external usage
type SLO struct {
	APIVersion string        `json:"apiVersion"`
	Kind       manifest.Kind `json:"kind"`
	Metadata   Metadata      `json:"metadata"`
	Spec       Spec          `json:"spec"`
	Status     *Status       `json:"status,omitempty"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
}

// Metadata provides identity information for SLO.
type Metadata struct {
	Name        string         `json:"name"`
	DisplayName string         `json:"displayName,omitempty"`
	Project     string         `json:"project,omitempty"`
	Labels      v1alpha.Labels `json:"labels,omitempty"`
}

// Spec holds detailed information specific to SLO.
type Spec struct {
	Description     string         `json:"description"`
	Indicator       Indicator      `json:"indicator"`
	BudgetingMethod string         `json:"budgetingMethod"`
	Objectives      []Objective    `json:"objectives"`
	Service         string         `json:"service"`
	TimeWindows     []TimeWindow   `json:"timeWindows"`
	AlertPolicies   []string       `json:"alertPolicies"`
	Attachments     []Attachment   `json:"attachments,omitempty"`
	CreatedAt       string         `json:"createdAt,omitempty"`
	Composite       *Composite     `json:"composite,omitempty"`
	AnomalyConfig   *AnomalyConfig `json:"anomalyConfig,omitempty"`
}

// TimeWindow represents content of time window
type TimeWindow struct {
	Unit      string    `json:"unit"`
	Count     int       `json:"count"`
	IsRolling bool      `json:"isRolling"`
	Calendar  *Calendar `json:"calendar,omitempty"`

	// Period is only returned in `/get/slo` requests it is ignored for `/apply`
	Period *Period `json:"period,omitempty"`
}

// GetType returns value of twindow.TimeWindowTypeEnum for given time window>
func (tw TimeWindow) GetType() twindow.TimeWindowTypeEnum {
	if tw.isCalendar() {
		return twindow.Calendar
	}
	return twindow.Rolling
}

func (tw TimeWindow) isCalendar() bool {
	return tw.Calendar != nil
}

// Calendar struct represents calendar time window
type Calendar struct {
	StartTime string `json:"startTime"`
	TimeZone  string `json:"timeZone"`
}

// Period represents period of time
type Period struct {
	Begin string `json:"begin"`
	End   string `json:"end"`
}

// Attachment represents user defined URL attached to SLO
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

// Objective represents single objective for SLO, for internal usage
type Objective struct {
	ObjectiveBase `json:",inline"`
	// <!-- Go struct field and type names renaming budgetTarget to target has been postponed after GA as requested
	// in PC-1240. -->
	BudgetTarget    *float64          `json:"target"`
	TimeSliceTarget *float64          `json:"timeSliceTarget,omitempty"`
	CountMetrics    *CountMetricsSpec `json:"countMetrics,omitempty"`
	RawMetric       *RawMetricSpec    `json:"rawMetric,omitempty"`
	Operator        *string           `json:"op,omitempty"`
}

// Indicator represents integration with metric source can be. e.g. Prometheus, Datadog, for internal usage
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
	NoData *AnomalyConfigNoData `json:"noData" validate:"omitempty"`
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
	ReplayStatus *ReplayStatus `json:"timeTravel,omitempty"`
}

type ReplayStatus struct {
	Status    string `json:"status"`
	Unit      string `json:"unit"`
	Value     int    `json:"value"`
	StartTime string `json:"startTime,omitempty"`
}