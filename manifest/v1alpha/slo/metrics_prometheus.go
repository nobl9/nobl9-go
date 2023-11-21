package slo

import "github.com/nobl9/nobl9-go/validation"

// PrometheusMetric represents metric from Prometheus
type PrometheusMetric struct {
	PromQL *string `json:"promql"`
}

var prometheusValidation = validation.New[PrometheusMetric](
	validation.ForPointer(func(p PrometheusMetric) *string { return p.PromQL }).
		WithName("promql").
		Required().
		Rules(validation.StringNotEmpty()),
)
