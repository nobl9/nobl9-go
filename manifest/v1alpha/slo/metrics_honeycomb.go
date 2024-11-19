package slo

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// HoneycombMetric represents metric from Honeycomb.
type HoneycombMetric struct {
	Attribute string `json:"attribute"`
}

var honeycombValidation = govy.New[HoneycombMetric](
	govy.For(func(h HoneycombMetric) string { return h.Attribute }).
		WithName("attribute").
		Required().
		Rules(
			rules.StringMaxLength(255),
			rules.StringNotEmpty()),
)

var honeycombCountMetricsValidation = govy.New[MetricSpec](
	govy.ForPointer(func(m MetricSpec) *HoneycombMetric { return m.Honeycomb }).
		WithName("honeycomb").
		Rules(rules.Forbidden[HoneycombMetric]()),
)

var honeycombRawMetricValidation = govy.New[MetricSpec](
	govy.ForPointer(func(m MetricSpec) *HoneycombMetric { return m.Honeycomb }).
		WithName("honeycomb").
		Rules(rules.Forbidden[HoneycombMetric]()),
)
