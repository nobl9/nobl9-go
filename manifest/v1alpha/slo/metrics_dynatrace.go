package slo

import "github.com/nobl9/nobl9-go/validation"

// DynatraceMetric represents metric from Dynatrace.
type DynatraceMetric struct {
	MetricSelector *string `json:"metricSelector"`
}

var dynatraceValidation = validation.New[DynatraceMetric](
	validation.ForPointer(func(d DynatraceMetric) *string { return d.MetricSelector }).
		WithName("metricSelector").
		Required().
		Rules(validation.StringNotEmpty()),
)
