package slo

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// SplunkObservabilityMetric represents metric from SplunkObservability
type SplunkObservabilityMetric struct {
	Program *string `json:"program"`
}

var splunkObservabilityValidation = govy.New(
	govy.ForPointer(func(s SplunkObservabilityMetric) *string { return s.Program }).
		WithName("program").
		Required().
		Rules(rules.StringNotEmpty()),
)
