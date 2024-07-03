package v1alphaExamples

import (
	"fmt"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaAgent "github.com/nobl9/nobl9-go/manifest/v1alpha/agent"
	"github.com/nobl9/nobl9-go/sdk"
)

type agentExample struct {
	standardExample
	typ v1alpha.DataSourceType
}

func Agent() []Example {
	types := v1alpha.DataSourceTypeValues()
	examples := make([]Example, 0, len(types))
	for _, typ := range types {
		example := agentExample{
			standardExample: standardExample{
				Variant: toKebabCase(typ.String()),
			},
			typ: typ,
		}
		example.Object = example.Generate()
		examples = append(examples, example)
	}
	return examples
}

func (a agentExample) Generate() v1alphaAgent.Agent {
	titleName := dataSourceTypePrettyName(a.typ)
	agent := v1alphaAgent.New(
		v1alphaAgent.Metadata{
			Name:        a.Variant,
			DisplayName: titleName + " Agent",
			Project:     sdk.DefaultProject,
		},
		v1alphaAgent.Spec{
			Description:    fmt.Sprintf("Example %s Agent", titleName),
			ReleaseChannel: v1alpha.ReleaseChannelStable,
		},
	)
	agent = a.generateVariant(agent)
	typ, _ := agent.Spec.GetType()
	if maxDuration, err := v1alpha.GetDataRetrievalMaxDuration(manifest.KindAgent, typ); err == nil {
		agent.Spec.HistoricalDataRetrieval = &v1alpha.HistoricalDataRetrieval{
			MaxDuration: maxDuration,
			DefaultDuration: v1alpha.HistoricalRetrievalDuration{
				Value: ptr(*maxDuration.Value / 2),
				Unit:  maxDuration.Unit,
			},
		}
	}
	defaultQueryDelay := v1alpha.GetQueryDelayDefaults()[typ]
	agent.Spec.QueryDelay = &v1alpha.QueryDelay{
		Duration: v1alpha.Duration{
			Value: ptr(*defaultQueryDelay.Value + 1),
			Unit:  defaultQueryDelay.Unit,
		},
	}
	return agent
}

func (a agentExample) generateVariant(agent v1alphaAgent.Agent) v1alphaAgent.Agent {
	switch a.typ {
	case v1alpha.AmazonPrometheus:
		agent.Spec.AmazonPrometheus = &v1alphaAgent.AmazonPrometheusConfig{
			URL:    "https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-f49ecf99-6dfa-4b00-9f94-a50b10a3010b",
			Region: "us-east-1",
		}
	case v1alpha.AppDynamics:
		agent.Spec.AppDynamics = &v1alphaAgent.AppDynamicsConfig{
			URL: "https://my-org.saas.appdynamics.com",
		}
	case v1alpha.AzureMonitor:
		agent.Spec.AzureMonitor = &v1alphaAgent.AzureMonitorConfig{
			TenantID: "5cdecca3-c2c5-4072-89dd-5555faf05202",
		}
	case v1alpha.AzurePrometheus:
		agent.Spec.AzurePrometheus = &v1alphaAgent.AzurePrometheusConfig{
			URL:      "https://defaultazuremonitorworkspace-westus2-szxw.westus2.prometheus.monitor.azure.com",
			TenantID: "41372654-f4b6-4bd1-a3fe-75629c024df1",
		}
	case v1alpha.BigQuery:
		agent.Spec.BigQuery = &v1alphaAgent.BigQueryConfig{}
	case v1alpha.CloudWatch:
		agent.Spec.CloudWatch = &v1alphaAgent.CloudWatchConfig{}
	case v1alpha.Datadog:
		agent.Spec.Datadog = &v1alphaAgent.DatadogConfig{
			Site: "com",
		}
	case v1alpha.Dynatrace:
		agent.Spec.Dynatrace = &v1alphaAgent.DynatraceConfig{
			URL: "https://zvf10945.live.dynatrace.com/",
		}
	case v1alpha.Elasticsearch:
		agent.Spec.Elasticsearch = &v1alphaAgent.ElasticsearchConfig{
			URL: "http://elasticsearch-main.elasticsearch:9200",
		}
	case v1alpha.GCM:
		agent.Spec.GCM = &v1alphaAgent.GCMConfig{}
	case v1alpha.Generic:
		agent.Spec.Generic = &v1alphaAgent.GenericConfig{}
	case v1alpha.GrafanaLoki:
		agent.Spec.GrafanaLoki = &v1alphaAgent.GrafanaLokiConfig{
			URL: "http://grafana-loki.loki:3100",
		}
	case v1alpha.Graphite:
		agent.Spec.Graphite = &v1alphaAgent.GraphiteConfig{
			URL: "http://graphite.graphite:8080/render",
		}
	case v1alpha.Honeycomb:
		agent.Spec.Honeycomb = &v1alphaAgent.HoneycombConfig{}
	case v1alpha.InfluxDB:
		agent.Spec.InfluxDB = &v1alphaAgent.InfluxDBConfig{
			URL: "https://us-west-2-2.aws.cloud2.influxdata.com",
		}
	case v1alpha.Instana:
		agent.Spec.Instana = &v1alphaAgent.InstanaConfig{
			URL: "https://orange-my-org12.instana.io",
		}
	case v1alpha.Lightstep:
		agent.Spec.Lightstep = &v1alphaAgent.LightstepConfig{
			Organization: "MyOrg",
			Project:      "prod-app",
			URL:          "https://api.lightstep.com",
		}
	case v1alpha.LogicMonitor:
		agent.Spec.LogicMonitor = &v1alphaAgent.LogicMonitorConfig{
			Account: "myaccountname",
		}
	case v1alpha.NewRelic:
		agent.Spec.NewRelic = &v1alphaAgent.NewRelicConfig{
			AccountID: 1234567,
		}
	case v1alpha.OpenTSDB:
		agent.Spec.OpenTSDB = &v1alphaAgent.OpenTSDBConfig{
			URL: "http://opentsdb.opentsdb:4242",
		}
	case v1alpha.Pingdom:
		agent.Spec.Pingdom = &v1alphaAgent.PingdomConfig{}
	case v1alpha.Prometheus:
		agent.Spec.Prometheus = &v1alphaAgent.PrometheusConfig{
			URL: "http://prometheus.prometheus:9090",
		}
	case v1alpha.Redshift:
		agent.Spec.Redshift = &v1alphaAgent.RedshiftConfig{}
	case v1alpha.Splunk:
		agent.Spec.Splunk = &v1alphaAgent.SplunkConfig{
			URL: "https://splunk.my-org.com/services",
		}
	case v1alpha.SplunkObservability:
		agent.Spec.SplunkObservability = &v1alphaAgent.SplunkObservabilityConfig{
			Realm: "us1",
		}
	case v1alpha.SumoLogic:
		agent.Spec.SumoLogic = &v1alphaAgent.SumoLogicConfig{
			URL: "https://service.sumologic.com",
		}
	case v1alpha.ThousandEyes:
		agent.Spec.ThousandEyes = &v1alphaAgent.ThousandEyesConfig{}
	default:
		panic(fmt.Sprintf("unexpected v1alpha.DataSourceType: %#v", a.typ))
	}
	return agent
}
