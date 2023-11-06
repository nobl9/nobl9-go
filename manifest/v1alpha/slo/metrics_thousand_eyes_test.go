package slo

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

func TestThousandEyes(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.ThousandEyes)
		err := validate(slo)
		assert.Empty(t, err)
	})
	t.Run("forbidden for count metrics", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.ThousandEyes)
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 2,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].countMetrics.total.thousandEyes",
				Code: validation.ErrorCodeForbidden,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].countMetrics.good.thousandEyes",
				Code: validation.ErrorCodeForbidden,
			},
		)
	})
	t.Run("required fields", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.ThousandEyes)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.ThousandEyes = &ThousandEyesMetric{}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 2,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.thousandEyes.testID",
				Code: validation.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.thousandEyes.testType",
				Code: validation.ErrorCodeRequired,
			},
		)
	})
	t.Run("invalid fields", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.ThousandEyes)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.ThousandEyes = &ThousandEyesMetric{
			TestID:   ptr[int64](-1),
			TestType: ptr("invalid"),
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 2,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.thousandEyes.testID",
				Code: validation.ErrorCodeGreaterThanOrEqualTo,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.thousandEyes.testType",
				Code: validation.ErrorCodeOneOf,
			},
		)
	})
	t.Run("valid testType", func(t *testing.T) {
		for _, testType := range supportedThousandEeyesTestTypes {
			slo := validRawMetricSLO(v1alpha.ThousandEyes)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.ThousandEyes.TestType = ptr(testType)
			err := validate(slo)
			assert.Empty(t, err)
		}
	})
}
