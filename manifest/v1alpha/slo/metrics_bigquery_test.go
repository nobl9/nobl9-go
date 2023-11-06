package slo

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

func TestBigQuery_CountMetrics(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.BigQuery)
		err := validate(slo)
		assert.Empty(t, err)
	})
	t.Run("projectId must be the same for good and total", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.BigQuery)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.BigQuery.ProjectID = "1"
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.BigQuery.ProjectID = "2"
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].countMetrics",
			Code: validation.ErrorCodeEqualTo,
		})
	})
	t.Run("location must be the same for good and total", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.BigQuery)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.BigQuery.Location = "1"
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.BigQuery.Location = "2"
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].countMetrics",
			Code: validation.ErrorCodeEqualTo,
		})
	})
}

func TestBigQuery(t *testing.T) {
	t.Run("required fields", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.BigQuery)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.BigQuery = &BigQueryMetric{}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 3,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.bigQuery.projectId",
				Code: validation.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.bigQuery.location",
				Code: validation.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.bigQuery.query",
				Code: validation.ErrorCodeRequired,
			},
		)
	})
	t.Run("invalid query", func(t *testing.T) {
		for expectedDetails, query := range map[string]string{
			"must contain 'n9date'": `
SELECT http_code AS n9value
FROM 'bdwtest-256112.metrics.http_response'
WHERE http_code = 200 AND created BETWEEN DATETIME(@n9date_from) AND DATETIME(@n9date_to)`,
			"must contain 'n9value'": `
SELECT created AS n9date
FROM 'bdwtest-256112.metrics.http_response'
WHERE http_code = 200 AND created BETWEEN DATETIME(@n9date_from) AND DATETIME(@n9date_to)`,
			"must have DATETIME placeholder with '@n9date_from'": `
SELECT http_code AS n9value, created AS n9date
FROM 'bdwtest-256112.metrics.http_response'
WHERE http_code = 200 AND created = DATETIME(@n9date_to)`,
			"must have DATETIME placeholder with '@n9date_to'": `
SELECT http_code AS n9value, created AS n9date
FROM 'bdwtest-256112.metrics.http_response'
WHERE http_code = 200 AND created = DATETIME(@n9date_from)`,
		} {
			slo := validRawMetricSLO(v1alpha.BigQuery)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.BigQuery.Query = query
			err := validate(slo)
			testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
				Prop:            "spec.objectives[0].rawMetric.query.bigQuery.query",
				ContainsMessage: expectedDetails,
			})
		}
	})
}
