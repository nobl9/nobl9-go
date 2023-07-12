package v1alpha

import (
	"time"

	"github.com/nobl9/nobl9-go/manifest/v1alpha/twindow"
)

type SLIAnalysis struct {
	Metadata        AnalysisMetadata     `json:"metadata"`
	MetricSpec      AnalysisMetricSpec   `json:"metricSpec"`
	CalculationSpec *AnalysisCalculation `json:"calculationSpec,omitempty"`
	Period          AnalysisPeriod       `json:"period"`
	Status          string               `json:"status"`
	CreatedAt       string               `json:"createdAt"`
	UpdatedAt       string               `json:"updatedAt"`
}

type AnalysisMetricSpec struct {
	Kind         string            `json:"kind" validate:"required,metricSourceKind"`
	MetricSource string            `json:"metricSource" validate:"required,objectName"`
	RawMetric    *MetricSpec       `json:"rawMetric,omitempty"`
	CountMetrics *CountMetricsSpec `json:"countMetrics,omitempty"`
}

type AnalysisPeriod struct {
	StartTime string `json:"startTime" validate:"required,dateWithTime"`
	EndTime   string `json:"endTime" validate:"required,dateWithTime"`
	TimeZone  string `json:"timeZone" validate:"required,timeZone"`
}

func (p *AnalysisPeriod) GetStartDate() (time.Time, error) {
	location, err := time.LoadLocation(p.TimeZone)
	if err != nil {
		return time.Time{}, err
	}
	startTime, err := time.ParseInLocation(twindow.IsoDateTimeOnlyLayout, p.StartTime, location)
	if err != nil {
		return time.Time{}, err
	}
	return startTime, nil
}

func (p *AnalysisPeriod) GetEndDate() (time.Time, error) {
	location, err := time.LoadLocation(p.TimeZone)
	if err != nil {
		return time.Time{}, err
	}
	endTime, err := time.ParseInLocation(twindow.IsoDateTimeOnlyLayout, p.EndTime, location)
	if err != nil {
		return time.Time{}, err
	}
	return endTime, nil
}

type AnalysisMetadata struct {
	Name string `json:"name,omitempty"`
	UpdatableAnalysisMetadata
}

// DataSourceType returns data source type for SLIAnalysis.
func (s *SLIAnalysis) DataSourceType() int32 {
	var dataSourceCode int32
	if s.MetricSpec.RawMetric != nil {
		dataSourceCode = int32(s.MetricSpec.RawMetric.DataSourceType())
	}
	if s.MetricSpec.CountMetrics != nil {
		dataSourceCode = int32(s.MetricSpec.CountMetrics.TotalMetric.DataSourceType())
	}
	return dataSourceCode
}

// AllMetricSpecs returns slice of all metrics defined in SLIAnalysis regardless of their type.
func (s *SLIAnalysis) AllMetricSpecs() []*MetricSpec {
	var metrics []*MetricSpec
	if s.MetricSpec.RawMetric != nil {
		metrics = append(metrics, s.MetricSpec.RawMetric)
	}
	if s.MetricSpec.CountMetrics != nil {
		metrics = append(metrics, s.MetricSpec.CountMetrics.GoodMetric)
		metrics = append(metrics, s.MetricSpec.CountMetrics.TotalMetric)
	}
	return metrics
}

func (s SLIAnalysis) IsValid() error {
	v := NewValidator()
	return v.Check(s)
}

type UpdatableAnalysisMetadata struct {
	Project     string `json:"project" validate:"required,objectName"`
	DisplayName string `json:"displayName" validate:"min=0,max=63,required"`
}

func (s UpdatableAnalysisMetadata) IsValid() error {
	v := NewValidator()
	return v.Check(s)
}

type AnalysisCalculation struct {
	Value           float64 `json:"value" validate:"required,numeric,gte=0" example:"2.9"`
	BudgetTarget    float64 `json:"target" validate:"required,numeric,gte=0,lt=1" example:"0.9"`
	TimeSliceTarget float64 `json:"timeSliceTarget,omitempty" example:"0.9"`
	BudgetingMethod string  `json:"budgetingMethod" validate:"required,budgetingMethod" example:"Occurrences"`
	Operator        string  `json:"op,omitempty" validate:"required,operator"  example:"lte"`
	CreatedAt       string  `json:"createdAt"`
}

type AnalysisCalculationSummary struct {
	GoodTotalRatio       float64 `json:"goodTotalRatio"`
	BudgetBurned         float64 `json:"budgetBurned"`
	TimeOfBadEvents      float64 `json:"timeOfBadEvents"`
	ErrorBudgetRemaining float64 `json:"errorBudgetRemaining"`
}

func (c AnalysisCalculation) IsValid() error {
	v := NewValidator()
	return v.Check(c)
}
