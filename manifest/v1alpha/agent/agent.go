package agent

import (
	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

//go:generate go run ../../../scripts/generate-object-impl.go Agent

// New creates new Agent instance.
func New(metadata Metadata, spec Spec) Agent {
	return Agent{
		APIVersion: v1alpha.APIVersion,
		Kind:       manifest.KindAgent,
		Metadata:   metadata,
		Spec:       spec,
	}
}

// Agent struct which mapped one to one with kind: Agent yaml definition
type Agent struct {
	APIVersion string        `json:"apiVersion"`
	Kind       manifest.Kind `json:"kind"`
	Metadata   Metadata      `json:"metadata"`
	Spec       Spec          `json:"spec"`
	Status     *Status       `json:"status,omitempty"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
	OktaClientID   string `json:"oktaClientID,omitempty"`
}

type Metadata struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`
	Project     string `json:"project,omitempty"`
}

// Spec represents content of Spec typical for Agent Object
type Spec struct {
	Description             string                           `json:"description,omitempty"`
	ReleaseChannel          v1alpha.ReleaseChannel           `json:"releaseChannel,omitempty"`
	Prometheus              *PrometheusConfig                `json:"prometheus,omitempty"`
	Datadog                 *DatadogConfig                   `json:"datadog,omitempty"`
	NewRelic                *NewRelicConfig                  `json:"newRelic,omitempty"`
	AppDynamics             *AppDynamicsConfig               `json:"appDynamics,omitempty"`
	Splunk                  *SplunkConfig                    `json:"splunk,omitempty"`
	Lightstep               *LightstepConfig                 `json:"lightstep,omitempty"`
	SplunkObservability     *SplunkObservabilityConfig       `json:"splunkObservability,omitempty"`
	Dynatrace               *DynatraceConfig                 `json:"dynatrace,omitempty"`
	Elasticsearch           *ElasticsearchConfig             `json:"elasticsearch,omitempty"`
	ThousandEyes            *ThousandEyesConfig              `json:"thousandEyes,omitempty"`
	Graphite                *GraphiteConfig                  `json:"graphite,omitempty"`
	BigQuery                *BigQueryConfig                  `json:"bigQuery,omitempty"`
	OpenTSDB                *OpenTSDBConfig                  `json:"opentsdb,omitempty"`
	GrafanaLoki             *GrafanaLokiConfig               `json:"grafanaLoki,omitempty"`
	CloudWatch              *CloudWatchConfig                `json:"cloudWatch,omitempty"`
	Pingdom                 *PingdomConfig                   `json:"pingdom,omitempty"`
	AmazonPrometheus        *AmazonPrometheusConfig          `json:"amazonPrometheus,omitempty"`
	Redshift                *RedshiftConfig                  `json:"redshift,omitempty"`
	SumoLogic               *SumoLogicConfig                 `json:"sumoLogic,omitempty"`
	Instana                 *InstanaConfig                   `json:"instana,omitempty"`
	InfluxDB                *InfluxDBConfig                  `json:"influxdb,omitempty"`
	AzureMonitor            *AzureMonitorConfig              `json:"azureMonitor,omitempty"`
	GCM                     *GCMConfig                       `json:"gcm,omitempty"`
	Generic                 *GenericConfig                   `json:"generic,omitempty"`
	Honeycomb               *HoneycombConfig                 `json:"honeycomb,omitempty"`
	HistoricalDataRetrieval *v1alpha.HistoricalDataRetrieval `json:"historicalDataRetrieval,omitempty"`
	QueryDelay              *v1alpha.QueryDelay              `json:"queryDelay,omitempty"`
}

// Status holds dynamic content which is not part of the static Agent definition.
type Status struct {
	AgentType      string `json:"agentType"`
	AgentVersion   string `json:"agentVersion,omitempty"`
	LastConnection string `json:"lastConnection,omitempty"`
}

func (spec Spec) GetType() (v1alpha.DataSourceType, error) {
	switch {
	case spec.Prometheus != nil:
		return v1alpha.Prometheus, nil
	case spec.Datadog != nil:
		return v1alpha.Datadog, nil
	case spec.NewRelic != nil:
		return v1alpha.NewRelic, nil
	case spec.AppDynamics != nil:
		return v1alpha.AppDynamics, nil
	case spec.Splunk != nil:
		return v1alpha.Splunk, nil
	case spec.Lightstep != nil:
		return v1alpha.Lightstep, nil
	case spec.SplunkObservability != nil:
		return v1alpha.SplunkObservability, nil
	case spec.Dynatrace != nil:
		return v1alpha.Dynatrace, nil
	case spec.Elasticsearch != nil:
		return v1alpha.Elasticsearch, nil
	case spec.ThousandEyes != nil:
		return v1alpha.ThousandEyes, nil
	case spec.Graphite != nil:
		return v1alpha.Graphite, nil
	case spec.BigQuery != nil:
		return v1alpha.BigQuery, nil
	case spec.OpenTSDB != nil:
		return v1alpha.OpenTSDB, nil
	case spec.GrafanaLoki != nil:
		return v1alpha.GrafanaLoki, nil
	case spec.CloudWatch != nil:
		return v1alpha.CloudWatch, nil
	case spec.Pingdom != nil:
		return v1alpha.Pingdom, nil
	case spec.AmazonPrometheus != nil:
		return v1alpha.AmazonPrometheus, nil
	case spec.Redshift != nil:
		return v1alpha.Redshift, nil
	case spec.SumoLogic != nil:
		return v1alpha.SumoLogic, nil
	case spec.Instana != nil:
		return v1alpha.Instana, nil
	case spec.InfluxDB != nil:
		return v1alpha.InfluxDB, nil
	case spec.GCM != nil:
		return v1alpha.GCM, nil
	case spec.AzureMonitor != nil:
		return v1alpha.AzureMonitor, nil
	case spec.Generic != nil:
		return v1alpha.Generic, nil
	case spec.Honeycomb != nil:
		return v1alpha.Honeycomb, nil
	}
	return 0, errors.New("unknown agent type")
}

// PrometheusConfig represents content of Prometheus Configuration typical for Agent Object.
type PrometheusConfig struct {
	URL string `json:"url"`
}

// DatadogConfig represents content of Datadog Configuration typical for Agent Object.
type DatadogConfig struct {
	Site string `json:"site"`
}

// NewRelicConfig represents content of NewRelic Configuration typical for Agent Object.
type NewRelicConfig struct {
	AccountID int `json:"accountId"`
}

// AppDynamicsConfig represents content of AppDynamics Configuration typical for Agent Object.
type AppDynamicsConfig struct {
	URL string `json:"url"`
}

// SplunkConfig represents content of Splunk Configuration typical for Agent Object.
type SplunkConfig struct {
	URL string `json:"url"`
}

// LightstepConfig represents content of Lightstep Configuration typical for Agent Object.
type LightstepConfig struct {
	Organization string `json:"organization"`
	Project      string `json:"project"`
}

// SplunkObservabilityConfig represents content of SplunkObservability Configuration typical for Agent Object.
type SplunkObservabilityConfig struct {
	Realm string `json:"realm"`
}

// DynatraceConfig represents content of Dynatrace Configuration typical for Agent Object.
type DynatraceConfig struct {
	URL string `json:"url"`
}

// ElasticsearchConfig represents content of Elasticsearch Configuration typical for Agent Object.
type ElasticsearchConfig struct {
	URL string `json:"url"`
}

// ThousandEyesConfig represents content of ThousandEyes Configuration typical for Agent Object.
type ThousandEyesConfig struct{}

// GraphiteConfig represents content of Graphite Configuration typical for Agent Object.
type GraphiteConfig struct {
	URL string `json:"url"`
}

// BigQueryConfig represents content of BigQuery configuration.
type BigQueryConfig struct{}

// OpenTSDBConfig represents content of OpenTSDBConfig Configuration typical for Agent Object.
type OpenTSDBConfig struct {
	URL string `json:"url"`
}

// GrafanaLokiConfig represents content of GrafanaLoki Configuration typical for Agent Object.
type GrafanaLokiConfig struct {
	URL string `json:"url"`
}

// CloudWatchConfig represents content of CloudWatch Configuration typical for Agent Object.
type CloudWatchConfig struct{}

// PingdomConfig represents content of Pingdom Configuration typical for Agent Object.
type PingdomConfig struct{}

// AmazonPrometheusConfig represents content of Amazon Managed Service Configuration typical for Agent Object.
type AmazonPrometheusConfig struct {
	URL    string `json:"url"`
	Region string `json:"region"`
}

// RedshiftConfig represents content of Redshift configuration typical for Agent Object
type RedshiftConfig struct{}

// SumoLogicConfig represents content of Sumo Logic configuration typical for Agent Object.
type SumoLogicConfig struct {
	URL string `json:"url"`
}

// InstanaConfig represents content of Instana configuration typical for Agent Object
type InstanaConfig struct {
	URL string `json:"url"`
}

// InfluxDBConfig represents content of InfluxDB configuration typical fo Agent Object
type InfluxDBConfig struct {
	URL string `json:"url"`
}

// GCMConfig represents content of GCM configuration.
type GCMConfig struct{}

// AzureMonitorConfig represents content of AzureMonitor Configuration typical for Agent Object.
type AzureMonitorConfig struct {
	TenantID string `json:"tenantId"`
}

// GenericConfig represents content of Generic Configuration typical for Agent Object.
type GenericConfig struct{}

// HoneycombConfig represents content of Honeycomb Configuration typical for Agent Object.
type HoneycombConfig struct{}
