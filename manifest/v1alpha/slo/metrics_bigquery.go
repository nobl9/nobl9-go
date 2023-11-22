package slo

import (
	"regexp"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

// BigQueryMetric represents metric from BigQuery
type BigQueryMetric struct {
	Query     string `json:"query"`
	ProjectID string `json:"projectId"`
	Location  string `json:"location"`
}

var bigQueryCountMetricsLevelValidation = validation.New[CountMetricsSpec](
	validation.For(validation.GetSelf[CountMetricsSpec]()).
		Rules(
			validation.NewSingleRule(func(c CountMetricsSpec) error {
				good := c.GoodMetric
				total := c.TotalMetric

				if good.BigQuery.ProjectID != total.BigQuery.ProjectID {
					return countMetricsPropertyEqualityError("bigQuery.projectId", goodMetric)
				}
				if good.BigQuery.Location != total.BigQuery.Location {
					return countMetricsPropertyEqualityError("bigQuery.location", goodMetric)
				}
				return nil
			}).WithErrorCode(validation.ErrorCodeEqualTo)),
).When(whenCountMetricsIs(v1alpha.BigQuery))

var bigQueryValidation = validation.New[BigQueryMetric](
	validation.For(func(b BigQueryMetric) string { return b.ProjectID }).
		WithName("projectId").
		Required().
		Rules(validation.StringMaxLength(255)),
	validation.For(func(b BigQueryMetric) string { return b.Location }).
		WithName("location").
		Required(),
	validation.For(func(b BigQueryMetric) string { return b.Query }).
		WithName("query").
		Required().
		Rules(
			validation.StringMatchRegexp(regexp.MustCompile(`\bn9date\b`)).
				WithDetails("must contain 'n9date'"),
			validation.StringMatchRegexp(regexp.MustCompile(`\bn9value\b`)).
				WithDetails("must contain 'n9value'"),
			validation.StringMatchRegexp(regexp.MustCompile(`DATETIME\(\s*@n9date_from\s*\)`)).
				WithDetails("must have DATETIME placeholder with '@n9date_from'"),
			validation.StringMatchRegexp(regexp.MustCompile(`DATETIME\(\s*@n9date_to\s*\)`)).
				WithDetails("must have DATETIME placeholder with '@n9date_to'")),
)
