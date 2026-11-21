package slo

import (
	"regexp"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// Column names a ClickHouse query must select and the parameters it must bind
// for the evaluation window.
const (
	ClickHouseDateColumn        = "n9date"
	ClickHouseValueColumn       = "n9value"
	ClickHouseDateFromParameter = "n9date_from"
	ClickHouseDateToParameter   = "n9date_to"
)

// ClickHouseMetric defines a parameterized SQL query for a ClickHouse SLI.
type ClickHouseMetric struct {
	Query      string            `json:"query"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

// maxClickHouseParameters is a high, defensive ceiling on the number of
// query parameters; it blocks abuse without limiting realistic queries.
const maxClickHouseParameters = 100

var clickHouseValidation = govy.New[ClickHouseMetric](
	govy.For(func(c ClickHouseMetric) string { return c.Query }).
		WithName("query").
		Required().
		Rules(
			rules.StringMatchRegexp(regexp.MustCompile(`(?i)\bSELECT\b`)).
				WithDetails("must contain a SELECT statement"),
			rules.StringMatchRegexp(regexp.MustCompile(`\b`+ClickHouseDateColumn+`\b`)).
				WithDetails("must contain '"+ClickHouseDateColumn+"' column"),
			rules.StringMatchRegexp(regexp.MustCompile(`\b`+ClickHouseValueColumn+`\b`)).
				WithDetails("must contain '"+ClickHouseValueColumn+"' column"),
			rules.StringMatchRegexp(regexp.MustCompile(`\b`+ClickHouseDateFromParameter+`\b`)).
				WithDetails("must contain '"+ClickHouseDateFromParameter+"' placeholder"),
			rules.StringMatchRegexp(regexp.MustCompile(`\b`+ClickHouseDateToParameter+`\b`)).
				WithDetails("must contain '"+ClickHouseDateToParameter+"' placeholder"),
		),
	govy.For(func(c ClickHouseMetric) map[string]string { return c.Parameters }).
		WithName("parameters").
		Rules(rules.MapMaxLength[map[string]string](maxClickHouseParameters)),
)
