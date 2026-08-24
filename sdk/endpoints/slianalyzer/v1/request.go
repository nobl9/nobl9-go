package v1

import (
	"github.com/nobl9/nobl9-go/manifest"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
)

// CreateAnalysisRequest defines an SLI analysis to create.
type CreateAnalysisRequest struct {
	Metadata   AnalysisMetadata   `json:"metadata"`
	MetricSpec AnalysisMetricSpec `json:"metricSpec"`
	Period     AnalysisPeriod     `json:"period"`
}

// UpdateAnalysisRequest defines the mutable metadata of an SLI analysis.
type UpdateAnalysisRequest struct {
	Project     string `json:"project"`
	DisplayName string `json:"displayName"`
}

// CreateCalculationRequest defines the threshold and budget settings for an analysis calculation.
type CreateCalculationRequest struct {
	Value           float64 `json:"value"`
	BudgetTarget    float64 `json:"target"`
	TimeSliceTarget float64 `json:"timeSliceTarget,omitempty"`
	BudgetingMethod string  `json:"budgetingMethod"`
	Operator        string  `json:"op,omitempty"`
}

// AnalysisMetadata identifies an SLI analysis and its project.
type AnalysisMetadata struct {
	Name        string `json:"name,omitempty"`
	DisplayName string `json:"displayName"`
	Project     string `json:"project"`
}

// AnalysisMetricSpec defines the data source and metric query for an SLI analysis.
// Exactly one of RawMetric and CountMetrics must be set.
type AnalysisMetricSpec struct {
	Kind         manifest.Kind                `json:"kind"`
	MetricSource string                       `json:"metricSource"`
	RawMetric    *v1alphaSLO.MetricSpec       `json:"rawMetric,omitempty"`
	CountMetrics *v1alphaSLO.CountMetricsSpec `json:"countMetrics,omitempty"`
}

// AnalysisPeriod defines the time range for an SLI analysis.
// StartTime and EndTime use the `2006-01-02 15:04:05` layout in TimeZone.
type AnalysisPeriod struct {
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	TimeZone  string `json:"timeZone"`
}
