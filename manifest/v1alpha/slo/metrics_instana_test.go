package slo

import (
	"testing"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestInstana_CountMetrics(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.Instana)
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
	t.Run("metricType must be the same for good and total", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.Instana)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.Instana.MetricType = instanaMetricTypeApplication
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.Instana.MetricType = instanaMetricTypeInfrastructure
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 2, testutils.ExpectedError{
			Prop: "spec.objectives[0].countMetrics",
			Code: rules.ErrorCodeEqualTo,
		})
	})
	t.Run("application metrics are not allowed", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.Instana)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.Instana = validInstanaApplicationMetric()
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.Instana = validInstanaApplicationMetric()
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 6,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].countMetrics.total.instana.application",
				Code: rules.ErrorCodeForbidden,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].countMetrics.total.instana.metricType",
				Code: rules.ErrorCodeEqualTo,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].countMetrics.total.instana.infrastructure",
				Code: rules.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].countMetrics.good.instana.application",
				Code: rules.ErrorCodeForbidden,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].countMetrics.good.instana.metricType",
				Code: rules.ErrorCodeEqualTo,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].countMetrics.good.instana.infrastructure",
				Code: rules.ErrorCodeRequired,
			},
		)
	})
	t.Run("metricType required", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.Instana)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.Instana.MetricType = ""
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.Instana.MetricType = ""
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 2,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].countMetrics.good.instana.metricType",
				Code: rules.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].countMetrics.good.instana.metricType",
				Code: rules.ErrorCodeRequired,
			},
		)
	})
}

func TestInstana_RawMetrics(t *testing.T) {
	t.Run("valid application metric", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Instana)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana = validInstanaApplicationMetric()
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
	t.Run("both application and infrastructure provided", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Instana)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana.Application = &InstanaApplicationMetricType{}
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana.Infrastructure = &InstanaInfrastructureMetricType{}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop:    "spec.objectives[0].rawMetric.query.instana",
			Message: "cannot use both 'instana.application' and 'instana.infrastructure'",
		})
	})
	t.Run("application missing for metricType", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Instana)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana.MetricType = instanaMetricTypeApplication
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana.Application = nil
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana.Infrastructure = nil
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop:    "spec.objectives[0].rawMetric.query.instana",
			Message: "when 'metricType' is 'application', 'instana.application' is required",
		})
	})
	t.Run("infrastructure missing for metricType", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Instana)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana.MetricType = instanaMetricTypeInfrastructure
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana.Application = nil
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana.Infrastructure = nil
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop:    "spec.objectives[0].rawMetric.query.instana",
			Message: "when 'metricType' is 'infrastructure', 'instana.infrastructure' is required",
		})
	})
	t.Run("invalid metricType", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Instana)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana.MetricType = "invalid"
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.instana.metricType",
			Code: rules.ErrorCodeOneOf,
		})
	})
	t.Run("metricType required", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Instana)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana.MetricType = ""
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.instana.metricType",
			Code: rules.ErrorCodeRequired,
		})
	})
}

func TestInstana_Infrastructure(t *testing.T) {
	t.Run("required fields", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Instana)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana = &InstanaMetric{
			MetricType:     instanaMetricTypeInfrastructure,
			Infrastructure: &InstanaInfrastructureMetricType{},
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 3,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.instana.infrastructure.metricRetrievalMethod",
				Code: rules.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.instana.infrastructure.metricId",
				Code: rules.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.instana.infrastructure.pluginId",
				Code: rules.ErrorCodeRequired,
			},
		)
	})
	t.Run("invalid metricRetrievalMethod", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Instana)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana.Infrastructure.MetricRetrievalMethod = "invalid"
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.instana.infrastructure.metricRetrievalMethod",
			Code: rules.ErrorCodeOneOf,
		})
	})
	t.Run("required query retrieval method", func(t *testing.T) {
		for name, test := range map[string]struct {
			Method       string
			Query        *string
			SnapshotID   *string
			ErrorMessage string
		}{
			"required query": {
				Method:       instanaMetricRetrievalMethodQuery,
				Query:        nil,
				SnapshotID:   nil,
				ErrorMessage: "when 'metricRetrievalMethod' is 'query', 'query' property must be provided",
			},
			"forbidden snapshot": {
				Method:       instanaMetricRetrievalMethodQuery,
				Query:        ptr("query"),
				SnapshotID:   ptr("123"),
				ErrorMessage: "when 'metricRetrievalMethod' is 'query', 'snapshotId' property is not allowed",
			},
			"required snapshot": {
				Method:       instanaMetricRetrievalMethodSnapshot,
				Query:        nil,
				SnapshotID:   nil,
				ErrorMessage: "when 'metricRetrievalMethod' is 'snapshot', 'snapshotId' property must be provided",
			},
			"forbidden query": {
				Method:       instanaMetricRetrievalMethodSnapshot,
				Query:        ptr("query"),
				SnapshotID:   ptr("123"),
				ErrorMessage: "when 'metricRetrievalMethod' is 'snapshot', 'query' property is not allowed",
			},
		} {
			t.Run(name, func(t *testing.T) {
				slo := validRawMetricSLO(v1alpha.Instana)
				slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana.Infrastructure.MetricRetrievalMethod = test.Method
				slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana.Infrastructure.Query = test.Query
				slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana.Infrastructure.SnapshotID = test.SnapshotID
				err := validate(slo)
				testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
					Prop:    "spec.objectives[0].rawMetric.query.instana.infrastructure",
					Message: test.ErrorMessage,
				})
			})
		}
	})
}

func TestInstana_Application(t *testing.T) {
	t.Run("required fields", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Instana)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana = &InstanaMetric{
			MetricType:     instanaMetricTypeApplication,
			Infrastructure: nil,
			Application:    &InstanaApplicationMetricType{},
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 4,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.instana.application.metricId",
				Code: rules.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.instana.application.aggregation",
				Code: rules.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.instana.application.groupBy",
				Code: rules.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.instana.application.apiQuery",
				Code: rules.ErrorCodeRequired,
			},
		)
	})
	t.Run("valid metricId", func(t *testing.T) {
		for metricID, aggregation := range map[string]string{
			"calls":          "sum",
			"erroneousCalls": "sum",
			"errors":         "mean",
			"latency":        "sum",
		} {
			slo := validRawMetricSLO(v1alpha.Instana)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana = validInstanaApplicationMetric()
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana.Application.MetricID = metricID
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana.Application.Aggregation = aggregation
			err := validate(slo)
			testutils.AssertNoError(t, slo, err)
		}
	})
	t.Run("invalid metricId", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Instana)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana = validInstanaApplicationMetric()
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana.Application.MetricID = "invalid"
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.instana.application.metricId",
			Code: rules.ErrorCodeOneOf,
		})
	})
	t.Run("invalid apiQuery", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Instana)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana = validInstanaApplicationMetric()
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana.Application.APIQuery = "{]}"
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.instana.application.apiQuery",
			Code: rules.ErrorCodeStringJSON,
		})
	})
	t.Run("missing fields for groupBy", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Instana)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana = validInstanaApplicationMetric()
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana.Application.GroupBy = InstanaApplicationMetricGroupBy{
			TagSecondLevelKey: ptr(""),
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 2,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.instana.application.groupBy.tag",
				Code: rules.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.instana.application.groupBy.tagEntity",
				Code: rules.ErrorCodeRequired,
			},
		)
	})
	t.Run("valid tagEntity", func(t *testing.T) {
		for _, tagEntity := range []string{
			"DESTINATION",
			"SOURCE",
			"NOT_APPLICABLE",
		} {
			slo := validRawMetricSLO(v1alpha.Instana)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana = validInstanaApplicationMetric()
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana.Application.GroupBy.TagEntity = tagEntity
			err := validate(slo)
			testutils.AssertNoError(t, slo, err)
		}
	})
	t.Run("invalid tagEntity", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Instana)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana = validInstanaApplicationMetric()
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana.Application.GroupBy.TagEntity = "invalid"
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.instana.application.groupBy.tagEntity",
			Code: rules.ErrorCodeOneOf,
		})
	})
	t.Run("metricId", func(t *testing.T) {
		for _, test := range []struct {
			MetricID    string
			Aggregation string
			IsValid     bool
		}{
			{
				MetricID:    "calls",
				Aggregation: "sum",
				IsValid:     true,
			},
			{
				MetricID:    "calls",
				Aggregation: "mean",
				IsValid:     false,
			},
			{
				MetricID:    "erroneousCalls",
				Aggregation: "sum",
				IsValid:     true,
			},
			{
				MetricID:    "erroneousCalls",
				Aggregation: "mean",
				IsValid:     false,
			},
			{
				MetricID:    "errors",
				Aggregation: "mean",
				IsValid:     true,
			},
			{
				MetricID:    "errors",
				Aggregation: "sum",
				IsValid:     false,
			},
		} {
			slo := validRawMetricSLO(v1alpha.Instana)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana = validInstanaApplicationMetric()
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana.Application.MetricID = test.MetricID
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana.Application.Aggregation = test.Aggregation
			err := validate(slo)
			if test.IsValid {
				testutils.AssertNoError(t, slo, err)
			} else {
				testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
					Prop: "spec.objectives[0].rawMetric.query.instana.application",
					Code: rules.ErrorCodeEqualTo,
				})
			}
		}
	})
	t.Run("metricId - valid latency", func(t *testing.T) {
		for _, agg := range validInstanaLatencyAggregations {
			slo := validRawMetricSLO(v1alpha.Instana)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana = validInstanaApplicationMetric()
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana.Application.MetricID = "latency"
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana.Application.Aggregation = agg
			err := validate(slo)
			testutils.AssertNoError(t, slo, err)
		}
	})
	t.Run("metricId - invalid latency", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Instana)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana = validInstanaApplicationMetric()
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana.Application.MetricID = "latency"
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Instana.Application.Aggregation = "invalid"
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.instana.application",
			Code: rules.ErrorCodeOneOf,
		})
	})
}

func validInstanaApplicationMetric() *InstanaMetric {
	return &InstanaMetric{
		MetricType: instanaMetricTypeApplication,
		Application: &InstanaApplicationMetricType{
			MetricID:    "latency",
			Aggregation: "p99",
			GroupBy: InstanaApplicationMetricGroupBy{
				Tag:       "endpoint.name",
				TagEntity: "DESTINATION",
			},
			APIQuery: `
{
  "type": "EXPRESSION",
  "logicalOperator": "AND",
  "elements": [
    {
      "type": "TAG_FILTER",
      "name": "service.name",
      "operator": "EQUALS",
      "entity": "DESTINATION",
      "value": "master"
    },
    {
      "type": "TAG_FILTER",
      "name": "call.type",
      "operator": "EQUALS",
      "entity": "NOT_APPLICABLE",
      "value": "HTTP"
    }
  ]
}
`,
		},
	}
}
