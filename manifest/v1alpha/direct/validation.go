package direct

import (
	"net/url"
	"slices"

	"github.com/pkg/errors"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func validate(d Direct) *v1alpha.ObjectError {
	return v1alpha.ValidateObject[Direct](validator, d, manifest.KindDirect)
}

var validator = govy.New[Direct](
	validationV1Alpha.FieldRuleAPIVersion(func(d Direct) manifest.Version { return d.APIVersion }),
	validationV1Alpha.FieldRuleKind(func(d Direct) manifest.Kind { return d.Kind }, manifest.KindDirect),
	validationV1Alpha.FieldRuleMetadataName(func(d Direct) string { return d.Metadata.Name }),
	validationV1Alpha.FieldRuleMetadataDisplayName(func(d Direct) string { return d.Metadata.DisplayName }),
	validationV1Alpha.FieldRuleMetadataProject(func(d Direct) string { return d.Metadata.Project }),
	validationV1Alpha.FieldRuleSpecDescription(func(d Direct) string { return d.Spec.Description }),
	govy.For(func(d Direct) Spec { return d.Spec }).
		WithName("spec").
		Include(specValidation),
)

var specValidation = govy.New[Spec](
	govy.For(govy.GetSelf[Spec]()).
		Cascade(govy.CascadeModeStop).
		Rules(exactlyOneDataSourceTypeValidationRule).
		Rules(
			historicalDataRetrievalValidationRule,
			queryDelayValidationRule,
			releaseChannelValidationRule),
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
	govy.ForPointer(func(s Spec) *DatadogConfig { return s.Datadog }).
		WithName("datadog").
		Include(datadogValidation),
	govy.ForPointer(func(s Spec) *NewRelicConfig { return s.NewRelic }).
		WithName("newRelic").
		Include(newRelicValidation),
	govy.ForPointer(func(s Spec) *AppDynamicsConfig { return s.AppDynamics }).
		WithName("appDynamics").
		Include(appDynamicsValidation),
	govy.ForPointer(func(s Spec) *SplunkObservabilityConfig { return s.SplunkObservability }).
		WithName("splunkObservability").
		Include(splunkObservabilityValidation),
	govy.ForPointer(func(s Spec) *ThousandEyesConfig { return s.ThousandEyes }).
		WithName("thousandEyes").
		Include(thousandEyesValidation),
	govy.ForPointer(func(s Spec) *BigQueryConfig { return s.BigQuery }).
		WithName("bigQuery").
		Include(bigQueryValidation),
	govy.ForPointer(func(s Spec) *SplunkConfig { return s.Splunk }).
		WithName("splunk").
		Include(splunkValidation),
	govy.ForPointer(func(s Spec) *CloudWatchConfig { return s.CloudWatch }).
		WithName("cloudWatch").
		Include(cloudWatchValidation),
	govy.ForPointer(func(s Spec) *PingdomConfig { return s.Pingdom }).
		WithName("pingdom").
		Include(pingdomValidation),
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
	govy.ForPointer(func(s Spec) *GCMConfig { return s.GCM }).
		WithName("gcm").
		Include(gcmValidation),
	govy.ForPointer(func(s Spec) *LightstepConfig { return s.Lightstep }).
		WithName("lightstep").
		Include(lightstepValidation),
	govy.ForPointer(func(s Spec) *DynatraceConfig { return s.Dynatrace }).
		WithName("dynatrace").
		Include(dynatraceValidation),
	govy.ForPointer(func(s Spec) *AzureMonitorConfig { return s.AzureMonitor }).
		WithName("azureMonitor").
		Include(azureMonitorValidation),
	govy.ForPointer(func(s Spec) *HoneycombConfig { return s.Honeycomb }).
		WithName("honeycomb").
		Include(honeycombValidation),
	govy.ForPointer(func(s Spec) *LogicMonitorConfig { return s.LogicMonitor }).
		WithName("logicMonitor").
		Include(logicMonitorValidation),
	govy.ForPointer(func(s Spec) *AzurePrometheusConfig { return s.AzurePrometheus }).
		WithName("azurePrometheus").
		Include(azurePrometheusValidation),
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
		govy.For(func(n NewRelicConfig) string { return n.InsightsQueryKey }).
			WithName("insightsQueryKey").
			HideValue().
			When(
				func(c NewRelicConfig) bool { return !isHiddenValue(c.InsightsQueryKey) },
				govy.WhenDescription("is empty or equal to '%s'", v1alpha.HiddenValue),
			).
			Rules(rules.StringStartsWith("NRIQ-")),
	)
	appDynamicsValidation = govy.New[AppDynamicsConfig](
		urlPropertyRules(func(a AppDynamicsConfig) string { return a.URL }),
		govy.For(func(a AppDynamicsConfig) string { return a.ClientName }).
			WithName("clientName").
			Required(),
		govy.For(func(a AppDynamicsConfig) string { return a.AccountName }).
			WithName("accountName").
			Required(),
	)
	splunkObservabilityValidation = govy.New[SplunkObservabilityConfig](
		govy.For(func(s SplunkObservabilityConfig) string { return s.Realm }).
			WithName("realm").
			Required(),
	)
	thousandEyesValidation = govy.New[ThousandEyesConfig]()
	bigQueryValidation     = govy.New[BigQueryConfig](
		govy.For(func(b BigQueryConfig) string { return b.ServiceAccountKey }).
			WithName("serviceAccountKey").
			HideValue().
			When(
				func(b BigQueryConfig) bool { return !isHiddenValue(b.ServiceAccountKey) },
				govy.WhenDescription("is empty or equal to '%s'", v1alpha.HiddenValue),
			).
			Rules(rules.StringJSON()),
	)
	splunkValidation = govy.New[SplunkConfig](
		urlPropertyRules(func(s SplunkConfig) string { return s.URL }),
	)
	cloudWatchValidation = govy.New[CloudWatchConfig]()
	pingdomValidation    = govy.New[PingdomConfig]()
	redshiftValidation   = govy.New[RedshiftConfig](
		govy.For(func(r RedshiftConfig) string { return r.SecretARN }).
			WithName("secretARN").
			Required(),
	)
	sumoLogicValidation = govy.New[SumoLogicConfig](
		urlPropertyRules(func(s SumoLogicConfig) string { return s.URL }),
	)
	instanaValidation = govy.New[InstanaConfig](
		urlPropertyRules(func(i InstanaConfig) string { return i.URL }),
	)
	influxDBValidation = govy.New[InfluxDBConfig](
		urlPropertyRules(func(i InfluxDBConfig) string { return i.URL }),
	)
	gcmValidation = govy.New[GCMConfig](
		govy.For(func(g GCMConfig) string { return g.ServiceAccountKey }).
			WithName("serviceAccountKey").
			HideValue().
			When(
				func(g GCMConfig) bool { return !isHiddenValue(g.ServiceAccountKey) },
				govy.WhenDescription("is empty or equal to '%s'", v1alpha.HiddenValue),
			).
			Rules(rules.StringJSON()),
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
			Rules(rules.URL()),
	)
	dynatraceValidation = govy.New[DynatraceConfig](
		urlPropertyRules(func(d DynatraceConfig) string { return d.URL }),
	)
	azureMonitorValidation = govy.New[AzureMonitorConfig](
		govy.For(func(a AzureMonitorConfig) string { return a.TenantID }).
			WithName("tenantId").
			Required().
			Rules(rules.StringUUID()),
	)
	honeycombValidation    = govy.New[HoneycombConfig]()
	logicMonitorValidation = govy.New[LogicMonitorConfig](
		govy.For(func(l LogicMonitorConfig) string { return l.Account }).
			WithName("account").
			Required().
			Rules(rules.StringNotEmpty()),
	)
	azurePrometheusValidation = govy.New[AzurePrometheusConfig](
		urlPropertyRules(func(s AzurePrometheusConfig) string { return s.URL }),
		govy.For(func(a AzurePrometheusConfig) string { return a.TenantID }).
			WithName("tenantId").
			Required().
			Rules(rules.StringUUID()),
	)
)

const (
	errCodeExactlyOneDataSourceType  = "exactly_one_data_source_type"
	errCodeQueryDelayOutOfBounds     = "query_delay_out_of_bounds"
	errCodeUnsupportedReleaseChannel = "unsupported_release_channel"
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
	if spec.BigQuery != nil {
		if err := typesMatch(v1alpha.BigQuery); err != nil {
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
	maxDuration, err := v1alpha.GetDataRetrievalMaxDuration(manifest.KindDirect, typ)
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

var releaseChannelValidationRule = govy.NewRule(func(spec Spec) error {
	typ, _ := spec.GetType()
	if spec.ReleaseChannel == v1alpha.ReleaseChannelAlpha &&
		!slices.Contains(v1alpha.GetReleaseChannelAlphaEnabledDataSources(), typ) {
		return govy.NewPropertyError(
			"releaseChannel",
			spec.ReleaseChannel,
			errors.New("must be one of [stable, beta]"),
		)
	}

	if typ == v1alpha.SplunkObservability && spec.ReleaseChannel != v1alpha.ReleaseChannelAlpha {
		return govy.NewPropertyError(
			"releaseChannel",
			spec.ReleaseChannel,
			errors.New("must be 'alpha' for Splunk Observability"),
		)
	}

	return nil
}).WithErrorCode(errCodeUnsupportedReleaseChannel)

const errorCodeHTTPSSchemeRequired = "https_scheme_required"

func urlPropertyRules[S any](getter govy.PropertyGetter[string, S]) govy.PropertyRules[*url.URL, S] {
	return govy.Transform(getter, url.Parse).
		WithName("url").
		Cascade(govy.CascadeModeStop).
		Required().
		Rules(rules.URL()).
		Rules(govy.NewRule(func(u *url.URL) error {
			if u.Scheme != "https" {
				return errors.New("requires https scheme")
			}
			return nil
		}).WithErrorCode(errorCodeHTTPSSchemeRequired))
}

func isHiddenValue(s string) bool { return s == "" || s == v1alpha.HiddenValue }
