package slo

import (
	"strings"
	"testing"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestRedshift_CountMetrics(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.Redshift)
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
	t.Run("region must be the same for good and total", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.Redshift)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.Redshift.Region = ptr("region-1")
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.Redshift.Region = ptr("region-2")
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].countMetrics",
			Code: validation.ErrorCodeEqualTo,
		})
	})
	t.Run("clusterId must be the same for good and total", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.Redshift)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.Redshift.ClusterID = ptr("1")
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.Redshift.ClusterID = ptr("2")
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].countMetrics",
			Code: validation.ErrorCodeEqualTo,
		})
	})
	t.Run("databaseName must be the same for good and total", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.Redshift)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.Redshift.DatabaseName = ptr("dev-db")
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.Redshift.DatabaseName = ptr("prod-db")
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].countMetrics",
			Code: validation.ErrorCodeEqualTo,
		})
	})
}

func TestRedshift(t *testing.T) {
	t.Run("required fields", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Redshift)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Redshift = &RedshiftMetric{}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 4,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.redshift.region",
				Code: validation.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.redshift.clusterId",
				Code: validation.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.redshift.databaseName",
				Code: validation.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.redshift.query",
				Code: validation.ErrorCodeRequired,
			},
		)
	})
	t.Run("invalid region", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Redshift)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Redshift.Region = ptr(strings.Repeat("a", 256))
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.redshift.region",
			Code: validation.ErrorCodeStringMaxLength,
		})
	})
	//nolint: lll
	t.Run("invalid query", func(t *testing.T) {
		for expectedDetails, query := range map[string]string{
			"must contain 'n9date' column":         "SELECT value as n9value FROM sinusoid WHERE timestamp BETWEEN :n9date_from AND :n9date_to",
			"must contain 'n9value' column":        "SELECT timestamp as n9date FROM sinusoid WHERE timestamp BETWEEN :n9date_from AND :n9date_to",
			"must filter by ':n9date_from' column": "SELECT value as n9value, timestamp as n9date FROM sinusoid WHERE timestamp = :n9date_to",
			"must filter by ':n9date_to' column":   "SELECT value as n9value, timestamp as n9date FROM sinusoid WHERE timestamp = :n9date_from",
		} {
			slo := validRawMetricSLO(v1alpha.Redshift)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Redshift.Query = ptr(query)
			err := validate(slo)
			testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
				Prop:            "spec.objectives[0].rawMetric.query.redshift.query",
				ContainsMessage: expectedDetails,
			})
		}
	})
}
