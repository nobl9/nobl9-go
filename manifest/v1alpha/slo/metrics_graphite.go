package slo

// GraphiteMetric represents metric from Graphite.
type GraphiteMetric struct {
	MetricPath *string `json:"metricPath" validate:"required,metricPathGraphite"`
}
