// Package v1alpha represents objects available in API n9/v1alpha
package v1alpha

import (
	"github.com/nobl9/nobl9-go/manifest"
)

// Time Series

type SLOTimeSeries struct {
	manifest.MetadataHolder
	TimeWindows []TimeWindowTimeSeries `json:"timewindows,omitempty"`
	Status      *SLOStatus             `json:"status,omitempty"`
}

type ThresholdTimeSeries struct {
	ThresholdBase
	InstantaneousBurnRateTimeSeries
	Status ThresholdTimeSeriesStatus `json:"status"`
	QueryValidationStatus
	Downsampling DownsamplingConfig `json:"downsampling,omitempty"`
	Operator     *string            `json:"op,omitempty"`
	CountsSLITimeSeries
	RawCountsSLITimeSeries
	BurnDownTimeSeries
	RawSLIPercentilesTimeSeries
	RawSLIThresholdTimeSeries
}

type ThresholdTimeSeriesStatus struct {
	BurnedBudget            *float64 `json:"burnedBudget,omitempty" example:"0.25"`
	RemainingBudget         *float64 `json:"errorBudgetRemainingPercentage,omitempty" example:"0.25"`
	RemainingBudgetDuration *float64 `json:"errorBudgetRemaining,omitempty" example:"300"`
	InstantaneousBurnRate   *float64 `json:"instantaneousBurnRate,omitempty" example:"1.25"`
	Condition               *string  `json:"condition,omitempty" example:"ok"`
}

type CompositeTimeSeriesStatus struct {
	BurnedBudget            *float64 `json:"burnedBudget,omitempty" example:"0.25"`
	RemainingBudget         *float64 `json:"errorBudgetRemainingPercentage,omitempty" example:"0.25"`
	RemainingBudgetDuration *float64 `json:"errorBudgetRemaining,omitempty" example:"300"`
	InstantaneousBurnRate   *float64 `json:"instantaneousBurnRate,omitempty" example:"1.25"`
	Condition               *string  `json:"condition,omitempty" example:"ok"`
}

type CompositeTimeSeries struct {
	InstantaneousBurnRateTimeSeries
	Status                 CompositeTimeSeriesStatus `json:"status"`
	Operator               *string                   `json:"op,omitempty"`
	CompositeBurnRateValue *float64                  `json:"compositeBurnRateValue,omitempty"`
	BudgetTarget           *float64                  `json:"budgetTarget,omitempty"`
	BurnDownTimeSeries
}

type TimeWindowTimeSeries struct {
	TimeWindow `json:"timewindow,omitempty"`
	// <!-- Go struct field and type names renaming thresholds to objectives has been postponed after GA as requested
	// in PC-1240. -->
	Thresholds []ThresholdTimeSeries `json:"objectives,omitempty"`
	Composite  *CompositeTimeSeries  `json:"composite,omitempty"`
}

const (
	P1  string = "p1"
	P5  string = "p5"
	P10 string = "p10"
	P50 string = "p50"
	P90 string = "p90"
	P95 string = "p95"
	P99 string = "p99"
)

func GetAvailablePercentiles() []string {
	return []string{P1, P5, P10, P50, P90, P95, P99}
}

type Percentiles struct {
	P1  TimeSeriesData `json:"p1,omitempty"`
	P5  TimeSeriesData `json:"p5,omitempty"`
	P10 TimeSeriesData `json:"p10,omitempty"`
	P50 TimeSeriesData `json:"p50,omitempty"`
	P90 TimeSeriesData `json:"p90,omitempty"`
	P95 TimeSeriesData `json:"p95,omitempty"`
	P99 TimeSeriesData `json:"p99,omitempty"`
}

type CountsSLITimeSeries struct {
	GoodCount  TimeSeriesData `json:"goodCount,omitempty"`
	BadCount   TimeSeriesData `json:"badCount,omitempty"`
	TotalCount TimeSeriesData `json:"totalCount,omitempty"`
}

type RawCountsSLITimeSeries struct {
	RawGoodCount  TimeSeriesData `json:"rawGoodCount,omitempty"`
	RawBadCount   TimeSeriesData `json:"rawBadCount,omitempty"`
	RawTotalCount TimeSeriesData `json:"rawTotalCount,omitempty"`
}

type InstantaneousBurnRateTimeSeries struct {
	InstantaneousBurnRate TimeSeriesData `json:"instantaneousBurnRate,omitempty"`
}

type DownsamplingConfig struct {
	WindowDuration    float64 `json:"windowDuration,omitempty"`
	RemainingDuration float64 `json:"remainingDuration,omitempty"`
}

// SLO History Report

type SLOHistoryReport struct {
	manifest.MetadataHolder
	TimeWindows []TimeWindowHistoryReport `json:"timewindows,omitempty"`
}

type TimeWindowHistoryReport struct {
	TimeWindow `json:"timewindow,omitempty"`
	Thresholds []ThresholdHistoryReport `json:"objectives,omitempty"`
}

type ThresholdHistoryReport struct {
	ThresholdBase
	BurnDownTimeSeries
}

// Common

type TimeSeriesData [][]interface{}

type BurnDownTimeSeries struct {
	BurnDown []TimeSeriesData `json:"burnDown,omitempty"`
}

type RawSLIPercentilesTimeSeries struct {
	Percentiles *Percentiles `json:"percentiles,omitempty"`
}

type RawSLIThresholdTimeSeries struct {
	Raw TimeSeriesData `json:"raw,omitempty"`
}
