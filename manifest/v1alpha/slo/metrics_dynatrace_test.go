package slo

import (
	"testing"

	"github.com/nobl9/govy/pkg/jsonpath"

	"github.com/nobl9/govy/pkg/rules"
	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestDynatrace(t *testing.T) {
	t.Run("passes with metric selector", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Dynatrace)
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
	t.Run("passes with dql", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Dynatrace)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Dynatrace.MetricSelector = nil
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Dynatrace.DQL = validDynatraceDQL()
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
	t.Run("rejects both query fields", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Dynatrace)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Dynatrace.DQL = validDynatraceDQL()
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: jsonpath.Parse("spec.objectives[0].rawMetric.query.dynatrace"),
			Code: rules.ErrorCodeMutuallyExclusive,
		})
	})
	t.Run("requires metric selector or dql", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Dynatrace)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Dynatrace.MetricSelector = nil
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: jsonpath.Parse("spec.objectives[0].rawMetric.query.dynatrace"),
			Code: rules.ErrorCodeRequired,
		})
	})
	t.Run("empty metric selector", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Dynatrace)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Dynatrace.MetricSelector = ptr("")
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 2, testutils.ExpectedError{
			Prop: jsonpath.Parse("spec.objectives[0].rawMetric.query.dynatrace"),
			Code: rules.ErrorCodeRequired,
		}, testutils.ExpectedError{
			Prop: jsonpath.Parse("spec.objectives[0].rawMetric.query.dynatrace.metricSelector"),
			Code: rules.ErrorCodeStringNotEmpty,
		})
	})
	t.Run("empty dql", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Dynatrace)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Dynatrace.MetricSelector = nil
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Dynatrace.DQL = &DynatraceDQL{Interval: "1m"}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 2, testutils.ExpectedError{
			Prop: jsonpath.Parse("spec.objectives[0].rawMetric.query.dynatrace"),
			Code: rules.ErrorCodeRequired,
		}, testutils.ExpectedError{
			Prop: jsonpath.Parse("spec.objectives[0].rawMetric.query.dynatrace.dql.query"),
			Code: rules.ErrorCodeStringNotEmpty,
		})
	})
	t.Run("empty dql interval", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Dynatrace)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Dynatrace.MetricSelector = nil
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Dynatrace.DQL = &DynatraceDQL{
			Query: "timeseries value = avg(dt.host.cpu.usage)",
		}
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
	t.Run("invalid dql interval", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Dynatrace)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Dynatrace.MetricSelector = nil
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Dynatrace.DQL = &DynatraceDQL{
			Query:    "timeseries value = avg(dt.host.cpu.usage)",
			Interval: "invalid",
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: jsonpath.Parse("spec.objectives[0].rawMetric.query.dynatrace.dql.interval"),
		})
	})
}

func TestDynatraceMetric_QueryType(t *testing.T) {
	t.Run("metric selector", func(t *testing.T) {
		metricSelector := "builtin:host.cpu.usage"
		metric := DynatraceMetric{MetricSelector: &metricSelector}

		assert.Equal(t, DynatraceMetricQueryTypeMetricSelector, metric.QueryType())
	})
	t.Run("dql", func(t *testing.T) {
		metric := DynatraceMetric{DQL: validDynatraceDQL()}

		assert.Equal(t, DynatraceMetricQueryTypeDQL, metric.QueryType())
	})
}

func TestDynatraceDQL_Query(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		isValid bool
	}{
		{
			name: "allows DQL without time range parameters",
			query: "timeseries response_time = avg(dt.service.request.response_time), by:{dt.entity.service} " +
				"| fields response_time, dt.entity.service, timeframe, interval",
			isValid: true,
		},
		{
			name:    "rejects interval parameter",
			query:   "timeseries response_time = avg(dt.service.request.response_time), by:{dt.entity.service}, interval:1m",
			isValid: false,
		},
		{
			name:    "rejects bins parameter",
			query:   "timeseries response_time = avg(dt.service.request.response_time), by:{dt.entity.service}, bins:120",
			isValid: false,
		},
		{
			name:    "rejects from parameter",
			query:   "timeseries response_time = avg(dt.service.request.response_time), by:{dt.entity.service}, from:-1h",
			isValid: false,
		},
		{
			name:    "rejects to parameter",
			query:   "timeseries response_time = avg(dt.service.request.response_time), by:{dt.entity.service}, to:now()",
			isValid: false,
		},
		{
			name:    "rejects timeframe parameter",
			query:   `timeseries response_time = avg(dt.service.request.response_time), timeframe:"2026-05-01/2026-05-02"`,
			isValid: false,
		},
		{
			name: "rejects shift parameter",
			query: "timeseries response_time_yesterday = avg(dt.service.request.response_time), " +
				"by:{dt.entity.service}, shift:-24h",
			isValid: false,
		},
		{
			name: "rejects forbidden parameters anywhere in the query",
			query: "fetch logs " +
				`| filter contains(content, "from: now()") ` +
				"| makeTimeseries errors = count(), interval:5m",
			isValid: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			slo := validRawMetricSLO(v1alpha.Dynatrace)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Dynatrace.MetricSelector = nil
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Dynatrace.DQL = &DynatraceDQL{
				Query:    test.query,
				Interval: "1m",
			}
			err := validate(slo)
			if test.isValid {
				testutils.AssertNoError(t, slo, err)
			} else {
				testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
					Prop: jsonpath.Parse("spec.objectives[0].rawMetric.query.dynatrace.dql.query"),
					Code: rules.ErrorCodeStringDenyRegexp,
				})
			}
		})
	}
}

func validDynatraceDQL() *DynatraceDQL {
	return &DynatraceDQL{
		Query:    "timeseries value = avg(dt.host.cpu.usage)",
		Interval: "1m",
	}
}
