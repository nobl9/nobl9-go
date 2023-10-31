package slo

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

func TestAzureMonitor_CounMetrics(t *testing.T) {
	t.Run("metricNamespace must be the same for good/bad and total", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.AzureMonitor)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.AzureMonitor = &AzureMonitorMetric{
			ResourceID:      "/subscriptions/1/resourceGroups/azure-monitor-test-sources/providers/Microsoft.Web/sites/app",
			MetricName:      "HttpResponseTime",
			Aggregation:     "Avg",
			MetricNamespace: "This",
		}
		// Good.
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.AzureMonitor = &AzureMonitorMetric{
			ResourceID:      "/subscriptions/1/resourceGroups/azure-monitor-test-sources/providers/Microsoft.Web/sites/app",
			MetricName:      "HttpResponseTime",
			Aggregation:     "Avg",
			MetricNamespace: "That",
		}
		err := validate(slo)
		assertContainsErrors(t, err, 1, expectedError{
			Prop:    "spec.objectives[0].countMetrics",
			Message: "'azureMonitor.metricNamespace' must be the same for both 'good' and 'total' metrics",
		})
		// Bad.
		slo.Spec.Objectives[0].CountMetrics.BadMetric = slo.Spec.Objectives[0].CountMetrics.GoodMetric
		slo.Spec.Objectives[0].CountMetrics.GoodMetric = nil
		err = validate(slo)
		assertContainsErrors(t, err, 1, expectedError{
			Prop:    "spec.objectives[0].countMetrics",
			Message: "'azureMonitor.metricNamespace' must be the same for both 'bad' and 'total' metrics",
		})
	})
	t.Run("resourceId must be the same for good/bad and total", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.AzureMonitor)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.AzureMonitor = &AzureMonitorMetric{
			ResourceID:  "/subscriptions/123/resourceGroups/azure-monitor-test-sources/providers/Microsoft.Web/sites/app",
			MetricName:  "HttpResponseTime",
			Aggregation: "Avg",
		}
		// Good.
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.AzureMonitor = &AzureMonitorMetric{
			ResourceID:  "/subscriptions/333/resourceGroups/azure-monitor-test-sources/providers/Microsoft.Web/sites/app",
			MetricName:  "HttpResponseTime",
			Aggregation: "Avg",
		}
		err := validate(slo)
		assertContainsErrors(t, err, 1, expectedError{
			Prop:    "spec.objectives[0].countMetrics",
			Message: "'azureMonitor.resourceId' must be the same for both 'good' and 'total' metrics",
		})
		// Bad.
		slo.Spec.Objectives[0].CountMetrics.BadMetric = slo.Spec.Objectives[0].CountMetrics.GoodMetric
		slo.Spec.Objectives[0].CountMetrics.GoodMetric = nil
		err = validate(slo)
		assertContainsErrors(t, err, 1, expectedError{
			Prop:    "spec.objectives[0].countMetrics",
			Message: "'azureMonitor.resourceId' must be the same for both 'bad' and 'total' metrics",
		})
	})
}

func TestAzureMonitor(t *testing.T) {
	t.Run("required fields", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.AzureMonitor)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor = &AzureMonitorMetric{
			ResourceID:  "",
			MetricName:  "",
			Aggregation: "",
		}
		err := validate(slo)
		assertContainsErrors(t, err, 3,
			expectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.resourceId",
				Code: validation.ErrorCodeRequired,
			},
			expectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.metricName",
				Code: validation.ErrorCodeRequired,
			},
			expectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.aggregation",
				Code: validation.ErrorCodeRequired,
			},
		)
	})
	t.Run("valid aggregations", func(t *testing.T) {
		for _, agg := range supportedAzureMonitorAggregations {
			slo := validRawMetricSLO(v1alpha.AzureMonitor)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor.Aggregation = agg
			err := validate(slo)
			assert.Empty(t, err)
		}
	})
	t.Run("invalid aggregations", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.AzureMonitor)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor.Aggregation = "invalid"
		err := validate(slo)
		assertContainsErrors(t, err, 1, expectedError{
			Prop: "spec.objectives[0].rawMetric.query.azureMonitor.aggregation",
			Code: validation.ErrorCodeOneOf,
		})
	})
}

func TestAzureMonitorDimension(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.AzureMonitor)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor.Dimensions = []AzureMonitorMetricDimension{
			{
				Name:  ptr("that"),
				Value: ptr("value-1"),
			},
			{
				Name:  ptr("this"),
				Value: ptr("value-2"),
			},
		}
		err := validate(slo)
		assert.Empty(t, err)
	})
	t.Run("invalid fields", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.AzureMonitor)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor.Dimensions = []AzureMonitorMetricDimension{
			{},
			{
				Name:  ptr(""),
				Value: ptr(""),
			},
			{
				Name:  ptr(strings.Repeat("l", 256)),
				Value: ptr(strings.Repeat("l", 256)),
			},
			{
				Name:  ptr("ｶﾀｶﾅ"),
				Value: ptr("ｶﾀｶﾅ"),
			},
		}
		err := validate(slo)
		assertContainsErrors(t, err, 8,
			expectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.dimensions[0].name",
				Code: validation.ErrorCodeRequired,
			},
			expectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.dimensions[0].value",
				Code: validation.ErrorCodeRequired,
			},
			expectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.dimensions[1].name",
				Code: validation.ErrorCodeStringNotEmpty,
			},
			expectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.dimensions[1].value",
				Code: validation.ErrorCodeStringNotEmpty,
			},
			expectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.dimensions[2].name",
				Code: validation.ErrorCodeStringMaxLength,
			},
			expectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.dimensions[2].value",
				Code: validation.ErrorCodeStringMaxLength,
			},
			expectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.dimensions[3].name",
				Code: validation.ErrorCodeStringASCII,
			},
			expectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.dimensions[3].value",
				Code: validation.ErrorCodeStringASCII,
			},
		)
	})
	t.Run("unique names", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.AzureMonitor)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor.Dimensions = []AzureMonitorMetricDimension{
			{
				Name:  ptr("this"),
				Value: ptr("value"),
			},
			{
				Name:  ptr("this"),
				Value: ptr("val"),
			},
		}
		err := validate(slo)
		assertContainsErrors(t, err, 1, expectedError{
			Prop: "spec.objectives[0].rawMetric.query.azureMonitor.dimensions",
			Code: validation.ErrorCodeSliceUnique,
		})
	})
}
