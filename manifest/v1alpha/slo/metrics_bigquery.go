package slo

import (
	"regexp"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

// BigQueryMetric represents metric from BigQuery
type BigQueryMetric struct {
	Query     string `json:"query"`
	ProjectID string `json:"projectId"`
	Location  string `json:"location"`
}

var bigQueryCountMetricsLevelValidation = govy.New(
	govy.For(govy.GetSelf[CountMetricsSpec]()).
		Rules(
			govy.NewRule(func(c CountMetricsSpec) error {
				good := c.GoodMetric
				total := c.TotalMetric

				if good.BigQuery.ProjectID != total.BigQuery.ProjectID {
					return countMetricsPropertyEqualityError("bigQuery.projectId", goodMetric)
				}
				if good.BigQuery.Location != total.BigQuery.Location {
					return countMetricsPropertyEqualityError("bigQuery.location", goodMetric)
				}
				return nil
			}).WithErrorCode(rules.ErrorCodeEqualTo)),
).When(
	whenCountMetricsIs(v1alpha.BigQuery),
	govy.WhenDescription("countMetrics is bigQuery"),
)

var bigQueryValidation = govy.New(
	govy.For(func(b BigQueryMetric) string { return b.ProjectID }).
		WithName("projectId").
		Required().
		Rules(rules.StringMaxLength(255)),
	govy.For(func(b BigQueryMetric) string { return b.Location }).
		WithName("location").
		Required(),
	govy.For(func(b BigQueryMetric) string { return b.Query }).
		WithName("query").
		Required().
		Rules(
			rules.StringMatchRegexp(regexp.MustCompile(`\bn9date\b`)).
				WithDetails("must contain 'n9date'"),
			rules.StringMatchRegexp(regexp.MustCompile(`\bn9value\b`)).
				WithDetails("must contain 'n9value'"),
			rules.StringMatchRegexp(regexp.MustCompile(`DATETIME\(\s*@n9date_from\s*\)`)).
				WithDetails("must have DATETIME placeholder with '@n9date_from'"),
			rules.StringMatchRegexp(regexp.MustCompile(`DATETIME\(\s*@n9date_to\s*\)`)).
				WithDetails("must have DATETIME placeholder with '@n9date_to'")),
)
