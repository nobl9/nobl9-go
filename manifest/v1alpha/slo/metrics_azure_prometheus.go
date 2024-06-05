package slo

import "github.com/nobl9/nobl9-go/internal/validation"

// AzurePrometheusMetric represents metric from Azure Monitor managed service for Prometheus
type AzurePrometheusMetric struct {
	PromQL string `json:"promql"`
}

var azurePrometheusValidation = validation.New[AzurePrometheusMetric](
	validation.For(func(p AzurePrometheusMetric) string { return p.PromQL }).
		WithName("promql").
		Required().
		Rules(validation.StringNotEmpty()),
)
