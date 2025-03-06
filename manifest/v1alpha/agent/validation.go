package agent

import (
	"net/url"
	"path"
	"strings"

	"github.com/pkg/errors"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func validate(a Agent) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(validator, a, manifest.KindAgent)
}

var validator = govy.New[Agent](
	validationV1Alpha.FieldRuleAPIVersion(func(a Agent) manifest.Version { return a.APIVersion }),
	validationV1Alpha.FieldRuleKind(func(a Agent) manifest.Kind { return a.Kind }, manifest.KindAgent),
	validationV1Alpha.FieldRuleMetadataName(func(a Agent) string { return a.Metadata.Name }),
	validationV1Alpha.FieldRuleMetadataDisplayName(func(a Agent) string { return a.Metadata.DisplayName }),
	validationV1Alpha.FieldRuleMetadataProject(func(a Agent) string { return a.Metadata.Project }),
	validationV1Alpha.FieldRuleSpecDescription(func(a Agent) string { return a.Spec.Description }),
	govy.For(func(a Agent) Spec { return a.Spec }).
		WithName("spec").
		Include(specValidation),
)

var specValidation = govy.New[Spec](
	govy.For(govy.GetSelf[Spec]()).
		Cascade(govy.CascadeModeStop).
		Rules(exactlyOneDataSourceTypeValidationRule).
		Rules(
			historicalDataRetrievalValidationRule,
			queryDelayValidationRule),
	govy.For(func(s Spec) v1alpha.ReleaseChannel { return s.ReleaseChannel }).
		WithName("releaseChannel").
		OmitEmpty().
		Rules(v1alpha.ReleaseChannelValidation()),
	govy.ForPointer(func(s Spec) *v1alpha.HistoricalDataRetrieval { return s.HistoricalDataRetrieval }).
		WithName("historicalDataRetrieval").
		Include(v1alpha.HistoricalDataRetrievalValidation()),
	govy.ForPointer(func(s Spec) *v1alpha.QueryDelay { return s.QueryDelay }).
		WithName("queryDelay").
		Include(v1alpha.QueryDelayValidation()),
	govy.ForPointer(func(s Spec) *PrometheusConfig { return s.Prometheus }).
		WithName("prometheus").
		Include(prometheusValidation),
	govy.ForPointer(func(s Spec) *DatadogConfig { return s.Datadog }).
		WithName("datadog").
		Include(datadogValidation),
	govy.ForPointer(func(s Spec) *NewRelicConfig { return s.NewRelic }).
		WithName("newRelic").
		Include(newRelicValidation),
	govy.ForPointer(func(s Spec) *AppDynamicsConfig { return s.AppDynamics }).
		WithName("appDynamics").
		Include(appDynamicsValidation),
	govy.ForPointer(func(s Spec) *SplunkConfig { return s.Splunk }).
		WithName("splunk").
		Include(splunkValidation),
	govy.ForPointer(func(s Spec) *LightstepConfig { return s.Lightstep }).
		WithName("lightstep").
		Include(lightstepValidation),
	govy.ForPointer(func(s Spec) *SplunkObservabilityConfig { return s.SplunkObservability }).
		WithName("splunkObservability").
		Include(splunkObservabilityValidation),
	govy.ForPointer(func(s Spec) *DynatraceConfig { return s.Dynatrace }).
		WithName("dynatrace").
		Include(dynatraceValidation),
	govy.ForPointer(func(s Spec) *ElasticsearchConfig { return s.Elasticsearch }).
		WithName("elasticsearch").
		Include(elasticsearchValidation),
	govy.ForPointer(func(s Spec) *ThousandEyesConfig { return s.ThousandEyes }).
		WithName("thousandEyes").
		Include(thousandEyesValidation),
	govy.ForPointer(func(s Spec) *GraphiteConfig { return s.Graphite }).
		WithName("graphite").
		Include(graphiteValidation),
	govy.ForPointer(func(s Spec) *BigQueryConfig { return s.BigQuery }).
		WithName("bigQuery").
		Include(bigQueryValidation),
	govy.ForPointer(func(s Spec) *OpenTSDBConfig { return s.OpenTSDB }).
		WithName("opentsdb").
		Include(openTSDBValidation),
	govy.ForPointer(func(s Spec) *GrafanaLokiConfig { return s.GrafanaLoki }).
		WithName("grafanaLoki").
		Include(grafanaLokiValidation),
	govy.ForPointer(func(s Spec) *CloudWatchConfig { return s.CloudWatch }).
		WithName("cloudWatch").
		Include(cloudWatchValidation),
	govy.ForPointer(func(s Spec) *PingdomConfig { return s.Pingdom }).
		WithName("pingdom").
		Include(pingdomValidation),
	govy.ForPointer(func(s Spec) *AmazonPrometheusConfig { return s.AmazonPrometheus }).
		WithName("amazonPrometheus").
		Include(amazonPrometheusValidation),
	govy.ForPointer(func(s Spec) *RedshiftConfig { return s.Redshift }).
		WithName("redshift").
		Include(redshiftValidation),
	govy.ForPointer(func(s Spec) *SumoLogicConfig { return s.SumoLogic }).
		WithName("sumoLogic").
		Include(sumoLogicValidation),
	govy.ForPointer(func(s Spec) *InstanaConfig { return s.Instana }).
		WithName("instana").
		Include(instanaValidation),
	govy.ForPointer(func(s Spec) *InfluxDBConfig { return s.InfluxDB }).
		WithName("influxdb").
		Include(influxDBValidation),
	govy.ForPointer(func(s Spec) *AzureMonitorConfig { return s.AzureMonitor }).
		WithName("azureMonitor").
		Include(azureMonitorValidation),
	govy.ForPointer(func(s Spec) *GCMConfig { return s.GCM }).
		WithName("gcm").
		Include(gcmValidation),
	govy.ForPointer(func(s Spec) *GenericConfig { return s.Generic }).
		WithName("generic").
		Include(genericValidation),
	govy.ForPointer(func(s Spec) *HoneycombConfig { return s.Honeycomb }).
		WithName("honeycomb").
		Include(honeycombValidation),
	govy.ForPointer(func(s Spec) *LogicMonitorConfig { return s.LogicMonitor }).
		WithName("logicMonitor").
		Include(logicMonitorValidation),
	govy.ForPointer(func(s Spec) *AzurePrometheusConfig { return s.AzurePrometheus }).
		WithName("azurePrometheus").
		Include(azurePrometheusValidation),
	govy.ForPointer(func(s Spec) *CoralogixConfig { return s.Coralogix }).
		WithName("coralogix").
		Include(coralogixValidation),
)

var (
	datadogValidation = govy.New[DatadogConfig](
		govy.For(func(d DatadogConfig) string { return d.Site }).
			WithName("site").
			Required().
			Rules(v1alpha.DataDogSiteValidationRule()),
	)
	newRelicValidation = govy.New[NewRelicConfig](
		govy.For(func(n NewRelicConfig) int { return n.AccountID }).
			WithName("accountId").
			Required().
			Rules(rules.GTE(1)),
	)
	lightstepValidation = govy.New[LightstepConfig](
		govy.For(func(l LightstepConfig) string { return l.Organization }).
			WithName("organization").
			Required(),
		govy.For(func(l LightstepConfig) string { return l.Project }).
			WithName("project").
			Required(),
		govy.Transform(func(l LightstepConfig) string { return l.URL }, url.Parse).
			WithName("url").
			OmitEmpty().
			Rules(
				rules.URL(),
			),
	)
	splunkObservabilityValidation = govy.New[SplunkObservabilityConfig](
		govy.For(func(s SplunkObservabilityConfig) string { return s.Realm }).
			WithName("realm").
			Required(),
	)
	dynatraceValidation = govy.New[DynatraceConfig](
		govy.Transform(func(d DynatraceConfig) string { return d.URL }, url.Parse).
			WithName("url").
			Required().
			Rules(
				rules.URL(),
				govy.NewRule(func(u *url.URL) error {
					// For SaaS type enforce https and land lack of path.
					// - Join instead of Clean (to avoid getting . for empty path),
					// - Trim to get rid of root.
					pathURL := strings.Trim(path.Join(u.Path), "/")
					if strings.HasSuffix(u.Host, "live.dynatrace.com") &&
						(u.Scheme != "https" || pathURL != "") {
						return errors.New(
							"Dynatrace SaaS URL (live.dynatrace.com) requires https scheme and empty URL path" +
								"; example: https://rxh50243.live.dynatrace.com/")
					}
					return nil
				}),
			),
	)
	amazonPrometheusValidation = govy.New[AmazonPrometheusConfig](
		govy.For(func(a AmazonPrometheusConfig) string { return a.URL }).
			WithName("url").
			Required().
			Rules(rules.StringURL()),
		govy.For(func(a AmazonPrometheusConfig) string { return a.Region }).
			WithName("region").
			Required().
			Rules(rules.StringMaxLength(255)),
	)
	azureMonitorValidation = govy.New[AzureMonitorConfig](
		govy.For(func(a AzureMonitorConfig) string { return a.TenantID }).
			WithName("tenantId").
			Required().
			Rules(rules.StringUUID()),
	)
	logicMonitorValidation = govy.New[LogicMonitorConfig](
		govy.For(func(l LogicMonitorConfig) string { return l.Account }).
			WithName("account").
			Required().
			Rules(rules.StringNotEmpty()),
	)
	azurePrometheusValidation = govy.New[AzurePrometheusConfig](
		govy.For(func(a AzurePrometheusConfig) string { return a.URL }).
			WithName("url").
			Required().
			Rules(rules.StringURL()),
		govy.For(func(a AzurePrometheusConfig) string { return a.TenantID }).
			WithName("tenantId").
			Required().
			Rules(rules.StringUUID()),
	)
	// URL only.
	prometheusValidation    = newURLValidator(func(p PrometheusConfig) string { return p.URL })
	appDynamicsValidation   = newURLValidator(func(a AppDynamicsConfig) string { return a.URL })
	splunkValidation        = newURLValidator(func(s SplunkConfig) string { return s.URL })
	elasticsearchValidation = newURLValidator(func(e ElasticsearchConfig) string { return e.URL })
	graphiteValidation      = newURLValidator(func(g GraphiteConfig) string { return g.URL })
	openTSDBValidation      = newURLValidator(func(o OpenTSDBConfig) string { return o.URL })
	grafanaLokiValidation   = newURLValidator(func(g GrafanaLokiConfig) string { return g.URL })
	sumoLogicValidation     = newURLValidator(func(s SumoLogicConfig) string { return s.URL })
	instanaValidation       = newURLValidator(func(i InstanaConfig) string { return i.URL })
	influxDBValidation      = newURLValidator(func(i InfluxDBConfig) string { return i.URL })
	coralogixValidation     = newURLValidator(func(c CoralogixConfig) string { return c.URL })
	// Empty configs.
	thousandEyesValidation = govy.New[ThousandEyesConfig]()
	bigQueryValidation     = govy.New[BigQueryConfig]()
	cloudWatchValidation   = govy.New[CloudWatchConfig]()
	pingdomValidation      = govy.New[PingdomConfig]()
	redshiftValidation     = govy.New[RedshiftConfig]()
	gcmValidation          = govy.New[GCMConfig]()
	genericValidation      = govy.New[GenericConfig]()
	honeycombValidation    = govy.New[HoneycombConfig]()
)

const (
	errCodeExactlyOneDataSourceType = "exactly_one_data_source_type"
	errCodeQueryDelayOutOfBounds    = "query_delay_out_of_bounds"
)

var exactlyOneDataSourceTypeValidationRule = govy.NewRule(func(spec Spec) error {
	var onlyType v1alpha.DataSourceType
	typesMatch := func(typ v1alpha.DataSourceType) error {
		if onlyType == 0 {
			onlyType = typ
		}
		if onlyType != typ {
			return errors.Errorf(
				"must have exactly one data source type, detected both %s and %s",
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
	if spec.LogicMonitor != nil {
		if err := typesMatch(v1alpha.LogicMonitor); err != nil {
			return err
		}
	}
	if spec.AzurePrometheus != nil {
		if err := typesMatch(v1alpha.AzurePrometheus); err != nil {
			return err
		}
	}
	if spec.Coralogix != nil {
		if err := typesMatch(v1alpha.Coralogix); err != nil {
			return err
		}
	}
	if onlyType == 0 {
		return errors.New("must have exactly one data source type, none were provided")
	}
	return nil
}).WithErrorCode(errCodeExactlyOneDataSourceType)

var historicalDataRetrievalValidationRule = govy.NewRule(func(spec Spec) error {
	if spec.HistoricalDataRetrieval == nil {
		return nil
	}
	typ, _ := spec.GetType()
	maxDuration, err := v1alpha.GetDataRetrievalMaxDuration(manifest.KindAgent, typ)
	if err != nil {
		return govy.NewPropertyError("historicalDataRetrieval", nil, err)
	}
	maxDurationAllowed := v1alpha.HistoricalRetrievalDuration{
		Value: maxDuration.Value,
		Unit:  maxDuration.Unit,
	}
	if spec.HistoricalDataRetrieval.MaxDuration.BiggerThan(maxDurationAllowed) {
		return govy.NewPropertyError(
			"historicalDataRetrieval.maxDuration",
			spec.HistoricalDataRetrieval.MaxDuration,
			errors.Errorf("must be less than or equal to %d %s",
				*maxDurationAllowed.Value, maxDurationAllowed.Unit))
	}

	return nil
})

var queryDelayValidationRule = govy.NewRule(func(spec Spec) error {
	if spec.QueryDelay == nil {
		return nil
	}
	typ, _ := spec.GetType()
	maxQueryDelay := v1alpha.GetQueryDelayMax(typ)
	maxQueryDelayAllowed := v1alpha.Duration{
		Value: maxQueryDelay.Value,
		Unit:  maxQueryDelay.Unit,
	}
	if spec.QueryDelay.Duration.GreaterThan(maxQueryDelayAllowed) {
		return govy.NewPropertyError(
			"queryDelay",
			spec.QueryDelay,
			errors.Errorf("must be less than or equal to %d %s",
				*maxQueryDelayAllowed.Value, maxQueryDelayAllowed.Unit))
	}
	agentDefault := v1alpha.GetQueryDelayDefaults()[typ]
	if spec.QueryDelay.LessThan(agentDefault) {
		return govy.NewPropertyError(
			"queryDelay",
			spec.QueryDelay,
			errors.Errorf("should be greater than or equal to %s", agentDefault),
		)
	}
	return nil
}).WithErrorCode(errCodeQueryDelayOutOfBounds)

// newURLValidator is a helper construct for Agent which only have a simple 'url' field govy.
func newURLValidator[S any](getter govy.PropertyGetter[string, S]) govy.Validator[S] {
	return govy.New[S](
		govy.For(getter).
			WithName("url").
			Required().
			Rules(rules.StringURL()),
	)
}
