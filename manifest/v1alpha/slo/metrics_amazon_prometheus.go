package slo

// AmazonPrometheusMetric represents metric from Amazon Managed Prometheus
type AmazonPrometheusMetric struct {
	PromQL *string `json:"promql" validate:"required" example:"cpu_usage_user{cpu=\"cpu-total\"}"`
}
