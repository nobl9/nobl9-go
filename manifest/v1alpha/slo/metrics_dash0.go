package slo

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// Dash0Metric represents metric from Dash0.
type Dash0Metric struct {
	PromQL *string `json:"promql"`
}

var dash0Validation = govy.New[Dash0Metric](
	govy.ForPointer(func(d Dash0Metric) *string { return d.PromQL }).
		WithName("promql").
		Required().
		Rules(rules.StringNotEmpty()),
)
