package slo

import "github.com/nobl9/nobl9-go/internal/validation"

// AmazonPrometheusMetric represents metric from Amazon Managed Prometheus
type AmazonPrometheusMetric struct {
	PromQL *string `json:"promql"`
}

var amazonPrometheusValidation = validation.New[AmazonPrometheusMetric](
	validation.ForPointer(func(p AmazonPrometheusMetric) *string { return p.PromQL }).
		WithName("promql").
		Required().
		Rules(validation.StringNotEmpty()),
)
