package slo

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// HoneycombMetric represents metric from Honeycomb.
type HoneycombMetric struct {
	// Deprecated: Once Honeycomb good/bad over total and raw metrics support will be discontinued,
	// this property will be removed.
	Calculation string `json:"calculation,omitempty"`
	Attribute   string `json:"attribute"`
}

var honeycombSingleQueryValidation = govy.New[HoneycombMetric](
	govy.For(func(h HoneycombMetric) string { return h.Calculation }).
		WithName("calculation").
		Rules(rules.Forbidden[string]()),
	govy.For(func(h HoneycombMetric) string { return h.Attribute }).
		WithName("attribute").
		Required().
		Rules(
			rules.StringMaxLength(255),
			rules.StringNotEmpty()),
)

// Deprecated: Honeycomb support for good/bad over total and raw metrics will no longer be supported in the future.
var honeycombLegacyValidation = govy.New[HoneycombMetric](
	govy.For(func(h HoneycombMetric) string { return h.Calculation }).
		WithName("calculation").
		Required().
		Rules(rules.OneOf(supportedHoneycombCalculationTypes...)),
	govy.For(func(h HoneycombMetric) string { return h.Attribute }).
		WithName("attribute").
		Required().
		Rules(
			rules.StringMaxLength(255),
			rules.StringNotEmpty()),
)

var supportedHoneycombCalculationTypes = []string{
	"CONCURRENCY", "COUNT", "SUM", "AVG", "COUNT_DISTINCT", "MAX", "MIN",
	"P001", "P01", "P05", "P10", "P25", "P50", "P75", "P90", "P95", "P99", "P999",
	"RATE_AVG", "RATE_SUM", "RATE_MAX",
}
