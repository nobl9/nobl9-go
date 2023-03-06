package nobl9

// CountMetricsSpec represents set of two time series of good and total counts
type CountMetricsSpec struct {
	Incremental *bool       `json:"incremental"`
	GoodMetric  *MetricSpec `json:"good"`
	TotalMetric *MetricSpec `json:"total"`
}

// RawMetricSpec represents integration with a metric source for a particular threshold
type RawMetricSpec struct {
	MetricQuery *MetricSpec `json:"query" validate:"required"`
}

// MetricSpec defines single time series obtained from data source
type MetricSpec struct {
	AmazonPrometheus    *AmazonPrometheusMetric    `json:"amazonPrometheus,omitempty"`
	AppDynamics         *AppDynamicsMetric         `json:"appDynamics,omitempty"`
	BigQuery            *BigQueryMetric            `json:"bigQuery,omitempty"`
	CloudWatch          *CloudWatchMetric          `json:"cloudWatch,omitempty"`
	Datadog             *DatadogMetric             `json:"datadog,omitempty"`
	Dynatrace           *DynatraceMetric           `json:"dynatrace,omitempty"`
	Elasticsearch       *ElasticsearchMetric       `json:"elasticsearch,omitempty"`
	GCM                 *GCMMetric                 `json:"gcm,omitempty"`
	GrafanaLoki         *GrafanaLokiMetric         `json:"grafanaLoki,omitempty"`
	Graphite            *GraphiteMetric            `json:"graphite,omitempty"`
	InfluxDB            *InfluxDBMetric            `json:"influxdb,omitempty"`
	Instana             *InstanaMetric             `json:"instana,omitempty"`
	Lightstep           *LightstepMetric           `json:"lightstep,omitempty"`
	NewRelic            *NewRelicMetric            `json:"newRelic,omitempty"`
	OpenTSDB            *OpenTSDBMetric            `json:"opentsdb,omitempty"`
	Pingdom             *PingdomMetric             `json:"pingdom,omitempty"`
	Prometheus          *PrometheusMetric          `json:"prometheus,omitempty"`
	Redshift            *RedshiftMetric            `json:"redshift,omitempty"`
	Splunk              *SplunkMetric              `json:"splunk,omitempty"`
	SplunkObservability *SplunkObservabilityMetric `json:"splunkObservability,omitempty"`
	SumoLogic           *SumoLogicMetric           `json:"sumoLogic,omitempty"`
	ThousandEyes        *ThousandEyesMetric        `json:"thousandEyes,omitempty"`
}

// AmazonPrometheusMetric describes metric from AmazonPrometheus server.
type AmazonPrometheusMetric struct {
	PromQL *string `json:"promql"`
}

// AppDynamicsMetric describes metric from AppDynamics
type AppDynamicsMetric struct {
	ApplicationName *string `json:"applicationName"`
	MetricPath      *string `json:"metricPath"`
}

// BigQueryMetric describes metric from BigQuery.
type BigQueryMetric struct {
	Query     string `json:"query"`
	ProjectID string `json:"projectId"`
	Location  string `json:"location"`
}

// CloudWatchMetric describes metric from CloudWatch.
type CloudWatchMetric struct {
	Region     *string                     `json:"region"`
	Namespace  *string                     `json:"namespace"`
	MetricName *string                     `json:"metricName"`
	Stat       *string                     `json:"stat"`
	Dimensions []CloudWatchMetricDimension `json:"dimensions"`
	SQL        *string                     `json:"sql"`
	JSON       *string                     `json:"json"`
}

// CloudWatchMetricDimension describes name/value pair that is part of the identity of a metric.
type CloudWatchMetricDimension struct {
	Name  *string `json:"name"`
	Value *string `json:"value"`
}

// DatadogMetric describes metric from Datadog
type DatadogMetric struct {
	Query *string `json:"query"`
}

// DynatraceMetric describes metric from Dynatrace.
type DynatraceMetric struct {
	MetricSelector *string `json:"metricSelector"`
}

// ElasticsearchMetric describes metric from Elasticsearch.
type ElasticsearchMetric struct {
	Query *string `json:"query"`
	Index *string `json:"index"`
}

// GCMMetric describes metadata used to fetch metrics from Google Cloud Monitoring API.
type GCMMetric struct {
	ProjectID string `json:"projectId"`
	Query     string `json:"query"`
}

// GrafanaLokiMetric describes metric from GrafanaLoki.
type GrafanaLokiMetric struct {
	Logql *string `json:"logql"`
}

// GraphiteMetric describes metric from Graphite.
type GraphiteMetric struct {
	MetricPath *string `json:"metricPath"`
}

// InfluxDBMetric describes metric from InfluxDB
type InfluxDBMetric struct {
	Query *string `json:"query"`
}

// InstanaMetric describes metric from Instana server.
type InstanaMetric struct {
	MetricType     string                           `json:"metricType"`
	Infrastructure *InstanaInfrastructureMetricType `json:"infrastructure"`
	Application    *InstanaApplicationMetricType    `json:"application"`
}

// InstanaInfrastructureMetricType describes Infrastructure for InstanaMetric.
type InstanaInfrastructureMetricType struct {
	MetricRetrievalMethod string  `json:"metricRetrievalMethod"`
	Query                 *string `json:"query"`
	SnapshotID            *string `json:"snapshotId"`
	MetricID              string  `json:"metricId"`
	PluginID              string  `json:"pluginId"`
}

// InstanaApplicationMetricType describes Application for InstanaMetric.
type InstanaApplicationMetricType struct {
	MetricID         string                          `json:"metricId"`
	Aggregation      string                          `json:"aggregation"`
	GroupBy          InstanaApplicationMetricGroupBy `json:"groupBy"`
	APIQuery         string                          `json:"apiQuery"`
	IncludeInternal  *bool                           `json:"includeInternal"`
	IncludeSynthetic *bool                           `json:"includeSynthetic"`
}

type InstanaApplicationMetricGroupBy struct {
	Tag               string  `json:"tag"`
	TagEntity         string  `json:"tagEntity"`
	TagSecondLevelKey *string `json:"tagSecondLevelKey"`
}

// LightstepMetric describes metric from Lightstep
type LightstepMetric struct {
	StreamID   *string  `json:"streamId"`
	TypeOfData *string  `json:"typeOfData"`
	Percentile *float64 `json:"percentile"`
	UQL        *string  `json:"uql"`
}

// NewRelicMetric describes metric from NewRelic
type NewRelicMetric struct {
	NRQL *string `json:"nrql"`
}

// OpenTSDBMetric describes metric from OpenTSDB.
type OpenTSDBMetric struct {
	Query *string `json:"query"`
}

// PingdomMetric describes metric from Pingdom.
type PingdomMetric struct {
	CheckID   *string `json:"checkId"`
	CheckType *string `json:"checkType,omitempty"`
	Status    *string `json:"status"`
}

// PrometheusMetric describes metric from Prometheus server
type PrometheusMetric struct {
	PromQL *string `json:"promql"`
}

// RedshiftMetric describes metric from Redshift server.
type RedshiftMetric struct {
	Region       *string `json:"region"`
	ClusterID    *string `json:"clusterId"`
	DatabaseName *string `json:"databaseName"`
	Query        *string `json:"query"`
}

// SplunkMetric describes metric from Splunk
type SplunkMetric struct {
	Query *string `json:"query"`
}

// SplunkObservabilityMetric describes metric from SplunkObservability
type SplunkObservabilityMetric struct {
	Program *string `json:"program"`
}

// SumoLogicMetric describes metric from SumoLogic server.
type SumoLogicMetric struct {
	Type         *string `json:"type"`
	Query        *string `json:"query"`
	Quantization *string `json:"quantization"`
	Rollup       *string `json:"rollup"`
}

// ThousandEyesMetric describes metric from ThousandEyes.
type ThousandEyesMetric struct {
	TestID   *int64  `json:"testID"`
	TestType *string `json:"testType"`
}
