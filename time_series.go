package nobl9

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

type SLOTimeSeries struct {
	MetadataHolder
	TimeWindows                 []TimeWindowTimeSeries `json:"timewindows,omitempty"`
	RawSLIPercentilesTimeSeries Percentile             `json:"percentiles,omitempty"`
}

type TimeWindowTimeSeries struct {
	TimeWindow `json:"timewindow,omitempty"`
	// <!-- Go struct field and type names renaming thresholds to objectives has been postponed after GA as requested
	// in PC-1240. -->
	Thresholds []ThresholdTimeSeries `json:"objectives,omitempty"`
}

type ThresholdTimeSeries struct {
	ThresholdBase
	InstantaneousBurnRateTimeSeries
	CumulativeBurnedTimeSeries
	Status   ThresholdTimeSeriesStatus `json:"status"`
	Operator *string                   `json:"op,omitempty"`
	CountsSLITimeSeries
	BurnDownTimeSeries
}

type ThresholdTimeSeriesStatus struct {
	BurnedBudget            *float64 `json:"burnedBudget,omitempty" example:"0.25"`
	RemainingBudget         *float64 `json:"errorBudgetRemainingPercentage,omitempty" example:"0.25"`
	RemainingBudgetDuration *float64 `json:"errorBudgetRemaining,omitempty" example:"300"`
	InstantaneousBurnRate   *float64 `json:"instantaneousBurnRate,omitempty" example:"1.25"`
	Condition               *string  `json:"condition,omitempty" example:"ok"`
}

type TimeSeriesData [][]interface{}

type Percentile struct {
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
	TotalCount TimeSeriesData `json:"totalCount,omitempty"`
}

type InstantaneousBurnRateTimeSeries struct {
	InstantaneousBurnRate TimeSeriesData `json:"instantaneousBurnRate,omitempty"`
}

type CumulativeBurnedTimeSeries struct {
	CumulativeBurned TimeSeriesData `json:"cumulativeBurned,omitempty"`
}

type BurnDownTimeSeries struct {
	BurnDown []TimeSeriesData `json:"burnDown,omitempty"`
}

type SLOHistoryReport struct {
	MetadataHolder
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
