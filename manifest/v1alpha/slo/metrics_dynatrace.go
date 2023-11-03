package slo

// DynatraceMetric represents metric from Dynatrace.
type DynatraceMetric struct {
	MetricSelector *string `json:"metricSelector" validate:"required"`
}
