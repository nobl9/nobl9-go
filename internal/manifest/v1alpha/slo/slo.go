package slo

import "github.com/nobl9/nobl9-go/manifest/v1alpha"

var BadOverTotalEnabledSources = []v1alpha.DataSourceType{
	v1alpha.CloudWatch,
	v1alpha.AppDynamics,
	v1alpha.AzureMonitor,
	v1alpha.LogicMonitor,
	v1alpha.AzurePrometheus,
}

var SingleQueryGoodOverTotalEnabledSources = []v1alpha.DataSourceType{
	v1alpha.Splunk,
	v1alpha.Honeycomb,
}
