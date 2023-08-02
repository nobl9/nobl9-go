package v1alpha

import (
	"github.com/nobl9/nobl9-go/manifest"
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
	Name        string `json:"name" validate:"required,objectName"`
	DisplayName string `json:"displayName,omitempty" validate:"omitempty,min=0,max=63"`
	Project     string `json:"project,omitempty" validate:"objectName"`
	Labels      Labels `json:"labels,omitempty" validate:"omitempty,labels"`
}

// SLOSpec represents content of Spec typical for SLO Object
type SLOSpec struct {
	Description     string         `json:"description,omitempty" validate:"description" example:"Total count of server requests"` //nolint:lll
	Indicator       Indicator      `json:"indicator"`
	BudgetingMethod string         `json:"budgetingMethod" validate:"required,budgetingMethod" example:"Occurrences"`
	Thresholds      []Threshold    `json:"objectives" validate:"required,dive"`
	Service         string         `json:"service" validate:"required,objectName" example:"webapp-service"`
	TimeWindows     []TimeWindow   `json:"timeWindows" validate:"required,len=1,dive"`
	AlertPolicies   []string       `json:"alertPolicies,omitempty" validate:"omitempty"`
	Attachments     []Attachment   `json:"attachments,omitempty" validate:"omitempty,max=20,dive"`
	CreatedAt       string         `json:"createdAt,omitempty"`
	Composite       *Composite     `json:"composite,omitempty" validate:"omitempty"`
	AnomalyConfig   *AnomalyConfig `json:"anomalyConfig,omitempty" validate:"omitempty"`
}

type SLOStatus struct {
	TimeTravelStatus *TimeTravelStatus `json:"timeTravel,omitempty"`
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

// ThresholdBase base structure representing a threshold
type ThresholdBase struct {
	DisplayName string  `json:"displayName" validate:"omitempty,min=0,max=63" example:"Good"`
	Value       float64 `json:"value" validate:"numeric" example:"100"`
	Name        string  `json:"name" validate:"omitempty,objectName"`
	NameChanged bool    `json:"-"`
}

// Threshold represents single threshold for SLO, for internal usage
type Threshold struct {
	ThresholdBase `json:",inline"`
	// <!-- Go struct field and type names renaming budgetTarget to target has been postponed after GA as requested
	// in PC-1240. -->
	BudgetTarget *float64 `json:"target" validate:"required,numeric,gte=0,lt=1" example:"0.9"`
	// <!-- Go struct field and type names renaming thresholds to objectives has been postponed after GA as requested
	// in PC-1240. -->
	TimeSliceTarget *float64          `json:"timeSliceTarget,omitempty" example:"0.9"`
	CountMetrics    *CountMetricsSpec `json:"countMetrics,omitempty"`
	RawMetric       *RawMetricSpec    `json:"rawMetric,omitempty"`
	Operator        *string           `json:"op,omitempty" example:"lte"`
}

// Indicator represents integration with metric source can be. e.g. Prometheus, Datadog, for internal usage
type Indicator struct {
	MetricSource *MetricSourceSpec `json:"metricSource" validate:"required"`
	RawMetric    *MetricSpec       `json:"rawMetric,omitempty"`
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

// postParse implements postParser interface.
func (s SLO) postParse() (SLO, error) {
	// to keep BC with the ThousandEyes initial implementation (that did not support passing TestType),
	// we default `res.Spec.Indicator.RawMetrics.ThousandEyes.TestType` to a value that, until now, was implicitly assumed
	if s.Spec.Indicator.RawMetric != nil &&
		s.Spec.Indicator.RawMetric.ThousandEyes != nil &&
		s.Spec.Indicator.RawMetric.ThousandEyes.TestType == nil {
		metricType := ThousandEyesNetLatency
		s.Spec.Indicator.RawMetric.ThousandEyes.TestType = &metricType
	}

	for i, threshold := range s.Spec.Thresholds {
		if threshold.RawMetric != nil &&
			threshold.RawMetric.MetricQuery != nil &&
			threshold.RawMetric.MetricQuery.ThousandEyes != nil &&
			threshold.RawMetric.MetricQuery.ThousandEyes.TestType == nil {
			metricType := ThousandEyesNetLatency
			s.Spec.Thresholds[i].RawMetric.MetricQuery.ThousandEyes.TestType = &metricType
		}
	}

	if s.Spec.Indicator.MetricSource.Project == "" {
		s.Spec.Indicator.MetricSource.Project = s.Metadata.Project
	}
	if !s.Spec.Indicator.MetricSource.Kind.IsValid() {
		s.Spec.Indicator.MetricSource.Kind = manifest.KindAgent
	}

	// we're moving towards the version where raw metrics are defined on each objective, but for now,
	// we have to make sure that old contract (with indicator defined directly on the SLO's spec) is also supported
	if s.Spec.Indicator.RawMetric != nil {
		for i := range s.Spec.Thresholds {
			s.Spec.Thresholds[i].RawMetric = &RawMetricSpec{
				MetricQuery: s.Spec.Indicator.RawMetric,
			}
		}
	}

	// AnomalyConfig will be moved into Anomaly Rules in PC-8502.
	// Set the default value of all alert methods defined in anomaly config to the same project
	// that is used by SLO.
	if s.Spec.AnomalyConfig != nil && s.Spec.AnomalyConfig.NoData != nil {
		for i := 0; i < len(s.Spec.AnomalyConfig.NoData.AlertMethods); i++ {
			if s.Spec.AnomalyConfig.NoData.AlertMethods[i].Project == "" {
				s.Spec.AnomalyConfig.NoData.AlertMethods[i].Project = s.Metadata.Project
			}
		}
	}

	return s, nil
}
