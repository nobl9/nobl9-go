package slo

import (
	"testing"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestElasticsearch(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Elasticsearch)
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
	t.Run("required", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Elasticsearch)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Elasticsearch = &ElasticsearchMetric{
			Index: nil,
			Query: nil,
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 2,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.elasticsearch.index",
				Code: validation.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.elasticsearch.query",
				Code: validation.ErrorCodeRequired,
			},
		)
	})
	t.Run("empty", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Elasticsearch)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Elasticsearch = &ElasticsearchMetric{
			Index: ptr(""),
			Query: ptr(""),
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 2,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.elasticsearch.index",
				Code: validation.ErrorCodeStringNotEmpty,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.elasticsearch.query",
				Code: validation.ErrorCodeStringNotEmpty,
			},
		)
	})
	t.Run("invalid query", func(t *testing.T) {
		for _, query := range []string{
			"invalid",
			"{{.EndTime}} got that",
			"{{.BeginTime}} got that",
		} {
			slo := validRawMetricSLO(v1alpha.Elasticsearch)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Elasticsearch = &ElasticsearchMetric{
				Index: ptr("index"),
				Query: ptr(query),
			}
			err := validate(slo)
			testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.elasticsearch.query",
				Code: validation.ErrorCodeStringContains,
			})
		}
	})
}
