package v1

// Analysis describes an SLI analysis and its current processing status.
type Analysis struct {
	Metadata        AnalysisMetadata     `json:"metadata"`
	MetricSpec      AnalysisMetricSpec   `json:"metricSpec"`
	CalculationSpec *AnalysisCalculation `json:"calculationSpec,omitempty"`
	Period          AnalysisPeriod       `json:"period"`
	Status          Status               `json:"status"`
	CreatedBy       *UserDetails         `json:"createdBy,omitempty"`
	CreatedAt       string               `json:"createdAt"`
	UpdatedAt       string               `json:"updatedAt"`
}

// AnalysisCalculation describes the threshold and budget settings used for a calculation.
type AnalysisCalculation struct {
	Value           float64 `json:"value"`
	BudgetTarget    float64 `json:"target"`
	TimeSliceTarget float64 `json:"timeSliceTarget,omitempty"`
	BudgetingMethod string  `json:"budgetingMethod"`
	Operator        string  `json:"op,omitempty"`
	CreatedAt       string  `json:"createdAt"`
}

// UserDetails identifies the account that created an analysis.
type UserDetails struct {
	ID                 string `json:"id"`
	FirstName          string `json:"firstName,omitempty"`
	LastName           string `json:"lastName,omitempty"`
	AccountType        string `json:"accountType,omitempty"`
	AccountDisplayName string `json:"accountDisplayName,omitempty"`
	IsRemoved          bool   `json:"isRemoved,omitempty"`
}

// AnalysisCalculationSummary summarizes the result of an SLI analysis calculation.
type AnalysisCalculationSummary struct {
	GoodTotalRatio       float64 `json:"goodTotalRatio"`
	BudgetBurned         float64 `json:"budgetBurned"`
	TimeOfBadEvents      float64 `json:"timeOfBadEvents"`
	ErrorBudgetRemaining float64 `json:"errorBudgetRemaining"`
}

// TimeseriesPoint contains a timestamp and its numeric value.
// The API encodes each point as a two-element JSON array.
type TimeseriesPoint []any

// Timeseries contains timestamp-value points in chronological order.
type Timeseries []TimeseriesPoint

// AggregatedTimeseries contains either raw metric percentiles or count metric series.
type AggregatedTimeseries struct {
	RawMetric    map[string]Timeseries   `json:"rawMetric,omitempty"`
	CountMetrics *CountMetricsTimeseries `json:"countMetrics,omitempty"`
}

// CountMetricsTimeseries contains the available count metric series.
type CountMetricsTimeseries struct {
	TotalCount Timeseries `json:"totalCount,omitempty"`
	GoodCount  Timeseries `json:"goodCount,omitempty"`
	BadCount   Timeseries `json:"badCount,omitempty"`
}

// AnalysisStats contains descriptive statistics for the available metric series.
type AnalysisStats struct {
	RawMetric  *AnalysisStatsData `json:"rawMetric,omitempty"`
	TotalCount *AnalysisStatsData `json:"totalCount,omitempty"`
	GoodCount  *AnalysisStatsData `json:"goodCount,omitempty"`
	BadCount   *AnalysisStatsData `json:"badCount,omitempty"`
}

// AnalysisStatsData contains descriptive statistics for one metric series.
type AnalysisStatsData struct {
	Min         *float64                  `json:"min,omitempty"`
	Mean        *float64                  `json:"mean,omitempty"`
	Max         *float64                  `json:"max,omitempty"`
	StdDev      *float64                  `json:"stdDev,omitempty"`
	Variance    *float64                  `json:"variance,omitempty"`
	Range       *float64                  `json:"range,omitempty"`
	Percentiles *AnalysisStatsPercentiles `json:"percentiles,omitempty"`
}

// AnalysisStatsPercentiles contains the supported percentile values.
type AnalysisStatsPercentiles struct {
	P1        *float64 `json:"p1,omitempty"`
	P5        *float64 `json:"p5,omitempty"`
	P10       *float64 `json:"p10,omitempty"`
	P50       *float64 `json:"p50,omitempty"`
	P90       *float64 `json:"p90,omitempty"`
	P95       *float64 `json:"p95,omitempty"`
	P99       *float64 `json:"p99,omitempty"`
	P99Point9 *float64 `json:"p99_9,omitempty"`
}

// AnalysisHistogram contains histogram buckets for an SLI analysis.
type AnalysisHistogram struct {
	Bins []AnalysisHistogramBucket `json:"bins"`
}

// AnalysisHistogramBucket contains one histogram interval and its metric counts.
type AnalysisHistogramBucket struct {
	GT    *float64 `json:"gt"`
	LTE   *float64 `json:"lte"`
	Count *float64 `json:"count,omitempty"`
	Good  *float64 `json:"good,omitempty"`
	Bad   *float64 `json:"bad,omitempty"`
	Total *float64 `json:"total,omitempty"`
}

// Status identifies the current processing stage of an SLI analysis.
type Status string

// SLI analysis processing statuses.
const (
	StatusFetchingHistoricalData Status = "fetching_historical_data"
	StatusSavingHistoricalData   Status = "saving_historical_data"
	StatusImportFailed           Status = "import_failed"
	StatusImportCompleted        Status = "import_completed"
	StatusCalculationInProgress  Status = "calculation_in_progress"
	StatusSavingCalculatedData   Status = "saving_calculated_data"
	StatusCalculationFailed      Status = "calculation_failed"
	StatusCalculationCompleted   Status = "calculation_completed"
)
