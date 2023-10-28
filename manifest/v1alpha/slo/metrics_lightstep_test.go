package slo

import (
	"reflect"
	"sort"
	"testing"

	v "github.com/go-playground/validator/v10"
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
			TypeOfData: nil,
		}
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.Lightstep = &LightstepMetric{
			StreamID:   ptr("different"),
			TypeOfData: nil,
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
		for _, typeOfData := range []string{
			LightstepErrorRateDataType,
			LightstepLatencyDataType,
			LightstepMetricDataType,
		} {
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Lightstep.TypeOfData = &typeOfData
			assert.NoError(t, validate(slo))
		}
	})
	t.Run("invalid typeOfData", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Lightstep)
		for _, typeOfData := range []string{
			LightstepGoodCountDataType,
			LightstepTotalCountDataType,
		} {
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Lightstep.TypeOfData = &typeOfData
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
		for _, typeOfData := range []string{
			LightstepTotalCountDataType,
			LightstepMetricDataType,
		} {
			slo.Spec.Objectives[0].CountMetrics.TotalMetric.Lightstep.TypeOfData = &typeOfData
			assert.NoError(t, validate(slo))
		}
	})
	t.Run("invalid typeOfData", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Lightstep)
		for _, typeOfData := range []string{
			LightstepGoodCountDataType,
			LightstepErrorRateDataType,
			LightstepLatencyDataType,
		} {
			slo.Spec.Objectives[0].CountMetrics.TotalMetric.Lightstep.TypeOfData = &typeOfData
			assertContainsErrors(t, validate(slo), 1, expectedError{
				Prop: "spec.objectives[0].countMetrics.total.lightstep.typeOfData",
				Code: validation.ErrorCodeOneOf,
			})
		}
	})
}

func TestLightstep_GoodMetricLevel(t *testing.T) {
	t.Run("valid typeOfData", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Lightstep)
		for _, typeOfData := range []string{
			LightstepGoodCountDataType,
			LightstepMetricDataType,
		} {
			slo.Spec.Objectives[0].CountMetrics.GoodMetric.Lightstep.TypeOfData = &typeOfData
			assert.NoError(t, validate(slo))
		}
	})
	t.Run("invalid typeOfData", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Lightstep)
		for _, typeOfData := range []string{
			LightstepTotalCountDataType,
			LightstepErrorRateDataType,
			LightstepLatencyDataType,
		} {
			slo.Spec.Objectives[0].CountMetrics.GoodMetric.Lightstep.TypeOfData = &typeOfData
			assertContainsErrors(t, validate(slo), 1, expectedError{
				Prop: "spec.objectives[0].countMetrics.good.lightstep.typeOfData",
				Code: validation.ErrorCodeOneOf,
			})
		}
	})
}

func TestLightstepMetric(t *testing.T) {
	negativePercentile := -1.0
	zeroPercentile := 0.0
	positivePercentile := 95.0
	overflowPercentile := 100.0
	streamID := "123"
	validUQL := `(
		metric cpu.utilization | rate | filter error == true && service == spans_sample | group_by [], min;
		spans count | rate | group_by [], sum
	) | join left/right * 100`
	forbiddenSpanSampleJoinedUQL := `(
	  spans_sample count | delta | filter error == true && service == android | group_by [], sum;
	  spans_sample count | delta | filter service == android | group_by [], sum
	) | join left/right * 100`
	forbiddenConstantUQL := "constant .5"
	forbiddenSpansSampleUQL := "spans_sample span filter"
	forbiddenAssembleUQL := "assemble span"
	createSpec := func(uql, streamID, dataType *string, percentile *float64) *MetricSpec {
		return &MetricSpec{
			Lightstep: &LightstepMetric{
				UQL:        uql,
				StreamID:   streamID,
				TypeOfData: dataType,
				Percentile: percentile,
			},
		}
	}
	getStringPointer := func(s string) *string { return &s }

	testCases := []struct {
		description string
		spec        *MetricSpec
		errors      []string
	}{
		{
			description: "Valid latency type spec",
			spec:        createSpec(nil, &streamID, getStringPointer(LightstepLatencyDataType), &positivePercentile),
			errors:      nil,
		},
		{
			description: "Invalid latency type spec",
			spec:        createSpec(&validUQL, nil, getStringPointer(LightstepLatencyDataType), nil),
			errors:      []string{"percentileRequired", "streamIDRequired", "uqlNotAllowed"},
		},
		{
			description: "Invalid latency type spec - negative percentile",
			spec:        createSpec(nil, &streamID, getStringPointer(LightstepLatencyDataType), &negativePercentile),
			errors:      []string{"invalidPercentile"},
		},
		{
			description: "Invalid latency type spec - zero percentile",
			spec:        createSpec(nil, &streamID, getStringPointer(LightstepLatencyDataType), &zeroPercentile),
			errors:      []string{"invalidPercentile"},
		},
		{
			description: "Invalid latency type spec - overflow percentile",
			spec:        createSpec(nil, &streamID, getStringPointer(LightstepLatencyDataType), &overflowPercentile),
			errors:      []string{"invalidPercentile"},
		},
		{
			description: "Valid error rate type spec",
			spec:        createSpec(nil, &streamID, getStringPointer(LightstepErrorRateDataType), nil),
			errors:      nil,
		},
		{
			description: "Invalid error rate type spec",
			spec:        createSpec(&validUQL, nil, getStringPointer(LightstepErrorRateDataType), &positivePercentile),
			errors:      []string{"streamIDRequired", "percentileNotAllowed", "uqlNotAllowed"},
		},
		{
			description: "Valid total count type spec",
			spec:        createSpec(nil, &streamID, getStringPointer(LightstepTotalCountDataType), nil),
			errors:      nil,
		},
		{
			description: "Invalid total count type spec",
			spec:        createSpec(&validUQL, nil, getStringPointer(LightstepTotalCountDataType), &positivePercentile),
			errors:      []string{"streamIDRequired", "uqlNotAllowed", "percentileNotAllowed"},
		},
		{
			description: "Valid good count type spec",
			spec:        createSpec(nil, &streamID, getStringPointer(LightstepGoodCountDataType), nil),
			errors:      nil,
		},
		{
			description: "Invalid good count type spec",
			spec:        createSpec(&validUQL, nil, getStringPointer(LightstepGoodCountDataType), &positivePercentile),
			errors:      []string{"streamIDRequired", "uqlNotAllowed", "percentileNotAllowed"},
		},
		{
			description: "Valid metric type spec",
			spec:        createSpec(&validUQL, nil, getStringPointer(LightstepMetricDataType), nil),
			errors:      nil,
		},
		{
			description: "Invalid metric type spec",
			spec:        createSpec(nil, &streamID, getStringPointer(LightstepMetricDataType), &positivePercentile),
			errors:      []string{"uqlRequired", "percentileNotAllowed", "streamIDNotAllowed"},
		},
		{
			description: "Invalid metric type spec - empty UQL",
			spec:        createSpec(getStringPointer(""), nil, getStringPointer(LightstepMetricDataType), nil),
			errors:      []string{"uqlRequired"},
		},
		{
			description: "Invalid metric type spec - not supported UQL",
			spec:        createSpec(&forbiddenSpanSampleJoinedUQL, nil, getStringPointer(LightstepMetricDataType), nil),
			errors:      []string{"onlyMetricAndSpansUQLQueriesAllowed"},
		},
		{
			description: "Invalid metric type spec - not supported UQL",
			spec:        createSpec(&forbiddenConstantUQL, nil, getStringPointer(LightstepMetricDataType), nil),
			errors:      []string{"onlyMetricAndSpansUQLQueriesAllowed"},
		},
		{
			description: "Invalid metric type spec - not supported UQL",
			spec:        createSpec(&forbiddenSpansSampleUQL, nil, getStringPointer(LightstepMetricDataType), nil),
			errors:      []string{"onlyMetricAndSpansUQLQueriesAllowed"},
		},
		{
			description: "Invalid metric type spec - not supported UQL",
			spec:        createSpec(&forbiddenAssembleUQL, nil, getStringPointer(LightstepMetricDataType), nil),
			errors:      []string{"onlyMetricAndSpansUQLQueriesAllowed"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			slo := validRawMetricSLO(v1alpha.Lightstep)
			slo.Spec.Objectives[0].RawMetric.MetricQuery = tc.spec
			err := validate(slo)
			if len(tc.errors) == 0 {
				assert.Empty(t, err)
				return
			}

			validationErrors, ok := err.(v.ValidationErrors)
			if !ok {
				assert.FailNow(t, "cannot cast error to validator.ValidatorErrors")
			}
			var errors []string
			for _, ve := range validationErrors {
				errors = append(errors, ve.Tag())
			}
			sort.Strings(tc.errors)
			sort.Strings(errors)
			assert.True(t, reflect.DeepEqual(tc.errors, errors))
		})
	}
}
