package slo

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// DatadogMetric represents metric from Datadog
type DatadogMetric struct {
	Query *string `json:"query"`
}

var datadogValidation = govy.New[DatadogMetric](
	govy.ForPointer(func(d DatadogMetric) *string { return d.Query }).
		WithName("query").
		Required().
		Rules(rules.StringNotEmpty()),
)
