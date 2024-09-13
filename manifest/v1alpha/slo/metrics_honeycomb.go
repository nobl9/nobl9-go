package slo

import (
	"strings"

	"golang.org/x/exp/slices"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// HoneycombMetric represents metric from Honeycomb.
type HoneycombMetric struct {
	Calculation string `json:"calculation"`
	Attribute   string `json:"attribute,omitempty"`
}

var honeycombValidation = govy.New[HoneycombMetric](
	govy.For(func(h HoneycombMetric) string { return h.Calculation }).
		WithName("calculation").
		Required().
		Rules(rules.OneOf(supportedHoneycombCalculationTypes...)),
)

var supportedHoneycombCalculationTypes = []string{
	"CONCURRENCY", "COUNT", "SUM", "AVG", "COUNT_DISTINCT", "MAX", "MIN",
	"P001", "P01", "P05", "P10", "P25", "P50", "P75", "P90", "P95", "P99", "P999",
	"RATE_AVG", "RATE_SUM", "RATE_MAX",
}

var attributeRequired = govy.New[HoneycombMetric](
	govy.For(func(h HoneycombMetric) string { return h.Attribute }).
		WithName("attribute").
		Required().
		Rules(
			rules.StringMaxLength(255),
			rules.StringNotEmpty()),
).When(
	func(h HoneycombMetric) bool {
		return slices.Contains([]string{
			"SUM", "AVG", "CONCURRENCY", "COUNT", "COUNT_DISTINCT", "MAX", "MIN",
			"P001", "P01", "P05", "P10", "P25", "P50", "P75", "P90", "P95", "P99", "P999",
			"RATE_AVG", "RATE_SUM", "RATE_MAX",
		}, h.Calculation)
	},
	govy.WhenDescription("calculation is one of: %s",
		strings.Join(supportedHoneycombCalculationTypes, ", ")),
)
