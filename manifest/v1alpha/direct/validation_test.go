package direct

import (
	_ "embed"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/validation"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

//go:embed test_data/expected_error.txt
var expectedError string

func TestValidate_AllErrors(t *testing.T) {
	err := validate(Direct{
		Kind: manifest.KindDirect,
		Metadata: Metadata{
			Name:        strings.Repeat("MY DIRECT", 20),
			DisplayName: strings.Repeat("my-direct", 10),
			Project:     strings.Repeat("MY PROJECT", 20),
		},
		Spec: Spec{
			Description: strings.Repeat("l", 2000),
			Datadog: &DatadogConfig{
				Site:           "datadoghq.com",
				APIKey:         "secret",
				ApplicationKey: "secret",
			},
		},
		ManifestSource: "/home/me/direct.yaml",
	})
	assert.Equal(t, strings.TrimSuffix(expectedError, "\n"), err.Error())
}

func TestValidateSpec(t *testing.T) {
	t.Run("exactly one data source - none provided", func(t *testing.T) {
		direct := validDirect(v1alpha.Datadog)
		direct.Spec.Datadog = nil
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec",
			Code: errCodeExactlyOneDataSourceType,
		})
	})
	t.Run("exactly one data source - both provided", func(t *testing.T) {
		for typ := range validDirectTypes {
			// We're using Prometheus as the offending data source type.
			// Any other source could've been used as well.
			if typ == v1alpha.Datadog {
				continue
			}
			direct := validDirect(typ)
			direct.Spec.Datadog = validDirectSpecs[v1alpha.Datadog].Datadog
			err := validate(direct)
			testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
				Prop: "spec",
				Code: errCodeExactlyOneDataSourceType,
			})
		}
	})
}

func TestValidateSpec_ReleaseChannel(t *testing.T) {
	t.Run("valid release channels", func(t *testing.T) {
		for _, rc := range []v1alpha.ReleaseChannel{
			v1alpha.ReleaseChannelStable,
			v1alpha.ReleaseChannelBeta,
			0, // empty field.
		} {
			direct := validDirect(v1alpha.Datadog)
			direct.Spec.ReleaseChannel = rc
			err := validate(direct)
			testutils.AssertNoError(t, direct, err)
		}
	})
	t.Run("invalid release channel", func(t *testing.T) {
		direct := validDirect(v1alpha.Datadog)
		direct.Spec.ReleaseChannel = -1
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.releaseChannel",
			Code: validation.ErrorCodeOneOf,
		})
	})
}

func TestValidateSpec_QueryDelay(t *testing.T) {
	t.Run("required", func(t *testing.T) {
		direct := validDirect(v1alpha.Datadog)
		direct.Spec.QueryDelay = &v1alpha.QueryDelay{}
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 2,
			testutils.ExpectedError{
				Prop: "spec.queryDelay.value",
				Code: validation.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.queryDelay.unit",
				Code: validation.ErrorCodeRequired,
			},
		)
	})
	t.Run("value too small", func(t *testing.T) {
		direct := validDirect(v1alpha.Datadog)
		direct.Spec.QueryDelay = &v1alpha.QueryDelay{Duration: v1alpha.Duration{
			Value: ptr(-1),
			Unit:  v1alpha.Minute,
		}}
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 2,
			testutils.ExpectedError{
				Prop: "spec.queryDelay",
				Message: "should be greater than or equal to " +
					v1alpha.GetQueryDelayDefaults()[v1alpha.Datadog.String()].String(),
			},
			testutils.ExpectedError{
				Prop: "spec.queryDelay.value",
				Code: validation.ErrorCodeGreaterThan,
			},
		)
	})
	t.Run("value too large", func(t *testing.T) {
		direct := validDirect(v1alpha.Datadog)
		direct.Spec.QueryDelay = &v1alpha.QueryDelay{Duration: v1alpha.Duration{
			Value: ptr(86400),
			Unit:  v1alpha.Second,
		}}
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.queryDelay.value",
			Code: validation.ErrorCodeLessThan,
		})
	})
	t.Run("valid units", func(t *testing.T) {
		for _, unit := range []v1alpha.DurationUnit{
			v1alpha.Minute,
			v1alpha.Second,
		} {
			direct := validDirect(v1alpha.Datadog)
			direct.Spec.QueryDelay = &v1alpha.QueryDelay{Duration: v1alpha.Duration{
				Value: ptr(10),
				Unit:  unit,
			}}
			err := validate(direct)
			testutils.AssertNoError(t, direct, err)
		}
	})
	t.Run("invalid units", func(t *testing.T) {
		for _, unit := range []v1alpha.DurationUnit{
			v1alpha.Millisecond,
			v1alpha.Hour,
			"invalid",
		} {
			direct := validDirect(v1alpha.Datadog)
			direct.Spec.QueryDelay = &v1alpha.QueryDelay{Duration: v1alpha.Duration{
				Value: ptr(10),
				Unit:  unit,
			}}
			err := validate(direct)
			testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
				Prop: "spec.queryDelay.unit",
				Code: validation.ErrorCodeOneOf,
			})
		}
	})
	t.Run("delay larger than max query delay", func(t *testing.T) {
		direct := validDirect(v1alpha.Datadog)
		direct.Spec.QueryDelay = &v1alpha.QueryDelay{Duration: v1alpha.Duration{
			Value: ptr(1441),
			Unit:  v1alpha.Minute,
		}}
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop:    "spec.queryDelay",
			Message: "must be smaller than or equal to 1440m",
		})
	})
}

func TestValidateSpec_HistoricalDataRetrieval(t *testing.T) {
	t.Run("required", func(t *testing.T) {
		direct := validDirect(v1alpha.Datadog)
		direct.Spec.HistoricalDataRetrieval = &v1alpha.HistoricalDataRetrieval{}
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 2,
			testutils.ExpectedError{
				Prop: "spec.historicalDataRetrieval.maxDuration",
				Code: validation.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.historicalDataRetrieval.defaultDuration",
				Code: validation.ErrorCodeRequired,
			},
		)
	})
	for name, test := range map[string]struct {
		Duration v1alpha.HistoricalRetrievalDuration
		Errors   []testutils.ExpectedError
	}{
		"required unit": {
			Duration: v1alpha.HistoricalRetrievalDuration{Value: ptr(10)},
			Errors: []testutils.ExpectedError{
				{
					Prop: "spec.historicalDataRetrieval.maxDuration.unit",
					Code: validation.ErrorCodeRequired,
				},
				{
					Prop: "spec.historicalDataRetrieval.defaultDuration.unit",
					Code: validation.ErrorCodeRequired,
				},
			},
		},
		"required value": {
			Duration: v1alpha.HistoricalRetrievalDuration{Unit: v1alpha.HRDHour},
			Errors: []testutils.ExpectedError{
				{
					Prop: "spec.historicalDataRetrieval.maxDuration.value",
					Code: validation.ErrorCodeRequired,
				},
				{
					Prop: "spec.historicalDataRetrieval.defaultDuration.value",
					Code: validation.ErrorCodeRequired,
				},
			},
		},
		"value too small": {
			Duration: v1alpha.HistoricalRetrievalDuration{
				Value: ptr(-1),
				Unit:  v1alpha.HRDHour,
			},
			Errors: []testutils.ExpectedError{
				{
					Prop: "spec.historicalDataRetrieval.maxDuration.value",
					Code: validation.ErrorCodeGreaterThan,
				},
				{
					Prop: "spec.historicalDataRetrieval.defaultDuration.value",
					Code: validation.ErrorCodeGreaterThan,
				},
			},
		},
		"value too large": {
			Duration: v1alpha.HistoricalRetrievalDuration{
				Value: ptr(43200),
				Unit:  v1alpha.HRDHour,
			},
			Errors: []testutils.ExpectedError{
				{
					Prop: "spec.historicalDataRetrieval.maxDuration.value",
					Code: validation.ErrorCodeLessThan,
				},
				{
					Prop: "spec.historicalDataRetrieval.defaultDuration.value",
					Code: validation.ErrorCodeLessThan,
				},
			},
		},
		"invalid unit": {
			Duration: v1alpha.HistoricalRetrievalDuration{
				Value: ptr(200),
				Unit:  "invalid",
			},
			Errors: []testutils.ExpectedError{
				{
					Prop: "spec.historicalDataRetrieval.maxDuration.unit",
					Code: validation.ErrorCodeOneOf,
				},
				{
					Prop: "spec.historicalDataRetrieval.defaultDuration.unit",
					Code: validation.ErrorCodeOneOf,
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			direct := validDirect(v1alpha.Datadog)
			direct.Spec.HistoricalDataRetrieval = &v1alpha.HistoricalDataRetrieval{
				MaxDuration:     test.Duration,
				DefaultDuration: test.Duration,
			}
			err := validate(direct)
			testutils.AssertContainsErrors(t, direct, err, len(test.Errors), test.Errors...)
		})
	}
	t.Run("valid units", func(t *testing.T) {
		for _, unit := range []v1alpha.HistoricalRetrievalDurationUnit{
			v1alpha.HRDMinute,
			v1alpha.HRDHour,
			v1alpha.HRDDay,
		} {
			direct := validDirect(v1alpha.Datadog)
			direct.Spec.HistoricalDataRetrieval = &v1alpha.HistoricalDataRetrieval{
				MaxDuration:     v1alpha.HistoricalRetrievalDuration{Value: ptr(10), Unit: unit},
				DefaultDuration: v1alpha.HistoricalRetrievalDuration{Value: ptr(10), Unit: unit},
			}
			err := validate(direct)
			testutils.AssertNoError(t, direct, err)
		}
	})
	t.Run("data retrieval disabled for data source type", func(t *testing.T) {
		direct := validDirect(v1alpha.Generic)
		direct.Spec.HistoricalDataRetrieval = &v1alpha.HistoricalDataRetrieval{
			MaxDuration: v1alpha.HistoricalRetrievalDuration{
				Value: ptr(20),
				Unit:  v1alpha.HRDHour,
			},
			DefaultDuration: v1alpha.HistoricalRetrievalDuration{
				Value: ptr(10),
				Unit:  v1alpha.HRDHour,
			},
		}
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop:    "spec.historicalDataRetrieval",
			Message: "historical data retrieval is not supported for Generic direct",
		})
	})
	t.Run("data retrieval default larger than max", func(t *testing.T) {
		direct := validDirect(v1alpha.Datadog)
		direct.Spec.HistoricalDataRetrieval = &v1alpha.HistoricalDataRetrieval{
			MaxDuration: v1alpha.HistoricalRetrievalDuration{
				Value: ptr(1),
				Unit:  v1alpha.HRDHour,
			},
			DefaultDuration: v1alpha.HistoricalRetrievalDuration{
				Value: ptr(2),
				Unit:  v1alpha.HRDHour,
			},
		}
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop:    "spec.historicalDataRetrieval.defaultDuration",
			Message: "must be smaller than or equal to 'maxDuration' (1 Hour)",
		})
	})
}

func TestValidateSpec_Datadog(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		direct := validDirect(v1alpha.Datadog)
		err := validate(direct)
		testutils.AssertNoError(t, direct, err)
	})
	t.Run("required site", func(t *testing.T) {
		direct := validDirect(v1alpha.Datadog)
		direct.Spec.Datadog.Site = ""
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.datadog.site",
			Code: validation.ErrorCodeRequired,
		})
	})
	t.Run("invalid site", func(t *testing.T) {
		direct := validDirect(v1alpha.Datadog)
		direct.Spec.Datadog.Site = "invalid"
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.datadog.site",
			Code: validation.ErrorCodeOneOf,
		})
	})
}

func TestValidateSpec_NewRelic(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		direct := validDirect(v1alpha.NewRelic)
		err := validate(direct)
		testutils.AssertNoError(t, direct, err)
	})
	t.Run("required account id", func(t *testing.T) {
		direct := validDirect(v1alpha.NewRelic)
		direct.Spec.NewRelic.AccountID = 0
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.newRelic.accountId",
			Code: validation.ErrorCodeRequired,
		})
	})
	t.Run("invalid account id", func(t *testing.T) {
		direct := validDirect(v1alpha.NewRelic)
		direct.Spec.NewRelic.AccountID = -1
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.newRelic.accountId",
			Code: validation.ErrorCodeGreaterThanOrEqualTo,
		})
	})
}

func TestValidateSpec_Lightstep(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		direct := validDirect(v1alpha.Lightstep)
		err := validate(direct)
		testutils.AssertNoError(t, direct, err)
	})
	t.Run("required fields", func(t *testing.T) {
		direct := validDirect(v1alpha.Lightstep)
		direct.Spec.Lightstep.Organization = ""
		direct.Spec.Lightstep.Project = ""
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 2,
			testutils.ExpectedError{
				Prop: "spec.lightstep.organization",
				Code: validation.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.lightstep.project",
				Code: validation.ErrorCodeRequired,
			},
		)
	})
}

func TestValidateSpec_SplunkObservability(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		direct := validDirect(v1alpha.SplunkObservability)
		err := validate(direct)
		testutils.AssertNoError(t, direct, err)
	})
	t.Run("required fields", func(t *testing.T) {
		direct := validDirect(v1alpha.SplunkObservability)
		direct.Spec.SplunkObservability.Realm = ""
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.splunkObservability.realm",
			Code: validation.ErrorCodeRequired,
		})
	})
}

func TestValidateSpec_Dynatrace(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		direct := validDirect(v1alpha.Dynatrace)
		err := validate(direct)
		testutils.AssertNoError(t, direct, err)
	})
	t.Run("required url", func(t *testing.T) {
		direct := validDirect(v1alpha.Dynatrace)
		direct.Spec.Dynatrace.URL = ""
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.dynatrace.url",
			Code: validation.ErrorCodeRequired,
		})
	})
	t.Run("invalid url", func(t *testing.T) {
		direct := validDirect(v1alpha.Dynatrace)
		direct.Spec.Dynatrace.URL = "h ttp"
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.dynatrace.url",
			Code: validation.ErrorCodeURL,
		})
	})
	urlTests := map[string]struct {
		url     string
		isValid bool
	}{
		"valid SaaS": {
			url:     "https://test.live.dynatrace.com",
			isValid: true,
		},
		"SaaS with port explicit speciefed": {
			url:     "https://test.live.dynatrace.com:433",
			isValid: true,
		},
		"valid SaaS multiple trailing /": {
			url:     "https://test.live.dynatrace.com///",
			isValid: true,
		},
		"invalid SaaS lack of https": {
			url:     "http://test.live.dynatrace.com",
			isValid: false,
		},
		"valid Managed/Environment ActiveGate lack of https": {
			url:     "http://test.com/e/environment-id",
			isValid: true,
		},
		"valid Managed/Environment ActiveGate wrong environment-id": {
			url:     "https://test.com/e/environment-id",
			isValid: true,
		},
		"valid Managed/Environment ActiveGate IP": {
			url:     "https://127.0.0.1/e/environment-id",
			isValid: true,
		},
		"valid Managed/Environment ActiveGate wrong environment-id, multiple /": {
			url:     "https://test.com///some-devops-path///e///environment-id///",
			isValid: true,
		},
	}
	for name, test := range urlTests {
		t.Run(name, func(t *testing.T) {
			direct := validDirect(v1alpha.Dynatrace)
			direct.Spec.Dynatrace.URL = test.url
			err := validate(direct)
			if test.isValid {
				testutils.AssertNoError(t, direct, err)
			} else {
				testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
					Prop: "spec.dynatrace.url",
					ContainsMessage: "Dynatrace SaaS URL (live.dynatrace.com)" +
						" requires https scheme and empty URL path",
				})
			}
		})
	}
}

func TestValidateSpec_AzureMonitor(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		direct := validDirect(v1alpha.AzureMonitor)
		err := validate(direct)
		testutils.AssertNoError(t, direct, err)
	})
	t.Run("required tenantId", func(t *testing.T) {
		direct := validDirect(v1alpha.AzureMonitor)
		direct.Spec.AzureMonitor.TenantID = ""
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.azureMonitor.tenantId",
			Code: validation.ErrorCodeRequired,
		})
	})
	t.Run("invalid tenantId", func(t *testing.T) {
		direct := validDirect(v1alpha.AzureMonitor)
		direct.Spec.AzureMonitor.TenantID = "invalid"
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.azureMonitor.tenantId",
			Code: validation.ErrorCodeStringUUID,
		})
	})
}

func validDirect(typ v1alpha.DataSourceType) Direct {
	spec := validDirectSpecs[typ]
	spec.Description = fmt.Sprintf("Example %s direct", typ)
	spec.ReleaseChannel = v1alpha.ReleaseChannelStable
	return New(Metadata{
		Name:        strings.ToLower(typ.String()),
		DisplayName: typ.String() + " Direct",
		Project:     "default",
	}, spec)
}

var validDirectSpecs = map[v1alpha.DataSourceType]Spec{
	v1alpha.Datadog: {
		Datadog: &DatadogConfig{
			Site: "datadoghq.com",
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
	v1alpha.ThousandEyes: {
		ThousandEyes: &ThousandEyesConfig{},
	},
	v1alpha.BigQuery: {
		BigQuery: &BigQueryConfig{},
	},
	v1alpha.CloudWatch: {
		CloudWatch: &CloudWatchConfig{},
	},
	v1alpha.Pingdom: {
		Pingdom: &PingdomConfig{},
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
	v1alpha.Honeycomb: {
		Honeycomb: &HoneycombConfig{},
	},
}

func ptr[T any](v T) *T { return &v }
