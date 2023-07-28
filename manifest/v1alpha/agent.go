package v1alpha

import (
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/manifest"
)

type AgentsSlice []Agent

func (agents AgentsSlice) Clone() AgentsSlice {
	clone := make([]Agent, len(agents))
	copy(clone, agents)
	return clone
}

// Agent struct which mapped one to one with kind: Agent yaml definition
type Agent struct {
	manifest.ObjectHeader
	Spec   AgentSpec   `json:"spec"`
	Status AgentStatus `json:"status"`
}

func (a *Agent) GetAPIVersion() string {
	return a.APIVersion
}

func (a *Agent) GetKind() manifest.Kind {
	return a.Kind
}

func (a *Agent) GetName() string {
	return a.Metadata.Name
}

func (a *Agent) Validate() error {
	//TODO implement me
	panic("implement me")
}

func (a *Agent) GetProject() string {
	return a.Metadata.Project
}

func (a *Agent) SetProject(project string) {
	a.Metadata.Project = project
}

// AgentSpec represents content of Spec typical for Agent Object
type AgentSpec struct {
	Description             string                          `json:"description,omitempty" validate:"description" example:"Prometheus description"` //nolint:lll
	SourceOf                []string                        `json:"sourceOf" example:"Metrics,Services"`
	ReleaseChannel          ReleaseChannel                  `json:"releaseChannel,omitempty" example:"beta,stable"`
	Prometheus              *PrometheusAgentConfig          `json:"prometheus,omitempty"`
	Datadog                 *DatadogAgentConfig             `json:"datadog,omitempty"`
	NewRelic                *NewRelicAgentConfig            `json:"newRelic,omitempty"`
	AppDynamics             *AppDynamicsAgentConfig         `json:"appDynamics,omitempty"`
	Splunk                  *SplunkAgentConfig              `json:"splunk,omitempty"`
	Lightstep               *LightstepAgentConfig           `json:"lightstep,omitempty"`
	SplunkObservability     *SplunkObservabilityAgentConfig `json:"splunkObservability,omitempty"`
	Dynatrace               *DynatraceAgentConfig           `json:"dynatrace,omitempty"`
	Elasticsearch           *ElasticsearchAgentConfig       `json:"elasticsearch,omitempty"`
	ThousandEyes            *ThousandEyesAgentConfig        `json:"thousandEyes,omitempty"`
	Graphite                *GraphiteAgentConfig            `json:"graphite,omitempty"`
	BigQuery                *BigQueryAgentConfig            `json:"bigQuery,omitempty"`
	OpenTSDB                *OpenTSDBAgentConfig            `json:"opentsdb,omitempty"`
	GrafanaLoki             *GrafanaLokiAgentConfig         `json:"grafanaLoki,omitempty"`
	CloudWatch              *CloudWatchAgentConfig          `json:"cloudWatch,omitempty"`
	Pingdom                 *PingdomAgentConfig             `json:"pingdom,omitempty"`
	AmazonPrometheus        *AmazonPrometheusAgentConfig    `json:"amazonPrometheus,omitempty"`
	Redshift                *RedshiftAgentConfig            `json:"redshift,omitempty"`
	SumoLogic               *SumoLogicAgentConfig           `json:"sumoLogic,omitempty"`
	Instana                 *InstanaAgentConfig             `json:"instana,omitempty"`
	InfluxDB                *InfluxDBAgentConfig            `json:"influxdb,omitempty"`
	GCM                     *GCMAgentConfig                 `json:"gcm,omitempty"`
	HistoricalDataRetrieval *HistoricalDataRetrieval        `json:"historicalDataRetrieval,omitempty"`
	QueryDelay              *QueryDelay                     `json:"queryDelay,omitempty"`
}

func (spec AgentSpec) GetType() (DataSourceType, error) {
	switch {
	case spec.Prometheus != nil:
		return Prometheus, nil
	case spec.Datadog != nil:
		return Datadog, nil
	case spec.NewRelic != nil:
		return NewRelic, nil
	case spec.AppDynamics != nil:
		return AppDynamics, nil
	case spec.Splunk != nil:
		return Splunk, nil
	case spec.Lightstep != nil:
		return Lightstep, nil
	case spec.SplunkObservability != nil:
		return SplunkObservability, nil
	case spec.Dynatrace != nil:
		return Dynatrace, nil
	case spec.Elasticsearch != nil:
		return Elasticsearch, nil
	case spec.ThousandEyes != nil:
		return ThousandEyes, nil
	case spec.Graphite != nil:
		return Graphite, nil
	case spec.BigQuery != nil:
		return BigQuery, nil
	case spec.OpenTSDB != nil:
		return OpenTSDB, nil
	case spec.GrafanaLoki != nil:
		return GrafanaLoki, nil
	case spec.CloudWatch != nil:
		return CloudWatch, nil
	case spec.Pingdom != nil:
		return Pingdom, nil
	case spec.AmazonPrometheus != nil:
		return AmazonPrometheus, nil
	case spec.Redshift != nil:
		return Redshift, nil
	case spec.SumoLogic != nil:
		return SumoLogic, nil
	case spec.Instana != nil:
		return Instana, nil
	case spec.InfluxDB != nil:
		return InfluxDB, nil
	case spec.GCM != nil:
		return GCM, nil
	}
	return 0, errors.New("unknown agent type")
}

// AgentStatus represents content of Status optional for Agent Object
type AgentStatus struct {
	AgentType      string `json:"agentType" example:"Prometheus"`
	AgentVersion   string `json:"agentVersion,omitempty" example:"0.0.9"`
	LastConnection string `json:"lastConnection,omitempty" example:"2020-08-31T14:26:13Z"`
}

// PrometheusAgentConfig represents content of Prometheus Configuration typical for Agent Object.
type PrometheusAgentConfig struct {
	URL    *string `json:"url,omitempty" example:"http://prometheus-service.monitoring:8080"`
	Region string  `json:"region,omitempty" example:"eu-cental-1"`
}

// DatadogAgentConfig represents content of Datadog Configuration typical for Agent Object.
type DatadogAgentConfig struct {
	Site string `json:"site,omitempty" validate:"site" example:"eu,us3.datadoghq.com"`
}

// NewRelicAgentConfig represents content of NewRelic Configuration typical for Agent Object.
type NewRelicAgentConfig struct {
	AccountID json.Number `json:"accountId,omitempty" example:"123654"`
}

// AmazonPrometheusAgentConfig represents content of Amazon Managed Service Configuration typical for Agent Object.
type AmazonPrometheusAgentConfig struct {
	URL    string `json:"url" validate:"required,url"`
	Region string `json:"region" validate:"required,max=255"`
}

// RedshiftAgentConfig represents content of Redshift configuration typical for Agent Object
// Since the agent does not require additional configuration this is just a marker struct.
type RedshiftAgentConfig struct {
}

// OpenTSDBAgentConfig represents content of OpenTSDB Configuration typical for Agent Object.
type OpenTSDBAgentConfig struct {
	URL string `json:"url,omitempty" validate:"required,url" example:"example of OpenTSDB cluster URL"` //nolint: lll
}

// GrafanaLokiAgentConfig represents content of GrafanaLoki Configuration typical for Agent Object.
type GrafanaLokiAgentConfig struct {
	URL string `json:"url,omitempty" validate:"required,url" example:"example of GrafanaLoki cluster URL"` //nolint: lll
}

// CloudWatchAgentConfig represents content of CloudWatch Configuration typical for Agent Object.
type CloudWatchAgentConfig struct {
	// CloudWatch agent doesn't require any additional parameters.
}

// SumoLogicAgentConfig represents content of Sumo Logic configuration typical for Agent Object.
type SumoLogicAgentConfig struct {
	URL string `json:"url" validate:"required,url"`
}

// InstanaAgentConfig represents content of Instana configuration typical for Agent Object
type InstanaAgentConfig struct {
	URL string `json:"url" validate:"required,url"`
}

// InfluxDBAgentConfig represents content of InfluxDB configuration typical fo Agent Object
type InfluxDBAgentConfig struct {
	URL string `json:"url" validate:"required,url"`
}

// PingdomAgentConfig represents content of Pingdom Configuration typical for Agent Object.
type PingdomAgentConfig struct {
	// Pingdom agent doesn't require any additional parameter
}

// GCMAgentConfig represents content of GCM configuration.
// Since the agent does not require additional configuration this is just a marker struct.
type GCMAgentConfig struct {
}

// DynatraceAgentConfig represents content of Dynatrace Configuration typical for Agent Object.
type DynatraceAgentConfig struct {
	URL string `json:"url,omitempty" validate:"required,url,urlDynatrace" example:"https://{your-environment-id}.live.dynatrace.com or https://{your-domain}/e/{your-environment-id}"` //nolint: lll
}

// ElasticsearchAgentConfig represents content of Elasticsearch Configuration typical for Agent Object.
type ElasticsearchAgentConfig struct {
	URL string `json:"url,omitempty" validate:"required,url,urlElasticsearch" example:"https://observability-deployment-946814.es.eu-central-1.aws.cloud.es.io:9243"` //nolint: lll
}

// GraphiteAgentConfig represents content of Graphite Configuration typical for Agent Object.
type GraphiteAgentConfig struct {
	URL string `json:"url,omitempty" validate:"required,url" example:"http://graphite.example.com"`
}

// BigQueryAgentConfig represents content of BigQuery configuration.
// Since the agent does not require additional configuration this is just a marker struct.
type BigQueryAgentConfig struct {
}

// ThousandEyesAgentConfig represents content of ThousandEyes Configuration typical for Agent Object.
type ThousandEyesAgentConfig struct {
	// ThousandEyes agent doesn't require any additional parameters.
}

// SplunkObservabilityAgentConfig represents content of SplunkObservability Configuration typical for Agent Object.
type SplunkObservabilityAgentConfig struct {
	Realm string `json:"realm,omitempty" validate:"required"  example:"us1"`
}

// LightstepAgentConfig represents content of Lightstep Configuration typical for Agent Object.
type LightstepAgentConfig struct {
	Organization string `json:"organization,omitempty" example:"LightStep-Play"`
	Project      string `json:"project,omitempty" example:"play"`
}

// AppDynamicsAgentConfig represents content of AppDynamics Configuration typical for Agent Object.
type AppDynamicsAgentConfig struct {
	URL string `json:"url,omitempty" example:"https://nobl9.saas.appdynamics.com"`
}

// SplunkAgentConfig represents content of Splunk Configuration typical for Agent Object.
type SplunkAgentConfig struct {
	URL string `json:"url,omitempty" example:"https://localhost:8089/servicesNS/admin/"`
}

// genericToAgent converts ObjectGeneric to ObjectAgent
func genericToAgent(o manifest.ObjectGeneric, v validator, onlyHeader bool) (Agent, error) {
	res := Agent{
		ObjectHeader: o.ObjectHeader,
	}
	if onlyHeader {
		return res, nil
	}
	var resSpec AgentSpec
	if err := json.Unmarshal(o.Spec, &resSpec); err != nil {
		err = manifest.EnhanceError(o, err)
		return res, err
	}
	res.Spec = resSpec
	if err := v.Check(res); err != nil {
		err = manifest.EnhanceError(o, err)
		return res, err
	}
	return res, nil
}

// AgentWithSLOs struct which mapped one to one with kind: agent and slo yaml definition
type AgentWithSLOs struct {
	Agent Agent `json:"agent"`
	SLOs  []SLO `json:"slos"`
}
