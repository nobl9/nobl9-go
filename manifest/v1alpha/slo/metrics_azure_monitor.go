package slo

import (
	"regexp"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

const (
	AzureMonitorDataTypeMetrics = "metrics"
	AzureMonitorDataTypeLogs    = "logs"
)

// AzureMonitorMetric represents metric from AzureMonitor
type AzureMonitorMetric struct {
	DataType        string                                   `json:"dataType"`
	ResourceID      string                                   `json:"resourceId,omitempty"`
	MetricName      string                                   `json:"metricName,omitempty"`
	Aggregation     string                                   `json:"aggregation,omitempty"`
	Dimensions      []AzureMonitorMetricDimension            `json:"dimensions,omitempty"`
	MetricNamespace string                                   `json:"metricNamespace,omitempty"`
	Workspace       *AzureMonitorMetricLogAnalyticsWorkspace `json:"workspace,omitempty"`
	KQLQuery        string                                   `json:"kqlQuery,omitempty"`
}

// AzureMonitorMetricLogAnalyticsWorkspace represents Azure Log Analytics Workspace
type AzureMonitorMetricLogAnalyticsWorkspace struct {
	SubscriptionID string `json:"subscriptionId"`
	ResourceGroup  string `json:"resourceGroup"`
	WorkspaceID    string `json:"workspaceId"`
}

// AzureMonitorMetricDimension represents name/value pair that is part of the identity of a metric.
type AzureMonitorMetricDimension struct {
	Name  *string `json:"name"`
	Value *string `json:"value"`
}

var supportedAzureMonitorAggregations = []string{
	"Avg",
	"Min",
	"Max",
	"Count",
	"Sum",
}

var supportedAzureMonitorDataTypes = []string{
	AzureMonitorDataTypeMetrics,
	AzureMonitorDataTypeLogs,
}

func azureMonitorMetricDataTypeIs(dataType string) validation.Predicate[AzureMonitorMetric] {
	return func(m AzureMonitorMetric) bool {
		return m.DataType == dataType
	}
}

var azureMonitorResourceIDRegex = regexp.MustCompile(`^/subscriptions/[a-zA-Z0-9-]+/resourceGroups/[a-zA-Z0-9-._()]+/providers/[a-zA-Z0-9-.()_]+/[a-zA-Z0-9-_()]+/[a-zA-Z0-9-_()]+$`) //nolint:lll

var azureMonitorValidation = validation.New[AzureMonitorMetric](
	validation.For(validation.GetSelf[AzureMonitorMetric]()).
		Include(azureMonitorMetricDataTypeValidation).
		Include(azureMonitorMetricLogsDataTypeValidation),
	validation.For(func(p AzureMonitorMetric) string { return p.DataType }).
		WithName("dataType").
		Required().
		Rules(validation.OneOf(supportedAzureMonitorDataTypes...)),
)

var azureMonitorMetricDataTypeValidation = validation.New[AzureMonitorMetric](
	validation.For(func(a AzureMonitorMetric) string { return a.MetricName }).
		WithName("metricName").
		Required(),
	validation.For(func(a AzureMonitorMetric) string { return a.ResourceID }).
		WithName("resourceId").
		Required().
		Rules(validation.StringMatchRegexp(azureMonitorResourceIDRegex)),
	validation.For(func(a AzureMonitorMetric) string { return a.Aggregation }).
		WithName("aggregation").
		Required().
		Rules(validation.OneOf(supportedAzureMonitorAggregations...)),
	validation.ForEach(func(a AzureMonitorMetric) []AzureMonitorMetricDimension { return a.Dimensions }).
		WithName("dimensions").
		IncludeForEach(azureMonitorMetricDimensionValidation).
		// We don't want to check names uniqueness if they're empty.
		StopOnError().
		Rules(validation.SliceUnique(func(d AzureMonitorMetricDimension) string {
			if d.Name == nil {
				return ""
			}
			return *d.Name
		}).WithDetails("dimension 'name' must be unique for all dimensions")),
).When(azureMonitorMetricDataTypeIs(AzureMonitorDataTypeMetrics))

var azureMonitorMetricLogAnalyticsWorkspaceValidation = validation.New[AzureMonitorMetricLogAnalyticsWorkspace](
	validation.For(func(a AzureMonitorMetricLogAnalyticsWorkspace) string { return a.SubscriptionID }).
		WithName("subscriptionId").
		OmitEmpty().
		Rules(validation.StringUUID()),
	validation.For(func(a AzureMonitorMetricLogAnalyticsWorkspace) string { return a.WorkspaceID }).
		WithName("workspaceId").
		Required().
		Rules(validation.StringUUID()),
)

var azureMonitorMetricLogsDataTypeValidation = validation.New[AzureMonitorMetric](
	validation.ForPointer(func(a AzureMonitorMetric) *AzureMonitorMetricLogAnalyticsWorkspace { return a.Workspace }).
		WithName("workspace").
		Required().
		Include(azureMonitorMetricLogAnalyticsWorkspaceValidation),
	validation.For(func(a AzureMonitorMetric) string { return a.KQLQuery }).
		WithName("kqlQuery").
		Required().
		Rules(
			validation.StringMatchRegexp(regexp.MustCompile(`(?m)\bn9_time\b`)).
				WithDetails("n9_time is required"),
			validation.StringMatchRegexp(regexp.MustCompile(`(?m)\bn9_value\b`)).
				WithDetails("n9_value is required"),
		),
).When(azureMonitorMetricDataTypeIs(AzureMonitorDataTypeLogs))

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
				if good.AzureMonitor.DataType != total.AzureMonitor.DataType {
					return countMetricsPropertyEqualityError("azureMonitor.dataType", goodMetric)
				}
				if good.AzureMonitor.Workspace != nil && total.AzureMonitor.Workspace != nil &&
					good.AzureMonitor.Workspace.WorkspaceID != total.AzureMonitor.Workspace.WorkspaceID {
					return countMetricsPropertyEqualityError("azureMonitor.workspace.workspaceId", goodMetric)
				}
				if good.AzureMonitor.MetricNamespace != total.AzureMonitor.MetricNamespace {
					return countMetricsPropertyEqualityError("azureMonitor.metricNamespace", goodMetric)
				}
				if good.AzureMonitor.ResourceID != total.AzureMonitor.ResourceID {
					return countMetricsPropertyEqualityError("azureMonitor.resourceId", goodMetric)
				}
			}
			if bad != nil {
				if bad.AzureMonitor.DataType != total.AzureMonitor.DataType {
					return countMetricsPropertyEqualityError("azureMonitor.dataType", badMetric)
				}
				if bad.AzureMonitor.Workspace != nil && total.AzureMonitor.Workspace != nil &&
					bad.AzureMonitor.Workspace.WorkspaceID != total.AzureMonitor.Workspace.WorkspaceID {
					return countMetricsPropertyEqualityError("azureMonitor.workspace.workspaceId", badMetric)
				}
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
