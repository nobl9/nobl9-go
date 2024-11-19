package slo

import (
	"testing"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func Test_whenCountMetricsIs(t *testing.T) {
	testCases := map[string]struct {
		datasource v1alpha.DataSourceType
		spec       CountMetricsSpec
		expected   bool
	}{
		"true - splunk - single query": {
			datasource: v1alpha.Splunk,
			spec: CountMetricsSpec{
				Incremental:     ptr(false),
				GoodTotalMetric: validSingleQueryMetricSpec(v1alpha.Splunk),
			},
			expected: true,
		},
		"false - splunk mixed with honeycomb - single query": {
			datasource: v1alpha.Splunk,
			spec: CountMetricsSpec{
				Incremental:     ptr(false),
				GoodTotalMetric: validSingleQueryMetricSpec(v1alpha.Honeycomb),
			},
			expected: false,
		},
		"true - cloudwatch": {
			datasource: v1alpha.CloudWatch,
			spec: CountMetricsSpec{
				Incremental: ptr(false),
				GoodMetric:  validMetricSpec(v1alpha.CloudWatch),
				TotalMetric: validMetricSpec(v1alpha.CloudWatch),
			},
			expected: true,
		},
		"true - cloudwatch - bad over total": {
			datasource: v1alpha.CloudWatch,
			spec: CountMetricsSpec{
				Incremental: ptr(false),
				BadMetric:   validMetricSpec(v1alpha.CloudWatch),
				TotalMetric: validMetricSpec(v1alpha.CloudWatch),
			},
			expected: true,
		},
		"false - newrelic - bad over total": {
			datasource: v1alpha.NewRelic,
			spec: CountMetricsSpec{
				Incremental: ptr(false),
				BadMetric:   validMetricSpec(v1alpha.NewRelic),
				TotalMetric: validMetricSpec(v1alpha.NewRelic),
			},
			expected: false,
		},
		"true - bigquery": {
			datasource: v1alpha.BigQuery,
			spec: CountMetricsSpec{
				Incremental: ptr(false),
				GoodMetric:  validMetricSpec(v1alpha.BigQuery),
				TotalMetric: validMetricSpec(v1alpha.BigQuery),
			},
			expected: true,
		},
		"false - mixed bigquery and cloudwatch": {
			datasource: v1alpha.CloudWatch,
			spec: CountMetricsSpec{
				Incremental: ptr(false),
				GoodMetric:  validMetricSpec(v1alpha.BigQuery),
				TotalMetric: validMetricSpec(v1alpha.BigQuery),
			},
			expected: false,
		},
	}

	for name, tc := range testCases {
		result := whenCountMetricsIs(tc.datasource)(tc.spec)
		if result != tc.expected {
			t.Errorf("%s: expected %v, got %v", name, tc.expected, result)
		}
	}
}
