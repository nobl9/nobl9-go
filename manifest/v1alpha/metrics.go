// Package v1alpha represents objects available in API n9/v1alpha
package v1alpha

import "sort"

// CountMetricsSpec represents set of two time series of good and total counts
type CountMetricsSpec struct {
	Incremental *bool       `json:"incremental" validate:"required"`
	GoodMetric  *MetricSpec `json:"good,omitempty"`
	BadMetric   *MetricSpec `json:"bad,omitempty"`
	TotalMetric *MetricSpec `json:"total" validate:"required"`
}

// RawMetricSpec represents integration with a metric source for a particular objective.
type RawMetricSpec struct {
	MetricQuery *MetricSpec `json:"query" validate:"required"`
}

// MetricSpec defines single time series obtained from data source
type MetricSpec struct {
	Prometheus          *PrometheusMetric          `json:"prometheus,omitempty"`
	Datadog             *DatadogMetric             `json:"datadog,omitempty"`
	NewRelic            *NewRelicMetric            `json:"newRelic,omitempty"`
	AppDynamics         *AppDynamicsMetric         `json:"appDynamics,omitempty"`
	Splunk              *SplunkMetric              `json:"splunk,omitempty"`
	Lightstep           *LightstepMetric           `json:"lightstep,omitempty"`
	SplunkObservability *SplunkObservabilityMetric `json:"splunkObservability,omitempty"`
	Dynatrace           *DynatraceMetric           `json:"dynatrace,omitempty"`
	Elasticsearch       *ElasticsearchMetric       `json:"elasticsearch,omitempty"`
	ThousandEyes        *ThousandEyesMetric        `json:"thousandEyes,omitempty"`
	Graphite            *GraphiteMetric            `json:"graphite,omitempty"`
	BigQuery            *BigQueryMetric            `json:"bigQuery,omitempty"`
	OpenTSDB            *OpenTSDBMetric            `json:"opentsdb,omitempty"`
	GrafanaLoki         *GrafanaLokiMetric         `json:"grafanaLoki,omitempty"`
	CloudWatch          *CloudWatchMetric          `json:"cloudWatch,omitempty"`
	Pingdom             *PingdomMetric             `json:"pingdom,omitempty"`
	AmazonPrometheus    *AmazonPrometheusMetric    `json:"amazonPrometheus,omitempty"`
	Redshift            *RedshiftMetric            `json:"redshift,omitempty"`
	SumoLogic           *SumoLogicMetric           `json:"sumoLogic,omitempty"`
	Instana             *InstanaMetric             `json:"instana,omitempty"`
	InfluxDB            *InfluxDBMetric            `json:"influxdb,omitempty"`
	GCM                 *GCMMetric                 `json:"gcm,omitempty"`
	AzureMonitor        *AzureMonitorMetric        `json:"azureMonitor,omitempty"`
	Generic             *GenericMetric             `json:"generic,omitempty"`
	Honeycomb           *HoneycombMetric           `json:"honeycomb,omitempty"`
}

// PrometheusMetric represents metric from Prometheus
type PrometheusMetric struct {
	PromQL *string `json:"promql" validate:"required" example:"cpu_usage_user{cpu=\"cpu-total\"}"`
}

// AmazonPrometheusMetric represents metric from Amazon Managed Prometheus
type AmazonPrometheusMetric struct {
	PromQL *string `json:"promql" validate:"required" example:"cpu_usage_user{cpu=\"cpu-total\"}"`
}

// DatadogMetric represents metric from Datadog
type DatadogMetric struct {
	Query *string `json:"query" validate:"required"`
}

// NewRelicMetric represents metric from NewRelic
type NewRelicMetric struct {
	NRQL *string `json:"nrql" validate:"required,noSinceOrUntil"`
}

const (
	ThousandEyesNetLatency              = "net-latency"
	ThousandEyesNetLoss                 = "net-loss"
	ThousandEyesWebPageLoad             = "web-page-load"
	ThousandEyesWebDOMLoad              = "web-dom-load"
	ThousandEyesHTTPResponseTime        = "http-response-time"
	ThousandEyesServerAvailability      = "http-server-availability"
	ThousandEyesServerThroughput        = "http-server-throughput"
	ThousandEyesServerTotalTime         = "http-server-total-time"
	ThousandEyesDNSServerResolutionTime = "dns-server-resolution-time"
	ThousandEyesDNSSECValid             = "dns-dnssec-valid"
)

// ThousandEyesMetric represents metric from ThousandEyes
type ThousandEyesMetric struct {
	TestID   *int64  `json:"testID" validate:"required,gte=0"`
	TestType *string `json:"testType" validate:"supportedThousandEyesTestType"`
}

// AppDynamicsMetric represents metric from AppDynamics
type AppDynamicsMetric struct {
	ApplicationName *string `json:"applicationName" validate:"required,notEmpty"`
	MetricPath      *string `json:"metricPath" validate:"required,unambiguousAppDynamicMetricPath"`
}

// SplunkMetric represents metric from Splunk
type SplunkMetric struct {
	Query *string `json:"query" validate:"required,notEmpty,splunkQueryValid"`
}

// LightstepMetric represents metric from Lightstep
type LightstepMetric struct {
	StreamID   *string  `json:"streamId,omitempty"`
	TypeOfData *string  `json:"typeOfData" validate:"required,oneof=latency error_rate good total metric"`
	Percentile *float64 `json:"percentile,omitempty"`
	UQL        *string  `json:"uql,omitempty"`
}

// SplunkObservabilityMetric represents metric from SplunkObservability
type SplunkObservabilityMetric struct {
	Program *string `json:"program" validate:"required"`
}

// DynatraceMetric represents metric from Dynatrace.
type DynatraceMetric struct {
	MetricSelector *string `json:"metricSelector" validate:"required"`
}

// ElasticsearchMetric represents metric from Elasticsearch.
type ElasticsearchMetric struct {
	Index *string `json:"index" validate:"required"`
	Query *string `json:"query" validate:"required,elasticsearchBeginEndTimeRequired"`
}

// CloudWatchMetric represents metric from CloudWatch.
type CloudWatchMetric struct {
	AccountID  *string                     `json:"accountId,omitempty"`
	Region     *string                     `json:"region" validate:"required,max=255"`
	Namespace  *string                     `json:"namespace,omitempty"`
	MetricName *string                     `json:"metricName,omitempty"`
	Stat       *string                     `json:"stat,omitempty"`
	Dimensions []CloudWatchMetricDimension `json:"dimensions,omitempty" validate:"max=10,uniqueDimensionNames,dive"`
	SQL        *string                     `json:"sql,omitempty"`
	JSON       *string                     `json:"json,omitempty"`
}

// RedshiftMetric represents metric from Redshift.
type RedshiftMetric struct {
	Region       *string `json:"region" validate:"required,max=255"`
	ClusterID    *string `json:"clusterId" validate:"required"`
	DatabaseName *string `json:"databaseName" validate:"required"`
	Query        *string `json:"query" validate:"required,redshiftRequiredColumns"`
}

// SumoLogicMetric represents metric from Sumo Logic.
type SumoLogicMetric struct {
	Type         *string `json:"type" validate:"required"`
	Query        *string `json:"query" validate:"required"`
	Quantization *string `json:"quantization,omitempty"`
	Rollup       *string `json:"rollup,omitempty"`
	// For struct level validation refer to sumoLogicStructValidation in pkg/manifest/v1alpha/validator.go
}

// InstanaMetric represents metric from Redshift.
type InstanaMetric struct {
	MetricType     string                           `json:"metricType" validate:"required,oneof=infrastructure application"` //nolint:lll
	Infrastructure *InstanaInfrastructureMetricType `json:"infrastructure,omitempty"`
	Application    *InstanaApplicationMetricType    `json:"application,omitempty"`
}

// InfluxDBMetric represents metric from InfluxDB
type InfluxDBMetric struct {
	Query *string `json:"query" validate:"required,influxDBRequiredPlaceholders"`
}

// GCMMetric represents metric from GCM
type GCMMetric struct {
	Query     string `json:"query" validate:"required"`
	ProjectID string `json:"projectId" validate:"required"`
}

type InstanaInfrastructureMetricType struct {
	MetricRetrievalMethod string  `json:"metricRetrievalMethod" validate:"required,oneof=query snapshot"`
	Query                 *string `json:"query,omitempty"`
	SnapshotID            *string `json:"snapshotId,omitempty"`
	MetricID              string  `json:"metricId" validate:"required"`
	PluginID              string  `json:"pluginId" validate:"required"`
}

type InstanaApplicationMetricType struct {
	MetricID         string                          `json:"metricId" validate:"required,oneof=calls erroneousCalls errors latency"` //nolint:lll
	Aggregation      string                          `json:"aggregation" validate:"required"`
	GroupBy          InstanaApplicationMetricGroupBy `json:"groupBy" validate:"required"`
	APIQuery         string                          `json:"apiQuery" validate:"required,json"`
	IncludeInternal  bool                            `json:"includeInternal,omitempty"`
	IncludeSynthetic bool                            `json:"includeSynthetic,omitempty"`
}

type InstanaApplicationMetricGroupBy struct {
	Tag               string  `json:"tag" validate:"required"`
	TagEntity         string  `json:"tagEntity" validate:"required,oneof=DESTINATION SOURCE NOT_APPLICABLE"`
	TagSecondLevelKey *string `json:"tagSecondLevelKey,omitempty"`
}

// IsStandardConfiguration returns true if the struct represents CloudWatch standard configuration.
func (c CloudWatchMetric) IsStandardConfiguration() bool {
	return c.Stat != nil || c.Dimensions != nil || c.MetricName != nil || c.Namespace != nil
}

// IsSQLConfiguration returns true if the struct represents CloudWatch SQL configuration.
func (c CloudWatchMetric) IsSQLConfiguration() bool {
	return c.SQL != nil
}

// IsJSONConfiguration returns true if the struct represents CloudWatch JSON configuration.
func (c CloudWatchMetric) IsJSONConfiguration() bool {
	return c.JSON != nil
}

// CloudWatchMetricDimension represents name/value pair that is part of the identity of a metric.
type CloudWatchMetricDimension struct {
	Name  *string `json:"name" validate:"required,max=255,ascii,notBlank"`
	Value *string `json:"value" validate:"required,max=255,ascii,notBlank"`
}

// PingdomMetric represents metric from Pingdom.
type PingdomMetric struct {
	CheckID   *string `json:"checkId" validate:"required,notBlank,numeric" example:"1234567"`
	CheckType *string `json:"checkType" validate:"required,pingdomCheckTypeFieldValid" example:"uptime"`
	Status    *string `json:"status,omitempty" validate:"omitempty,pingdomStatusValid" example:"up,down"`
}

// GraphiteMetric represents metric from Graphite.
type GraphiteMetric struct {
	MetricPath *string `json:"metricPath" validate:"required,metricPathGraphite"`
}

// BigQueryMetric represents metric from BigQuery
type BigQueryMetric struct {
	Query     string `json:"query" validate:"required,bigQueryRequiredColumns"`
	ProjectID string `json:"projectId" validate:"required"`
	Location  string `json:"location" validate:"required"`
}

// OpenTSDBMetric represents metric from OpenTSDB.
type OpenTSDBMetric struct {
	Query *string `json:"query" validate:"required"`
}

// GrafanaLokiMetric represents metric from GrafanaLokiMetric.
type GrafanaLokiMetric struct {
	Logql *string `json:"logql" validate:"required"`
}

// AzureMonitorMetric represents metric from AzureMonitor
type AzureMonitorMetric struct {
	ResourceID      string                        `json:"resourceId" validate:"required"`
	MetricName      string                        `json:"metricName" validate:"required"`
	Aggregation     string                        `json:"aggregation" validate:"required"`
	Dimensions      []AzureMonitorMetricDimension `json:"dimensions,omitempty" validate:"uniqueDimensionNames,dive"`
	MetricNamespace string                        `json:"metricNamespace,omitempty"`
}

// AzureMonitorMetricDimension represents name/value pair that is part of the identity of a metric.
type AzureMonitorMetricDimension struct {
	Name  *string `json:"name" validate:"required,max=255,ascii,notBlank"`
	Value *string `json:"value" validate:"required,max=255,ascii,notBlank"`
}

type GenericMetric struct {
	Query *string `json:"query" validate:"required"`
}

// HoneycombMetric represents metric from Honeycomb.
type HoneycombMetric struct {
	// FIXME PC-10654: Check if the entire struct is correct.
	Dataset     string `json:"dataset" validate:"required,max=255,ascii,notBlank"`
	Calculation string `json:"calculation" validate:"required,max=30,ascii,notBlank,supportedHoneycombCalculationType"`
	// FIXME PC-10654: "with validation based on Honeycomb API docs", "max length - 100 characters (TBC)"
	Attribute string          `json:"attribute" validate:"required,max=255,ascii,notBlank"`
	Filter    HoneycombFilter `json:"filter"`
}

// HoneycombFilter represents filter for Honeycomb metric. It has custom struct validation.
type HoneycombFilter struct {
	Operator   string                     `json:"operator" validate:"max=30,ascii"`
	Conditions []HoneycombFilterCondition `json:"conditions" validate:"required,gt=0,lte=100,dive"`
}

// HoneycombFilterCondition represents single condition for Honeycomb filter.
type HoneycombFilterCondition struct {
	Attribute string `json:"attribute" validate:"required,max=255,ascii,notBlank"`
	Operator  string `json:"operator" validate:"required,max=30,ascii,notBlank,supportedHoneycombFilterConditionOperator"`
	Value     string `json:"value" validate:"max=255,ascii"`
}

func (slo *SLOSpec) containsIndicatorRawMetric() bool {
	return slo.Indicator.RawMetric != nil
}

// IsComposite returns true if SLOSpec contains composite type.
func (slo *SLOSpec) IsComposite() bool {
	return slo.Composite != nil
}

// HasRawMetric returns true if SLOSpec has raw metric.
func (slo *SLOSpec) HasRawMetric() bool {
	if slo.containsIndicatorRawMetric() {
		return true
	}
	for _, objective := range slo.Objectives {
		if objective.HasRawMetricQuery() {
			return true
		}
	}
	return false
}

// RawMetrics returns raw metric spec.
func (slo *SLOSpec) RawMetrics() []*MetricSpec {
	if slo.containsIndicatorRawMetric() {
		return []*MetricSpec{slo.Indicator.RawMetric}
	}
	rawMetrics := make([]*MetricSpec, 0, slo.ObjectivesRawMetricsCount())
	for _, objective := range slo.Objectives {
		if objective.RawMetric != nil {
			rawMetrics = append(rawMetrics, objective.RawMetric.MetricQuery)
		}
	}
	return rawMetrics
}

// HasRawMetricQuery returns true if Objective has raw metric with query set.
func (o *Objective) HasRawMetricQuery() bool {
	return o.RawMetric != nil && o.RawMetric.MetricQuery != nil
}

// ObjectivesRawMetricsCount returns total number of all raw metrics defined in this SLO Spec's objectives.
func (slo *SLOSpec) ObjectivesRawMetricsCount() int {
	var count int
	for _, objective := range slo.Objectives {
		if objective.HasRawMetricQuery() {
			count++
		}
	}
	return count
}

// HasCountMetrics returns true if SLOSpec has count metrics.
func (slo *SLOSpec) HasCountMetrics() bool {
	for _, objective := range slo.Objectives {
		if objective.HasCountMetrics() {
			return true
		}
	}
	return false
}

// HasCountMetrics returns true if Objective has count metrics.
func (o *Objective) HasCountMetrics() bool {
	return o.CountMetrics != nil
}

// CountMetricsCount returns total number of all count metrics defined in this SLOSpec's objectives.
func (slo *SLOSpec) CountMetricsCount() int {
	var count int
	for _, objective := range slo.Objectives {
		if objective.CountMetrics != nil {
			if objective.CountMetrics.GoodMetric != nil {
				count++
			}
			if objective.CountMetrics.TotalMetric != nil {
				count++
			}
			if objective.CountMetrics.BadMetric != nil && isBadOverTotalEnabledForDataSourceType(objective) {
				count++
			}
		}
	}
	return count
}

// CountMetrics returns a flat slice of all count metrics defined in this SLOSpec's objectives.
func (slo *SLOSpec) CountMetrics() []*MetricSpec {
	countMetrics := make([]*MetricSpec, slo.CountMetricsCount())
	var i int
	for _, objective := range slo.Objectives {
		if objective.CountMetrics == nil {
			continue
		}
		if objective.CountMetrics.GoodMetric != nil {
			countMetrics[i] = objective.CountMetrics.GoodMetric
			i++
		}
		if objective.CountMetrics.TotalMetric != nil {
			countMetrics[i] = objective.CountMetrics.TotalMetric
			i++
		}
		if objective.CountMetrics.BadMetric != nil && isBadOverTotalEnabledForDataSourceType(objective) {
			countMetrics[i] = objective.CountMetrics.BadMetric
			i++
		}
	}
	return countMetrics
}

// CountMetricPairs returns a slice of all count metrics defined in this SLOSpec's objectives.
func (slo *SLOSpec) CountMetricPairs() []*CountMetricsSpec {
	countMetrics := make([]*CountMetricsSpec, slo.CountMetricsCount())
	var i int
	for _, objective := range slo.Objectives {
		if objective.CountMetrics == nil {
			continue
		}
		if objective.CountMetrics.GoodMetric != nil && objective.CountMetrics.TotalMetric != nil {
			countMetrics[i] = objective.CountMetrics
			i++
		}
	}
	return countMetrics
}

func (slo *SLOSpec) GoodTotalCountMetrics() (good, total []*MetricSpec) {
	for _, objective := range slo.Objectives {
		if objective.CountMetrics == nil {
			continue
		}
		if objective.CountMetrics.GoodMetric != nil {
			good = append(good, objective.CountMetrics.GoodMetric)
		}
		if objective.CountMetrics.TotalMetric != nil {
			total = append(total, objective.CountMetrics.TotalMetric)
		}
	}
	return
}

// AllMetricSpecs returns slice of all metrics defined in SLO regardless of their type.
func (slo *SLOSpec) AllMetricSpecs() []*MetricSpec {
	var metrics []*MetricSpec
	metrics = append(metrics, slo.RawMetrics()...)
	metrics = append(metrics, slo.CountMetrics()...)
	return metrics
}

// DataSourceType returns a type of data source.
func (m *MetricSpec) DataSourceType() DataSourceType {
	switch {
	case m.Prometheus != nil:
		return Prometheus
	case m.Datadog != nil:
		return Datadog
	case m.NewRelic != nil:
		return NewRelic
	case m.AppDynamics != nil:
		return AppDynamics
	case m.Splunk != nil:
		return Splunk
	case m.Lightstep != nil:
		return Lightstep
	case m.SplunkObservability != nil:
		return SplunkObservability
	case m.Dynatrace != nil:
		return Dynatrace
	case m.Elasticsearch != nil:
		return Elasticsearch
	case m.ThousandEyes != nil:
		return ThousandEyes
	case m.Graphite != nil:
		return Graphite
	case m.BigQuery != nil:
		return BigQuery
	case m.OpenTSDB != nil:
		return OpenTSDB
	case m.GrafanaLoki != nil:
		return GrafanaLoki
	case m.CloudWatch != nil:
		return CloudWatch
	case m.Pingdom != nil:
		return Pingdom
	case m.AmazonPrometheus != nil:
		return AmazonPrometheus
	case m.Redshift != nil:
		return Redshift
	case m.SumoLogic != nil:
		return SumoLogic
	case m.Instana != nil:
		return Instana
	case m.InfluxDB != nil:
		return InfluxDB
	case m.GCM != nil:
		return GCM
	case m.AzureMonitor != nil:
		return AzureMonitor
	case m.Generic != nil:
		return Generic
	case m.Honeycomb != nil:
		return Honeycomb
	default:
		return 0
	}
}

// Query returns interface containing metric query for this MetricSpec.
func (m *MetricSpec) Query() interface{} {
	switch m.DataSourceType() {
	case Prometheus:
		return m.Prometheus
	case Datadog:
		return m.Datadog
	case NewRelic:
		return m.NewRelic
	case AppDynamics:
		return m.AppDynamics
	case Splunk:
		return m.Splunk
	case Lightstep:
		return m.Lightstep
	case SplunkObservability:
		return m.SplunkObservability
	case Dynatrace:
		return m.Dynatrace
	case Elasticsearch:
		return m.Elasticsearch
	case ThousandEyes:
		return m.ThousandEyes
	case Graphite:
		return m.Graphite
	case BigQuery:
		return m.BigQuery
	case OpenTSDB:
		return m.OpenTSDB
	case GrafanaLoki:
		return m.GrafanaLoki
	case CloudWatch:
		// To be clean, entire metric spec is copied so that original value is not mutated.
		var cloudWatchCopy CloudWatchMetric
		cloudWatchCopy = *m.CloudWatch
		// Dimension list is optional. This is done so that during upsert empty slice and nil slice are treated equally.
		if cloudWatchCopy.Dimensions == nil {
			cloudWatchCopy.Dimensions = []CloudWatchMetricDimension{}
		}
		// Dimensions are sorted so that metric_query = '...':jsonb comparison was insensitive to the order in slice.
		// It assumes that all dimensions' names are unique (ensured by validation).
		sort.Slice(cloudWatchCopy.Dimensions, func(i, j int) bool {
			return *cloudWatchCopy.Dimensions[i].Name < *cloudWatchCopy.Dimensions[j].Name
		})
		return cloudWatchCopy
	case Pingdom:
		return m.Pingdom
	case AmazonPrometheus:
		return m.AmazonPrometheus
	case Redshift:
		return m.Redshift
	case SumoLogic:
		return m.SumoLogic
	case Instana:
		return m.Instana
	case InfluxDB:
		return m.InfluxDB
	case GCM:
		return m.GCM
	case AzureMonitor:
		// To be clean, entire metric spec is copied so that original value is not mutated.
		var azureMonitorCopy AzureMonitorMetric
		azureMonitorCopy = *m.AzureMonitor
		// Dimension list is optional. This is done so that during upsert empty slice and nil slice are treated equally.
		if azureMonitorCopy.Dimensions == nil {
			azureMonitorCopy.Dimensions = []AzureMonitorMetricDimension{}
		}
		// Dimensions are sorted so that metric_query = '...':jsonb comparison was insensitive to the order in slice.
		// It assumes that all dimensions' names are unique (ensured by validation).
		sort.Slice(azureMonitorCopy.Dimensions, func(i, j int) bool {
			return *azureMonitorCopy.Dimensions[i].Name < *azureMonitorCopy.Dimensions[j].Name
		})
		return azureMonitorCopy
	case Generic:
		return m.Generic
	case Honeycomb:
		return m.Honeycomb
	default:
		return nil
	}
}
