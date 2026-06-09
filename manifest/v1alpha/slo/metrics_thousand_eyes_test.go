package slo

import (
	"testing"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestThousandEyes(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.ThousandEyes)
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
	t.Run("forbidden for count metrics", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.ThousandEyes)
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 2,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].countMetrics.total.thousandEyes",
				Code: rules.ErrorCodeForbidden,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].countMetrics.good.thousandEyes",
				Code: rules.ErrorCodeForbidden,
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
				Code: rules.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.thousandEyes.testType",
				Code: rules.ErrorCodeRequired,
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
				Code: rules.ErrorCodeGreaterThanOrEqualTo,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.thousandEyes.testType",
				Code: rules.ErrorCodeOneOf,
			},
		)
	})
	t.Run("valid testType", func(t *testing.T) {
		for _, testType := range supportedThousandEyesTestTypes {
			slo := validRawMetricSLO(v1alpha.ThousandEyes)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.ThousandEyes.TestType = ptr(testType)
			err := validate(slo)
			testutils.AssertNoError(t, slo, err)
		}
	})
	t.Run("invalid accountGroupID", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.ThousandEyes)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.ThousandEyes.AccountGroupID = ptr[int64](-1)
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.thousandEyes.accountGroupID",
				Code: rules.ErrorCodeGreaterThanOrEqualTo,
			},
		)
	})
	t.Run("valid accountGroupID", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.ThousandEyes)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.ThousandEyes.AccountGroupID = ptr[int64](2114119)
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
}
