package slo

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

func TestLightstep_CountMetricLevel(t *testing.T) {
	t.Run("streamId must be the same for good and total", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.Lightstep)
		slo.Spec.Objectives[0].CountMetrics.Incremental = ptr(false)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.Lightstep = &LightstepMetric{
			StreamID:   ptr("streamId"),
			TypeOfData: ptr(LightstepTotalCountDataType),
		}
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.Lightstep = &LightstepMetric{
			StreamID:   ptr("different"),
			TypeOfData: ptr(LightstepGoodCountDataType),
		}
		err := validate(slo)
		assertContainsErrors(t, err, 1, expectedError{
			Prop: "spec.objectives[0].countMetrics",
			Code: validation.ErrorCodeEqualTo,
		})
	})
	t.Run("incremental must be set to false", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.Lightstep)
		slo.Spec.Objectives[0].CountMetrics.Incremental = ptr(true)
		err := validate(slo)
		assertContainsErrors(t, err, 1, expectedError{
			Prop: "spec.objectives[0].countMetrics.incremental",
			Code: validation.ErrorCodeEqualTo,
		})
	})
}

func TestLightstep_RawMetricLevel(t *testing.T) {
	t.Run("valid typeOfData", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Lightstep)
		for _, metric := range []*LightstepMetric{
			{
				StreamID:   ptr("123"),
				TypeOfData: ptr(LightstepErrorRateDataType),
			},
			{
				StreamID:   ptr("123"),
				TypeOfData: ptr(LightstepLatencyDataType),
				Percentile: ptr(92.1),
			},
			{
				TypeOfData: ptr(LightstepMetricDataType),
				UQL:        ptr("metric"),
			},
		} {
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Lightstep = metric
			assert.NoError(t, validate(slo))
		}
	})
	t.Run("invalid typeOfData", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Lightstep)
		for _, metric := range []*LightstepMetric{
			{
				StreamID:   ptr("123"),
				TypeOfData: ptr(LightstepTotalCountDataType),
			},
			{
				StreamID:   ptr("123"),
				TypeOfData: ptr(LightstepGoodCountDataType),
			},
		} {
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Lightstep = metric
			assertContainsErrors(t, validate(slo), 1, expectedError{
				Prop: "spec.objectives[0].rawMetric.query.lightstep.typeOfData",
				Code: validation.ErrorCodeOneOf,
			})
		}
	})
}

func TestLightstep_TotalMetricLevel(t *testing.T) {
	t.Run("valid typeOfData", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.Lightstep)
		for _, metric := range []*LightstepMetric{
			{
				StreamID:   ptr("123"),
				TypeOfData: ptr(LightstepTotalCountDataType),
			},
			{
				UQL:        ptr("metric"),
				TypeOfData: ptr(LightstepMetricDataType),
			},
		} {
			slo.Spec.Objectives[0].CountMetrics.TotalMetric.Lightstep = metric
			assert.NoError(t, validate(slo))
		}
	})
	t.Run("invalid typeOfData", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.Lightstep)
		for _, metric := range []*LightstepMetric{
			{
				StreamID:   ptr("123"),
				TypeOfData: ptr(LightstepErrorRateDataType),
			},
			{
				StreamID:   ptr("123"),
				TypeOfData: ptr(LightstepLatencyDataType),
				Percentile: ptr(92.1),
			},
			{
				StreamID:   ptr("123"),
				TypeOfData: ptr(LightstepGoodCountDataType),
			},
		} {
			slo.Spec.Objectives[0].CountMetrics.TotalMetric.Lightstep = metric
			assertContainsErrors(t, validate(slo), 1, expectedError{
				Prop: "spec.objectives[0].countMetrics.total.lightstep.typeOfData",
				Code: validation.ErrorCodeOneOf,
			})
		}
	})
}

func TestLightstep_GoodMetricLevel(t *testing.T) {
	t.Run("valid typeOfData", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.Lightstep)
		for _, metric := range []*LightstepMetric{
			{
				StreamID:   ptr("123"),
				TypeOfData: ptr(LightstepGoodCountDataType),
			},
			{
				UQL:        ptr("metric"),
				TypeOfData: ptr(LightstepMetricDataType),
			},
		} {
			slo.Spec.Objectives[0].CountMetrics.GoodMetric.Lightstep = metric
			assert.NoError(t, validate(slo))
		}
	})
	t.Run("invalid typeOfData", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.Lightstep)
		for _, metric := range []*LightstepMetric{
			{
				StreamID:   ptr("123"),
				TypeOfData: ptr(LightstepErrorRateDataType),
			},
			{
				StreamID:   ptr("123"),
				TypeOfData: ptr(LightstepLatencyDataType),
				Percentile: ptr(92.1),
			},
			{
				TypeOfData: ptr(LightstepTotalCountDataType),
				StreamID:   ptr("123"),
			},
		} {
			slo.Spec.Objectives[0].CountMetrics.GoodMetric.Lightstep = metric
			assertContainsErrors(t, validate(slo), 1, expectedError{
				Prop: "spec.objectives[0].countMetrics.good.lightstep.typeOfData",
				Code: validation.ErrorCodeOneOf,
			})
		}
	})
}

func TestLightstepLatencyTypeOfData(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Lightstep)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Lightstep = &LightstepMetric{
			TypeOfData: ptr(LightstepLatencyDataType),
			StreamID:   ptr("123"),
			Percentile: ptr(99.99),
		}
		err := validate(slo)
		assert.Empty(t, err)
	})
	t.Run("fails", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Lightstep)
		slo.Spec.Objectives = []Objective{
			{
				ObjectiveBase: ObjectiveBase{Name: "test", Value: ptr(10.0)},
				BudgetTarget:  ptr(0.9),

				RawMetric: &RawMetricSpec{MetricQuery: &MetricSpec{Lightstep: &LightstepMetric{
					StreamID:   ptr("123"),
					TypeOfData: ptr(LightstepLatencyDataType),
					UQL:        ptr("metric"),
				}}},
				Operator: ptr(v1alpha.GreaterThan.String()),
			},
			{
				ObjectiveBase: ObjectiveBase{Name: "test1", Value: ptr(11.0)},
				BudgetTarget:  ptr(0.8),
				RawMetric: &RawMetricSpec{MetricQuery: &MetricSpec{Lightstep: &LightstepMetric{
					StreamID:   nil,
					TypeOfData: ptr(LightstepLatencyDataType),
					Percentile: ptr(0.0),
				}}},
				Operator: ptr(v1alpha.GreaterThan.String()),
			},
			{
				ObjectiveBase: ObjectiveBase{Name: "test2", Value: ptr(12.0)},
				BudgetTarget:  ptr(0.7),
				RawMetric: &RawMetricSpec{MetricQuery: &MetricSpec{Lightstep: &LightstepMetric{
					StreamID:   ptr("123"),
					TypeOfData: ptr(LightstepLatencyDataType),
					Percentile: ptr(100.0),
				}}},
				Operator: ptr(v1alpha.GreaterThan.String()),
			},
		}
		err := validate(slo)
		assertContainsErrors(t, err, 5,
			expectedError{
				Prop: "spec.objectives[0].rawMetric.query.lightstep.percentile",
				Code: validation.ErrorCodeRequired,
			},
			expectedError{
				Prop: "spec.objectives[0].rawMetric.query.lightstep.uql",
				Code: validation.ErrorCodeForbidden,
			},
			expectedError{
				Prop: "spec.objectives[1].rawMetric.query.lightstep.streamId",
				Code: validation.ErrorCodeRequired,
			},
			expectedError{
				Prop: "spec.objectives[1].rawMetric.query.lightstep.percentile",
				Code: validation.ErrorCodeGreaterThan,
			},
			expectedError{
				Prop: "spec.objectives[2].rawMetric.query.lightstep.percentile",
				Code: validation.ErrorCodeLessThanOrEqualTo,
			},
		)
	})
}

func TestLightstepErrorRateTypeOfData(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Lightstep)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Lightstep = &LightstepMetric{
			TypeOfData: ptr(LightstepErrorRateDataType),
			StreamID:   ptr("123"),
		}
		err := validate(slo)
		assert.Empty(t, err)
	})
	t.Run("fails", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Lightstep)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Lightstep = &LightstepMetric{
			TypeOfData: ptr(LightstepErrorRateDataType),
			StreamID:   nil,
			Percentile: ptr(0.1),
			UQL:        ptr("this"),
		}
		err := validate(slo)
		assertContainsErrors(t, err, 3,
			expectedError{
				Prop: "spec.objectives[0].rawMetric.query.lightstep.percentile",
				Code: validation.ErrorCodeForbidden,
			},
			expectedError{
				Prop: "spec.objectives[0].rawMetric.query.lightstep.uql",
				Code: validation.ErrorCodeForbidden,
			},
			expectedError{
				Prop: "spec.objectives[0].rawMetric.query.lightstep.streamId",
				Code: validation.ErrorCodeRequired,
			},
		)
	})
}

func TestLightstepMetricTypeOfData(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Lightstep)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Lightstep = &LightstepMetric{
			TypeOfData: ptr(LightstepMetricDataType),
			UQL: ptr(`(
metric cpu.utilization | rate | filter error == true && service == spans_sample | group_by [], min;
spans count | rate | group_by [], sum
) | join left/right * 100`),
		}
		err := validate(slo)
		assert.Empty(t, err)
	})
	t.Run("fails", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Lightstep)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Lightstep = &LightstepMetric{
			TypeOfData: ptr(LightstepMetricDataType),
			UQL:        nil,
			Percentile: ptr(0.1),
			StreamID:   ptr("this"),
		}
		err := validate(slo)
		assertContainsErrors(t, err, 3,
			expectedError{
				Prop: "spec.objectives[0].rawMetric.query.lightstep.uql",
				Code: validation.ErrorCodeRequired,
			},
			expectedError{
				Prop: "spec.objectives[0].rawMetric.query.lightstep.percentile",
				Code: validation.ErrorCodeForbidden,
			},
			expectedError{
				Prop: "spec.objectives[0].rawMetric.query.lightstep.streamId",
				Code: validation.ErrorCodeForbidden,
			},
		)
	})
	t.Run("invalid metrics", func(t *testing.T) {
		for name, uql := range map[string]string{
			"spans_sample joined UQL": `(
spans_sample count | delta | filter error == true && service == android | group_by [], sum;
spans_sample count | delta | filter service == android | group_by [], sum) | join left/right * 100`,
			"constant UQL":     "constant .5",
			"spans_sample UQL": "spans_sample span filter",
			"assemble UQL":     "assemble span",
		} {
			t.Run(name, func(t *testing.T) {
				slo := validRawMetricSLO(v1alpha.Lightstep)
				slo.Spec.Objectives[0].RawMetric.MetricQuery.Lightstep = &LightstepMetric{
					TypeOfData: ptr(LightstepMetricDataType),
					UQL:        ptr(uql),
				}
				err := validate(slo)
				assertContainsErrors(t, err, 1,
					expectedError{
						Prop: "spec.objectives[0].rawMetric.query.lightstep.uql",
						Code: validation.ErrorCodeStringDenyRegexp,
					},
				)
			})
		}
	})
}

func TestLightstepGoodTotalTypeOfData(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.Lightstep)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.Lightstep = &LightstepMetric{
			TypeOfData: ptr(LightstepTotalCountDataType),
			StreamID:   ptr("123"),
		}
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.Lightstep = &LightstepMetric{
			TypeOfData: ptr(LightstepGoodCountDataType),
			StreamID:   ptr("123"),
		}
		err := validate(slo)
		assert.Empty(t, err)
	})
	t.Run("fails", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.Lightstep)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.Lightstep = &LightstepMetric{
			TypeOfData: ptr(LightstepTotalCountDataType),
			StreamID:   nil,
			Percentile: ptr(0.1),
			UQL:        ptr("this"),
		}
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.Lightstep = &LightstepMetric{
			TypeOfData: ptr(LightstepGoodCountDataType),
			StreamID:   nil,
			Percentile: ptr(0.1),
			UQL:        ptr("this"),
		}
		err := validate(slo)
		assertContainsErrors(t, err, 6,
			expectedError{
				Prop: "spec.objectives[0].countMetrics.total.lightstep.percentile",
				Code: validation.ErrorCodeForbidden,
			},
			expectedError{
				Prop: "spec.objectives[0].countMetrics.total.lightstep.uql",
				Code: validation.ErrorCodeForbidden,
			},
			expectedError{
				Prop: "spec.objectives[0].countMetrics.total.lightstep.streamId",
				Code: validation.ErrorCodeRequired,
			},
			expectedError{
				Prop: "spec.objectives[0].countMetrics.good.lightstep.percentile",
				Code: validation.ErrorCodeForbidden,
			},
			expectedError{
				Prop: "spec.objectives[0].countMetrics.good.lightstep.uql",
				Code: validation.ErrorCodeForbidden,
			},
			expectedError{
				Prop: "spec.objectives[0].countMetrics.good.lightstep.streamId",
				Code: validation.ErrorCodeRequired,
			},
		)
	})
}
