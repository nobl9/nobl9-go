package slo

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// DynatraceMetric represents metric from Dynatrace.
type DynatraceMetric struct {
	MetricSelector *string `json:"metricSelector"`
}

var dynatraceValidation = govy.New(
	govy.ForPointer(func(d DynatraceMetric) *string { return d.MetricSelector }).
		WithName("metricSelector").
		Required().
		Rules(rules.StringNotEmpty()),
)
