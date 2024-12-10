package v1alphaExamples

import (
	"embed"
	"fmt"
	"path/filepath"
	"reflect"
	"slices"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/twindow"
	"github.com/nobl9/nobl9-go/sdk"
)

//go:embed queries
var queriesFS embed.FS

type sloBaseExample struct {
	BudgetingMethod v1alphaSLO.BudgetingMethod
	TimeWindowType  twindow.TimeWindowTypeEnum
}

type sloExample struct {
	sloBaseExample

	DataSourceType   v1alpha.DataSourceType
	MetricVariant    metricVariant
	MetricSubVariant metricSubVariant
}

func (s sloExample) GetObject() any {
	return s.SLO()
}

func (s sloExample) GetVariant() string {
	return toKebabCase(s.DataSourceType.String())
}

func (s sloExample) GetSubVariant() string {
	return s.String()
}

func (s sloExample) GetYAMLComments() []string {
	comments := []string{
		fmt.Sprintf("Metric type: %s", s.MetricVariant),
		fmt.Sprintf("Budgeting method: %s", s.BudgetingMethod),
		fmt.Sprintf("Time window type: %s", s.TimeWindowType),
	}
	if s.MetricSubVariant != "" {
		comments = slices.Insert(comments, 1, fmt.Sprintf("Metric variant: %s", s.MetricSubVariant))
	}
	return comments
}

func (s sloExample) GetDataSourceType() v1alpha.DataSourceType {
	return s.DataSourceType
}

func (s sloExample) String() string {
	subVariantStr := s.MetricSubVariant
	if subVariantStr != "" {
		subVariantStr = subVariantStr + " "
	}
	return fmt.Sprintf(
		"%s %s %sSLO using %s budgeting method and %s time window",
		s.DataSourceType,
		s.MetricVariant,
		subVariantStr,
		s.BudgetingMethod,
		s.TimeWindowType,
	)
}

func (s sloExample) SLO() v1alphaSLO.SLO {
	slo := v1alphaSLO.New(
		v1alphaSLO.Metadata{
			Name:        "api-server-slo",
			DisplayName: "API Server SLO",
			Project:     sdk.DefaultProject,
			Labels:      exampleLabels(),
			Annotations: exampleMetadataAnnotations(),
		},
		v1alphaSLO.Spec{
			Description: fmt.Sprintf("Example %s SLO", dataSourceTypePrettyName(s.DataSourceType)),
			Service:     "api-server",
			Indicator: &v1alphaSLO.Indicator{
				MetricSource: v1alphaSLO.MetricSourceSpec{
					Name:    toKebabCase(s.DataSourceType.String()),
					Project: sdk.DefaultProject,
					Kind:    manifest.KindAgent,
				},
			},
			BudgetingMethod: s.BudgetingMethod.String(),
			Attachments:     exampleAttachments(),
			AlertPolicies:   exampleAlertPolicies(),
			AnomalyConfig:   exampleAnomalyConfig(),
			TimeWindows:     exampleTimeWindows(s.TimeWindowType),
		})
	objective := v1alphaSLO.Objective{
		ObjectiveBase: v1alphaSLO.ObjectiveBase{
			DisplayName: "Good response (200)",
			Name:        "ok",
			Value:       nil,
		},
		BudgetTarget:    ptr(0.95),
		TimeSliceTarget: exampleTimeSliceTarget(s.BudgetingMethod),
		Primary:         ptr(true),
		RawMetric:       &v1alphaSLO.RawMetricSpec{},
		CountMetrics: &v1alphaSLO.CountMetricsSpec{
			Incremental: ptr(true),
		},
	}
	switch s.MetricVariant {
	case metricVariantThreshold:
		objective.Value = ptr(200.0)
		objective.Operator = ptr(v1alpha.LessThanEqual.String())
	case metricVariantGoodRatio:
		objective.Value = ptr(1.0)
	case metricVariantBadRatio:
		objective.Value = ptr(1.0)
	case metricVariantSingleQueryGoodRatio:
		objective.Value = ptr(1.0)
	default:
		panic(fmt.Sprintf("unsupported metric variant: %s", s.MetricVariant))
	}
	slo.Spec.Objectives = append(slo.Spec.Objectives, objective)
	// Set the metric spec variant as the last step.
	// This way the getVariant function can modify the SLO object as needed
	// without any chance that these changes would be overwritten.
	return s.generateMetricVariant(slo)
}

func exampleTimeWindows(timeWindowType twindow.TimeWindowTypeEnum) []v1alphaSLO.TimeWindow {
	var timeWindow []v1alphaSLO.TimeWindow
	switch timeWindowType {
	case twindow.Calendar:
		timeWindow = []v1alphaSLO.TimeWindow{
			{
				Unit:  "Month",
				Count: 1.0,
				Calendar: &v1alphaSLO.Calendar{
					StartTime: "2022-12-01 00:00:00",
					TimeZone:  "UTC",
				},
			},
		}
	case twindow.Rolling:
		timeWindow = []v1alphaSLO.TimeWindow{
			{
				Unit:      twindow.Hour.String(),
				Count:     1.0,
				IsRolling: true,
			},
		}
	}
	return timeWindow
}

func exampleTimeSliceTarget(method v1alphaSLO.BudgetingMethod) *float64 {
	var target *float64
	if method == v1alphaSLO.BudgetingMethodTimeslices {
		target = ptr(0.90)
	} else {
		target = nil
	}
	return target
}

func exampleAlertPolicies() []string {
	return []string{"fast-burn-5x-for-last-10m"}
}

func exampleAnomalyConfig() *v1alphaSLO.AnomalyConfig {
	return &v1alphaSLO.AnomalyConfig{
		NoData: &v1alphaSLO.AnomalyConfigNoData{
			AlertMethods: []v1alphaSLO.AnomalyConfigAlertMethod{
				{
					Name:    "slack-notification",
					Project: sdk.DefaultProject,
				},
			},
		},
	}
}

func exampleAttachments() []v1alphaSLO.Attachment {
	return []v1alphaSLO.Attachment{
		{
			URL:         "https://docs.nobl9.com",
			DisplayName: ptr("Nobl9 Documentation"),
		},
	}
}

// metricVariant lists the standard metric variants.
// If a metric source has non-standard variants (e.g. Lightstep), it can extend with metricSubVariant.
type metricVariant = string

const (
	metricVariantThreshold            metricVariant = "threshold"
	metricVariantGoodRatio            metricVariant = "good over total"
	metricVariantBadRatio             metricVariant = "bad over total"
	metricVariantSingleQueryGoodRatio metricVariant = "single query good over total"
)

// metricSubVariant allows extending standard metric variants with metric source specific sub-variants.
type metricSubVariant = string

const (
	// Lightstep.
	metricSubVariantLightstepMetrics metricVariant = "metrics"
	metricSubVariantLightstepLatency metricVariant = "latency"
	metricSubVariantLightstepError   metricVariant = "error"
	// ThousandEyes.
	metricSubVariantThousandEyesWebPageLoad        metricVariant = "web page load"
	metricSubVariantThousandEyesResponseTime       metricVariant = "response time"
	metricSubVariantThousandEyesNetLatency         metricVariant = "net latency"
	metricSubVariantThousandEyesNetLoss            metricVariant = "net loss"
	metricSubVariantThousandEyesDOMLoad            metricVariant = "DOM load"
	metricSubVariantThousandEyesServerAvailability metricVariant = "server availability"
	metricSubVariantThousandEyesServerThroughput   metricVariant = "server throughput"
	// CloudWatch.
	metricSubVariantCloudWatchStandard metricVariant = "standard configuration"
	metricSubVariantCloudWatchSQLQuery metricVariant = "sql query"
	metricSubVariantCloudWatchJSON     metricVariant = "JSON"
	// Pingdom.
	metricSubVariantPingdomUptime      metricVariant = "uptime"
	metricSubVariantPingdomTransaction metricVariant = "transaction"
	// SumoLogic.
	metricSubVariantSumoLogicMetrics metricVariant = "metrics"
	metricSubVariantSumoLogicLogs    metricVariant = "logs"
	// Instana.
	metricSubVariantInstanaInfrastructureQuery      metricVariant = "infrastructure query"
	metricSubVariantInstanaInfrastructureSnapshotID metricVariant = "infrastructure snapshot id"
	metricSubVariantInstanaApplication              metricVariant = "application"
	// AzureMonitor.
	metricSubVariantAzureMonitorMetrics metricVariant = "metrics"
	metricSubVariantAzureMonitorLogs    metricVariant = "logs"
)

// generateMetricVariant returns a [v1alphaSLO.SLO] with all [v1alphaSLO.MetricSpec] variants filled out.
// The standard variants are: raw, good/total, and bad/total (only supported sources).
// If a metric source has non-standard variants (e.g. Lightstep), it can extend metricVariant with it's own types.
// It is up to the caller to nil-out the unwanted fields.
func (s sloExample) generateMetricVariant(slo v1alphaSLO.SLO) v1alphaSLO.SLO {
	switch s.DataSourceType {
	case v1alpha.Prometheus:
		switch s.MetricVariant {
		case metricVariantThreshold:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.PrometheusMetric{
				PromQL: ptr(`api_server_requestMsec{host="*",job="nginx"}`),
			}))
		case metricVariantGoodRatio:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.PrometheusMetric{
				PromQL: ptr(`sum(http_request_duration_seconds_bucket{handler="/api/v1/slos",le="2.5"})`),
			}), newMetricSpec(v1alphaSLO.PrometheusMetric{
				PromQL: ptr(`sum(http_request_duration_seconds_count{handler="/api/v1/slos"})`),
			}))
		}
	case v1alpha.Datadog:
		switch s.MetricVariant {
		case metricVariantThreshold:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.DatadogMetric{
				Query: ptr(`avg:trace.http.request.duration{*}`),
			}))
		case metricVariantGoodRatio:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.DatadogMetric{
				Query: ptr(`sum:trace.http.request.hits.by_http_status{http.status_class:2xx}.as_count()`),
			}), newMetricSpec(v1alphaSLO.DatadogMetric{
				Query: ptr(`sum:trace.http.request.hits.by_http_status{*}.as_count()`),
			}))
		}
	case v1alpha.NewRelic:
		switch s.MetricVariant {
		case metricVariantThreshold:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.NewRelicMetric{
				NRQL: ptr(`select average(duration) from transaction timeseries`),
			}))
		case metricVariantGoodRatio:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.NewRelicMetric{
				NRQL: ptr(`SELECT count(*) FROM Transaction WHERE httpResponseCode IN ('200','301','302') TIMESERIES`),
			}), newMetricSpec(v1alphaSLO.NewRelicMetric{
				NRQL: ptr(`SELECT count(*) FROM Transaction TIMESERIES`),
			}))
		}
	case v1alpha.AppDynamics:
		total := newMetricSpec(&v1alphaSLO.AppDynamicsMetric{
			ApplicationName: ptr("api-server"),
			MetricPath:      ptr("End User Experience|App|Normal Requests"),
		})
		switch s.MetricVariant {
		case metricVariantThreshold:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.AppDynamicsMetric{
				ApplicationName: ptr("api-server"),
				MetricPath:      ptr("Overall Application Performance|Average Response Time (ms)"),
			}))
		case metricVariantGoodRatio:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.AppDynamicsMetric{
				ApplicationName: ptr("api-server"),
				MetricPath:      ptr("End User Experience|App|Slow Requests"),
			}), total)
		case metricVariantBadRatio:
			return setBadOverTotalMetric(slo, newMetricSpec(v1alphaSLO.AppDynamicsMetric{
				ApplicationName: ptr("api-server"),
				MetricPath:      ptr("End User Experience|App|Slow Requests"),
			}), total)
		}
	case v1alpha.Splunk:
		switch s.MetricVariant {
		case metricVariantThreshold:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.SplunkMetric{
				Query: ptr(`index=* source=udp:5072 sourcetype=syslog status<400 | bucket _time span=1m | stats avg(response_time) as n9value by _time | rename _time as n9time | fields n9time n9value`),
			}))
		case metricVariantGoodRatio:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.SplunkMetric{
				Query: ptr(`index=* source=udp:5072 sourcetype=syslog status<400 | bucket _time span=1m | stats count as n9value by _time | rename _time as n9time | fields n9time n9value`),
			}), newMetricSpec(v1alphaSLO.SplunkMetric{
				Query: ptr(`index=* source=udp:5072 sourcetype=syslog | bucket _time span=1m | stats count as n9value by _time | rename _time as n9time | fields n9time n9value`),
			}))
		case metricVariantSingleQueryGoodRatio:
			return setSingleQueryGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.SplunkMetric{
				Query: ptr(`| mstats avg("spl.intr.resource_usage.IOWait.data.avg_cpu_pct") as n9good WHERE index="_metrics" span=15s
| join type=left _time [
| mstats avg("spl.intr.resource_usage.IOWait.data.max_cpus_pct") as n9total WHERE index="_metrics" span=15s
]
| rename _time as n9time
| fields n9time n9good n9total`),
			}))
		}
	case v1alpha.Lightstep:
		slo.Spec.Objectives[0].CountMetrics.Incremental = ptr(false)
		switch s.MetricVariant + s.MetricSubVariant {
		case metricVariantThreshold + metricSubVariantLightstepMetrics:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.LightstepMetric{
				TypeOfData: ptr(v1alphaSLO.LightstepMetricDataType),
				UQL:        ptr(`metric cpu.utilization | rate | group_by [], mean`),
			}))
		case metricVariantThreshold + metricSubVariantLightstepLatency:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.LightstepMetric{
				TypeOfData: ptr(v1alphaSLO.LightstepLatencyDataType),
				StreamID:   ptr("DzpxcSRh"),
				Percentile: ptr(95.0),
			}))
		case metricVariantThreshold + metricSubVariantLightstepError:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.LightstepMetric{
				TypeOfData: ptr(v1alphaSLO.LightstepErrorRateDataType),
				StreamID:   ptr("DzpxcSRh"),
			}))
		case metricVariantGoodRatio + metricSubVariantLightstepMetrics:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.LightstepMetric{
				TypeOfData: ptr(v1alphaSLO.LightstepMetricDataType),
				UQL:        ptr(`metric cpu.utilization | rate | group_by [], mean`),
			}), newMetricSpec(v1alphaSLO.LightstepMetric{
				TypeOfData: ptr(v1alphaSLO.LightstepMetricDataType),
				UQL:        ptr(`metric cpu.utilization | rate | group_by [], max`),
			}))
		case metricVariantGoodRatio + metricSubVariantLightstepError:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.LightstepMetric{
				TypeOfData: ptr(v1alphaSLO.LightstepGoodCountDataType),
				StreamID:   ptr("DzpxcSRh"),
			}), newMetricSpec(v1alphaSLO.LightstepMetric{
				TypeOfData: ptr(v1alphaSLO.LightstepTotalCountDataType),
				StreamID:   ptr("DzpxcSRh"),
			}))
		}
	case v1alpha.SplunkObservability:
		switch s.MetricVariant {
		case metricVariantThreshold:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.SplunkObservabilityMetric{
				Program: ptr(`data('demo.trans.count', filter=filter('api_server'), rollup='rate').mean().publish()`),
			}))
		case metricVariantGoodRatio:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.SplunkObservabilityMetric{
				Program: ptr(`data('demo.trans.count', filter=filter('api_server'), rollup='rate').stddev().publish()`),
			}), newMetricSpec(v1alphaSLO.SplunkObservabilityMetric{
				Program: ptr(`data('demo.trans.count', filter=filter('api_server'), rollup='rate').mean().publish()`),
			}))
		}
	case v1alpha.Dynatrace:
		switch s.MetricVariant {
		case metricVariantThreshold:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.DynatraceMetric{
				MetricSelector: ptr(`builtin:service.response.server:filter(and(or(in("dt.entity.service",entitySelector("type(service),entityName.equals(~"APIServer~")"))))):splitBy("dt.entity.service"):sort(value(auto,descending)):limit(100)`),
			}))
		case metricVariantGoodRatio:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.DynatraceMetric{
				MetricSelector: ptr(`builtin:synthetic.http.request.statusCode:filter(and(or(eq("Status code",SC_2xx)))):splitBy():sort(value(auto,descending)):limit(20)`),
			}), newMetricSpec(v1alphaSLO.DynatraceMetric{
				MetricSelector: ptr(`builtin:synthetic.http.request.statusCode:splitBy():sort(value(auto,descending)):limit(20)`),
			}))
		}
	case v1alpha.Elasticsearch:
		switch s.MetricVariant {
		case metricVariantThreshold:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.ElasticsearchMetric{
				Index: ptr("apm-7.13.3-transaction"),
				Query: ptr(mustLoadQuery("elasticsearch_threshold.json")),
			}))
		case metricVariantGoodRatio:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.ElasticsearchMetric{
				Index: ptr("apm-7.13.3-transaction"),
				Query: ptr(mustLoadQuery("elasticsearch_count_good.json")),
			}), newMetricSpec(v1alphaSLO.ElasticsearchMetric{
				Index: ptr("apm-7.13.3-transaction"),
				Query: ptr(mustLoadQuery("elasticsearch_count_total.json")),
			}))
		}
	case v1alpha.ThousandEyes:
		switch s.MetricVariant + s.MetricSubVariant {
		case metricVariantThreshold + metricSubVariantThousandEyesWebPageLoad:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.ThousandEyesMetric{
				TestID:   ptr[int64](2280492),
				TestType: ptr(v1alphaSLO.ThousandEyesWebPageLoad),
			}))
		case metricVariantThreshold + metricSubVariantThousandEyesResponseTime:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.ThousandEyesMetric{
				TestID:   ptr[int64](2280492),
				TestType: ptr(v1alphaSLO.ThousandEyesHTTPResponseTime),
			}))
		case metricVariantThreshold + metricSubVariantThousandEyesNetLatency:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.ThousandEyesMetric{
				TestID:   ptr[int64](2280492),
				TestType: ptr(v1alphaSLO.ThousandEyesNetLatency),
			}))
		case metricVariantThreshold + metricSubVariantThousandEyesNetLoss:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.ThousandEyesMetric{
				TestID:   ptr[int64](2280492),
				TestType: ptr(v1alphaSLO.ThousandEyesNetLoss),
			}))
		case metricVariantThreshold + metricSubVariantThousandEyesDOMLoad:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.ThousandEyesMetric{
				TestID:   ptr[int64](2280492),
				TestType: ptr(v1alphaSLO.ThousandEyesWebDOMLoad),
			}))
		case metricVariantThreshold + metricSubVariantThousandEyesServerAvailability:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.ThousandEyesMetric{
				TestID:   ptr[int64](2280492),
				TestType: ptr(v1alphaSLO.ThousandEyesServerAvailability),
			}))
		case metricVariantThreshold + metricSubVariantThousandEyesServerThroughput:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.ThousandEyesMetric{
				TestID:   ptr[int64](2280492),
				TestType: ptr(v1alphaSLO.ThousandEyesServerThroughput),
			}))
		}
	case v1alpha.Graphite:
		switch s.MetricVariant {
		case metricVariantThreshold:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.GraphiteMetric{
				MetricPath: ptr(`carbon.agents.9b365cce.cpuUsage`),
			}))
		case metricVariantGoodRatio:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.GraphiteMetric{
				MetricPath: ptr(`stats_counts.response.200`),
			}), newMetricSpec(v1alphaSLO.GraphiteMetric{
				MetricPath: ptr(`stats_counts.response.all`),
			}))
		}
	case v1alpha.BigQuery:
		projectID := "api-server-256112"
		switch s.MetricVariant {
		case metricVariantThreshold:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.BigQueryMetric{
				ProjectID: projectID,
				Location:  "US",
				Query:     fmt.Sprintf("SELECT response_time AS n9value, created AS n9date FROM `%s.metrics.http_response` WHERE created BETWEEN DATETIME(@n9date_from) AND DATETIME(@n9date_to)`", projectID),
			}))
		case metricVariantGoodRatio:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.BigQueryMetric{
				ProjectID: projectID,
				Location:  "US",
				Query:     fmt.Sprintf("SELECT http_code AS n9value, created AS n9date FROM `%s.metrics.http_response` WHERE http_code = 200 AND created BETWEEN DATETIME(@n9date_from) AND DATETIME(@n9date_to)", projectID),
			}), newMetricSpec(v1alphaSLO.BigQueryMetric{
				ProjectID: projectID,
				Location:  "US",
				Query:     fmt.Sprintf("SELECT http_code AS n9value, created AS n9date FROM `%s.metrics.http_response` WHERE created BETWEEN DATETIME(@n9date_from) AND DATETIME(@n9date_to)", projectID),
			}))
		}
	case v1alpha.OpenTSDB:
		switch s.MetricVariant {
		case metricVariantThreshold:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.OpenTSDBMetric{
				Query: ptr(`start={{.BeginTime}}&end={{.EndTime}}&ms=true&m=none:{{.Resolution}}-avg-zero:transaction.duration{host=host.01}`),
			}))
		case metricVariantGoodRatio:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.OpenTSDBMetric{
				Query: ptr(`start={{.BeginTime}}&end={{.EndTime}}&ms=true&m=none:{{.Resolution}}-count-zero:http.code{code=2xx}`),
			}), newMetricSpec(v1alphaSLO.OpenTSDBMetric{
				Query: ptr(`start={{.BeginTime}}&end={{.EndTime}}&ms=true&m=none:{{.Resolution}}-count-zero:http.code{type=http.status_code}`),
			}))
		}
	case v1alpha.GrafanaLoki:
		switch s.MetricVariant {
		case metricVariantThreshold:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.GrafanaLokiMetric{
				Logql: ptr(`sum(sum_over_time({topic="cdc"} |= "kafka_consumergroup_lag" | logfmt | line_format "{{.kafka_consumergroup_lag}}" | unwrap kafka_consumergroup_lag [1m]))`),
			}))
		case metricVariantGoodRatio:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.GrafanaLokiMetric{
				Logql: ptr(`count(count_over_time(({component="api-server"} | json | line_format "{{.log}}" | json | http_status_code >= 200 and http_status_code < 300)[1m]))`),
			}), newMetricSpec(v1alphaSLO.GrafanaLokiMetric{
				Logql: ptr(`count(count_over_time(({component="api-server"} | json | line_format "{{.log}}" | json | http_status_code > 0)[1m]))`),
			}))
		}
	case v1alpha.CloudWatch:
		switch s.MetricVariant + s.MetricSubVariant {
		case metricVariantThreshold + metricSubVariantCloudWatchStandard:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.CloudWatchMetric{
				AccountID:  ptr("123456789012"),
				Region:     ptr("us-west-2"),
				Namespace:  ptr("AWS/RDS"),
				MetricName: ptr("ReadLatency"),
				Stat:       ptr("Average"),
				Dimensions: []v1alphaSLO.CloudWatchMetricDimension{
					{
						Name:  ptr("LoadBalancer"),
						Value: ptr("app/api-server"),
					},
				},
			}))
		case metricVariantGoodRatio + metricSubVariantCloudWatchStandard:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.CloudWatchMetric{
				AccountID:  ptr("123456789012"),
				Region:     ptr("us-west-2"),
				Namespace:  ptr("AWS/ApplicationELB"),
				MetricName: ptr("HTTPCode_Target_2XX_Count"),
				Stat:       ptr("SampleCount"),
				Dimensions: []v1alphaSLO.CloudWatchMetricDimension{
					{
						Name:  ptr("LoadBalancer"),
						Value: ptr("app/api-server"),
					},
				},
			}), newMetricSpec(v1alphaSLO.CloudWatchMetric{
				AccountID:  ptr("123456789012"),
				Region:     ptr("us-west-2"),
				Namespace:  ptr("AWS/ApplicationELB"),
				MetricName: ptr("RequestCount"),
				Stat:       ptr("SampleCount"),
				Dimensions: []v1alphaSLO.CloudWatchMetricDimension{
					{
						Name:  ptr("LoadBalancer"),
						Value: ptr("app/api-server"),
					},
				},
			}))
		case metricVariantBadRatio + metricSubVariantCloudWatchStandard:
			return setBadOverTotalMetric(slo, newMetricSpec(v1alphaSLO.CloudWatchMetric{
				AccountID:  ptr("123456789012"),
				Region:     ptr("us-west-2"),
				Namespace:  ptr("AWS/ApplicationELB"),
				MetricName: ptr("HTTPCode_Target_5XX_Count"),
				Stat:       ptr("SampleCount"),
				Dimensions: []v1alphaSLO.CloudWatchMetricDimension{
					{
						Name:  ptr("LoadBalancer"),
						Value: ptr("app/api-server"),
					},
				},
			}), newMetricSpec(v1alphaSLO.CloudWatchMetric{
				AccountID:  ptr("123456789012"),
				Region:     ptr("us-west-2"),
				Namespace:  ptr("AWS/ApplicationELB"),
				MetricName: ptr("RequestCount"),
				Stat:       ptr("SampleCount"),
				Dimensions: []v1alphaSLO.CloudWatchMetricDimension{
					{
						Name:  ptr("LoadBalancer"),
						Value: ptr("app/api-server"),
					},
				},
			}))
		case metricVariantThreshold + metricSubVariantCloudWatchSQLQuery:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.CloudWatchMetric{
				Region: ptr("us-west-2"),
				SQL:    ptr(`SELECT AVG(CPUUtilization) FROM "AWS/EC2”`),
			}))
		case metricVariantGoodRatio + metricSubVariantCloudWatchSQLQuery:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.CloudWatchMetric{
				Region: ptr("us-west-2"),
				SQL:    ptr(`SELECT AVG(CPUUtilization) FROM "AWS/EC2"`),
			}), newMetricSpec(v1alphaSLO.CloudWatchMetric{
				Region: ptr("us-west-2"),
				SQL:    ptr(`SELECT MAX(CPUUtilization) FROM "AWS/EC2"`),
			}))
			// TODO this needs to be better adjusted.
		case metricVariantBadRatio + metricSubVariantCloudWatchSQLQuery:
			return setBadOverTotalMetric(slo, newMetricSpec(v1alphaSLO.CloudWatchMetric{
				Region: ptr("us-west-2"),
				SQL:    ptr(`SELECT AVG(CPUUtilization) FROM "AWS/EC2"`),
			}), newMetricSpec(v1alphaSLO.CloudWatchMetric{
				Region: ptr("us-west-2"),
				SQL:    ptr(`SELECT MAX(CPUUtilization) FROM "AWS/EC2"`),
			}))
		case metricVariantThreshold + metricSubVariantCloudWatchJSON:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.CloudWatchMetric{
				Region: ptr("us-west-2"),
				JSON:   ptr(mustLoadQuery("cloudwatch_threshold.json")),
			}))
		case metricVariantGoodRatio + metricSubVariantCloudWatchJSON:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.CloudWatchMetric{
				Region: ptr("us-west-2"),
				JSON:   ptr(mustLoadQuery("cloudwatch_count_good.json")),
			}), newMetricSpec(v1alphaSLO.CloudWatchMetric{
				Region: ptr("us-west-2"),
				JSON:   ptr(mustLoadQuery("cloudwatch_count_total.json")),
			}))
			// TODO this needs to be better adjusted.
		case metricVariantBadRatio + metricSubVariantCloudWatchJSON:
			return setBadOverTotalMetric(slo, newMetricSpec(v1alphaSLO.CloudWatchMetric{
				Region: ptr("us-west-2"),
				JSON:   ptr(mustLoadQuery("cloudwatch_count_bad.json")),
			}), newMetricSpec(v1alphaSLO.CloudWatchMetric{
				Region: ptr("us-west-2"),
				JSON:   ptr(mustLoadQuery("cloudwatch_count_total.json")),
			}))
		}
	case v1alpha.Pingdom:
		switch s.MetricVariant + s.MetricSubVariant {
		case metricVariantThreshold + metricSubVariantPingdomUptime:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.PingdomMetric{
				CheckID:   ptr("1234567"),
				CheckType: ptr(v1alphaSLO.PingdomTypeUptime),
				Status:    ptr("up"),
			}))
		case metricVariantGoodRatio + metricSubVariantPingdomUptime:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.PingdomMetric{
				CheckID:   ptr("1234567"),
				CheckType: ptr(v1alphaSLO.PingdomTypeUptime),
				Status:    ptr("up"),
			}), newMetricSpec(v1alphaSLO.PingdomMetric{
				CheckID:   ptr("1234567"),
				CheckType: ptr(v1alphaSLO.PingdomTypeUptime),
				Status:    ptr("up,down"),
			}))
		case metricVariantGoodRatio + metricSubVariantPingdomTransaction:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.PingdomMetric{
				CheckID:   ptr("1234567"),
				CheckType: ptr(v1alphaSLO.PingdomTypeTransaction),
			}), newMetricSpec(v1alphaSLO.PingdomMetric{
				CheckID:   ptr("1234567"),
				CheckType: ptr(v1alphaSLO.PingdomTypeTransaction),
			}))
		}
	case v1alpha.AmazonPrometheus:
		switch s.MetricVariant {
		case metricVariantThreshold:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.AmazonPrometheusMetric{
				PromQL: ptr(`api_server_requestMsec{host="*",job="nginx"}`),
			}))
		case metricVariantGoodRatio:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.AmazonPrometheusMetric{
				PromQL: ptr(`sum(http_request_duration_seconds_bucket{handler="/api/v1/slos",le="2.5"})`),
			}), newMetricSpec(v1alphaSLO.AmazonPrometheusMetric{
				PromQL: ptr(`sum(http_request_duration_seconds_count{handler="/api/v1/slos"})`),
			}))
		}
	case v1alpha.Redshift:
		switch s.MetricVariant {
		case metricVariantThreshold:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.RedshiftMetric{
				Region:       ptr("eu-central-1"),
				ClusterID:    ptr("prod-cluster"),
				DatabaseName: ptr("db"),
				Query:        ptr(`SELECT value as n9value, timestamp as n9date FROM sinusoid WHERE timestamp BETWEEN :n9date_from AND :n9date_to`),
			}))
		case metricVariantGoodRatio:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.RedshiftMetric{
				Region:       ptr("eu-central-1"),
				ClusterID:    ptr("prod-cluster"),
				DatabaseName: ptr("db"),
				Query:        ptr(`SELECT value as n9value, timestamp as n9date FROM http_status_codes WHERE value = '200' AND timestamp BETWEEN :n9date_from AND :n9date_to`),
			}), newMetricSpec(v1alphaSLO.RedshiftMetric{
				Region:       ptr("eu-central-1"),
				ClusterID:    ptr("prod-cluster"),
				DatabaseName: ptr("db"),
				Query:        ptr(`SELECT value as n9value, timestamp as n9date FROM http_status_codes WHERE timestamp BETWEEN :n9date_from AND :n9date_to`),
			}))
		}
	case v1alpha.SumoLogic:
		switch s.MetricVariant + s.MetricSubVariant {
		case metricVariantThreshold + metricSubVariantSumoLogicMetrics:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.SumoLogicMetric{
				Type:         ptr("metrics"),
				Rollup:       ptr("Avg"),
				Quantization: ptr("15s"),
				Query:        ptr(`metric=CPU_Usage`),
			}))
		case metricVariantGoodRatio + metricSubVariantSumoLogicMetrics:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.SumoLogicMetric{
				Type:         ptr("metrics"),
				Rollup:       ptr("Avg"),
				Quantization: ptr("15s"),
				Query:        ptr(`metric=Mem_Used`),
			}), newMetricSpec(v1alphaSLO.SumoLogicMetric{
				Type:         ptr("metrics"),
				Rollup:       ptr("Avg"),
				Quantization: ptr("15s"),
				Query:        ptr(`metric=Mem_Total`),
			}))
		case metricVariantThreshold + metricSubVariantSumoLogicLogs:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.SumoLogicMetric{
				Type: ptr("logs"),
				Query: ptr(`_sourceCategory=uploads/nginx
| timeslice 1m as n9_time
| parse "HTTP/1.1" * * " as (status_code, size, tail)
| if (status_code matches "20" or status_code matches "30*",1,0) as resp_ok
| sum(resp_ok) as n9_value by n9_time
| sort by n9_time asc`),
			}))
		case metricVariantGoodRatio + metricSubVariantSumoLogicLogs:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.SumoLogicMetric{
				Type: ptr("logs"),
				Query: ptr(`_collector="app-cluster" _source="logs"
| json "log"
| timeslice 15s as n9_time
| parse "level=* *" as (log_level, tail)
| if (log_level matches "error" ,0,1) as log_level_not_error
| sum(log_level_not_error) as n9_value by n9_time
| sort by n9_time asc`),
			}), newMetricSpec(v1alphaSLO.SumoLogicMetric{
				Type: ptr("logs"),
				Query: ptr(`_collector="app-cluster" _source="logs"
| json "log"
| timeslice 15s as n9_time
| parse "level=* *" as (log_level, tail)
| count(*) as n9_value by n9_time
| sort by n9_time asc`),
			}))
		}
	case v1alpha.Instana:
		switch s.MetricVariant + s.MetricSubVariant {
		case metricVariantThreshold + metricSubVariantInstanaInfrastructureQuery:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.InstanaMetric{
				MetricType: "infrastructure",
				Infrastructure: &v1alphaSLO.InstanaInfrastructureMetricType{
					MetricRetrievalMethod: "query",
					Query:                 ptr("entity.selfType:zookeeper AND entity.label:replica.1"),
					MetricID:              "max_request_latency",
					PluginID:              "zooKeeper",
				},
			}))
		case metricVariantThreshold + metricSubVariantInstanaInfrastructureSnapshotID:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.InstanaMetric{
				MetricType: "infrastructure",
				Infrastructure: &v1alphaSLO.InstanaInfrastructureMetricType{
					MetricRetrievalMethod: "snapshot",
					SnapshotID:            ptr("00u2y4e4atkzaYkXP4x8"),
					MetricID:              "max_request_latency",
					PluginID:              "zooKeeper",
				},
			}))
		case metricVariantThreshold + metricSubVariantInstanaApplication:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.InstanaMetric{
				MetricType: "application",
				Application: &v1alphaSLO.InstanaApplicationMetricType{
					MetricID:    "calls",
					Aggregation: "sum",
					GroupBy: v1alphaSLO.InstanaApplicationMetricGroupBy{
						Tag:       "application.name",
						TagEntity: "DESTINATION",
					},
					APIQuery: mustLoadQuery("instana_application_query.json"),
				},
			}))
		case metricVariantGoodRatio + metricSubVariantInstanaInfrastructureQuery:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.InstanaMetric{
				MetricType: "infrastructure",
				Infrastructure: &v1alphaSLO.InstanaInfrastructureMetricType{
					MetricRetrievalMethod: "query",
					Query:                 ptr("entity.selfType:zookeeper AND entity.label:replica.1"),
					MetricID:              "error_requests_count",
					PluginID:              "zooKeeper",
				},
			}), newMetricSpec(v1alphaSLO.InstanaMetric{
				MetricType: "infrastructure",
				Infrastructure: &v1alphaSLO.InstanaInfrastructureMetricType{
					MetricRetrievalMethod: "query",
					Query:                 ptr("entity.selfType:zookeeper AND entity.label:replica.1"),
					MetricID:              "total_requests_count",
					PluginID:              "zooKeeper",
				},
			}))
		case metricVariantGoodRatio + metricSubVariantInstanaInfrastructureSnapshotID:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.InstanaMetric{
				MetricType: "infrastructure",
				Infrastructure: &v1alphaSLO.InstanaInfrastructureMetricType{
					MetricRetrievalMethod: "snapshot",
					SnapshotID:            ptr("00u2y4e4atkzaYkXP4x8"),
					MetricID:              "error_requests_count",
					PluginID:              "zooKeeper",
				},
			}), newMetricSpec(v1alphaSLO.InstanaMetric{
				MetricType: "infrastructure",
				Infrastructure: &v1alphaSLO.InstanaInfrastructureMetricType{
					MetricRetrievalMethod: "snapshot",
					SnapshotID:            ptr("00u2y4e4atkzaYkXP4x8"),
					MetricID:              "total_requests_count",
					PluginID:              "zooKeeper",
				},
			}))
		}
	case v1alpha.InfluxDB:
		switch s.MetricVariant {
		case metricVariantThreshold:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.InfluxDBMetric{
				Query: ptr(`from(bucket: "integrations")
|> range(start: time(v: params.n9time_start), stop: time(v: params.n9time_stop))
|> aggregateWindow(every: 15s, fn: mean, createEmpty: false)
|> filter(fn: (r) => r["_measurement"] == "internal_write")
|> filter(fn: (r) => r["_field"] == "write_time_ns")`),
			}))
		case metricVariantGoodRatio:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.InfluxDBMetric{
				Query: ptr(`from(bucket: "integrations")
|> range(start: time(v: params.n9time_start), stop: time(v: params.n9time_stop))
|> aggregateWindow(every: 15s, fn: mean, createEmpty: false)
|> filter(fn: (r) => r["_measurement"] == "internal_write")
|> filter(fn: (r) => r["_field"] == "write_time_ns")`),
			}), newMetricSpec(v1alphaSLO.InfluxDBMetric{
				Query: ptr(`from(bucket: "integrations")
|> range(start: time(v: params.n9time_start), stop: time(v: params.n9time_stop))
|> aggregateWindow(every: 15s, fn: mean, createEmpty: false)
|> filter(fn: (r) => r["_measurement"] == "internal_write")
|> filter(fn: (r) => r["_field"] == "write_time_ns")`),
			}))
		}
	case v1alpha.GCM:
		switch s.MetricVariant {
		case metricVariantThreshold:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.GCMMetric{
				ProjectID: "my-project-id",
				Query: `fetch api-server
| metric 'serviceruntime.googleapis.com/api/request_latencies'
| filter (resource.service == 'monitoring.googleapis.com')
| align delta(1m)
| every 1m
| group_by [resource.service],
    [value_request_latencies_mean: mean(value.request_latencies)]`}))
		case metricVariantGoodRatio:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.GCMMetric{
				ProjectID: "my-project-id",
				Query: `fetch api-server
| metric 'serviceruntime.googleapis.com/api/request_count'
| filter
    (resource.service == 'monitoring.googleapis.com')
    && (metric.response_code == '200')
| align rate(1m)
| every 1m
| group_by [resource.service],
    [value_request_count_aggregate: aggregate(value.request_count)]`,
			}), newMetricSpec(v1alphaSLO.GCMMetric{
				ProjectID: "my-project-id",
				Query: `fetch api-server
| metric 'serviceruntime.googleapis.com/api/request_count'
| filter
    (resource.service == 'monitoring.googleapis.com')
| align rate(1m)
| every 1m
| group_by [resource.service],
    [value_request_count_aggregate: aggregate(value.request_count)]`}))
		}
	case v1alpha.AzureMonitor:
		switch s.MetricVariant + s.MetricSubVariant {
		case metricVariantThreshold + metricSubVariantAzureMonitorMetrics:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.AzureMonitorMetric{
				DataType:        v1alphaSLO.AzureMonitorDataTypeMetrics,
				ResourceID:      "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/myResourceGroup/providers/Microsoft.Compute/virtualMachines/api-server",
				MetricName:      "Percentage CPU",
				MetricNamespace: "azure.applicationinsights",
				Aggregation:     "Avg",
			}))
		case metricVariantGoodRatio + metricSubVariantAzureMonitorMetrics:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.AzureMonitorMetric{
				DataType:    v1alphaSLO.AzureMonitorDataTypeMetrics,
				ResourceID:  "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/myResourceGroup/providers/Microsoft.Compute/virtualMachines/api-server",
				MetricName:  "Http2xx",
				Aggregation: "Sum",
			}), newMetricSpec(v1alphaSLO.AzureMonitorMetric{
				DataType:    v1alphaSLO.AzureMonitorDataTypeMetrics,
				ResourceID:  "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/myResourceGroup/providers/Microsoft.Compute/virtualMachines/api-server",
				MetricName:  "Requests",
				Aggregation: "Sum",
			}))
		case metricVariantBadRatio + metricSubVariantAzureMonitorMetrics:
			return setBadOverTotalMetric(slo, newMetricSpec(v1alphaSLO.AzureMonitorMetric{
				DataType:    v1alphaSLO.AzureMonitorDataTypeMetrics,
				ResourceID:  "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/myResourceGroup/providers/Microsoft.Compute/virtualMachines/api-server",
				MetricName:  "Http4xx",
				Aggregation: "Sum",
			}), newMetricSpec(v1alphaSLO.AzureMonitorMetric{
				DataType:    v1alphaSLO.AzureMonitorDataTypeMetrics,
				ResourceID:  "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/myResourceGroup/providers/Microsoft.Compute/virtualMachines/api-server",
				MetricName:  "Requests",
				Aggregation: "Sum",
			}))
		case metricVariantThreshold + metricSubVariantAzureMonitorLogs:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.AzureMonitorMetric{
				DataType: v1alphaSLO.AzureMonitorDataTypeLogs,
				KQLQuery: `AppRequests
| where AppRoleName == "api-server"
| summarize n9_value = avg(DurationMs) by bin(TimeGenerated, 15s)
| project n9_time = TimeGenerated, n9_value`,
				Workspace: &v1alphaSLO.AzureMonitorMetricLogAnalyticsWorkspace{
					SubscriptionID: "00000000-0000-0000-0000-000000000000",
					ResourceGroup:  "myResourceGroup",
					WorkspaceID:    "11111111-1111-1111-1111-111111111111",
				},
			}))
		case metricVariantGoodRatio + metricSubVariantAzureMonitorLogs:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.AzureMonitorMetric{
				DataType: v1alphaSLO.AzureMonitorDataTypeLogs,
				KQLQuery: `AppRequests
| where AppRoleName == "my-app"
| where ResultCode >= 200 and ResultCode < 400
| summarize n9_value = count() by bin(TimeGenerated, 15s)
| project n9_time = TimeGenerated, n9_value`,
				Workspace: &v1alphaSLO.AzureMonitorMetricLogAnalyticsWorkspace{
					SubscriptionID: "00000000-0000-0000-0000-000000000000",
					ResourceGroup:  "myResourceGroup",
					WorkspaceID:    "11111111-1111-1111-1111-111111111111",
				},
			}), newMetricSpec(v1alphaSLO.AzureMonitorMetric{
				DataType: v1alphaSLO.AzureMonitorDataTypeLogs,
				KQLQuery: `AppRequests
| where AppRoleName == "my-app"
| summarize n9_value = count() by bin(TimeGenerated, 15s)
| project n9_time = TimeGenerated, n9_value`,
				Workspace: &v1alphaSLO.AzureMonitorMetricLogAnalyticsWorkspace{
					SubscriptionID: "00000000-0000-0000-0000-000000000000",
					ResourceGroup:  "myResourceGroup",
					WorkspaceID:    "11111111-1111-1111-1111-111111111111",
				},
			}))
		case metricVariantBadRatio + metricSubVariantAzureMonitorLogs:
			return setBadOverTotalMetric(slo, newMetricSpec(v1alphaSLO.AzureMonitorMetric{
				DataType: v1alphaSLO.AzureMonitorDataTypeLogs,
				KQLQuery: `AppRequests
| where AppRoleName == "my-app"
| where ResultCode == 0 or ResultCode >= 400
| summarize n9_value = count() by bin(TimeGenerated, 15s)
| project n9_time = TimeGenerated, n9_value`,
				Workspace: &v1alphaSLO.AzureMonitorMetricLogAnalyticsWorkspace{
					SubscriptionID: "00000000-0000-0000-0000-000000000000",
					ResourceGroup:  "myResourceGroup",
					WorkspaceID:    "11111111-1111-1111-1111-111111111111",
				},
			}), newMetricSpec(v1alphaSLO.AzureMonitorMetric{
				DataType: v1alphaSLO.AzureMonitorDataTypeLogs,
				KQLQuery: `AppRequests
| where AppRoleName == "my-app"
| summarize n9_value = count() by bin(TimeGenerated, 15s)
| project n9_time = TimeGenerated, n9_value`,
				Workspace: &v1alphaSLO.AzureMonitorMetricLogAnalyticsWorkspace{
					SubscriptionID: "00000000-0000-0000-0000-000000000000",
					ResourceGroup:  "myResourceGroup",
					WorkspaceID:    "11111111-1111-1111-1111-111111111111",
				},
			}))
		}
	case v1alpha.Generic:
		switch s.MetricVariant {
		case metricVariantThreshold:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.GenericMetric{
				Query: ptr(`SINCE N9FROM UNTIL N9TO FROM a1: entities(aws:postgresql:123) FETCH a1.metrics("infra:database.cpu.utilization", "aws-cloudwatch"){timestamp, value} LIMITS metrics.granularityDuration(PT1M)`),
			}))
		case metricVariantGoodRatio:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.GenericMetric{
				Query: ptr(`SINCE N9FROM UNTIL N9TO FROM a1: entities(aws:postgresql:123) FETCH a1.metrics("infra:database.requests.good", "aws-cloudwatch"){timestamp, value} LIMITS metrics.granularityDuration(PT1M)`),
			}), newMetricSpec(v1alphaSLO.GenericMetric{
				Query: ptr(`SINCE N9FROM UNTIL N9TO FROM a1: entities(aws:postgresql:123) FETCH a1.metrics("infra:database.requests.total", "aws-cloudwatch"){timestamp, value} LIMITS metrics.granularityDuration(PT1M)`),
			}))
		}
	case v1alpha.Honeycomb:
		switch s.MetricVariant {
		case metricVariantThreshold:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.HoneycombMetric{
				Calculation: "AVG",
				Attribute:   "requestsLatency",
			}))
		case metricVariantGoodRatio:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.HoneycombMetric{
				Calculation: "SUM",
				Attribute:   "counterGood",
			}), newMetricSpec(v1alphaSLO.HoneycombMetric{
				Calculation: "SUM",
				Attribute:   "counterTotal",
			}))
		case metricVariantBadRatio:
			return setBadOverTotalMetric(slo, newMetricSpec(v1alphaSLO.HoneycombMetric{
				Calculation: "SUM",
				Attribute:   "counterBad",
			}), newMetricSpec(v1alphaSLO.HoneycombMetric{
				Calculation: "SUM",
				Attribute:   "counterTotal",
			}))
		case metricVariantSingleQueryGoodRatio:
			return setSingleQueryGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.HoneycombMetric{
				Attribute: "dc.sli.some-service-availability",
			}))
		}
	case v1alpha.LogicMonitor:
		switch s.MetricVariant {
		case metricVariantThreshold:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.LogicMonitorMetric{
				QueryType:    v1alphaSLO.LMQueryTypeWebsiteMetrics,
				WebsiteID:    "1",
				CheckpointID: "1044712023",
				GraphName:    "responseTime",
				Line:         "MIN RTT",
			}))
		case metricVariantGoodRatio:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.LogicMonitorMetric{
				QueryType:                  v1alphaSLO.LMQueryTypeDeviceMetrics,
				DeviceDataSourceInstanceID: 933147615,
				GraphID:                    11438,
				Line:                       "CONNECTIONSUCCESSES",
			}), newMetricSpec(v1alphaSLO.LogicMonitorMetric{
				QueryType:                  v1alphaSLO.LMQueryTypeDeviceMetrics,
				DeviceDataSourceInstanceID: 933147615,
				GraphID:                    11436,
				Line:                       "CONNECTIONSESTABLISHED",
			}))
		case metricVariantBadRatio:
			return setBadOverTotalMetric(slo, newMetricSpec(v1alphaSLO.LogicMonitorMetric{
				QueryType:                  v1alphaSLO.LMQueryTypeDeviceMetrics,
				DeviceDataSourceInstanceID: 933147615,
				GraphID:                    11437,
				Line:                       "CONNECTIONFAILURES",
			}), newMetricSpec(v1alphaSLO.LogicMonitorMetric{
				QueryType:                  v1alphaSLO.LMQueryTypeDeviceMetrics,
				DeviceDataSourceInstanceID: 933147615,
				GraphID:                    11436,
				Line:                       "CONNECTIONSESTABLISHED",
			}))
		}
	case v1alpha.AzurePrometheus:
		switch s.MetricVariant {
		case metricVariantThreshold:
			return setThresholdMetric(slo, newMetricSpec(v1alphaSLO.AzurePrometheusMetric{
				PromQL: `sum((rate(container_cpu_usage_seconds_total{container!="POD",container!=""}[30m])
- on (namespace,pod,container) group_left avg by (namespace,pod,container)(kube_pod_container_resource_requests{resource="cpu"}))
* -1 >0)`,
			}))
		case metricVariantGoodRatio:
			return setGoodOverTotalMetric(slo, newMetricSpec(v1alphaSLO.AzurePrometheusMetric{
				PromQL: `sum(api_server_requests_total{code="2xx"})`,
			}), newMetricSpec(v1alphaSLO.AzurePrometheusMetric{
				PromQL: `sum(api_server_requests_total{})`,
			}))
		case metricVariantBadRatio:
			return setBadOverTotalMetric(slo, newMetricSpec(v1alphaSLO.AzurePrometheusMetric{
				PromQL: `sum(api_server_requests_total{code="5xx"})`,
			}), newMetricSpec(v1alphaSLO.AzurePrometheusMetric{
				PromQL: `sum(api_server_requests_total{})`,
			}))
		}
	default:
		panic(fmt.Sprintf("unsupported data source type: %s", s.DataSourceType))
	}
	panic(fmt.Sprintf("unsupported data source type and/or variants: %s %s %s",
		s.DataSourceType, s.MetricVariant, s.MetricSubVariant))
}

func setThresholdMetric(slo v1alphaSLO.SLO, metricSpec *v1alphaSLO.MetricSpec) v1alphaSLO.SLO {
	slo.Spec.Objectives[0].CountMetrics = nil
	slo.Spec.Objectives[0].RawMetric.MetricQuery = metricSpec
	return slo
}

func setGoodOverTotalMetric(slo v1alphaSLO.SLO, good, total *v1alphaSLO.MetricSpec) v1alphaSLO.SLO {
	slo.Spec.Objectives[0].RawMetric = nil
	slo.Spec.Objectives[0].CountMetrics.GoodMetric = good
	slo.Spec.Objectives[0].CountMetrics.TotalMetric = total
	slo.Spec.Objectives[0].CountMetrics.GoodTotalMetric = nil
	return slo
}

func setBadOverTotalMetric(slo v1alphaSLO.SLO, bad, total *v1alphaSLO.MetricSpec) v1alphaSLO.SLO {
	slo.Spec.Objectives[0].RawMetric = nil
	slo.Spec.Objectives[0].CountMetrics.BadMetric = bad
	slo.Spec.Objectives[0].CountMetrics.TotalMetric = total
	slo.Spec.Objectives[0].CountMetrics.GoodTotalMetric = nil
	return slo
}

func setSingleQueryGoodOverTotalMetric(slo v1alphaSLO.SLO, goodTotal *v1alphaSLO.MetricSpec) v1alphaSLO.SLO {
	slo.Spec.Objectives[0].RawMetric = nil
	slo.Spec.Objectives[0].CountMetrics.GoodMetric = nil
	slo.Spec.Objectives[0].CountMetrics.TotalMetric = nil
	slo.Spec.Objectives[0].CountMetrics.GoodTotalMetric = goodTotal
	return slo
}

func newMetricSpec(metric any) *v1alphaSLO.MetricSpec {
	v := reflect.ValueOf(metric)
	if v.Kind() == reflect.Ptr {
		metric = v.Elem().Interface()
	}
	spec := &v1alphaSLO.MetricSpec{}
	switch v := metric.(type) {
	case v1alphaSLO.PrometheusMetric:
		spec.Prometheus = &v
	case v1alphaSLO.DatadogMetric:
		spec.Datadog = &v
	case v1alphaSLO.NewRelicMetric:
		spec.NewRelic = &v
	case v1alphaSLO.AppDynamicsMetric:
		spec.AppDynamics = &v
	case v1alphaSLO.SplunkMetric:
		spec.Splunk = &v
	case v1alphaSLO.LightstepMetric:
		spec.Lightstep = &v
	case v1alphaSLO.SplunkObservabilityMetric:
		spec.SplunkObservability = &v
	case v1alphaSLO.DynatraceMetric:
		spec.Dynatrace = &v
	case v1alphaSLO.ElasticsearchMetric:
		spec.Elasticsearch = &v
	case v1alphaSLO.ThousandEyesMetric:
		spec.ThousandEyes = &v
	case v1alphaSLO.GraphiteMetric:
		spec.Graphite = &v
	case v1alphaSLO.BigQueryMetric:
		spec.BigQuery = &v
	case v1alphaSLO.OpenTSDBMetric:
		spec.OpenTSDB = &v
	case v1alphaSLO.GrafanaLokiMetric:
		spec.GrafanaLoki = &v
	case v1alphaSLO.CloudWatchMetric:
		spec.CloudWatch = &v
	case v1alphaSLO.PingdomMetric:
		spec.Pingdom = &v
	case v1alphaSLO.AmazonPrometheusMetric:
		spec.AmazonPrometheus = &v
	case v1alphaSLO.RedshiftMetric:
		spec.Redshift = &v
	case v1alphaSLO.SumoLogicMetric:
		spec.SumoLogic = &v
	case v1alphaSLO.InstanaMetric:
		spec.Instana = &v
	case v1alphaSLO.InfluxDBMetric:
		spec.InfluxDB = &v
	case v1alphaSLO.GCMMetric:
		spec.GCM = &v
	case v1alphaSLO.AzureMonitorMetric:
		spec.AzureMonitor = &v
	case v1alphaSLO.GenericMetric:
		spec.Generic = &v
	case v1alphaSLO.HoneycombMetric:
		spec.Honeycomb = &v
	case v1alphaSLO.LogicMonitorMetric:
		spec.LogicMonitor = &v
	case v1alphaSLO.AzurePrometheusMetric:
		spec.AzurePrometheus = &v
	default:
		panic(fmt.Sprintf("unsupported metric type: %T", metric))
	}
	return spec
}

func mustLoadQuery(name string) string {
	data, err := queriesFS.ReadFile(filepath.Join("queries", name))
	if err != nil {
		panic(fmt.Sprintf("failed to load query: %s", err))
	}
	return string(data)
}

func ptr[T any](v T) *T { return &v }
