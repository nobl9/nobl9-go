package slo

import (
	"strings"
	"testing"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestAzureMonitor_CountMetrics(t *testing.T) {
	t.Run("metricNamespace must be the same for good/bad and total", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.AzureMonitor)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.AzureMonitor = &AzureMonitorMetric{
			DataType:        AzureMonitorDataTypeMetrics,
			ResourceID:      "/subscriptions/1/resourceGroups/azure-monitor-test-sources/providers/Microsoft.Web/sites/app",
			MetricName:      "HttpResponseTime",
			Aggregation:     "Avg",
			MetricNamespace: "This",
		}
		// Good.
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.AzureMonitor = &AzureMonitorMetric{
			DataType:        AzureMonitorDataTypeMetrics,
			ResourceID:      "/subscriptions/1/resourceGroups/azure-monitor-test-sources/providers/Microsoft.Web/sites/app",
			MetricName:      "HttpResponseTime",
			Aggregation:     "Avg",
			MetricNamespace: "That",
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop:    "spec.objectives[0].countMetrics",
			Message: "'azureMonitor.metricNamespace' must be the same for both 'good' and 'total' metrics",
		})
		// Bad.
		slo.Spec.Objectives[0].CountMetrics.BadMetric = slo.Spec.Objectives[0].CountMetrics.GoodMetric
		slo.Spec.Objectives[0].CountMetrics.GoodMetric = nil
		err = validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop:    "spec.objectives[0].countMetrics",
			Message: "'azureMonitor.metricNamespace' must be the same for both 'bad' and 'total' metrics",
		})
	})
	t.Run("resourceId must be the same for good/bad and total", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.AzureMonitor)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.AzureMonitor = &AzureMonitorMetric{
			DataType:    AzureMonitorDataTypeMetrics,
			ResourceID:  "/subscriptions/123/resourceGroups/azure-monitor-test-sources/providers/Microsoft.Web/sites/app",
			MetricName:  "HttpResponseTime",
			Aggregation: "Avg",
		}
		// Good.
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.AzureMonitor = &AzureMonitorMetric{
			DataType:    AzureMonitorDataTypeMetrics,
			ResourceID:  "/subscriptions/333/resourceGroups/azure-monitor-test-sources/providers/Microsoft.Web/sites/app",
			MetricName:  "HttpResponseTime",
			Aggregation: "Avg",
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop:    "spec.objectives[0].countMetrics",
			Message: "'azureMonitor.resourceId' must be the same for both 'good' and 'total' metrics",
		})
		// Bad.
		slo.Spec.Objectives[0].CountMetrics.BadMetric = slo.Spec.Objectives[0].CountMetrics.GoodMetric
		slo.Spec.Objectives[0].CountMetrics.GoodMetric = nil
		err = validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop:    "spec.objectives[0].countMetrics",
			Message: "'azureMonitor.resourceId' must be the same for both 'bad' and 'total' metrics",
		})
	})
	t.Run("dataType must be the same for good/bad and total", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.AzureMonitor)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.AzureMonitor = getValidAzureMetric(AzureMonitorDataTypeMetrics)
		// Good.
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.AzureMonitor = getValidAzureMetric(AzureMonitorDataTypeLogs)
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop:    "spec.objectives[0].countMetrics",
			Message: "'azureMonitor.dataType' must be the same for both 'good' and 'total' metrics",
		})
		// Bad.
		slo.Spec.Objectives[0].CountMetrics.BadMetric = slo.Spec.Objectives[0].CountMetrics.GoodMetric
		slo.Spec.Objectives[0].CountMetrics.GoodMetric = nil
		err = validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop:    "spec.objectives[0].countMetrics",
			Message: "'azureMonitor.dataType' must be the same for both 'bad' and 'total' metrics",
		})
	})
	t.Run("workspace.subscriptionId must be the same for good/bad and total", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.AzureMonitor)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.AzureMonitor = getValidAzureMetric(AzureMonitorDataTypeLogs)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.AzureMonitor.Workspace.
			SubscriptionID = "44444444-4444-4444-4444-444444444444"
		// Good.
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.AzureMonitor = getValidAzureMetric(AzureMonitorDataTypeLogs)
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.AzureMonitor.Workspace.
			SubscriptionID = "11111111-1111-1111-1111-111111111111"
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop:    "spec.objectives[0].countMetrics",
			Message: "'azureMonitor.workspace.subscriptionId' must be the same for both 'good' and 'total' metrics",
		})
		// Bad.
		slo.Spec.Objectives[0].CountMetrics.BadMetric = slo.Spec.Objectives[0].CountMetrics.GoodMetric
		slo.Spec.Objectives[0].CountMetrics.GoodMetric = nil
		err = validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop:    "spec.objectives[0].countMetrics",
			Message: "'azureMonitor.workspace.subscriptionId' must be the same for both 'bad' and 'total' metrics",
		})
	})
	t.Run("workspace.resourceGroup must be the same for good/bad and total", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.AzureMonitor)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.AzureMonitor = getValidAzureMetric(AzureMonitorDataTypeLogs)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.AzureMonitor.Workspace.
			ResourceGroup = "rg-1"
		// Good.
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.AzureMonitor = getValidAzureMetric(AzureMonitorDataTypeLogs)
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.AzureMonitor.Workspace.
			ResourceGroup = "rg-2"
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop:    "spec.objectives[0].countMetrics",
			Message: "'azureMonitor.workspace.resourceGroup' must be the same for both 'good' and 'total' metrics",
		})
		// Bad.
		slo.Spec.Objectives[0].CountMetrics.BadMetric = slo.Spec.Objectives[0].CountMetrics.GoodMetric
		slo.Spec.Objectives[0].CountMetrics.GoodMetric = nil
		err = validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop:    "spec.objectives[0].countMetrics",
			Message: "'azureMonitor.workspace.resourceGroup' must be the same for both 'bad' and 'total' metrics",
		})
	})
	t.Run("workspace.workspaceId must be the same for good/bad and total", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.AzureMonitor)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.AzureMonitor = getValidAzureMetric(AzureMonitorDataTypeLogs)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.AzureMonitor.Workspace.
			WorkspaceID = "00000000-0000-0000-0000-000000000000"
		// Good.
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.AzureMonitor = getValidAzureMetric(AzureMonitorDataTypeLogs)
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.AzureMonitor.Workspace.
			WorkspaceID = "11111111-1111-1111-1111-111111111111"
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop:    "spec.objectives[0].countMetrics",
			Message: "'azureMonitor.workspace.workspaceId' must be the same for both 'good' and 'total' metrics",
		})
		// Bad.
		slo.Spec.Objectives[0].CountMetrics.BadMetric = slo.Spec.Objectives[0].CountMetrics.GoodMetric
		slo.Spec.Objectives[0].CountMetrics.GoodMetric = nil
		err = validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop:    "spec.objectives[0].countMetrics",
			Message: "'azureMonitor.workspace.workspaceId' must be the same for both 'bad' and 'total' metrics",
		})
	})
}

func TestAzureMonitor_DataType(t *testing.T) {
	t.Run("required", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.AzureMonitor)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor = &AzureMonitorMetric{
			DataType: "",
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.dataType",
				Code: rules.ErrorCodeRequired,
			},
		)
	})
	t.Run("valid", func(t *testing.T) {
		for _, dt := range supportedAzureMonitorDataTypes {
			slo := validRawMetricSLO(v1alpha.AzureMonitor)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor = getValidAzureMetric(dt)
			err := validate(slo)
			testutils.AssertNoError(t, slo, err)
		}
	})
	t.Run("invalid", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.AzureMonitor)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor.DataType = "invalid"
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.azureMonitor.dataType",
			Code: rules.ErrorCodeOneOf,
		})
	})
}

func TestAzureMonitor_LogsDataType(t *testing.T) {
	t.Run("required fields", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.AzureMonitor)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor = &AzureMonitorMetric{
			DataType:  AzureMonitorDataTypeLogs,
			Workspace: nil,
			KQLQuery:  "",
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 2,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.workspace",
				Code: rules.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.kqlQuery",
				Code: rules.ErrorCodeRequired,
			},
		)
	})
	t.Run("required query token - n9_time", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.AzureMonitor)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor = getValidAzureMetric(AzureMonitorDataTypeLogs)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.
			AzureMonitor.KQLQuery = "logs | summarize n9_value = sum(value) | project TimeGenerated as n9_missing_time, 1 as n9_value"
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.kqlQuery",
				Code: rules.ErrorCodeStringMatchRegexp,
			},
		)
	})
	t.Run("required query token - n9_value", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.AzureMonitor)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor = getValidAzureMetric(AzureMonitorDataTypeLogs)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.
			AzureMonitor.KQLQuery = "logs | summarize n9_val = sum(value) | project TimeGenerated as n9_time, 1 as n9_missing_value"
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.kqlQuery",
				Code: rules.ErrorCodeStringMatchRegexp,
			},
		)
	})
}

func TestAzureMonitor_MetricDataType(t *testing.T) {
	t.Run("required fields", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.AzureMonitor)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor = &AzureMonitorMetric{
			DataType:    AzureMonitorDataTypeMetrics,
			ResourceID:  "",
			MetricName:  "",
			Aggregation: "",
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 3,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.resourceId",
				Code: rules.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.metricName",
				Code: rules.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.aggregation",
				Code: rules.ErrorCodeRequired,
			},
		)
	})
	t.Run("forbidden fields", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.AzureMonitor)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor = getValidAzureMetric(AzureMonitorDataTypeMetrics)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor.Workspace = &AzureMonitorMetricLogAnalyticsWorkspace{
			SubscriptionID: "00000000-0000-0000-0000-000000000000",
			ResourceGroup:  "rg1",
			WorkspaceID:    "00000000-0000-0000-0000-000000000001",
		}
		slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor.KQLQuery =
			"logs | project TimeGenerated as n9_time, 1 as n9_value"
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 2,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.workspace",
				Code: rules.ErrorCodeForbidden,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.kqlQuery",
				Code: rules.ErrorCodeForbidden,
			},
		)
	})
	t.Run("valid aggregations", func(t *testing.T) {
		for _, agg := range supportedAzureMonitorAggregations {
			slo := validRawMetricSLO(v1alpha.AzureMonitor)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor.Aggregation = agg
			err := validate(slo)
			testutils.AssertNoError(t, slo, err)
		}
	})
	t.Run("invalid aggregations", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.AzureMonitor)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor.Aggregation = "invalid"
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.azureMonitor.aggregation",
			Code: rules.ErrorCodeOneOf,
		})
	})
}

func TestAzureMonitorLogAnalyticsWorkspace(t *testing.T) {
	t.Run("required fields", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.AzureMonitor)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor = getValidAzureMetric(AzureMonitorDataTypeLogs)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor.Workspace = &AzureMonitorMetricLogAnalyticsWorkspace{
			SubscriptionID: "",
			ResourceGroup:  "",
			WorkspaceID:    "",
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 3,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.workspace.subscriptionId",
				Code: rules.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.workspace.resourceGroup",
				Code: rules.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.workspace.workspaceId",
				Code: rules.ErrorCodeRequired,
			})
	})
	t.Run("forbidden fields", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.AzureMonitor)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor = getValidAzureMetric(AzureMonitorDataTypeLogs)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor.ResourceID =
			"/subscriptions/1/resourceGroups/azure-monitor-test-sources/providers/Microsoft.Web/sites/app"
		slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor.MetricName = "HttpResponseTime"
		slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor.Aggregation = "Avg"
		slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor.MetricNamespace = "Microsoft.Web/sites"
		slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor.Dimensions = []AzureMonitorMetricDimension{
			{
				Name:  ptr("that"),
				Value: ptr("value-1"),
			},
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 5,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.resourceId",
				Code: rules.ErrorCodeForbidden,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.metricName",
				Code: rules.ErrorCodeForbidden,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.metricNamespace",
				Code: rules.ErrorCodeForbidden,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.dimensions",
				Code: rules.ErrorCodeForbidden,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.aggregation",
				Code: rules.ErrorCodeForbidden,
			})
	})
	t.Run("subscriptionId must be uuid if defined validation", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.AzureMonitor)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor = getValidAzureMetric(AzureMonitorDataTypeLogs)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor.Workspace.SubscriptionID = "invalid"
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.workspace.subscriptionId",
				Code: rules.ErrorCodeStringUUID,
			})
	})
	t.Run("resourceGroup must match regex if defined validation", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.AzureMonitor)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor = getValidAzureMetric(AzureMonitorDataTypeLogs)
		testCases := []struct {
			desc          string
			resourceGroup string
			isValid       bool
		}{
			{
				desc:          "unsupported character",
				resourceGroup: "azure-monitor-!test-sources",
				isValid:       false,
			},
			{
				desc:          "spaces",
				resourceGroup: "azure-monitor test-sources",
				isValid:       false,
			},
			{
				desc:          "valid azure resource group 1",
				resourceGroup: "azure-monitor-test-sources",
				isValid:       true,
			},
			{
				desc:          "valid azure resource group 2",
				resourceGroup: "azure-monitor-test-source)s",
				isValid:       true,
			},
			{
				desc:          "valid azure resource group 3",
				resourceGroup: "MC-azure-monitor-test-sources_aks-cluster_west_europe",
				isValid:       true,
			},
		}
		for _, tC := range testCases {
			t.Run(tC.desc, func(t *testing.T) {
				slo := validRawMetricSLO(v1alpha.AzureMonitor)
				slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor = getValidAzureMetric(AzureMonitorDataTypeLogs)
				slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor.Workspace.ResourceGroup = tC.resourceGroup
				err := validate(slo)
				if tC.isValid {
					testutils.AssertNoError(t, slo, err)
				} else {
					testutils.AssertContainsErrors(t, slo, err, 1,
						testutils.ExpectedError{
							Prop: "spec.objectives[0].rawMetric.query.azureMonitor.workspace.resourceGroup",
							Code: rules.ErrorCodeStringMatchRegexp,
						})
				}
			})
		}
	})
	t.Run("workspaceId must be uuid validation", func(t *testing.T) {
		testCases := []struct {
			desc        string
			workspaceId string
			isValid     bool
		}{
			{
				desc:        "one letter",
				workspaceId: "a",
				isValid:     false,
			},
			{
				desc:        "non hex number used",
				workspaceId: "XXXXXXX-0000-0000-0000-00000000000",
				isValid:     false,
			},
			{
				desc:        "to short",
				workspaceId: "0000000-0000-0000-0000-00000000000",
				isValid:     false,
			},
			{
				desc:        "valid rfc4122 uuid",
				workspaceId: "00000000-0000-0000-0000-000000000000",
				isValid:     true,
			},
			{
				desc:        "valid rfc4122 uuid lowercase",
				workspaceId: "abcdefab-cdef-abcd-efab-cdefabcdefab",
				isValid:     true,
			},
			{
				desc:        "valid rfc4122 uuid uppercase",
				workspaceId: "ABCDEFAB-CDEF-ABCD-EFAB-CDEFABCDEFAB",
				isValid:     true,
			},
		}
		for _, tC := range testCases {
			t.Run(tC.desc, func(t *testing.T) {
				slo := validRawMetricSLO(v1alpha.AzureMonitor)
				slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor = getValidAzureMetric(AzureMonitorDataTypeLogs)
				slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor.Workspace.WorkspaceID = tC.workspaceId
				err := validate(slo)
				if tC.isValid {
					testutils.AssertNoError(t, slo, err)
				} else {
					testutils.AssertContainsErrors(t, slo, err, 1,
						testutils.ExpectedError{
							Prop: "spec.objectives[0].rawMetric.query.azureMonitor.workspace.workspaceId",
							Code: rules.ErrorCodeStringUUID,
						})
				}
			})
		}
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
		testutils.AssertNoError(t, slo, err)
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
		testutils.AssertContainsErrors(t, slo, err, 9,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.dimensions[0].name",
				Code: rules.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.dimensions[0].value",
				Code: rules.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.dimensions[1].name",
				Code: rules.ErrorCodeStringNotEmpty,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.dimensions[1].value",
				Code: rules.ErrorCodeStringNotEmpty,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.dimensions[2].name",
				Code: rules.ErrorCodeStringMaxLength,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.dimensions[2].value",
				Code: rules.ErrorCodeStringMaxLength,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.dimensions[3].name",
				Code: rules.ErrorCodeStringASCII,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.azureMonitor.dimensions[3].value",
				Code: rules.ErrorCodeStringASCII,
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
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.azureMonitor.dimensions",
			Code: rules.ErrorCodeSliceUnique,
		})
	})
}

func TestAzureMonitor_ResourceID(t *testing.T) {
	testCases := []struct {
		desc       string
		resourceID string
		isValid    bool
	}{
		{
			desc:       "one letter",
			resourceID: "a",
			isValid:    false,
		},
		{
			desc:       "incomplete resource provider",
			resourceID: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/vm",
			isValid:    false,
		},
		{
			desc:       "missing resource providerNamespace",
			resourceID: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/Test-RG1/providers/virtualMachines/vm", //nolint:lll
			isValid:    false,
		},
		{
			desc:       "missing resource type",
			resourceID: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Compute/vm", //nolint:lll
			isValid:    false,
		},
		{
			desc:       "missing resource name",
			resourceID: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines", //nolint:lll
			isValid:    false,
		},
		{
			desc:       "valid resource id",
			resourceID: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm", //nolint:lll
			isValid:    true,
		},
		{
			desc:       "valid resource id with _",
			resourceID: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm-123_x", //nolint:lll
			isValid:    true,
		},
		{
			desc:       "valid resource id with _ in rg",
			resourceID: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/mc_().rg-xxx-01_ups-aks_eu_west/providers/Microsoft.()Network/loadBalancers1_-()/kubernetes", //nolint:lll
			isValid:    true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			slo := validRawMetricSLO(v1alpha.AzureMonitor)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor.ResourceID = tC.resourceID
			err := validate(slo)
			if tC.isValid {
				testutils.AssertNoError(t, slo, err)
			} else {
				testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
					Prop: "spec.objectives[0].rawMetric.query.azureMonitor.resourceId",
					Code: rules.ErrorCodeStringMatchRegexp,
				})
			}
		})
	}
}

func TestAzureMonitor_kqlQuery(t *testing.T) {
	testCases := []struct {
		desc         string
		kqlQuery     string
		isValid      bool
		errorMessage string
	}{
		{
			"valid query without bin",
			"Logs | summarize n9_value = max(value) | project TimeGenerated as n9_time, 1 as n9_value",
			true,
			"",
		},
		{
			"valid query with bin",
			"Logs | summarize n9_value = max(value) by bin(time, 15s) | project TimeGenerated as n9_time, 1 as n9_value",
			true,
			"",
		},
		{
			"no summarize",
			"Logs | project TimeGenerated as n9_time, 1 as n9_value",
			false,
			"summarize is required",
		},
		{
			"summarize without bin",
			"Logs | summarize n9_value = avg(value) | project TimeGenerated as n9_time, 1 as n9_value",
			true,
			"",
		},
		{
			"summarize without bin with time aggregation",
			"Logs | summarize n9_value = avg(value) by time | project TimeGenerated as n9_time, 1 as n9_value",
			false,
			"'summarize .* by' requires 'bin'(time, resolution) clause",
		},
		{
			"invalid aggregation resolution",
			"Logs | summarize n9_value = avg(value) by bin(time, 15) | project TimeGenerated as n9_time, 1 as n9_value",
			false,
			"bin duration is required in short 'timespan' format. E.g. '15s'",
		},
		{
			"aggregation resolution to small",
			"Logs | summarize n9_value = avg(value) by bin(time, 10ms) | project TimeGenerated as n9_time, 1 as n9_value",
			false,
			"bin duration must be at least 15s but was 10ms",
		},
		{
			"summarize used two times - valid",
			"Logs | summarize n9_value = avg(value) by time | summarize n9_value = avg(value) | project TimeGenerated as n9_time, 1 as n9_value",
			true,
			"",
		},
		{
			"summarize used two times - invalid",
			"Logs | summarize n9_value = avg(value) | summarize n9_value = avg(value) by time | project TimeGenerated as n9_time, 1 as n9_value",
			false,
			"'summarize .* by' requires 'bin'(time, resolution) clause",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			slo := validRawMetricSLO(v1alpha.AzureMonitor)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor = getValidAzureMetric(AzureMonitorDataTypeLogs)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.AzureMonitor.KQLQuery = tC.kqlQuery

			err := validate(slo)
			if tC.isValid {
				testutils.AssertNoError(t, slo, err)
			} else {
				testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
					Prop:    "spec.objectives[0].rawMetric.query.azureMonitor.kqlQuery",
					Message: tC.errorMessage,
				})
			}
		})
	}
}

func validAzureMonitorMetricsDataType() *AzureMonitorMetric {
	return &AzureMonitorMetric{DataType: AzureMonitorDataTypeMetrics,
		ResourceID:  "/subscriptions/123/resourceGroups/azure-monitor-test-sources/providers/Microsoft.Web/sites/app",
		MetricName:  "HttpResponseTime",
		Aggregation: "Avg",
	}
}

func validAzureMonitorLogsDataType() *AzureMonitorMetric {
	return &AzureMonitorMetric{
		DataType: AzureMonitorDataTypeLogs,
		Workspace: &AzureMonitorMetricLogAnalyticsWorkspace{
			SubscriptionID: "00000000-0000-0000-0000-000000000000",
			ResourceGroup:  "rg",
			WorkspaceID:    "11111111-1111-1111-1111-111111111111",
		},
		KQLQuery: "A | summarize n9_value = max(value) | project TimeGenerated as n9_time, 1 as n9_value",
	}
}

func getValidAzureMetric(dataType string) *AzureMonitorMetric {
	if dataType == AzureMonitorDataTypeMetrics {
		return validAzureMonitorMetricsDataType()
	}
	return validAzureMonitorLogsDataType()
}
