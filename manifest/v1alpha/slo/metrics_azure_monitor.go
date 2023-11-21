package slo

import (
	"regexp"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

// AzureMonitorMetric represents metric from AzureMonitor
type AzureMonitorMetric struct {
	ResourceID      string                        `json:"resourceId"`
	MetricName      string                        `json:"metricName"`
	Aggregation     string                        `json:"aggregation"`
	Dimensions      []AzureMonitorMetricDimension `json:"dimensions,omitempty"`
	MetricNamespace string                        `json:"metricNamespace,omitempty"`
}

// AzureMonitorMetricDimension represents name/value pair that is part of the identity of a metric.
type AzureMonitorMetricDimension struct {
	Name  *string `json:"name"`
	Value *string `json:"value"`
}

var azureMonitorCountMetricsLevelValidation = validation.New[CountMetricsSpec](
	validation.For(validation.GetSelf[CountMetricsSpec]()).Rules(
		validation.NewSingleRule(func(c CountMetricsSpec) error {
			total := c.TotalMetric
			good := c.GoodMetric
			bad := c.BadMetric

			if total == nil {
				return nil
			}
			if good != nil {
				if good.AzureMonitor.MetricNamespace != total.AzureMonitor.MetricNamespace {
					return countMetricsPropertyEqualityError("azureMonitor.metricNamespace", goodMetric)
				}
				if good.AzureMonitor.ResourceID != total.AzureMonitor.ResourceID {
					return countMetricsPropertyEqualityError("azureMonitor.resourceId", goodMetric)
				}
			}
			if bad != nil {
				if bad.AzureMonitor.MetricNamespace != total.AzureMonitor.MetricNamespace {
					return countMetricsPropertyEqualityError("azureMonitor.metricNamespace", badMetric)
				}
				if bad.AzureMonitor.ResourceID != total.AzureMonitor.ResourceID {
					return countMetricsPropertyEqualityError("azureMonitor.resourceId", badMetric)
				}
			}
			return nil
		}).WithErrorCode(validation.ErrorCodeNotEqualTo)),
).When(whenCountMetricsIs(v1alpha.AzureMonitor))

var supportedAzureMonitorAggregations = []string{
	"Avg",
	"Min",
	"Max",
	"Count",
	"Sum",
}

var azureMonitorResourceIDRegexp = regexp.MustCompile(`^\/subscriptions\/[a-zA-Z0-9-]+\/resourceGroups\/[a-zA-Z0-9-._()]+\/providers\/[a-zA-Z0-9-.()_]+\/[a-zA-Z0-9-_()]+\/[a-zA-Z0-9-_()]+$`) //nolint:lll

var azureMonitorValidation = validation.New[AzureMonitorMetric](
	validation.For(func(a AzureMonitorMetric) string { return a.MetricName }).
		WithName("metricName").
		Required(),
	validation.For(func(a AzureMonitorMetric) string { return a.ResourceID }).
		WithName("resourceId").
		Required().
		Rules(validation.StringMatchRegexp(azureMonitorResourceIDRegexp)),
	validation.For(func(a AzureMonitorMetric) string { return a.Aggregation }).
		WithName("aggregation").
		Required().
		Rules(validation.OneOf(supportedAzureMonitorAggregations...)),
	validation.ForEach(func(a AzureMonitorMetric) []AzureMonitorMetricDimension { return a.Dimensions }).
		WithName("dimensions").
		IncludeForEach(azureMonitorMetricDimensionValidation).
		// We don't want to check names uniqueness if for exsample names are empty.
		StopOnError().
		Rules(validation.SliceUnique(func(d AzureMonitorMetricDimension) string {
			if d.Name == nil {
				return ""
			}
			return *d.Name
		}).WithDetails("dimension 'name' must be unique for all dimensions")),
)

var azureMonitorMetricDimensionValidation = validation.New[AzureMonitorMetricDimension](
	validation.ForPointer(func(a AzureMonitorMetricDimension) *string { return a.Name }).
		WithName("name").
		Required().
		Rules(
			validation.StringNotEmpty(),
			validation.StringMaxLength(255),
			validation.StringASCII()),
	validation.ForPointer(func(a AzureMonitorMetricDimension) *string { return a.Value }).
		WithName("value").
		Required().
		Rules(
			validation.StringNotEmpty(),
			validation.StringMaxLength(255),
			validation.StringASCII()),
)
