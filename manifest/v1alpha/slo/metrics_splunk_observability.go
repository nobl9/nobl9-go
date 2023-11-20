package slo

import "github.com/nobl9/nobl9-go/validation"

// SplunkObservabilityMetric represents metric from SplunkObservability
type SplunkObservabilityMetric struct {
	Program *string `json:"program"`
}

var splunkObservabilityValidation = validation.New[SplunkObservabilityMetric](
	validation.ForPointer(func(s SplunkObservabilityMetric) *string { return s.Program }).
		WithName("program").
		Required().
		Rules(validation.StringNotEmpty()),
)
