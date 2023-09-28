package v1alpha

import (
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/labels"
)

//go:generate go run ../../scripts/generate-object-impl.go SLO

// SLO struct which mapped one to one with kind: slo yaml definition, external usage
type SLO struct {
	APIVersion string        `json:"apiVersion"`
	Kind       manifest.Kind `json:"kind"`
	Metadata   SLOMetadata   `json:"metadata"`
	Spec       SLOSpec       `json:"spec"`
	Status     *SLOStatus    `json:"status,omitempty"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
}

type SLOMetadata struct {
	Name        string        `json:"name" validate:"required,objectName"`
	DisplayName string        `json:"displayName,omitempty" validate:"omitempty,min=0,max=63"`
	Project     string        `json:"project,omitempty" validate:"objectName"`
	Labels      labels.Labels `json:"labels,omitempty" validate:"omitempty,labels"`
}

// SLOSpec represents content of Spec typical for SLO Object
type SLOSpec struct {
	Description     string         `json:"description" validate:"description" example:"Total count of server requests"`
	Indicator       Indicator      `json:"indicator"`
	BudgetingMethod string         `json:"budgetingMethod" validate:"required,budgetingMethod" example:"Occurrences"`
	Objectives      []Objective    `json:"objectives" validate:"required,dive"`
	Service         string         `json:"service" validate:"required,objectName" example:"webapp-service"`
	TimeWindows     []TimeWindow   `json:"timeWindows" validate:"required,len=1,dive"`
	AlertPolicies   []string       `json:"alertPolicies" validate:"omitempty"`
	Attachments     []Attachment   `json:"attachments,omitempty" validate:"omitempty,max=20,dive"`
	CreatedAt       string         `json:"createdAt,omitempty"`
	Composite       *Composite     `json:"composite,omitempty" validate:"omitempty"`
	AnomalyConfig   *AnomalyConfig `json:"anomalyConfig,omitempty" validate:"omitempty"`
}

type SLOStatus struct {
	ReplayStatus *ReplayStatus `json:"timeTravel,omitempty"`
}

// Calendar struct represents calendar time window
type Calendar struct {
	StartTime string `json:"startTime" validate:"required,dateWithTime,minDateTime" example:"2020-01-21 12:30:00"`
	TimeZone  string `json:"timeZone" validate:"required,timeZone" example:"America/New_York"`
}

// Period represents period of time
type Period struct {
	Begin string `json:"begin"`
	End   string `json:"end"`
}

// TimeWindow represents content of time window
type TimeWindow struct {
	Unit      string    `json:"unit" validate:"required,timeUnit" example:"Week"`
	Count     int       `json:"count" validate:"required,gt=0" example:"1"`
	IsRolling bool      `json:"isRolling" example:"true"`
	Calendar  *Calendar `json:"calendar,omitempty"`

	// Period is only returned in `/get/slo` requests it is ignored for `/apply`
	Period *Period `json:"period,omitempty"`
}

// Attachment represents user defined URL attached to SLO
type Attachment struct {
	URL         string  `json:"url" validate:"required,url"`
	DisplayName *string `json:"displayName,omitempty" validate:"max=63"`
}

// ObjectiveBase base structure representing an objective.
type ObjectiveBase struct {
	DisplayName string  `json:"displayName" validate:"omitempty,min=0,max=63" example:"Good"`
	Value       float64 `json:"value" validate:"numeric" example:"100"`
	Name        string  `json:"name" validate:"omitempty,objectName"`
	NameChanged bool    `json:"-"`
}

// Objective represents single objective for SLO, for internal usage
type Objective struct {
	ObjectiveBase `json:",inline"`
	// <!-- Go struct field and type names renaming budgetTarget to target has been postponed after GA as requested
	// in PC-1240. -->
	BudgetTarget    *float64          `json:"target" validate:"required,numeric,gte=0,lt=1" example:"0.9"`
	TimeSliceTarget *float64          `json:"timeSliceTarget,omitempty" example:"0.9"`
	CountMetrics    *CountMetricsSpec `json:"countMetrics,omitempty"`
	RawMetric       *RawMetricSpec    `json:"rawMetric,omitempty"`
	Operator        *string           `json:"op,omitempty" example:"lte"`
}

// Indicator represents integration with metric source can be. e.g. Prometheus, Datadog, for internal usage
type Indicator struct {
	MetricSource MetricSourceSpec `json:"metricSource" validate:"required"`
	RawMetric    *MetricSpec      `json:"rawMetric,omitempty"`
}

type MetricSourceSpec struct {
	Project string        `json:"project,omitempty" validate:"omitempty,objectName" example:"default"`
	Name    string        `json:"name" validate:"required,objectName" example:"prometheus-source"`
	Kind    manifest.Kind `json:"kind,omitempty" validate:"omitempty,metricSourceKind" example:"Agent"`
}

// Composite represents configuration for Composite SLO.
type Composite struct {
	BudgetTarget      float64                     `json:"target" validate:"required,numeric,gte=0,lt=1" example:"0.9"`
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
	Value    float64 `json:"value" validate:"numeric,gte=0,lte=1000" example:"2"`
	Operator string  `json:"op" validate:"required,oneof=gt" example:"gt"`
}

// AnomalyConfig represents relationship between anomaly type and selected notification methods.
// This will be removed (moved into Anomaly Policy) in PC-8502
type AnomalyConfig struct {
	NoData *AnomalyConfigNoData `json:"noData" validate:"omitempty"`
}

// AnomalyConfigNoData contains alertMethods used for No Data anomaly type.
type AnomalyConfigNoData struct {
	AlertMethods []AnomalyConfigAlertMethod `json:"alertMethods" validate:"required"`
}

// AnomalyConfigAlertMethod represents a single alert method used in AnomalyConfig
// defined by name and project.
type AnomalyConfigAlertMethod struct {
	Name    string `json:"name" validate:"required,objectName" example:"slack-monitoring-channel"`
	Project string `json:"project,omitempty" validate:"objectName" example:"default"`
}
