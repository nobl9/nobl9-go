package slo

import (
	"testing"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
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
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
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
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
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
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.sumoLogic.type",
			Code: validation.ErrorCodeRequired,
		})
	})
	t.Run("invalid type", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.SumoLogic)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.SumoLogic.Type = ptr("invalid")
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.sumoLogic.type",
			Code: validation.ErrorCodeOneOf,
		})
	})
	t.Run("missing query", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.SumoLogic)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.SumoLogic.Query = nil
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.sumoLogic.query",
			Code: validation.ErrorCodeRequired,
		})
	})
}

func TestSumoLogic_MetricType(t *testing.T) {
	t.Run("required values", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.SumoLogic)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.SumoLogic.Quantization = nil
		slo.Spec.Objectives[0].RawMetric.MetricQuery.SumoLogic.Rollup = nil
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 2,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.sumoLogic.quantization",
				Code: validation.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.sumoLogic.rollup",
				Code: validation.ErrorCodeRequired,
			},
		)
	})
	t.Run("invalid quantization", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.SumoLogic)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.SumoLogic.Quantization = ptr("invalid")
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop:    "spec.objectives[0].rawMetric.query.sumoLogic.quantization",
			Message: `error parsing quantization string to duration - time: invalid duration "invalid"`,
		})
	})
	t.Run("minimum quantization", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.SumoLogic)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.SumoLogic.Quantization = ptr("14s")
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop:    "spec.objectives[0].rawMetric.query.sumoLogic.quantization",
			Message: "minimum quantization value is [15s], got: [14s]",
		})
	})
	t.Run("valid rollups", func(t *testing.T) {
		for _, rollup := range sumoLogicValidRollups {
			slo := validRawMetricSLO(v1alpha.SumoLogic)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.SumoLogic.Rollup = ptr(rollup)
			err := validate(slo)
			testutils.AssertNoError(t, slo, err)
		}
	})
	t.Run("invalid rollup", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.SumoLogic)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.SumoLogic.Rollup = ptr("invalid")
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.sumoLogic.rollup",
			Code: validation.ErrorCodeOneOf,
		})
	})
}

func TestSumoLogic_LogsType(t *testing.T) {
	t.Run("forbidden values", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.SumoLogic)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.SumoLogic = &SumoLogicMetric{
			Type: ptr(sumoLogicTypeLogs),
			Query: ptr(`
_collector="n9-dev-tooling-cluster" _source="logs"
  | json "log"
  | timeslice 20s as n9_time
  | parse "level=* *" as (log_level, tail)
  | if (log_level matches "error" ,0,1) as log_level_not_error
  | sum(log_level_not_error) as n9_value by n9_time
  | sort by n9_time asc`),
			Quantization: ptr("20s"),
			Rollup:       ptr("None"),
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 2,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.sumoLogic.quantization",
				Code: validation.ErrorCodeForbidden,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.sumoLogic.rollup",
				Code: validation.ErrorCodeForbidden,
			},
		)
	})
	tests := map[string]struct {
		Query string
		Error testutils.ExpectedError
	}{
		"no timeslice segment": {
			Query: `
_collector="n9-dev-tooling-cluster" _source="logs"
  | json "log"
  | parse "level=* *" as (log_level, tail)
  | if (log_level matches "error" ,0,1) as log_level_not_error
  | sum(log_level_not_error) as n9_value by n9_time
  | sort by n9_time asc`,
			Error: testutils.ExpectedError{
				Prop:    "spec.objectives[0].rawMetric.query.sumoLogic.query",
				Message: "exactly one timeslice declaration is required in the query",
			},
		},
		"two timeslice segments": {
			Query: `
_collector="n9-dev-tooling-cluster" _source="logs"
  | json "log"
  | timeslice 30s
  | timeslice 20s as n9_time
  | parse "level=* *" as (log_level, tail)
  | if (log_level matches "error" ,0,1) as log_level_not_error
  | sum(log_level_not_error) as n9_value by n9_time
  | sort by n9_time asc`,
			Error: testutils.ExpectedError{
				Prop:    "spec.objectives[0].rawMetric.query.sumoLogic.query",
				Message: "exactly one timeslice declaration is required in the query",
			},
		},
		"invalid timeslice segment": {
			Query: `
_collector="n9-dev-tooling-cluster" _source="logs"
  | json "log"
  | timeslice 20x as n9_time
  | parse "level=* *" as (log_level, tail)
  | if (log_level matches "error" ,0,1) as log_level_not_error
  | sum(log_level_not_error) as n9_value by n9_time
  | sort by n9_time asc`,
			Error: testutils.ExpectedError{
				Prop:    "spec.objectives[0].rawMetric.query.sumoLogic.query",
				Message: `error parsing timeslice duration: time: unknown unit "x" in duration "20x"`,
			},
		},
		"minimum timeslice value": {
			Query: `
_collector="n9-dev-tooling-cluster" _source="logs"
  | json "log"
  | timeslice 14s as n9_time
  | parse "level=* *" as (log_level, tail)
  | if (log_level matches "error" ,0,1) as log_level_not_error
  | sum(log_level_not_error) as n9_value by n9_time
  | sort by n9_time asc`,
			Error: testutils.ExpectedError{
				Prop:    "spec.objectives[0].rawMetric.query.sumoLogic.query",
				Message: `minimum timeslice value is [15s], got: [14s]`,
			},
		},
		"missing n9_value": {
			Query: `
_collector="n9-dev-tooling-cluster" _source="logs"
  | json "log"
  | timeslice 20s as n9_time
  | parse "level=* *" as (log_level, tail)
  | if (log_level matches "error" ,0,1) as log_level_not_error
  | sum(log_level_not_error) by n9_time
  | sort by n9_time asc`,
			Error: testutils.ExpectedError{
				Prop:            "spec.objectives[0].rawMetric.query.sumoLogic.query",
				ContainsMessage: "n9_value is required",
			},
		},
		"missing n9_time": {
			Query: `
_collector="n9-dev-tooling-cluster" _source="logs"
  | json "log"
  | timeslice 20s
  | parse "level=* *" as (log_level, tail)
  | if (log_level matches "error" ,0,1) as log_level_not_error
  | sum(log_level_not_error) as n9_value by time
  | sort by time asc`,
			Error: testutils.ExpectedError{
				Prop:            "spec.objectives[0].rawMetric.query.sumoLogic.query",
				ContainsMessage: "n9_time is required",
			},
		},
		"missing aggregation function": {
			Query: `
_collector="n9-dev-tooling-cluster" _source="logs"
  | json "log"
  | timeslice 20s as n9_time
  | parse "level=* *" as (log_level, tail)
  | if (log_level matches "error" ,0,1) as log_level_not_error
  | sum(log_level_not_error) as n9_value`,
			Error: testutils.ExpectedError{
				Prop:            "spec.objectives[0].rawMetric.query.sumoLogic.query",
				ContainsMessage: "aggregation function is required",
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			slo := validRawMetricSLO(v1alpha.SumoLogic)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.SumoLogic = &SumoLogicMetric{
				Type:  ptr(sumoLogicTypeLogs),
				Query: ptr(test.Query),
			}
			err := validate(slo)
			testutils.AssertContainsErrors(t, slo, err, 1, test.Error)
		})
	}
}
