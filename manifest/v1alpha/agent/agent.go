package agent

import (
	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

//go:generate go run ../../../internal/cmd/objectimpl Agent

// New creates new Agent instance.
func New(metadata Metadata, spec Spec) Agent {
	return Agent{
		APIVersion: manifest.VersionV1alpha,
		Kind:       manifest.KindAgent,
		Metadata:   metadata,
		Spec:       spec,
	}
}

// Agent struct which mapped one to one with kind: Agent yaml definition
type Agent struct {
	APIVersion manifest.Version `json:"apiVersion"`
	Kind       manifest.Kind    `json:"kind"`
	Metadata   Metadata         `json:"metadata"`
	Spec       Spec             `json:"spec"`
	Status     *Status          `json:"status,omitempty" nobl9:"computed"`

	Organization   string `json:"organization,omitempty" nobl9:"computed"`
	ManifestSource string `json:"manifestSrc,omitempty" nobl9:"computed"`
	OktaClientID   string `json:"oktaClientID,omitempty" nobl9:"computed"`
}

type Metadata struct {
	Name        string                      `json:"name"`
	DisplayName string                      `json:"displayName,omitempty"`
	Project     string                      `json:"project,omitempty"`
	Annotations v1alpha.MetadataAnnotations `json:"annotations,omitempty"`
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
	LogicMonitor            *LogicMonitorConfig              `json:"logicMonitor,omitempty"`
	AzurePrometheus         *AzurePrometheusConfig           `json:"azurePrometheus,omitempty"`
	Coralogix               *CoralogixConfig                 `json:"coralogix,omitempty"`
	HistoricalDataRetrieval *v1alpha.HistoricalDataRetrieval `json:"historicalDataRetrieval,omitempty"`
	QueryDelay              *v1alpha.QueryDelay              `json:"queryDelay,omitempty"`
	// Interval, Timeout and Jitter are readonly and cannot be set via API
	Interval *v1alpha.Interval `json:"interval,omitempty" nobl9:"computed"`
	Timeout  *v1alpha.Timeout  `json:"timeout,omitempty" nobl9:"computed"`
	Jitter   *v1alpha.Jitter   `json:"jitter,omitempty" nobl9:"computed"`
}

// Status holds dynamic content which is not part of the static Agent definition.
type Status struct {
	AgentType                string      `json:"agentType"`
	AgentVersion             string      `json:"agentVersion,omitempty"`
	LastConnection           string      `json:"lastConnection,omitempty"`
	NewestStableAgentVersion string      `json:"newestStableAgentVersion,omitempty"`
	NewestBetaAgentVersion   string      `json:"newestBetaAgentVersion,omitempty"`
	Environment              Environment `json:"environment"`
}

// Environment holds environment-specific variables for the Agent.
type Environment struct {
	JitterOverride   *v1alpha.Jitter   `json:"jitterOverride,omitempty"`
	IntervalOverride *v1alpha.Interval `json:"intervalOverride,omitempty"`
}

func (s Spec) GetType() (v1alpha.DataSourceType, error) {
	switch {
	case s.Prometheus != nil:
		return v1alpha.Prometheus, nil
	case s.Datadog != nil:
		return v1alpha.Datadog, nil
	case s.NewRelic != nil:
		return v1alpha.NewRelic, nil
	case s.AppDynamics != nil:
		return v1alpha.AppDynamics, nil
	case s.Splunk != nil:
		return v1alpha.Splunk, nil
	case s.Lightstep != nil:
		return v1alpha.Lightstep, nil
	case s.SplunkObservability != nil:
		return v1alpha.SplunkObservability, nil
	case s.Dynatrace != nil:
		return v1alpha.Dynatrace, nil
	case s.Elasticsearch != nil:
		return v1alpha.Elasticsearch, nil
	case s.ThousandEyes != nil:
		return v1alpha.ThousandEyes, nil
	case s.Graphite != nil:
		return v1alpha.Graphite, nil
	case s.BigQuery != nil:
		return v1alpha.BigQuery, nil
	case s.OpenTSDB != nil:
		return v1alpha.OpenTSDB, nil
	case s.GrafanaLoki != nil:
		return v1alpha.GrafanaLoki, nil
	case s.CloudWatch != nil:
		return v1alpha.CloudWatch, nil
	case s.Pingdom != nil:
		return v1alpha.Pingdom, nil
	case s.AmazonPrometheus != nil:
		return v1alpha.AmazonPrometheus, nil
	case s.Redshift != nil:
		return v1alpha.Redshift, nil
	case s.SumoLogic != nil:
		return v1alpha.SumoLogic, nil
	case s.Instana != nil:
		return v1alpha.Instana, nil
	case s.InfluxDB != nil:
		return v1alpha.InfluxDB, nil
	case s.GCM != nil:
		return v1alpha.GCM, nil
	case s.AzureMonitor != nil:
		return v1alpha.AzureMonitor, nil
	case s.Generic != nil:
		return v1alpha.Generic, nil
	case s.Honeycomb != nil:
		return v1alpha.Honeycomb, nil
	case s.LogicMonitor != nil:
		return v1alpha.LogicMonitor, nil
	case s.AzurePrometheus != nil:
		return v1alpha.AzurePrometheus, nil
	case s.Coralogix != nil:
		return v1alpha.Coralogix, nil
	}
	return 0, errors.New("unknown agent type")
}

// PrometheusConfig represents content of Prometheus Configuration typical for Agent Object.
type PrometheusConfig struct {
	URL  string `json:"url"`
	Step int    `json:"step,omitempty"`
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
	URL          string `json:"url"`
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
	Step   int    `json:"step,omitempty"`
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
type GCMConfig struct {
	Step int `json:"step,omitempty"`
}

// AzureMonitorConfig represents content of AzureMonitor Configuration typical for Agent Object.
type AzureMonitorConfig struct {
	TenantID string `json:"tenantId"`
}

// GenericConfig represents content of Generic Configuration typical for Agent Object.
type GenericConfig struct{}

// HoneycombConfig represents content of Honeycomb Configuration typical for Agent Object.
type HoneycombConfig struct{}

type LogicMonitorConfig struct {
	Account string `json:"account"`
}

// AzurePrometheusConfig represents content of Azure Monitor managed service for Prometheus typical for Agent Object.
type AzurePrometheusConfig struct {
	URL      string `json:"url"`
	TenantID string `json:"tenantId"`
	Step     int    `json:"step,omitempty"`
}

type CoralogixConfig struct {
	// Domain is the Coralogix domain as defined [here].
	//
	// [here]: https://coralogix.com/docs/user-guides/account-management/account-settings/coralogix-domain/#domains
	Domain string `json:"domain"`
	Step   int    `json:"step,omitempty"`
}
