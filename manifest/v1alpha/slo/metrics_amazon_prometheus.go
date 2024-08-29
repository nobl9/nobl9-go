package slo

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// AmazonPrometheusMetric represents metric from Amazon Managed Prometheus
type AmazonPrometheusMetric struct {
	PromQL *string `json:"promql"`
}

var amazonPrometheusValidation = govy.New(
	govy.ForPointer(func(p AmazonPrometheusMetric) *string { return p.PromQL }).
		WithName("promql").
		Required().
		Rules(rules.StringNotEmpty()),
)
