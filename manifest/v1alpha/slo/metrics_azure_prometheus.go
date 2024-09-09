package slo

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// AzurePrometheusMetric represents metric from Azure Monitor managed service for Prometheus
type AzurePrometheusMetric struct {
	PromQL string `json:"promql"`
}

var azurePrometheusValidation = govy.New[AzurePrometheusMetric](
	govy.For(func(p AzurePrometheusMetric) string { return p.PromQL }).
		WithName("promql").
		Required().
		Rules(rules.StringNotEmpty()),
)
