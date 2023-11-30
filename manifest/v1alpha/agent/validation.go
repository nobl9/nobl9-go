package agent

import (
	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

var agentValidation = validation.New[Agent](
	v1alpha.FieldRuleMetadataName(func(a Agent) string { return a.Metadata.Name }),
	v1alpha.FieldRuleMetadataDisplayName(func(a Agent) string { return a.Metadata.DisplayName }),
	v1alpha.FieldRuleMetadataProject(func(a Agent) string { return a.Metadata.Project }),
	v1alpha.FieldRuleSpecDescription(func(a Agent) string { return a.Spec.Description }),
	validation.For(func(a Agent) Spec { return a.Spec }).
		WithName("spec").
		Include(specValidation),
)

var specValidation = validation.New[Spec](
	validation.For(validation.GetSelf[Spec]()).
		Rules(exactlyOneDataSourceTypeValidationRule).
		StopOnError().
		Rules(
			historicalDataRetrievalValidationRule,
			queryDelayGreaterThanOrEqualToDefaultValidationRule),
	validation.For(func(s Spec) v1alpha.ReleaseChannel { return s.ReleaseChannel }).
		WithName("releaseChannel").
		Omitempty().
		Rules(v1alpha.ReleaseChannelValidation()),
	validation.ForPointer(func(s Spec) *v1alpha.HistoricalDataRetrieval { return s.HistoricalDataRetrieval }).
		WithName("historicalDataRetrieval").
		Omitempty().
		Include(v1alpha.HistoricalDataRetrievalValidation()),
	validation.ForPointer(func(s Spec) *v1alpha.QueryDelay { return s.QueryDelay }).
		WithName("queryDelay").
		Omitempty().
		Include(v1alpha.QueryDelayValidation()),
	validation.ForPointer(func(s Spec) *PrometheusConfig { return s.Prometheus }).
		WithName("prometheus").
		Include(prometheusValidation),
	validation.ForPointer(func(s Spec) *DatadogConfig { return s.Datadog }).
		WithName("datadog").
		Include(datadogValidation),
	validation.ForPointer(func(s Spec) *NewRelicConfig { return s.NewRelic }).
		WithName("newRelic").
		Include(newRelicValidation),
	validation.ForPointer(func(s Spec) *AppDynamicsConfig { return s.AppDynamics }).
		WithName("appDynamics").
		Include(appDynamicsValidation),
	validation.ForPointer(func(s Spec) *SplunkConfig { return s.Splunk }).
		WithName("splunk").
		Include(splunkValidation),
	validation.ForPointer(func(s Spec) *LightstepConfig { return s.Lightstep }).
		WithName("lightstep").
		Include(lightstepValidation),
	validation.ForPointer(func(s Spec) *SplunkObservabilityConfig { return s.SplunkObservability }).
		WithName("splunkObservability").
		Include(splunkObservabilityValidation),
	validation.ForPointer(func(s Spec) *DynatraceConfig { return s.Dynatrace }).
		WithName("dynatrace").
		Include(dynatraceValidation),
	validation.ForPointer(func(s Spec) *ElasticsearchConfig { return s.Elasticsearch }).
		WithName("elasticsearch").
		Include(elasticsearchValidation),
	validation.ForPointer(func(s Spec) *ThousandEyesConfig { return s.ThousandEyes }).
		WithName("thousandEyes").
		Include(thousandEyesValidation),
	validation.ForPointer(func(s Spec) *GraphiteConfig { return s.Graphite }).
		WithName("graphite").
		Include(graphiteValidation),
	validation.ForPointer(func(s Spec) *BigQueryConfig { return s.BigQuery }).
		WithName("bigQuery").
		Include(bigQueryValidation),
	validation.ForPointer(func(s Spec) *OpenTSDBConfig { return s.OpenTSDB }).
		WithName("opentsdb").
		Include(openTSDBValidation),
	validation.ForPointer(func(s Spec) *GrafanaLokiConfig { return s.GrafanaLoki }).
		WithName("grafanaLoki").
		Include(grafanaLokiValidation),
	validation.ForPointer(func(s Spec) *CloudWatchConfig { return s.CloudWatch }).
		WithName("cloudWatch").
		Include(cloudWatchValidation),
	validation.ForPointer(func(s Spec) *PingdomConfig { return s.Pingdom }).
		WithName("pingdom").
		Include(pingdomValidation),
	validation.ForPointer(func(s Spec) *AmazonPrometheusConfig { return s.AmazonPrometheus }).
		WithName("amazonPrometheus").
		Include(amazonPrometheusValidation),
	validation.ForPointer(func(s Spec) *RedshiftConfig { return s.Redshift }).
		WithName("redshift").
		Include(redshiftValidation),
	validation.ForPointer(func(s Spec) *SumoLogicConfig { return s.SumoLogic }).
		WithName("sumoLogic").
		Include(sumoLogicValidation),
	validation.ForPointer(func(s Spec) *InstanaConfig { return s.Instana }).
		WithName("instana").
		Include(instanaValidation),
	validation.ForPointer(func(s Spec) *InfluxDBConfig { return s.InfluxDB }).
		WithName("influxdb").
		Include(influxDBValidation),
	validation.ForPointer(func(s Spec) *AzureMonitorConfig { return s.AzureMonitor }).
		WithName("azureMonitor").
		Include(azureMonitorValidation),
	validation.ForPointer(func(s Spec) *GCMConfig { return s.GCM }).
		WithName("gcm").
		Include(gcmValidation),
	validation.ForPointer(func(s Spec) *GenericConfig { return s.Generic }).
		WithName("generic").
		Include(genericValidation),
	validation.ForPointer(func(s Spec) *HoneycombConfig { return s.Honeycomb }).
		WithName("honeycomb").
		Include(honeycombValidation),
)

var prometheusValidation = validation.New[PrometheusConfig](
	validation.For(func(p PrometheusConfig) string { return p.URL }).
		WithName("url").
		Required().
		Rules(validation.StringURL()),
)

var datadogValidation = validation.New[DatadogConfig](
	validation.For(func(d DatadogConfig) string { return d.Site }).
		WithName("site").
		Required().
		Rules(v1alpha.DataDogSiteValidationRule()),
)

var newRelicValidation = validation.New[NewRelicConfig](
	validation.For(func(n NewRelicConfig) int { return n.AccountID }).
		WithName("url").
		Required().
		Rules(validation.GreaterThanOrEqualTo(1)),
)

var appDynamicsValidation = validation.New[AppDynamicsConfig]()

var splunkValidation = validation.New[SplunkConfig]()

var lightstepValidation = validation.New[LightstepConfig]()

var splunkObservabilityValidation = validation.New[SplunkObservabilityConfig]()

var dynatraceValidation = validation.New[DynatraceConfig]()

var elasticsearchValidation = validation.New[ElasticsearchConfig]()

var thousandEyesValidation = validation.New[ThousandEyesConfig]()

var graphiteValidation = validation.New[GraphiteConfig]()

var bigQueryValidation = validation.New[BigQueryConfig]()

var openTSDBValidation = validation.New[OpenTSDBConfig]()

var grafanaLokiValidation = validation.New[GrafanaLokiConfig]()

var cloudWatchValidation = validation.New[CloudWatchConfig]()

var pingdomValidation = validation.New[PingdomConfig]()

var amazonPrometheusValidation = validation.New[AmazonPrometheusConfig]()

var redshiftValidation = validation.New[RedshiftConfig]()

var sumoLogicValidation = validation.New[SumoLogicConfig]()

var instanaValidation = validation.New[InstanaConfig]()

var influxDBValidation = validation.New[InfluxDBConfig]()

var azureMonitorValidation = validation.New[AzureMonitorConfig]()

var gcmValidation = validation.New[GCMConfig]()

var genericValidation = validation.New[GenericConfig]()

var honeycombValidation = validation.New[HoneycombConfig]()

const errCodeExactlyOneDataSourceType = "exactly_one_data_source_type"

var exactlyOneDataSourceTypeValidationRule = validation.NewSingleRule(func(spec Spec) error {
	var onlyType v1alpha.DataSourceType
	typesMatch := func(typ v1alpha.DataSourceType) error {
		if onlyType == 0 {
			onlyType = typ
		}
		if onlyType != typ {
			return errors.Errorf(
				"must have exactly one datas source type, detected both %s and %s",
				onlyType, typ)
		}
		return nil
	}
	if spec.Prometheus != nil {
		if err := typesMatch(v1alpha.Prometheus); err != nil {
			return err
		}
	}
	if spec.Datadog != nil {
		if err := typesMatch(v1alpha.Datadog); err != nil {
			return err
		}
	}
	if spec.NewRelic != nil {
		if err := typesMatch(v1alpha.NewRelic); err != nil {
			return err
		}
	}
	if spec.AppDynamics != nil {
		if err := typesMatch(v1alpha.AppDynamics); err != nil {
			return err
		}
	}
	if spec.Splunk != nil {
		if err := typesMatch(v1alpha.Splunk); err != nil {
			return err
		}
	}
	if spec.Lightstep != nil {
		if err := typesMatch(v1alpha.Lightstep); err != nil {
			return err
		}
	}
	if spec.SplunkObservability != nil {
		if err := typesMatch(v1alpha.SplunkObservability); err != nil {
			return err
		}
	}
	if spec.ThousandEyes != nil {
		if err := typesMatch(v1alpha.ThousandEyes); err != nil {
			return err
		}
	}
	if spec.Dynatrace != nil {
		if err := typesMatch(v1alpha.Dynatrace); err != nil {
			return err
		}
	}
	if spec.Elasticsearch != nil {
		if err := typesMatch(v1alpha.Elasticsearch); err != nil {
			return err
		}
	}
	if spec.Graphite != nil {
		if err := typesMatch(v1alpha.Graphite); err != nil {
			return err
		}
	}
	if spec.BigQuery != nil {
		if err := typesMatch(v1alpha.BigQuery); err != nil {
			return err
		}
	}
	if spec.OpenTSDB != nil {
		if err := typesMatch(v1alpha.OpenTSDB); err != nil {
			return err
		}
	}
	if spec.GrafanaLoki != nil {
		if err := typesMatch(v1alpha.GrafanaLoki); err != nil {
			return err
		}
	}
	if spec.CloudWatch != nil {
		if err := typesMatch(v1alpha.CloudWatch); err != nil {
			return err
		}
	}
	if spec.Pingdom != nil {
		if err := typesMatch(v1alpha.Pingdom); err != nil {
			return err
		}
	}
	if spec.AmazonPrometheus != nil {
		if err := typesMatch(v1alpha.AmazonPrometheus); err != nil {
			return err
		}
	}
	if spec.Redshift != nil {
		if err := typesMatch(v1alpha.Redshift); err != nil {
			return err
		}
	}
	if spec.SumoLogic != nil {
		if err := typesMatch(v1alpha.SumoLogic); err != nil {
			return err
		}
	}
	if spec.Instana != nil {
		if err := typesMatch(v1alpha.Instana); err != nil {
			return err
		}
	}
	if spec.InfluxDB != nil {
		if err := typesMatch(v1alpha.InfluxDB); err != nil {
			return err
		}
	}
	if spec.GCM != nil {
		if err := typesMatch(v1alpha.GCM); err != nil {
			return err
		}
	}
	if spec.AzureMonitor != nil {
		if err := typesMatch(v1alpha.AzureMonitor); err != nil {
			return err
		}
	}
	if spec.Generic != nil {
		if err := typesMatch(v1alpha.Generic); err != nil {
			return err
		}
	}
	if spec.Honeycomb != nil {
		if err := typesMatch(v1alpha.Honeycomb); err != nil {
			return err
		}
	}
	if onlyType == 0 {
		return errors.New("must have exactly one data source type, none were provided")
	}
	return nil
}).WithErrorCode(errCodeExactlyOneDataSourceType)

var historicalDataRetrievalValidationRule = validation.NewSingleRule(func(spec Spec) error {
	typ, _ := spec.GetType()
	if _, err := v1alpha.GetDataRetrievalMaxDuration(manifest.KindAgent, typ); err != nil {
		return validation.NewPropertyError("historicalDataRetrieval", nil, err)
	}
	return nil
})

var queryDelayGreaterThanOrEqualToDefaultValidationRule = validation.NewSingleRule(func(spec Spec) error {
	if spec.QueryDelay == nil {
		return nil
	}
	typ, _ := spec.GetType()
	agentDefault := v1alpha.GetQueryDelayDefaults()[typ.String()]
	if spec.QueryDelay.LessThan(agentDefault) {
		return validation.NewPropertyError(
			"queryDelay",
			spec.QueryDelay,
			errors.Errorf("should be greater than or equal to %s", agentDefault),
		)
	}
	return nil
})

func validate(a Agent) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(agentValidation, a)
}
