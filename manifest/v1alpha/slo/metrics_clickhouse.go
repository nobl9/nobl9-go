package slo

import (
	"regexp"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// ClickHouseMetric represents metric from ClickHouse.
type ClickHouseMetric struct {
	Query      string            `json:"query"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

var clickHouseValidation = govy.New[ClickHouseMetric](
	govy.For(func(c ClickHouseMetric) string { return c.Query }).
		WithName("query").
		Required().
		Rules(
			rules.StringMatchRegexp(regexp.MustCompile(`(?i)\bSELECT\b`)).
				WithDetails("must contain a SELECT statement"),
			rules.StringMatchRegexp(regexp.MustCompile(`\bn9date\b`)).
				WithDetails("must contain 'n9date' column"),
			rules.StringMatchRegexp(regexp.MustCompile(`\bn9value\b`)).
				WithDetails("must contain 'n9value' column"),
			rules.StringMatchRegexp(regexp.MustCompile(`\bn9date_from\b`)).
				WithDetails("must contain 'n9date_from' placeholder"),
			rules.StringMatchRegexp(regexp.MustCompile(`\bn9date_to\b`)).
				WithDetails("must contain 'n9date_to' placeholder"),
		),
)
