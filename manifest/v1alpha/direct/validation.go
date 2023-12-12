package direct

import (
	"net/url"
	"path"
	"strings"

	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

var directValidation = validation.New[Direct](
	v1alpha.FieldRuleMetadataName(func(d Direct) string { return d.Metadata.Name }),
	v1alpha.FieldRuleMetadataDisplayName(func(d Direct) string { return d.Metadata.DisplayName }),
	v1alpha.FieldRuleMetadataProject(func(d Direct) string { return d.Metadata.Project }),
	v1alpha.FieldRuleSpecDescription(func(d Direct) string { return d.Spec.Description }),
	validation.For(func(d Direct) Spec { return d.Spec }).
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
	validation.ForPointer(func(s Spec) *ThousandEyesConfig { return s.ThousandEyes }).
		WithName("thousandEyes").
		Include(thousandEyesValidation),
	validation.ForPointer(func(s Spec) *BigQueryConfig { return s.BigQuery }).
		WithName("bigQuery").
		Include(bigQueryValidation),
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
	validation.ForPointer(func(s Spec) *AzureMonitorConfig { return s.AzureMonitor }).
		WithName("azureMonitor").
		Include(azureMonitorValidation),
	validation.ForPointer(func(s Spec) *GCMConfig { return s.GCM }).
		WithName("gcm").
		Include(gcmValidation),
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
		validation.Transform(func(a AppDynamicsConfig) string { return a.URL }, url.Parse).
			WithName("url").
			Required().
			Rules(validation.URL()).
			StopOnError().
			Rules(
				validation.NewSingleRule(func(u *url.URL) error {
					if u.Scheme != "https" {
						return errors.New("requires https scheme")
					}
					return nil
				}),
			),
		validation.For(func(a AppDynamicsConfig) string { return a.ClientName }).
			WithName("clientName").
			Required(),
		validation.For(func(a AppDynamicsConfig) string { return a.AccountName }).
			WithName("accountName").
			Required(),
	)
	lightstepValidation = validation.New[LightstepConfig](
		validation.For(func(l LightstepConfig) string { return l.Organization }).
			WithName("organization").
			Required(),
		validation.For(func(l LightstepConfig) string { return l.Project }).
			WithName("project").
			Required(),
	)
	splunkObservabilityValidation = validation.New[SplunkObservabilityConfig](
		validation.For(func(s SplunkObservabilityConfig) string { return s.Realm }).
			WithName("realm").
			Required(),
	)
	dynatraceValidation = validation.New[DynatraceConfig](
		validation.Transform(func(d DynatraceConfig) string { return d.URL }, url.Parse).
			WithName("url").
			Required().
			Rules(
				validation.URL(),
				validation.NewSingleRule(func(u *url.URL) error {
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
	azureMonitorValidation = validation.New[AzureMonitorConfig](
		validation.For(func(a AzureMonitorConfig) string { return a.TenantID }).
			WithName("tenantId").
			Required().
			Rules(validation.StringUUID()),
	)
	// URL only.
	splunkValidation    = newURLValidator(func(s SplunkConfig) string { return s.URL })
	sumoLogicValidation = newURLValidator(func(s SumoLogicConfig) string { return s.URL })
	instanaValidation   = newURLValidator(func(i InstanaConfig) string { return i.URL })
	influxDBValidation  = newURLValidator(func(i InfluxDBConfig) string { return i.URL })
	// Empty configs.
	thousandEyesValidation = validation.New[ThousandEyesConfig]()
	bigQueryValidation     = validation.New[BigQueryConfig]()
	cloudWatchValidation   = validation.New[CloudWatchConfig]()
	pingdomValidation      = validation.New[PingdomConfig]()
	redshiftValidation     = validation.New[RedshiftConfig]()
	gcmValidation          = validation.New[GCMConfig]()
	honeycombValidation    = validation.New[HoneycombConfig]()
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
				"must have exactly one datas source type, detected both %s and %s",
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
	agentDefault := v1alpha.GetQueryDelayDefaults()[typ.String()]
	if spec.QueryDelay.LessThan(agentDefault) {
		return validation.NewPropertyError(
			"queryDelay",
			spec.QueryDelay,
			errors.Errorf("should be greater than or equal to %s", agentDefault),
		)
	}
	return nil
}).WithErrorCode(errCodeQueryDelayGreaterThanOrEqualToDefault)

// newURLValidator is a helper construct for Agent which only have a simple 'url' field validation.
func newURLValidator[S any](getter validation.PropertyGetter[string, S]) validation.Validator[S] {
	return validation.New[S](
		validation.For(getter).
			WithName("url").
			Required().
			Rules(validation.StringURL()),
	)
}

func isHiddenValue(s string) bool { return s == "" || s == v1alpha.HiddenValue }

func validate(d Direct) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(directValidation, d)
}
