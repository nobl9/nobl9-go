package slo

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

func TestValidate_AppDynamics_ObjectiveLevel(t *testing.T) {
	t.Run("appDynamics applicationName mismatch for bad over total", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.Objectives[0].CountMetrics.TotalMetric = validMetricSpec(v1alpha.AppDynamics)
		slo.Spec.Objectives[0].CountMetrics.GoodMetric = validMetricSpec(v1alpha.AppDynamics)
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.AppDynamics.ApplicationName = ptr("different")
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].countMetrics",
			Code: validation.ErrorCodeNotEqualTo,
		})
	})
	t.Run("appDynamics applicationName mismatch for bad over total", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.Objectives[0].CountMetrics.TotalMetric = validMetricSpec(v1alpha.AppDynamics)
		slo.Spec.Objectives[0].CountMetrics.GoodMetric = nil
		slo.Spec.Objectives[0].CountMetrics.BadMetric = validMetricSpec(v1alpha.AppDynamics)
		slo.Spec.Objectives[0].CountMetrics.BadMetric.AppDynamics.ApplicationName = ptr("different")
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].countMetrics",
			Code: validation.ErrorCodeNotEqualTo,
		})
	})
}

func TestValidate_AppDynamics_Valid(t *testing.T) {
	for _, slo := range []SLO{
		validRawMetricSLO(v1alpha.AppDynamics),
		validCountMetricSLO(v1alpha.AppDynamics),
		func() SLO {
			slo := validRawMetricSLO(v1alpha.AppDynamics)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.AppDynamics.MetricPath = ptr("App | * | Latency")
			return slo
		}(),
	} {
		err := validate(slo)
		testutils.AssertNoErrors(t, slo, err)
	}
}

func TestValidate_AppDynamics_Invalid(t *testing.T) {
	for name, test := range map[string]struct {
		Spec           *AppDynamicsMetric
		ExpectedErrors []testutils.ExpectedError
	}{
		"required fields": {
			Spec: &AppDynamicsMetric{},
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "applicationName",
					Code: validation.ErrorCodeRequired,
				},
				{
					Prop: "metricPath",
					Code: validation.ErrorCodeRequired,
				},
			},
		},
		"application name non empty": {
			Spec: &AppDynamicsMetric{
				ApplicationName: ptr("     "),
				MetricPath:      ptr("path"),
			},
			ExpectedErrors: []testutils.ExpectedError{{
				Prop: "applicationName",
				Code: validation.ErrorCodeStringNotEmpty,
			}},
		},
		"metric path wildcard not supported": {
			Spec: &AppDynamicsMetric{
				ApplicationName: ptr("name"),
				MetricPath:      ptr("App | This* | Latency"),
			},
			ExpectedErrors: []testutils.ExpectedError{{
				Prop: "metricPath",
				Code: errCodeAppDynamicsWildcardNotSupported,
			}},
		},
	} {
		t.Run("rawMetric "+name, func(t *testing.T) {
			slo := validRawMetricSLO(v1alpha.AppDynamics)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.AppDynamics = test.Spec
			err := validate(slo)

			raw := make([]testutils.ExpectedError, len(test.ExpectedErrors))
			copy(raw, test.ExpectedErrors)
			raw = testutils.PrependPropertyPath(raw, "spec.objectives[0].rawMetric.query.appDynamics")
			testutils.AssertContainsErrors(t, slo, err, len(test.ExpectedErrors), raw...)
		})
		t.Run("countMetric "+name, func(t *testing.T) {
			slo := validCountMetricSLO(v1alpha.AppDynamics)
			slo.Spec.Objectives[0].CountMetrics.TotalMetric.AppDynamics = test.Spec
			slo.Spec.Objectives[0].CountMetrics.GoodMetric.AppDynamics = test.Spec
			err := validate(slo)

			total := make([]testutils.ExpectedError, len(test.ExpectedErrors))
			copy(total, test.ExpectedErrors)
			good := make([]testutils.ExpectedError, len(test.ExpectedErrors))
			copy(good, test.ExpectedErrors)
			total = testutils.PrependPropertyPath(total, "spec.objectives[0].countMetrics.total.appDynamics")
			good = testutils.PrependPropertyPath(good, "spec.objectives[0].countMetrics.good.appDynamics")
			testutils.AssertContainsErrors(t, slo, err, len(test.ExpectedErrors)*2, append(total, good...)...) //nolint: makezero
		})
	}
}

func TestValidate_AppDynamics_MetricPathRegex(t *testing.T) {
	for _, test := range []struct {
		metricPath string
		isValid    bool
	}{
		// Valid
		{isValid: true, metricPath: "App | * | Latency"},
		{isValid: true, metricPath: "App |*| Latency"},
		{isValid: true, metricPath: "App|* | Latency"},
		{isValid: true, metricPath: "App | *|Latency"},
		{isValid: true, metricPath: "App|*|Latency"},
		// Invalid
		{isValid: false, metricPath: "App*|Latency"},
		{isValid: false, metricPath: "Ap*p|Latency"},
		{isValid: false, metricPath: "*p|Latency"},
		{isValid: false, metricPath: "App|*p|Latency"},
		{isValid: false, metricPath: "App| *p |Latency"},
		{isValid: false, metricPath: "App|Latency|p*"},
	} {
		slo := validRawMetricSLO(v1alpha.AppDynamics)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.AppDynamics = &AppDynamicsMetric{
			ApplicationName: ptr("name"),
			MetricPath:      ptr(test.metricPath),
		}
		err := validate(slo)
		if test.isValid {
			testutils.AssertNoErrors(t, slo, err)
		} else {
			assert.Error(t, err)
		}
	}
}
