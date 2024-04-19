package slo

import (
	"strings"

	"golang.org/x/exp/slices"

	"github.com/nobl9/nobl9-go/internal/validation"
)

// HoneycombMetric represents metric from Honeycomb.
type HoneycombMetric struct {
	Calculation string `json:"calculation"`
	Attribute   string `json:"attribute,omitempty"`
}

var honeycombValidation = validation.New[HoneycombMetric](
	validation.For(func(h HoneycombMetric) string { return h.Calculation }).
		WithName("calculation").
		Required().
		Rules(validation.OneOf(supportedHoneycombCalculationTypes...)),
)

var supportedHoneycombCalculationTypes = []string{
	"CONCURRENCY", "COUNT", "SUM", "AVG", "COUNT_DISTINCT", "MAX", "MIN",
	"P001", "P01", "P05", "P10", "P25", "P50", "P75", "P90", "P95", "P99", "P999",
	"RATE_AVG", "RATE_SUM", "RATE_MAX",
}

var attributeRequired = validation.New[HoneycombMetric](
	validation.For(func(h HoneycombMetric) string { return h.Attribute }).
		WithName("attribute").
		Required().
		Rules(
			validation.StringMaxLength(255),
			validation.StringNotEmpty()),
).When(
	func(h HoneycombMetric) bool {
		return slices.Contains([]string{
			"SUM", "AVG", "CONCURRENCY", "COUNT", "COUNT_DISTINCT", "MAX", "MIN",
			"P001", "P01", "P05", "P10", "P25", "P50", "P75", "P90", "P95", "P99", "P999",
			"RATE_AVG", "RATE_SUM", "RATE_MAX",
		}, h.Calculation)
	},
	validation.WhenDescription("calculation is one of: %s",
		strings.Join(supportedHoneycombCalculationTypes, ", ")),
)
