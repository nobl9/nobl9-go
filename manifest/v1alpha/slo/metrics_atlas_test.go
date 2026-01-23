package slo

import (
	"testing"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestAtlas_rawMetric(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Atlas)
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
	t.Run("required promql", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Atlas)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Atlas.PromQL = ""
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.atlas.promql",
			Code: rules.ErrorCodeRequired,
		})
	})
	t.Run("required dataReplay", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Atlas)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Atlas.DataReplay = nil
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.atlas.dataReplay",
			Code: rules.ErrorCodeRequired,
		})
	})
	t.Run("required dataReplay.parameters", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Atlas)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Atlas.DataReplay.Parameters = nil
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.atlas.dataReplay.parameters",
			Code: rules.ErrorCodeRequired,
		})
	})
	t.Run("dataReplay.parameters must have at least one element", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Atlas)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Atlas.DataReplay.Parameters = map[string]string{}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.atlas.dataReplay.parameters",
			Code: rules.ErrorCodeMapMinLength,
		})
	})
	t.Run("goodSeriesLabel forbidden for raw metrics", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Atlas)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Atlas.DataReplay.GoodSeriesLabel = "good"
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop:    "spec.objectives[0].rawMetric.query.atlas",
			Message: "goodSeriesLabel and totalSeriesLabel are forbidden for raw metrics",
		})
	})
	t.Run("totalSeriesLabel forbidden for raw metrics", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Atlas)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Atlas.DataReplay.TotalSeriesLabel = "total"
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop:    "spec.objectives[0].rawMetric.query.atlas",
			Message: "goodSeriesLabel and totalSeriesLabel are forbidden for raw metrics",
		})
	})
}

func TestAtlas_singleQueryGoodOverTotal(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validSingleQueryGoodOverTotalCountMetricSLO(v1alpha.Atlas)
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
	t.Run("required promql", func(t *testing.T) {
		slo := validSingleQueryGoodOverTotalCountMetricSLO(v1alpha.Atlas)
		slo.Spec.Objectives[0].CountMetrics.GoodTotalMetric.Atlas.PromQL = ""
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].countMetrics.goodTotal.atlas.promql",
			Code: rules.ErrorCodeRequired,
		})
	})
	t.Run("required dataReplay", func(t *testing.T) {
		slo := validSingleQueryGoodOverTotalCountMetricSLO(v1alpha.Atlas)
		slo.Spec.Objectives[0].CountMetrics.GoodTotalMetric.Atlas.DataReplay = nil
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].countMetrics.goodTotal.atlas.dataReplay",
			Code: rules.ErrorCodeRequired,
		})
	})
	t.Run("required dataReplay.goodSeriesLabel", func(t *testing.T) {
		slo := validSingleQueryGoodOverTotalCountMetricSLO(v1alpha.Atlas)
		slo.Spec.Objectives[0].CountMetrics.GoodTotalMetric.Atlas.DataReplay.GoodSeriesLabel = ""
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].countMetrics.goodTotal.atlas.dataReplay.goodSeriesLabel",
			Code: rules.ErrorCodeRequired,
		})
	})
	t.Run("required dataReplay.totalSeriesLabel", func(t *testing.T) {
		slo := validSingleQueryGoodOverTotalCountMetricSLO(v1alpha.Atlas)
		slo.Spec.Objectives[0].CountMetrics.GoodTotalMetric.Atlas.DataReplay.TotalSeriesLabel = ""
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].countMetrics.goodTotal.atlas.dataReplay.totalSeriesLabel",
			Code: rules.ErrorCodeRequired,
		})
	})
	t.Run("required dataReplay.parameters", func(t *testing.T) {
		slo := validSingleQueryGoodOverTotalCountMetricSLO(v1alpha.Atlas)
		slo.Spec.Objectives[0].CountMetrics.GoodTotalMetric.Atlas.DataReplay.Parameters = nil
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].countMetrics.goodTotal.atlas.dataReplay.parameters",
			Code: rules.ErrorCodeRequired,
		})
	})
	t.Run("dataReplay.parameters must have at least one element", func(t *testing.T) {
		slo := validSingleQueryGoodOverTotalCountMetricSLO(v1alpha.Atlas)
		slo.Spec.Objectives[0].CountMetrics.GoodTotalMetric.Atlas.DataReplay.Parameters = map[string]string{}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].countMetrics.goodTotal.atlas.dataReplay.parameters",
			Code: rules.ErrorCodeMapMinLength,
		})
	})
}

func TestAtlas_countMetrics_forbidden(t *testing.T) {
	slo := validCountMetricSLO(v1alpha.Atlas)
	err := validate(slo)
	testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
		Prop: "spec.objectives[0].countMetrics",
		Code: rules.ErrorCodeForbidden,
	})
}
