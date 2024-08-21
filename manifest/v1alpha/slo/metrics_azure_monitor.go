package slo

import (
	"regexp"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
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

var azureMonitorResourceIDRegex = regexp.MustCompile(`^/subscriptions/[a-zA-Z0-9-]+/resourceGroups/[a-zA-Z0-9-._()]+/providers/[a-zA-Z0-9-.()_]+/[a-zA-Z0-9-_()]+/[a-zA-Z0-9-_()]+$`) //nolint:lll

var azureMonitorValidation = govy.New(
	govy.For(govy.GetSelf[AzureMonitorMetric]()).
		Include(azureMonitorMetricDataTypeValidation).
		Include(azureMonitorMetricLogsDataTypeValidation),
	govy.For(func(p AzureMonitorMetric) string { return p.DataType }).
		WithName("dataType").
		Required().
		Rules(rules.OneOf(supportedAzureMonitorDataTypes...)),
)

var azureMonitorMetricDataTypeValidation = govy.New(
	govy.For(func(a AzureMonitorMetric) string { return a.MetricName }).
		WithName("metricName").
		Required(),
	govy.For(func(a AzureMonitorMetric) string { return a.ResourceID }).
		WithName("resourceId").
		Required().
		Rules(rules.StringMatchRegexp(azureMonitorResourceIDRegex)),
	govy.For(func(a AzureMonitorMetric) string { return a.Aggregation }).
		WithName("aggregation").
		Required().
		Rules(rules.OneOf(supportedAzureMonitorAggregations...)),
	govy.ForSlice(func(a AzureMonitorMetric) []AzureMonitorMetricDimension { return a.Dimensions }).
		WithName("dimensions").
		IncludeForEach(azureMonitorMetricDimensionValidation).
		Rules(rules.SliceUnique(func(d AzureMonitorMetricDimension) string {
			if d.Name == nil {
				return ""
			}
			return *d.Name
		}).WithDetails("dimension 'name' must be unique for all dimensions")),
	govy.ForPointer(func(a AzureMonitorMetric) *AzureMonitorMetricLogAnalyticsWorkspace { return a.Workspace }).
		WithName("workspace").
		Rules(rules.Forbidden[AzureMonitorMetricLogAnalyticsWorkspace]()),
	govy.For(func(a AzureMonitorMetric) string { return a.KQLQuery }).
		WithName("kqlQuery").
		Rules(rules.Forbidden[string]()),
).When(
	func(a AzureMonitorMetric) bool { return a.DataType == AzureMonitorDataTypeMetrics },
	govy.WhenDescription("dataType is '%s'", AzureMonitorDataTypeMetrics),
)

var validAzureResourceGroupRegex = regexp.MustCompile(`^[a-zA-Z0-9-._()]+$`)

var azureMonitorMetricLogAnalyticsWorkspaceValidation = govy.New(
	govy.For(func(a AzureMonitorMetricLogAnalyticsWorkspace) string { return a.SubscriptionID }).
		WithName("subscriptionId").
		Required().
		Rules(rules.StringUUID()),
	govy.For(func(a AzureMonitorMetricLogAnalyticsWorkspace) string { return a.ResourceGroup }).
		WithName("resourceGroup").
		Required().
		Rules(rules.StringMatchRegexp(validAzureResourceGroupRegex)),
	govy.For(func(a AzureMonitorMetricLogAnalyticsWorkspace) string { return a.WorkspaceID }).
		WithName("workspaceId").
		Required().
		Rules(rules.StringUUID()),
)

var azureMonitorMetricLogsDataTypeValidation = govy.New(
	govy.ForPointer(func(a AzureMonitorMetric) *AzureMonitorMetricLogAnalyticsWorkspace { return a.Workspace }).
		WithName("workspace").
		Required().
		Include(azureMonitorMetricLogAnalyticsWorkspaceValidation),
	govy.For(func(a AzureMonitorMetric) string { return a.KQLQuery }).
		WithName("kqlQuery").
		Required().
		Rules(
			rules.StringMatchRegexp(regexp.MustCompile(`(?m)\bn9_time\b`)).
				WithDetails("n9_time is required"),
			rules.StringMatchRegexp(regexp.MustCompile(`(?m)\bn9_value\b`)).
				WithDetails("n9_value is required"),
		),
	govy.For(func(a AzureMonitorMetric) string { return a.ResourceID }).
		WithName("resourceId").
		Rules(rules.Forbidden[string]()),
	govy.For(func(a AzureMonitorMetric) string { return a.MetricNamespace }).
		WithName("metricNamespace").
		Rules(rules.Forbidden[string]()),
	govy.For(func(a AzureMonitorMetric) string { return a.MetricName }).
		WithName("metricName").
		Rules(rules.Forbidden[string]()),
	govy.For(func(a AzureMonitorMetric) string { return a.Aggregation }).
		WithName("aggregation").
		Rules(rules.Forbidden[string]()),
	govy.ForSlice(func(a AzureMonitorMetric) []AzureMonitorMetricDimension { return a.Dimensions }).
		WithName("dimensions").
		Rules(rules.Forbidden[[]AzureMonitorMetricDimension]()),
).When(
	func(a AzureMonitorMetric) bool { return a.DataType == AzureMonitorDataTypeLogs },
	govy.WhenDescription("dataType is '%s'", AzureMonitorDataTypeLogs),
)

var azureMonitorMetricDimensionValidation = govy.New(
	govy.ForPointer(func(a AzureMonitorMetricDimension) *string { return a.Name }).
		WithName("name").
		Required().
		Rules(
			rules.StringNotEmpty(),
			rules.StringMaxLength(255),
			rules.StringASCII()),
	govy.ForPointer(func(a AzureMonitorMetricDimension) *string { return a.Value }).
		WithName("value").
		Required().
		Rules(
			rules.StringNotEmpty(),
			rules.StringMaxLength(255),
			rules.StringASCII()),
)

var azureMonitorCountMetricsLevelValidation = govy.New(
	govy.For(govy.GetSelf[CountMetricsSpec]()).Rules(
		govy.NewRule(func(c CountMetricsSpec) error {
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
				if good.AzureMonitor.Workspace != nil && total.AzureMonitor.Workspace != nil {
					if good.AzureMonitor.Workspace.SubscriptionID != total.AzureMonitor.Workspace.SubscriptionID {
						return countMetricsPropertyEqualityError("azureMonitor.workspace.subscriptionId", goodMetric)
					}
					if good.AzureMonitor.Workspace.ResourceGroup != total.AzureMonitor.Workspace.ResourceGroup {
						return countMetricsPropertyEqualityError("azureMonitor.workspace.resourceGroup", goodMetric)
					}
					if good.AzureMonitor.Workspace.WorkspaceID != total.AzureMonitor.Workspace.WorkspaceID {
						return countMetricsPropertyEqualityError("azureMonitor.workspace.workspaceId", goodMetric)
					}
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
				if bad.AzureMonitor.Workspace != nil && total.AzureMonitor.Workspace != nil {
					if bad.AzureMonitor.Workspace.SubscriptionID != total.AzureMonitor.Workspace.SubscriptionID {
						return countMetricsPropertyEqualityError("azureMonitor.workspace.subscriptionId", badMetric)
					}
					if bad.AzureMonitor.Workspace.ResourceGroup != total.AzureMonitor.Workspace.ResourceGroup {
						return countMetricsPropertyEqualityError("azureMonitor.workspace.resourceGroup", badMetric)
					}
					if bad.AzureMonitor.Workspace.WorkspaceID != total.AzureMonitor.Workspace.WorkspaceID {
						return countMetricsPropertyEqualityError("azureMonitor.workspace.workspaceId", badMetric)
					}
				}
				if bad.AzureMonitor.MetricNamespace != total.AzureMonitor.MetricNamespace {
					return countMetricsPropertyEqualityError("azureMonitor.metricNamespace", badMetric)
				}
				if bad.AzureMonitor.ResourceID != total.AzureMonitor.ResourceID {
					return countMetricsPropertyEqualityError("azureMonitor.resourceId", badMetric)
				}
			}
			return nil
		}).WithErrorCode(rules.ErrorCodeNotEqualTo)),
).When(
	whenCountMetricsIs(v1alpha.AzureMonitor),
	govy.WhenDescription("countMetrics is azureMonitor"),
)
