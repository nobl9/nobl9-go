package slo

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// PrometheusMetric represents metric from Prometheus
type PrometheusMetric struct {
	PromQL *string `json:"promql"`
}

var prometheusValidation = govy.New[PrometheusMetric](
	govy.ForPointer(func(p PrometheusMetric) *string { return p.PromQL }).
		WithName("promql").
		Required().
		Rules(rules.StringNotEmpty()),
)
