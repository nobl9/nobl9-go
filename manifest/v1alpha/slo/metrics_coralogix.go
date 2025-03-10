package slo

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// CoralogixMetric represents metric from Coralogix
type CoralogixMetric struct {
	PromQL *string `json:"promql"`
}

var coralogixValidation = govy.New[CoralogixMetric](
	govy.ForPointer(func(p CoralogixMetric) *string { return p.PromQL }).
		WithName("promql").
		Required().
		Rules(rules.StringNotEmpty()),
)
