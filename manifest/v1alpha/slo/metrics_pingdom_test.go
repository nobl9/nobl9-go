package slo

import (
	"testing"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestPingdom_CountMetricsLevel(t *testing.T) {
	t.Run("checkId must be the same for good and total", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.Pingdom)
		slo.Spec.Objectives[0].CountMetrics.Incremental = ptr(false)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.Pingdom = &PingdomMetric{
			CheckID:   ptr("123"),
			CheckType: ptr(PingdomTypeTransaction),
		}
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.Pingdom = &PingdomMetric{
			CheckID:   ptr("333"),
			CheckType: ptr(PingdomTypeTransaction),
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].countMetrics",
			Code: rules.ErrorCodeEqualTo,
		})
	})
	t.Run("checkType must be the same for good and total", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.Pingdom)
		slo.Spec.Objectives[0].CountMetrics.Incremental = ptr(false)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.Pingdom = &PingdomMetric{
			CheckID:   ptr("123"),
			CheckType: ptr(PingdomTypeUptime),
			Status:    ptr(pingdomStatusDown),
		}
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.Pingdom = &PingdomMetric{
			CheckID:   ptr("123"),
			CheckType: ptr(PingdomTypeTransaction),
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].countMetrics",
			Code: rules.ErrorCodeEqualTo,
		})
	})
	t.Run("required status", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.Pingdom)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.Pingdom.CheckType = ptr(PingdomTypeUptime)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.Pingdom.Status = nil
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.Pingdom.CheckType = ptr(PingdomTypeUptime)
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.Pingdom.Status = nil
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 2,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].countMetrics.total.pingdom.status",
				Code: rules.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].countMetrics.good.pingdom.status",
				Code: rules.ErrorCodeRequired,
			},
		)
	})
}

func TestPingdom_RawMetricLevel(t *testing.T) {
	t.Run("valid checkType", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Pingdom)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Pingdom.CheckType = ptr(PingdomTypeUptime)
		testutils.AssertNoError(t, slo, validate(slo))
	})
	t.Run("invalid checkType", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Pingdom)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Pingdom.CheckType = ptr(PingdomTypeTransaction)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Pingdom.Status = nil
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.pingdom.checkType",
			Code: rules.ErrorCodeEqualTo,
		})
	})
	t.Run("omit empty status", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Pingdom)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Pingdom.CheckType = ptr(PingdomTypeUptime)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Pingdom.Status = nil
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
}

func TestPingdom(t *testing.T) {
	t.Run("required checkType", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Pingdom)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Pingdom.CheckType = nil
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.pingdom.checkType",
			Code: rules.ErrorCodeRequired,
		})
	})
	t.Run("required checkId", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Pingdom)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Pingdom.CheckID = nil
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.pingdom.checkId",
			Code: rules.ErrorCodeRequired,
		})
	})
	t.Run("missing checkId", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Pingdom)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Pingdom.CheckID = ptr("")
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.pingdom.checkId",
			Code: rules.ErrorCodeStringNotEmpty,
		})
	})
	t.Run("invalid checkId", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Pingdom)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Pingdom.CheckID = ptr("a12393")
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.pingdom.checkId",
			Code: rules.ErrorCodeStringMatchRegexp,
		})
	})
}

func TestPingdom_CheckTypeTransaction(t *testing.T) {
	t.Run("forbidden status", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.Pingdom)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.Pingdom.CheckType = ptr(PingdomTypeTransaction)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.Pingdom.Status = ptr(pingdomStatusDown)
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.Pingdom.CheckType = ptr(PingdomTypeTransaction)
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.Pingdom.Status = nil
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].countMetrics.total.pingdom.status",
			Code: rules.ErrorCodeForbidden,
		})
	})
}

func TestPingdom_CheckTypeUptime(t *testing.T) {
	t.Run("valid status", func(t *testing.T) {
		for _, status := range []string{
			pingdomStatusUp,
			pingdomStatusDown,
			pingdomStatusUnconfirmed,
			pingdomStatusUnknown,
			pingdomStatusDown + "," + pingdomStatusUnconfirmed,
		} {
			slo := validRawMetricSLO(v1alpha.Pingdom)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Pingdom.CheckType = ptr(PingdomTypeUptime)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Pingdom.Status = ptr(status)
			err := validate(slo)
			testutils.AssertNoError(t, slo, err)
		}
	})
	t.Run("invalid status", func(t *testing.T) {
		for _, status := range []string{
			",",
			"",
			"",
			pingdomStatusDown + "," + "invalid",
			"invalid" + "," + pingdomStatusUp,
		} {
			slo := validRawMetricSLO(v1alpha.Pingdom)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Pingdom.CheckType = ptr(PingdomTypeUptime)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Pingdom.Status = ptr(status)
			err := validate(slo)
			testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.pingdom.status",
				Code: rules.ErrorCodeOneOf,
			})
		}
	})
}
