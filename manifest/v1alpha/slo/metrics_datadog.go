package slo

import "github.com/nobl9/nobl9-go/validation"

// DatadogMetric represents metric from Datadog
type DatadogMetric struct {
	Query *string `json:"query" validate:"required"`
}

var datadogValidation = validation.New[DatadogMetric](
	validation.ForPointer(func(d DatadogMetric) *string { return d.Query }).
		WithName("query").
		Required().
		Rules(validation.StringNotEmpty()),
)
