package slo

import (
	"testing"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

func TestSplunk(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Splunk)
		err := validate(slo)
		testutils.AssertNoErrors(t, slo, err)
	})
	t.Run("required", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Splunk)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Splunk.Query = nil
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.splunk.query",
			Code: validation.ErrorCodeRequired,
		})
	})
	t.Run("empty", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Splunk)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Splunk.Query = ptr("")
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.splunk.query",
			Code: validation.ErrorCodeStringNotEmpty,
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
				ExpectedCode: validation.ErrorCodeStringContains,
			},
			"missing n9value": {
				Query: `
search index=svc-events source=udp:5072 sourcetype=syslog status<400 |
bucket _time span=1m |
stats avg(response_time) as value by _time |
rename _time as n9time |
fields n9time value`,
				ExpectedCode: validation.ErrorCodeStringContains,
			},
			"missing index": {
				Query: `
search source=udp:5072 sourcetype=syslog status<400 |
bucket _time span=1m |
stats avg(response_time) as n9value by _time |
rename _time as n9time |
fields n9time n9value`,
				ExpectedCode: validation.ErrorCodeStringMatchRegexp,
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
