package nobl9

import "encoding/json"

// Agent struct which mapped one to one with kind: Agent yaml definition.
type Agent struct {
	ObjectHeader
	Spec   AgentSpec   `json:"spec"`
	Status AgentStatus `json:"status"`
}

// AgentWithSLOs struct which mapped one to one with kind: agent and slo yaml definition.
type AgentWithSLOs struct {
	Agent Agent `json:"agent"`
	SLOs  []SLO `json:"slos"`
}

// AgentStatus represents content of Status optional for Agent Object.
type AgentStatus struct {
	AgentType      string `json:"agentType" example:"Prometheus"`
	AgentVersion   string `json:"agentVersion,omitempty" example:"0.0.9"`
	LastConnection string `json:"lastConnection,omitempty" example:"2020-08-31T14:26:13Z"`
}

// AgentSpec represents content of Spec typical for Agent Object.
type AgentSpec struct {
	Description         string                          `json:"description,omitempty"`
	SourceOf            []string                        `json:"sourceOf" example:"Metrics,Services"`
	QueryDelay          *QueryDelayDuration             `json:"queryDelay"`
	AmazonPrometheus    *AmazonPrometheusAgentConfig    `json:"amazonPrometheus,omitempty"`
	AppDynamics         *AppDynamicsAgentConfig         `json:"appDynamics,omitempty"`
	BigQuery            *BigQueryAgentConfig            `json:"bigQuery,omitempty"`
	CloudWatch          *CloudWatchAgentConfig          `json:"cloudWatch,omitempty"`
	Datadog             *DatadogAgentConfig             `json:"datadog,omitempty"`
	Dynatrace           *DynatraceAgentConfig           `json:"dynatrace,omitempty"`
	Elasticsearch       *ElasticsearchAgentConfig       `json:"elasticsearch,omitempty"`
	GCM                 *GCMAgentConfig                 `json:"gcm,omitempty"`
	GrafanaLoki         *GrafanaLokiAgentConfig         `json:"grafanaLoki,omitempty"`
	Graphite            *GraphiteAgentConfig            `json:"graphite,omitempty"`
	InfluxDB            *InfluxDBAgentConfig            `json:"influxdb,omitempty"`
	Instana             *InstanaAgentConfig             `json:"instana,omitempty"`
	Lightstep           *LightstepAgentConfig           `json:"lightstep,omitempty"`
	NewRelic            *NewRelicAgentConfig            `json:"newRelic,omitempty"`
	OpenTSDB            *OpenTSDBAgentConfig            `json:"opentsdb,omitempty"`
	Pingdom             *PingdomAgentConfig             `json:"pingdom,omitempty"`
	Prometheus          *PrometheusAgentConfig          `json:"prometheus,omitempty"`
	Redshift            *RedshiftAgentConfig            `json:"redshift,omitempty"`
	Splunk              *SplunkAgentConfig              `json:"splunk,omitempty"`
	SplunkObservability *SplunkObservabilityAgentConfig `json:"splunkObservability,omitempty"`
	SumoLogic           *SumoLogicAgentConfig           `json:"sumoLogic,omitempty"`
	ThousandEyes        *ThousandEyesAgentConfig        `json:"thousandEyes,omitempty"`
}

// DataSourceStatus represents content of Status optional for DataSource Object.
type DataSourceStatus struct {
	DataSourceType string `json:"dataSourceType" example:"Prometheus"`
}

// PrometheusAgentConfig represents content of Prometheus Configuration typical for DataSource Object.
type PrometheusAgentConfig struct {
	URL              *string                     `json:"url,omitempty" example:"http://prometheus-service.monitoring:8080"`
	ServiceDiscovery *PrometheusServiceDiscovery `json:"serviceDiscovery,omitempty"`
}

// PrometheusServiceDiscovery provides settings for mechanism of auto Service discovery.
type PrometheusServiceDiscovery struct {
	// empty is treated as once, later support 1m, 2d, etc. (for now not validated, skipped)
	Interval string                    `json:"interval,omitempty"`
	Rules    []PrometheusDiscoveryRule `json:"rules,omitempty"`
}

// PrometheusDiscoveryRule provides struct for storing rule for single Service discovery rule from Prometheus.
type PrometheusDiscoveryRule struct {
	Discovery          string        `json:"discovery"`
	ServiceNamePattern string        `json:"serviceNamePattern"`
	Filter             []FilterEntry `json:"filter,omitempty"`
}

// FilterEntry represents single metric label to be matched against value.
type FilterEntry struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// DatadogConfig represents content of Datadog Configuration typical for DataSource Object.
type DatadogConfig struct {
	Site string `json:"site,omitempty"`
}

// DatadogAgentConfig represents content of Datadog Configuration typical for Agent Object.
type DatadogAgentConfig struct {
	Site string `json:"site,omitempty"`
}

// NewRelicConfig represents content of NewRelic Configuration typical for DataSource Object.
type NewRelicConfig struct {
	AccountID json.Number `json:"accountId,omitempty" example:"123654"`
}

// NewRelicAgentConfig represents content of NewRelic Configuration typical for Agent Object.
type NewRelicAgentConfig struct {
	AccountID json.Number `json:"accountId,omitempty" example:"123654"`
}

// AppDynamicsConfig represents content of AppDynamics Configuration typical for DataSource Object.
type AppDynamicsConfig struct {
	URL string `json:"url,omitempty" example:"https://nobl9.saas.appdynamics.com"`
}

// AppDynamicsAgentConfig represents content of AppDynamics Configuration typical for Agent Object.
type AppDynamicsAgentConfig struct {
	URL *string `json:"url,omitempty" example:"https://nobl9.saas.appdynamics.com"`
}

// SplunkConfig represents content of Splunk Configuration typical for DataSource Object.
type SplunkConfig struct {
	URL string `json:"url,omitempty" example:"https://localhost:8089/servicesNS/admin/"`
}

// SplunkAgentConfig represents content of Splunk Configuration typical for Agent Object.
type SplunkAgentConfig struct {
	URL string `json:"url,omitempty" example:"https://localhost:8089/servicesNS/admin/"`
}

// LightstepConfig represents content of Lightstep Configuration typical for DataSource Object.
type LightstepConfig struct {
	Organization string `json:"organization,omitempty" example:"LightStep-Play"`
	Project      string `json:"project,omitempty" example:"play"`
}

// LightstepAgentConfig represents content of Lightstep Configuration typical for Agent Object.
type LightstepAgentConfig struct {
	Organization string `json:"organization,omitempty" example:"LightStep-Play"`
	Project      string `json:"project,omitempty" example:"play"`
}

// SplunkObservabilityAgentConfig represents content of SplunkObservability Configuration typical for Agent Object.
type SplunkObservabilityAgentConfig struct {
	Realm string `json:"realm,omitempty" example:"us1"`
}

// ThousandEyesAgentConfig represents content of ThousandEyes Configuration typical for Agent Object.
type ThousandEyesAgentConfig struct {
	// ThousandEyes agent doesn't require any additional parameters.
}

// DynatraceAgentConfig represents content of Dynatrace Configuration typical for Agent Object.
type DynatraceAgentConfig struct {
	URL string `json:"url,omitempty"` //nolint: lll
}

// DynatraceConfig represents content of Dynatrace Configuration typical for DataSource Object.
type DynatraceConfig struct {
	URL string `json:"url,omitempty"` //nolint: lll
}

// ElasticsearchAgentConfig represents content of Elasticsearch Configuration typical for Agent Object.
type ElasticsearchAgentConfig struct {
	URL string `json:"url,omitempty"`
}

// ElasticsearchConfig represents content of Elasticsearch Configuration typical for DataSource Object.
type ElasticsearchConfig struct {
	URL string `json:"url,omitempty"`
}

// GraphiteAgentConfig represents content of Graphite Configuration typical for Agent Object.
type GraphiteAgentConfig struct {
	URL string `json:"url,omitempty"`
}

// BigQueryAgentConfig represents content of BigQuery configuration.
// Since the agent does not require additional configuration this is just a marker struct.
type BigQueryAgentConfig struct{}

// OpenTSDBAgentConfig represents content of OpenTSDB Configuration typical for Agent Object.
type OpenTSDBAgentConfig struct {
	URL string `json:"url,omitempty" example:"example of OpenTSDB cluster URL"`
}

// GrafanaLokiAgentConfig represents content of GrafanaLoki Configuration typical for Agent Object.
type GrafanaLokiAgentConfig struct {
	URL string `json:"url,omitempty" example:"example of GrafanaLoki cluster URL"`
}

// CloudWatchAgentConfig represents content of CloudWatch Configuration typical for Agent Object.
type CloudWatchAgentConfig struct {
	// CloudWatch agent doesn't require any additional parameters.
}

// PingdomAgentConfig represents content of Pingdom Configuration typical for Agent Object.
type PingdomAgentConfig struct {
	// Pingdom agent doesn't require any additional parameter
}

// AmazonPrometheusAgentConfig represents content of Amazon Managed Service Configuration typical for Agent Object.
type AmazonPrometheusAgentConfig struct {
	URL    string `json:"url" validate:"required,url"`
	Region string `json:"region" validate:"required,max=255"`
}

// RedshiftAgentConfig represents content of Redshift configuration typical for Agent Object.
type RedshiftAgentConfig struct {
	// RedshiftAgentConfig agent doesn't require any additional parameter
}

// SumoLogicAgentConfig represents content of Sumo Logic configuration typical for Agent Object.
type SumoLogicAgentConfig struct {
	URL string `json:"url" validate:"required,url"`
}

// InstanaAgentConfig represents content of Instana configuration typical for Agent Object.
type InstanaAgentConfig struct {
	URL string `json:"url" validate:"required,url"`
}

// InfluxDBAgentConfig represents content of InfluxDB configuration typical fo Agent Object.
type InfluxDBAgentConfig struct {
	URL string `json:"url" validate:"required,url"`
}

// GCMAgentConfig represents content of GCM configuration.
type GCMAgentConfig struct {
	// GCMAgentConfig agent doesn't require any additional parameters.
}

// genericToAgent converts ObjectGeneric to ObjectAgent.
func genericToAgent(o ObjectGeneric, onlyHeader bool) (Agent, error) {
	res := Agent{
		ObjectHeader: o.ObjectHeader,
	}
	if onlyHeader {
		return res, nil
	}
	var resSpec AgentSpec
	if err := json.Unmarshal(o.Spec, &resSpec); err != nil {
		err = EnhanceError(o, err)
		return res, err
	}
	res.Spec = resSpec
	return res, nil
}
