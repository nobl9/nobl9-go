package v1alphaExamples

import (
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/twindow"
)

var standardGoodOverTotalMetrics = []v1alpha.DataSourceType{
	v1alpha.Prometheus,
	v1alpha.Datadog,
	v1alpha.NewRelic,
	v1alpha.Splunk,
	v1alpha.SplunkObservability,
	v1alpha.Dynatrace,
	v1alpha.Elasticsearch,
	v1alpha.Graphite,
	v1alpha.BigQuery,
	v1alpha.OpenTSDB,
	v1alpha.GrafanaLoki,
	v1alpha.AmazonPrometheus,
	v1alpha.Redshift,
	v1alpha.InfluxDB,
	v1alpha.GCM,
	v1alpha.Generic,
}

var standardBadOverTotalMetrics = []v1alpha.DataSourceType{
	v1alpha.AppDynamics,
	v1alpha.LogicMonitor,
	v1alpha.Honeycomb,
	v1alpha.AzurePrometheus,
}

var customMetricExamples = map[v1alpha.DataSourceType]map[metricVariant][]metricSubVariant{
	v1alpha.Lightstep: {
		metricVariantThreshold: []metricSubVariant{
			metricSubVariantLightstepMetrics,
			metricSubVariantLightstepLatency,
			metricSubVariantLightstepError,
		},
		metricVariantGoodRatio: []metricSubVariant{
			metricSubVariantLightstepMetrics,
			metricSubVariantLightstepError,
		},
	},
	v1alpha.ThousandEyes: {
		metricVariantThreshold: []metricSubVariant{
			metricSubVariantThousandEyesWebPageLoad,
			metricSubVariantThousandEyesResponseTime,
			metricSubVariantThousandEyesNetLatency,
			metricSubVariantThousandEyesNetLoss,
			metricSubVariantThousandEyesDOMLoad,
			metricSubVariantThousandEyesServerAvailability,
			metricSubVariantThousandEyesServerThroughput,
		},
	},
	v1alpha.CloudWatch: {
		metricVariantThreshold: []metricSubVariant{
			metricSubVariantCloudWatchStandard,
			metricSubVariantCloudWatchJSON,
			metricSubVariantCloudWatchSQLQuery,
		},
		metricVariantGoodRatio: []metricSubVariant{
			metricSubVariantCloudWatchStandard,
			metricSubVariantCloudWatchJSON,
			metricSubVariantCloudWatchSQLQuery,
		},
		metricVariantBadRatio: []metricSubVariant{
			metricSubVariantCloudWatchStandard,
			metricSubVariantCloudWatchJSON,
			metricSubVariantCloudWatchSQLQuery,
		},
	},
	v1alpha.Pingdom: {
		metricVariantThreshold: []metricSubVariant{
			metricSubVariantPingdomUptime,
		},
		metricVariantGoodRatio: []metricSubVariant{
			metricSubVariantPingdomUptime,
			metricSubVariantPingdomTransaction,
		},
	},
	v1alpha.SumoLogic: {
		metricVariantThreshold: []metricSubVariant{
			metricSubVariantSumoLogicMetrics,
			metricSubVariantSumoLogicLogs,
		},
		metricVariantGoodRatio: []metricSubVariant{
			metricSubVariantSumoLogicMetrics,
			metricSubVariantSumoLogicLogs,
		},
	},
	v1alpha.Instana: {
		metricVariantThreshold: []metricSubVariant{
			metricSubVariantInstanaInfrastructureQuery,
			metricSubVariantInstanaInfrastructureSnapshotID,
			metricSubVariantInstanaApplication,
		},
		metricVariantGoodRatio: []metricSubVariant{
			metricSubVariantInstanaInfrastructureQuery,
			metricSubVariantInstanaInfrastructureSnapshotID,
		},
	},
	v1alpha.AzureMonitor: {
		metricVariantThreshold: []metricSubVariant{
			metricSubVariantAzureMonitorMetrics,
			metricSubVariantAzureMonitorLogs,
		},
		metricVariantGoodRatio: []metricSubVariant{
			metricSubVariantAzureMonitorMetrics,
			metricSubVariantAzureMonitorLogs,
		},
		metricVariantBadRatio: []metricSubVariant{
			metricSubVariantAzureMonitorMetrics,
			metricSubVariantAzureMonitorLogs,
		},
	},
}

var goodOverTotalVariants = []string{
	metricVariantThreshold,
	metricVariantGoodRatio,
}

var badOverTotalVariants = []string{
	metricVariantThreshold,
	metricVariantGoodRatio,
	metricVariantBadRatio,
}

func SLO() []Example {
	baseExamples := make([]sloExample, 0)
	for _, dataSourceType := range standardGoodOverTotalMetrics {
		baseExamples = append(baseExamples, createVariants(dataSourceType, goodOverTotalVariants, nil)...)
	}
	for _, dataSourceType := range standardBadOverTotalMetrics {
		baseExamples = append(baseExamples, createVariants(dataSourceType, badOverTotalVariants, nil)...)
	}
	for dataSourceType, customExamples := range customMetricExamples {
		for variant, subVariants := range customExamples {
			baseExamples = append(baseExamples, createVariants(
				dataSourceType,
				[]metricVariant{variant},
				subVariants,
			)...)
		}
	}
	variants := make([]sloExample, 0, len(baseExamples)*4)
	for _, example := range baseExamples {
		for _, timeWindow := range []twindow.TimeWindowTypeEnum{
			twindow.Rolling,
			twindow.Calendar,
		} {
			for _, method := range []v1alphaSLO.BudgetingMethod{
				v1alphaSLO.BudgetingMethodTimeslices,
				v1alphaSLO.BudgetingMethodOccurrences,
			} {
				example = sloExample{
					DataSourceType:   example.DataSourceType,
					BudgetingMethod:  method,
					TimeWindowType:   timeWindow,
					MetricVariant:    example.MetricVariant,
					MetricSubVariant: example.MetricSubVariant,
				}
				example.SLO = example.Generate()
				variants = append(variants, example)
			}
		}
	}
	return newExampleSlice(variants...)
}

func createVariants(
	dataSourceType v1alpha.DataSourceType,
	metricVariants []metricVariant,
	metricSubVariants []metricSubVariant,
) []sloExample {
	examples := make([]sloExample, 0, len(metricVariants)*(1+len(metricSubVariants)))
	for _, example := range metricVariants {
		if len(metricSubVariants) == 0 {
			examples = append(examples, sloExample{
				DataSourceType: dataSourceType,
				MetricVariant:  example,
			})
			continue
		}
		for _, subVariant := range metricSubVariants {
			examples = append(examples, sloExample{
				DataSourceType:   dataSourceType,
				MetricVariant:    example,
				MetricSubVariant: subVariant,
			})
		}
	}
	return examples
}
