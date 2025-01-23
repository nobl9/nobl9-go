package slo

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// HoneycombMetric represents metric from Honeycomb.
type HoneycombMetric struct {
	Attribute string `json:"attribute"`
}

var honeycombSingleQueryValidation = govy.New[HoneycombMetric](
	govy.For(func(h HoneycombMetric) string { return h.Attribute }).
		WithName("attribute").
		Required().
		Rules(
			rules.StringMaxLength(255),
			rules.StringNotEmpty()),
)
