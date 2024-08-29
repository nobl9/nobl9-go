package slo

import (
	"regexp"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

// RedshiftMetric represents metric from Redshift.
type RedshiftMetric struct {
	Region       *string `json:"region"`
	ClusterID    *string `json:"clusterId"`
	DatabaseName *string `json:"databaseName"`
	Query        *string `json:"query"`
}

var redshiftCountMetricsLevelValidation = govy.New(
	govy.For(govy.GetSelf[CountMetricsSpec]()).
		Rules(
			govy.NewRule(func(c CountMetricsSpec) error {
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
			}).WithErrorCode(rules.ErrorCodeEqualTo)),
).When(
	whenCountMetricsIs(v1alpha.Redshift),
	govy.WhenDescription("countMetrics is redshift"),
)

var redshiftValidation = govy.New(
	govy.ForPointer(func(r RedshiftMetric) *string { return r.Region }).
		WithName("region").
		Required().
		Rules(rules.StringMaxLength(255)),
	govy.ForPointer(func(r RedshiftMetric) *string { return r.ClusterID }).
		WithName("clusterId").
		Required(),
	govy.ForPointer(func(r RedshiftMetric) *string { return r.DatabaseName }).
		WithName("databaseName").
		Required(),
	govy.ForPointer(func(r RedshiftMetric) *string { return r.Query }).
		WithName("query").
		Required().
		Rules(
			rules.StringMatchRegexp(regexp.MustCompile(`^SELECT[\s\S]*\bn9date\b[\s\S]*FROM`)).
				WithDetails("must contain 'n9date' column"),
			rules.StringMatchRegexp(regexp.MustCompile(`^SELECT\s[\s\S]*\bn9value\b[\s\S]*\sFROM`)).
				WithDetails("must contain 'n9value' column"),
			rules.StringMatchRegexp(regexp.MustCompile(`WHERE[\s\S]*\W:n9date_from\b[\s\S]*`)).
				WithDetails("must filter by ':n9date_from' column"),
			rules.StringMatchRegexp(regexp.MustCompile(`WHERE[\s\S]*\W:n9date_to\b[\s\S]*`)).
				WithDetails("must filter by ':n9date_to' column")),
)
