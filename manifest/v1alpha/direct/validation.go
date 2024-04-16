package direct

import (
	"net/url"

	"github.com/pkg/errors"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

var directValidation = validation.New[Direct](
	validationV1Alpha.FieldRuleMetadataName(func(d Direct) string { return d.Metadata.Name }),
	validationV1Alpha.FieldRuleMetadataDisplayName(func(d Direct) string { return d.Metadata.DisplayName }),
	validationV1Alpha.FieldRuleMetadataProject(func(d Direct) string { return d.Metadata.Project }),
	validationV1Alpha.FieldRuleSpecDescription(func(d Direct) string { return d.Spec.Description }),
	validation.For(func(d Direct) Spec { return d.Spec }).
		WithName("spec").
		Include(specValidation),
)

var specValidation = validation.New[Spec](
	validation.For(validation.GetSelf[Spec]()).
		Rules(exactlyOneDataSourceTypeValidationRule).
		CascadeMode(validation.CascadeModeStop).
		Rules(
			historicalDataRetrievalValidationRule,
			queryDelayGreaterThanOrEqualToDefaultValidationRule),
	validation.For(func(s Spec) v1alpha.ReleaseChannel { return s.ReleaseChannel }).
		WithName("releaseChannel").
		OmitEmpty().
		Rules(v1alpha.ReleaseChannelValidation()),
	validation.ForPointer(func(s Spec) *v1alpha.HistoricalDataRetrieval { return s.HistoricalDataRetrieval }).
		WithName("historicalDataRetrieval").
		Include(v1alpha.HistoricalDataRetrievalValidation()),
	validation.ForPointer(func(s Spec) *v1alpha.QueryDelay { return s.QueryDelay }).
		WithName("queryDelay").
		Include(v1alpha.QueryDelayValidation()),
	validation.ForPointer(func(s Spec) *DatadogConfig { return s.Datadog }).
		WithName("datadog").
		Include(datadogValidation),
	validation.ForPointer(func(s Spec) *NewRelicConfig { return s.NewRelic }).
		WithName("newRelic").
		Include(newRelicValidation),
	validation.ForPointer(func(s Spec) *AppDynamicsConfig { return s.AppDynamics }).
		WithName("appDynamics").
		Include(appDynamicsValidation),
	validation.ForPointer(func(s Spec) *SplunkObservabilityConfig { return s.SplunkObservability }).
		WithName("splunkObservability").
		Include(splunkObservabilityValidation),
	validation.ForPointer(func(s Spec) *ThousandEyesConfig { return s.ThousandEyes }).
		WithName("thousandEyes").
		Include(thousandEyesValidation),
	validation.ForPointer(func(s Spec) *BigQueryConfig { return s.BigQuery }).
		WithName("bigQuery").
		Include(bigQueryValidation),
	validation.ForPointer(func(s Spec) *SplunkConfig { return s.Splunk }).
		WithName("splunk").
		Include(splunkValidation),
	validation.ForPointer(func(s Spec) *CloudWatchConfig { return s.CloudWatch }).
		WithName("cloudWatch").
		Include(cloudWatchValidation),
	validation.ForPointer(func(s Spec) *PingdomConfig { return s.Pingdom }).
		WithName("pingdom").
		Include(pingdomValidation),
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
	validation.ForPointer(func(s Spec) *GCMConfig { return s.GCM }).
		WithName("gcm").
		Include(gcmValidation),
	validation.ForPointer(func(s Spec) *LightstepConfig { return s.Lightstep }).
		WithName("lightstep").
		Include(lightstepValidation),
	validation.ForPointer(func(s Spec) *DynatraceConfig { return s.Dynatrace }).
		WithName("dynatrace").
		Include(dynatraceValidation),
	validation.ForPointer(func(s Spec) *AzureMonitorConfig { return s.AzureMonitor }).
		WithName("azureMonitor").
		Include(azureMonitorValidation),
	validation.ForPointer(func(s Spec) *HoneycombConfig { return s.Honeycomb }).
		WithName("honeycomb").
		Include(honeycombValidation),
)

var (
	datadogValidation = validation.New[DatadogConfig](
		validation.For(func(d DatadogConfig) string { return d.Site }).
			WithName("site").
			Required().
			Rules(v1alpha.DataDogSiteValidationRule()),
	)
	newRelicValidation = validation.New[NewRelicConfig](
		validation.For(func(n NewRelicConfig) int { return n.AccountID }).
			WithName("accountId").
			Required().
			Rules(validation.GreaterThanOrEqualTo(1)),
		validation.For(func(n NewRelicConfig) string { return n.InsightsQueryKey }).
			WithName("insightsQueryKey").
			HideValue().
			When(func(c NewRelicConfig) bool { return !isHiddenValue(c.InsightsQueryKey) }).
			Rules(validation.StringStartsWith("NRIQ-")),
	)
	appDynamicsValidation = validation.New[AppDynamicsConfig](
		urlPropertyRules(func(a AppDynamicsConfig) string { return a.URL }),
		validation.For(func(a AppDynamicsConfig) string { return a.ClientName }).
			WithName("clientName").
			Required(),
		validation.For(func(a AppDynamicsConfig) string { return a.AccountName }).
			WithName("accountName").
			Required(),
	)
	splunkObservabilityValidation = validation.New[SplunkObservabilityConfig](
		validation.For(func(s SplunkObservabilityConfig) string { return s.Realm }).
			WithName("realm").
			Required(),
	)
	thousandEyesValidation = validation.New[ThousandEyesConfig]()
	bigQueryValidation     = validation.New[BigQueryConfig](
		validation.For(func(b BigQueryConfig) string { return b.ServiceAccountKey }).
			WithName("serviceAccountKey").
			HideValue().
			When(func(b BigQueryConfig) bool { return !isHiddenValue(b.ServiceAccountKey) }).
			Rules(validation.StringJSON()),
	)
	splunkValidation = validation.New[SplunkConfig](
		urlPropertyRules(func(s SplunkConfig) string { return s.URL }),
	)
	cloudWatchValidation = validation.New[CloudWatchConfig]()
	pingdomValidation    = validation.New[PingdomConfig]()
	redshiftValidation   = validation.New[RedshiftConfig](
		validation.For(func(r RedshiftConfig) string { return r.SecretARN }).
			WithName("secretARN").
			Required(),
	)
	sumoLogicValidation = validation.New[SumoLogicConfig](
		urlPropertyRules(func(s SumoLogicConfig) string { return s.URL }),
	)
	instanaValidation = validation.New[InstanaConfig](
		urlPropertyRules(func(i InstanaConfig) string { return i.URL }),
	)
	influxDBValidation = validation.New[InfluxDBConfig](
		urlPropertyRules(func(i InfluxDBConfig) string { return i.URL }),
	)
	gcmValidation = validation.New[GCMConfig](
		validation.For(func(g GCMConfig) string { return g.ServiceAccountKey }).
			WithName("serviceAccountKey").
			HideValue().
			When(func(g GCMConfig) bool { return !isHiddenValue(g.ServiceAccountKey) }).
			Rules(validation.StringJSON()),
	)
	lightstepValidation = validation.New[LightstepConfig](
		validation.For(func(l LightstepConfig) string { return l.Organization }).
			WithName("organization").
			Required(),
		validation.For(func(l LightstepConfig) string { return l.Project }).
			WithName("project").
			Required(),
		validation.Transform(func(l LightstepConfig) string { return l.URL }, url.Parse).
			WithName("url").
			OmitEmpty().
			Rules(
				validation.URL(),
			),
	)
	dynatraceValidation = validation.New[DynatraceConfig](
		urlPropertyRules(func(d DynatraceConfig) string { return d.URL }),
	)
	azureMonitorValidation = validation.New[AzureMonitorConfig](
		validation.For(func(a AzureMonitorConfig) string { return a.TenantID }).
			WithName("tenantId").
			Required().
			Rules(validation.StringUUID()),
	)
	honeycombValidation = validation.New[HoneycombConfig]()
)

const (
	errCodeExactlyOneDataSourceType              = "exactly_one_data_source_type"
	errCodeQueryDelayGreaterThanOrEqualToDefault = "query_delay_greater_than_or_equal_to_default"
)

var exactlyOneDataSourceTypeValidationRule = validation.NewSingleRule(func(spec Spec) error {
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
	if onlyType == 0 {
		return errors.New("must have exactly one data source type, none were provided")
	}
	return nil
}).WithErrorCode(errCodeExactlyOneDataSourceType)

var historicalDataRetrievalValidationRule = validation.NewSingleRule(func(spec Spec) error {
	if spec.HistoricalDataRetrieval == nil {
		return nil
	}
	typ, _ := spec.GetType()
	maxDuration, err := v1alpha.GetDataRetrievalMaxDuration(manifest.KindDirect, typ)
	if err != nil {
		return validation.NewPropertyError("historicalDataRetrieval", nil, err)
	}
	maxDurationAllowed := v1alpha.HistoricalRetrievalDuration{
		Value: maxDuration.Value,
		Unit:  maxDuration.Unit,
	}
	if spec.HistoricalDataRetrieval.MaxDuration.BiggerThan(maxDurationAllowed) {
		return validation.NewPropertyError(
			"historicalDataRetrieval.maxDuration",
			spec.HistoricalDataRetrieval.MaxDuration,
			errors.Errorf("must be less than or equal to %d %s",
				*maxDurationAllowed.Value, maxDurationAllowed.Unit))
	}
	return nil
})

var queryDelayGreaterThanOrEqualToDefaultValidationRule = validation.NewSingleRule(func(spec Spec) error {
	if spec.QueryDelay == nil {
		return nil
	}
	typ, _ := spec.GetType()
	directDefault := v1alpha.GetQueryDelayDefaults()[typ]
	if spec.QueryDelay.LessThan(directDefault) {
		return validation.NewPropertyError(
			"queryDelay",
			spec.QueryDelay,
			errors.Errorf("should be greater than or equal to %s", directDefault),
		)
	}
	return nil
}).WithErrorCode(errCodeQueryDelayGreaterThanOrEqualToDefault)

const errorCodeHTTPSSchemeRequired = "https_scheme_required"

func urlPropertyRules[S any](getter validation.PropertyGetter[string, S]) validation.PropertyRules[*url.URL, S] {
	return validation.Transform(getter, url.Parse).
		WithName("url").
		CascadeMode(validation.CascadeModeStop).
		Required().
		Rules(validation.URL()).
		Rules(validation.NewSingleRule(func(u *url.URL) error {
			if u.Scheme != "https" {
				return errors.New("requires https scheme")
			}
			return nil
		}).WithErrorCode(errorCodeHTTPSSchemeRequired))
}

func isHiddenValue(s string) bool { return s == "" || s == v1alpha.HiddenValue }

func validate(d Direct) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(directValidation, d)
}
