package slo

import (
	"regexp"

	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

// RedshiftMetric represents metric from Redshift.
type RedshiftMetric struct {
	Region       *string `json:"region"`
	ClusterID    *string `json:"clusterId"`
	DatabaseName *string `json:"databaseName"`
	Query        *string `json:"query"`
}

var redshiftCountMetricsLevelValidation = validation.New[CountMetricsSpec](
	validation.For(validation.GetSelf[CountMetricsSpec]()).
		Rules(
			validation.NewSingleRule(func(c CountMetricsSpec) error {
				good := c.GoodMetric
				total := c.TotalMetric

				if !arePointerValuesEqual(good.Redshift.Region, total.Redshift.Region) {
					return countMetricsPropertyEqualityError("redshift.region", goodMetric)
				}
				if !arePointerValuesEqual(good.Redshift.ClusterID, total.Redshift.ClusterID) {
					return countMetricsPropertyEqualityError("redshift.clusterId", goodMetric)
				}
				if !arePointerValuesEqual(good.Redshift.DatabaseName, total.Redshift.DatabaseName) {
					return countMetricsPropertyEqualityError("redshift.databaseName", goodMetric)
				}
				return nil
			}).WithErrorCode(validation.ErrorCodeEqualTo)),
).When(whenCountMetricsIs(v1alpha.Redshift))

var redshiftValidation = validation.New[RedshiftMetric](
	validation.ForPointer(func(r RedshiftMetric) *string { return r.Region }).
		WithName("region").
		Required().
		Rules(validation.StringMaxLength(255)),
	validation.ForPointer(func(r RedshiftMetric) *string { return r.ClusterID }).
		WithName("clusterId").
		Required(),
	validation.ForPointer(func(r RedshiftMetric) *string { return r.DatabaseName }).
		WithName("databaseName").
		Required(),
	validation.ForPointer(func(r RedshiftMetric) *string { return r.Query }).
		WithName("query").
		Required().
		Rules(
			validation.StringMatchRegexp(regexp.MustCompile(`^SELECT[\s\S]*\bn9date\b[\s\S]*FROM`)).
				WithDetails("must contain 'n9date' column"),
			validation.StringMatchRegexp(regexp.MustCompile(`^SELECT\s[\s\S]*\bn9value\b[\s\S]*\sFROM`)).
				WithDetails("must contain 'n9value' column"),
			validation.StringMatchRegexp(regexp.MustCompile(`WHERE[\s\S]*\W:n9date_from\b[\s\S]*`)).
				WithDetails("must filter by ':n9date_from' column"),
			validation.StringMatchRegexp(regexp.MustCompile(`WHERE[\s\S]*\W:n9date_to\b[\s\S]*`)).
				WithDetails("must filter by ':n9date_to' column")),
)
