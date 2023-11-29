package agent

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

//go:embed test_data/expected_error.txt
var expectedError string

func TestValidate_AllErrors(t *testing.T) {
	err := validate(Agent{
		Kind: manifest.KindProject,
		Metadata: Metadata{
			Name:        strings.Repeat("MY AGENT", 20),
			DisplayName: strings.Repeat("my-agent", 10),
			Project:     strings.Repeat("MY PROJECT", 20),
		},
		Spec: Spec{
			Description: strings.Repeat("l", 2000),
			Prometheus: &PrometheusConfig{
				URL: "https://prometheus-service.monitoring:8080",
			},
		},
		ManifestSource: "/home/me/agent.yaml",
	})
	assert.Equal(t, strings.TrimSuffix(expectedError, "\n"), err.Error())
}

func TestValidateSpec(t *testing.T) {
	t.Run("exactly one data source - none provided", func(t *testing.T) {
		agent := validAgent(v1alpha.Prometheus)
		agent.Spec.Prometheus = nil
		err := validate(agent)
		testutils.AssertContainsErrors(t, agent, err, 1, testutils.ExpectedError{
			Prop: "spec",
			Code: errCodeExactlyOneDataSourceType,
		})
	})
	t.Run("exactly one data source - both provided", func(t *testing.T) {
		for _, typ := range v1alpha.DataSourceTypeValues() {
			// We're using Prometheus as the offending data source type.
			// Any other source could've been used as well.
			if typ == v1alpha.Prometheus {
				continue
			}
			agent := validAgent(typ)
			agent.Spec.Prometheus = validAgentSpecs[v1alpha.Prometheus].Prometheus
			err := validate(agent)
			testutils.AssertContainsErrors(t, agent, err, 1, testutils.ExpectedError{
				Prop: "spec",
				Code: errCodeExactlyOneDataSourceType,
			})
		}
	})
}

func validAgent(typ v1alpha.DataSourceType) Agent {
	spec := validAgentSpecs[typ]
	spec.Description = "Example Prometheus Agent"
	spec.ReleaseChannel = v1alpha.ReleaseChannelStable
	return New(Metadata{
		Name:        "prometheus",
		DisplayName: "Prometheus Agent",
		Project:     "default",
	}, spec)
}

var validAgentSpecs = map[v1alpha.DataSourceType]Spec{
	v1alpha.Prometheus: {
		Prometheus: &PrometheusConfig{
			URL: "https://prometheus-service.monitoring:8080",
		},
	},
	v1alpha.Datadog: {
		Datadog: &DatadogConfig{
			Site: "https://datadog-service.monitoring:8125",
		},
	},
	v1alpha.NewRelic: {
		NewRelic: &NewRelicConfig{
			AccountID: 123,
		},
	},
	v1alpha.AppDynamics: {
		AppDynamics: &AppDynamicsConfig{
			URL: "https://nobl9.saas.appdynamics.com",
		},
	},
	v1alpha.Splunk: {
		Splunk: &SplunkConfig{
			URL: "https://localhost:8089/servicesNS/admin/",
		},
	},
	v1alpha.Lightstep: {
		Lightstep: &LightstepConfig{
			Organization: "LightStep-Play",
			Project:      "play",
		},
	},
	v1alpha.SplunkObservability: {
		SplunkObservability: &SplunkObservabilityConfig{
			Realm: "us-1",
		},
	},
	v1alpha.Dynatrace: {
		Dynatrace: &DynatraceConfig{
			URL: "https://rxh70845.live.dynatrace.com/",
		},
	},
	v1alpha.Elasticsearch: {
		Elasticsearch: &ElasticsearchConfig{
			URL: "https://observability-deployment-946814.es.eu-central-1.aws.cloud.es.io:9243",
		},
	},
	v1alpha.ThousandEyes: {
		ThousandEyes: &ThousandEyesConfig{},
	},
	v1alpha.Graphite: {
		Graphite: &GraphiteConfig{
			URL: "http://graphite.example.com",
		},
	},
	v1alpha.BigQuery: {
		BigQuery: &BigQueryConfig{},
	},
	v1alpha.OpenTSDB: {
		OpenTSDB: &OpenTSDBConfig{
			URL: "http://opentsdb.example.com",
		},
	},
	v1alpha.GrafanaLoki: {
		GrafanaLoki: &GrafanaLokiConfig{
			URL: "http://loki.example.com",
		},
	},
	v1alpha.CloudWatch: {
		CloudWatch: &CloudWatchConfig{},
	},
	v1alpha.Pingdom: {
		Pingdom: &PingdomConfig{},
	},
	v1alpha.AmazonPrometheus: {
		AmazonPrometheus: &AmazonPrometheusConfig{
			URL:    "https://prometheus-service.monitoring:8080",
			Region: "us-east-1",
		},
	},
	v1alpha.Redshift: {
		Redshift: &RedshiftConfig{},
	},
	v1alpha.SumoLogic: {
		SumoLogic: &SumoLogicConfig{
			URL: "https://sumologic-service.monitoring:443",
		},
	},
	v1alpha.Instana: {
		Instana: &InstanaConfig{
			URL: "https://instana-service.monitoring:443",
		},
	},
	v1alpha.InfluxDB: {
		InfluxDB: &InfluxDBConfig{
			URL: "https://influxdb-service.monitoring:8086",
		},
	},
	v1alpha.GCM: {
		GCM: &GCMConfig{},
	},
	v1alpha.AzureMonitor: {
		AzureMonitor: &AzureMonitorConfig{
			TenantID: "abf988bf-86f1-41af-91ab-2d7cd011db46",
		},
	},
	v1alpha.Generic: {
		Generic: &GenericConfig{},
	},
	v1alpha.Honeycomb: {
		Honeycomb: &HoneycombConfig{},
	},
}
