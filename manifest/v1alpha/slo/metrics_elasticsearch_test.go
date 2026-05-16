package slo

import (
	"testing"

	"github.com/nobl9/govy/pkg/jsonpath"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
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
				Prop: jsonpath.New().
					Name("spec").
					Name("objectives").
					Index(0).
					Name("rawMetric").
					Name("query").
					Name("elasticsearch").
					Name("index"),
				Code: rules.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: jsonpath.New().
					Name("spec").
					Name("objectives").
					Index(0).
					Name("rawMetric").
					Name("query").
					Name("elasticsearch").
					Name("query"),
				Code: rules.ErrorCodeRequired,
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
				Prop: jsonpath.New().
					Name("spec").
					Name("objectives").
					Index(0).
					Name("rawMetric").
					Name("query").
					Name("elasticsearch").
					Name("index"),
				Code: rules.ErrorCodeStringNotEmpty,
			},
			testutils.ExpectedError{
				Prop: jsonpath.New().
					Name("spec").
					Name("objectives").
					Index(0).
					Name("rawMetric").
					Name("query").
					Name("elasticsearch").
					Name("query"),
				Code: rules.ErrorCodeStringNotEmpty,
			},
		)
	})
	t.Run("invalid query", func(t *testing.T) {
		for _, query := range []string{
			"invalid",
			"{{.EndTime}} got that",
			"{{.BeginTime}} got that",
			"{{.BeginTimeInMilliseconds}} {{.EndTime}} got that",
			"{{.BeginTime}} {{.EndTimeInMilliseconds}} got that",
			"{{.EndTimeInMilliseconds}} got that",
			"{{.BeginTimeInMilliseconds}} got that",
		} {
			slo := validRawMetricSLO(v1alpha.Elasticsearch)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Elasticsearch = &ElasticsearchMetric{
				Index: ptr("index"),
				Query: ptr(query),
			}
			err := validate(slo)
			testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
				Prop: jsonpath.New().
					Name("spec").
					Name("objectives").
					Index(0).
					Name("rawMetric").
					Name("query").
					Name("elasticsearch").
					Name("query"),
				Code: rules.ErrorCodeStringContains,
			})
		}
	})
}
