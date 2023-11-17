package slo

import (
	"github.com/nobl9/nobl9-go/validation"
)

// HoneycombMetric represents metric from Honeycomb.
type HoneycombMetric struct {
	Dataset     string          `json:"dataset"`
	Calculation string          `json:"calculation"`
	Attribute   string          `json:"attribute"`
	Filter      HoneycombFilter `json:"filter"`
}

// HoneycombFilter represents filter for Honeycomb metric. It has custom struct validation.
type HoneycombFilter struct {
	Operator   string                     `json:"op"`
	Conditions []HoneycombFilterCondition `json:"conditions"`
}

// HoneycombFilterCondition represents single condition for Honeycomb filter.
type HoneycombFilterCondition struct {
	Attribute string `json:"attribute"`
	Operator  string `json:"op"`
	Value     string `json:"value"`
}

var honeycombValidation = validation.New[HoneycombMetric](
	validation.For(func(h HoneycombMetric) string { return h.Dataset }).
		WithName("dataset").
		Required().
		Rules(
			validation.StringMaxLength(255),
			validation.StringNotEmpty(),
			validation.StringASCII()),
	validation.For(func(h HoneycombMetric) string { return h.Calculation }).
		WithName("calculation").
		Required().
		Rules(validation.OneOf(supportedHoneycombCalculationTypes...)),
	validation.For(func(h HoneycombMetric) string { return h.Attribute }).
		WithName("attribute").
		Required().
		Rules(
			validation.StringMaxLength(255),
			validation.StringNotEmpty(),
			validation.StringASCII()),
	validation.For(func(h HoneycombMetric) HoneycombFilter { return h.Filter }).
		WithName("filter").
		Omitempty().
		Include(honeycombFilterValidation),
)

var honeycombFilterValidation = validation.New[HoneycombFilter](
	validation.For(validation.GetSelf[HoneycombFilter]()).
		Rules(validation.NewSingleRule(func(h HoneycombFilter) error {
			if len(h.Conditions) > 1 && h.Operator == "" {
				return validation.NewPropertyError("op", h.Operator, validation.NewRequiredError())
			}
			return nil
		})),
	validation.For(func(h HoneycombFilter) string { return h.Operator }).
		WithName("op").
		Omitempty().
		Rules(validation.OneOf(supportedHoneycombFilterOperators...)).
		Include(),
	validation.ForEach(func(h HoneycombFilter) []HoneycombFilterCondition { return h.Conditions }).
		WithName("conditions").
		Rules(validation.SliceMaxLength[[]HoneycombFilterCondition](100)).
		// We don't want to spend too much time here if someone is spamming us with more than 100 conditions.
		StopOnError().
		Rules(validation.SliceUnique(validation.SelfHashFunc[HoneycombFilterCondition]())).
		IncludeForEach(honeycombFilterConditionValidation),
)

var honeycombFilterConditionValidation = validation.New[HoneycombFilterCondition](
	validation.For(func(h HoneycombFilterCondition) string { return h.Attribute }).
		WithName("attribute").
		Required().
		Rules(
			validation.StringMaxLength(255),
			validation.StringNotEmpty(),
			validation.StringASCII()),
	validation.For(func(h HoneycombFilterCondition) string { return h.Operator }).
		WithName("op").
		Required().
		Rules(validation.OneOf(supportedHoneycombFilterConditionOperators...)),
	validation.For(func(h HoneycombFilterCondition) string { return h.Value }).
		WithName("value").
		Rules(
			validation.StringMaxLength(255),
			validation.StringASCII()),
)

var supportedHoneycombFilterOperators = []string{"AND", "OR"}

var supportedHoneycombCalculationTypes = []string{
	"COUNT", "SUM", "AVG", "COUNT_DISTINCT", "MAX", "MIN",
	"P001", "P01", "P05", "P10", "P25", "P50", "P75", "P90", "P95", "P99", "P999",
	"RATE_AVG", "RATE_SUM", "RATE_MAX",
}

var supportedHoneycombFilterConditionOperators = []string{
	"=", "!=", ">", ">=", "<", "<=",
	"starts-with", "does-not-start-with", "exists", "does-not-exist",
	"contains", "does-not-contain", "in", "not-in",
}
