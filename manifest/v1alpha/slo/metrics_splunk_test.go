package slo

import (
	"testing"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestSplunk(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Splunk)
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
	t.Run("required", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Splunk)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Splunk.Query = nil
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.splunk.query",
			Code: rules.ErrorCodeRequired,
		})
	})
	t.Run("empty", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Splunk)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Splunk.Query = ptr("")
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.splunk.query",
			Code: rules.ErrorCodeStringNotEmpty,
		})
	})
	t.Run("invalid query", func(t *testing.T) {
		tests := map[string]struct {
			Query        string
			ExpectedCode string
		}{
			"missing n9time": {
				Query: `
search index=svc-events source=udp:5072 sourcetype=syslog status<400 |
bucket _time span=1m |
stats avg(response_time) as n9value by _time |
fields _time n9value`,
				ExpectedCode: rules.ErrorCodeStringContains,
			},
			"missing n9value": {
				Query: `
search index=svc-events source=udp:5072 sourcetype=syslog status<400 |
bucket _time span=1m |
stats avg(response_time) as value by _time |
rename _time as n9time |
fields n9time value`,
				ExpectedCode: rules.ErrorCodeStringContains,
			},
			"missing index": {
				Query: `
search source=udp:5072 sourcetype=syslog status<400 |
bucket _time span=1m |
stats avg(response_time) as n9value by _time |
rename _time as n9time |
fields n9time n9value`,
				ExpectedCode: rules.ErrorCodeStringMatchRegexp,
			},
		}
		for name, test := range tests {
			t.Run(name, func(t *testing.T) {
				slo := validRawMetricSLO(v1alpha.Splunk)
				slo.Spec.Objectives[0].RawMetric.MetricQuery.Splunk.Query = ptr(test.Query)
				err := validate(slo)
				testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
					Prop: "spec.objectives[0].rawMetric.query.splunk.query",
					Code: test.ExpectedCode,
				})
			})
		}
	})
}

func TestSplunk_CountMetrics_SingleQuery(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validSingleQueryGoodOverTotalCountMetricSLO(v1alpha.Splunk)
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
	t.Run("required", func(t *testing.T) {
		slo := validSingleQueryGoodOverTotalCountMetricSLO(v1alpha.Splunk)
		slo.Spec.Objectives[0].CountMetrics.GoodTotalMetric.Splunk.Query = nil
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].countMetrics.goodTotal.splunk.query",
			Code: rules.ErrorCodeRequired,
		})
	})
	t.Run("empty", func(t *testing.T) {
		slo := validSingleQueryGoodOverTotalCountMetricSLO(v1alpha.Splunk)
		slo.Spec.Objectives[0].CountMetrics.GoodTotalMetric.Splunk.Query = ptr("")
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].countMetrics.goodTotal.splunk.query",
			Code: rules.ErrorCodeStringNotEmpty,
		})
	})
	t.Run("goodTotal mixed with total", func(t *testing.T) {
		slo := validSingleQueryGoodOverTotalCountMetricSLO(v1alpha.Splunk)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric = validMetricSpec(v1alpha.Splunk)
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].countMetrics",
			Code: rules.ErrorCodeMutuallyExclusive,
		})
	})
	t.Run("goodTotal mixed with good", func(t *testing.T) {
		slo := validSingleQueryGoodOverTotalCountMetricSLO(v1alpha.Splunk)
		slo.Spec.Objectives[0].CountMetrics.GoodMetric = validMetricSpec(v1alpha.Splunk)
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].countMetrics",
			Code: rules.ErrorCodeMutuallyExclusive,
		})
	})
	t.Run("invalid query", func(t *testing.T) {
		tests := map[string]struct {
			Query        string
			ExpectedCode string
		}{
			"missing n9time": {
				Query: `
    | mstats avg("spl.intr.resource_usage.IOWait.data.avg_cpu_pct") as n9good WHERE index="_metrics" span=15s
    | join type=left _time [
    | mstats avg("spl.intr.resource_usage.IOWait.data.max_cpus_pct") as n9total WHERE index="_metrics" span=15s
    ]
    | fields _time n9good n9total`,
				ExpectedCode: rules.ErrorCodeStringContains,
			},
			"missing n9good": {
				Query: `
    | mstats avg("spl.intr.resource_usage.IOWait.data.avg_cpu_pct") as good WHERE index="_metrics" span=15s
    | join type=left _time [
    | mstats avg("spl.intr.resource_usage.IOWait.data.max_cpus_pct") as n9total WHERE index="_metrics" span=15s
    ]
    | rename _time as n9time
    | fields n9time good n9total`,
				ExpectedCode: rules.ErrorCodeStringContains,
			},
			"missing n9total": {
				Query: `
    | mstats avg("spl.intr.resource_usage.IOWait.data.avg_cpu_pct") as n9good WHERE index="_metrics" span=15s
    | join type=left _time [
    | mstats avg("spl.intr.resource_usage.IOWait.data.max_cpus_pct") as total WHERE index="_metrics" span=15s
    ]
    | rename _time as n9time
    | fields n9time n9good total`,
				ExpectedCode: rules.ErrorCodeStringContains,
			},
			"missing index": {
				Query: `
    | mstats avg("spl.intr.resource_usage.IOWait.data.avg_cpu_pct") as n9good span=15s
    | join type=left _time [
    | mstats avg("spl.intr.resource_usage.IOWait.data.max_cpus_pct") as n9total span=15s
    ]
    | rename _time as n9time
    | fields n9time n9good n9total`,
				ExpectedCode: rules.ErrorCodeStringMatchRegexp,
			},
		}
		for name, test := range tests {
			t.Run(name, func(t *testing.T) {
				slo := validSingleQueryGoodOverTotalCountMetricSLO(v1alpha.Splunk)
				slo.Spec.Objectives[0].CountMetrics.GoodTotalMetric.Splunk.Query = ptr(test.Query)
				err := validate(slo)
				testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
					Prop: "spec.objectives[0].countMetrics.goodTotal.splunk.query",
					Code: test.ExpectedCode,
				})
			})
		}
	})
}
