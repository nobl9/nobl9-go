package slo

import (
	"testing"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

func TestSumoLogic_CountMetricsLevel(t *testing.T) {
	t.Run("quantization must be the same for good and total", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.SumoLogic)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.SumoLogic = &SumoLogicMetric{
			Type:         ptr(sumoLogicTypeMetric),
			Query:        ptr("kube_node_status_condition | min"),
			Quantization: ptr("20s"),
			Rollup:       ptr("None"),
		}
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.SumoLogic = &SumoLogicMetric{
			Type:         ptr(sumoLogicTypeMetric),
			Query:        ptr("kube_node_status_condition | min"),
			Quantization: ptr("25s"),
			Rollup:       ptr("None"),
		}
		err := validate(slo)
		assertContainsErrors(t, err, 1, expectedError{
			Prop: "spec.objectives[0].countMetrics",
			Code: validation.ErrorCodeEqualTo,
		})
	})
	t.Run("query timeslice duration must be the same for good and total", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.SumoLogic)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.SumoLogic = &SumoLogicMetric{
			Type: ptr(sumoLogicTypeLogs),
			Query: ptr(`
_collector="n9-dev-tooling-cluster" _source="logs"
  | json "log"
  | timeslice 20s as n9_time
  | parse "level=* *" as (log_level, tail)
  | if (log_level matches "error" ,0,1) as log_level_not_error
  | sum(log_level_not_error) as n9_value by n9_time
  | sort by n9_time asc`),
		}
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.SumoLogic = &SumoLogicMetric{
			Type: ptr(sumoLogicTypeLogs),
			Query: ptr(`
_collector="n9-dev-tooling-cluster" _source="logs"
  | json "log"
  | timeslice 25s as n9_time
  | parse "level=* *" as (log_level, tail)
  | if (log_level matches "error" ,0,1) as log_level_not_error
  | sum(log_level_not_error) as n9_value by n9_time
  | sort by n9_time asc`),
		}
		err := validate(slo)
		assertContainsErrors(t, err, 1, expectedError{
			Prop: "spec.objectives[0].countMetrics",
			Code: validation.ErrorCodeEqualTo,
		})
	})
}

func TestSumoLogic(t *testing.T) {
	t.Run("missing type", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.SumoLogic)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.SumoLogic.Type = nil
		err := validate(slo)
		assertContainsErrors(t, err, 1, expectedError{
			Prop: "spec.objectives[0].rawMetric.query.sumoLogic.type",
			Code: validation.ErrorCodeRequired,
		})
	})
	t.Run("invalid type", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.SumoLogic)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.SumoLogic.Type = ptr("invalid")
		err := validate(slo)
		assertContainsErrors(t, err, 1, expectedError{
			Prop: "spec.objectives[0].rawMetric.query.sumoLogic.type",
			Code: validation.ErrorCodeOneOf,
		})
	})
	t.Run("missing query", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.SumoLogic)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.SumoLogic.Query = nil
		err := validate(slo)
		assertContainsErrors(t, err, 1, expectedError{
			Prop: "spec.objectives[0].rawMetric.query.sumoLogic.query",
			Code: validation.ErrorCodeRequired,
		})
	})
}

func TestSumoLogic_MetricType(t *testing.T) {
	t.Run("missing type", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.SumoLogic)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.SumoLogic.Type = nil
		err := validate(slo)
		assertContainsErrors(t, err, 1, expectedError{
			Prop: "spec.objectives[0].rawMetric.query.sumoLogic.type",
			Code: validation.ErrorCodeRequired,
		})
	})
}

func TestSumoLogic_LogsType(t *testing.T) {
}
