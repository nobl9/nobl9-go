package slo

// PrometheusMetric represents metric from Prometheus
type PrometheusMetric struct {
	PromQL *string `json:"promql" validate:"required" example:"cpu_usage_user{cpu=\"cpu-total\"}"`
}
