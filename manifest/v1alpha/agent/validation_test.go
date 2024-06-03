package agent

import (
	_ "embed"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/internal/validation"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

var validationMessageRegexp = regexp.MustCompile(strings.TrimSpace(`
(?s)Validation for Agent '.*' in project '.*' has failed for the following fields:
.*
Manifest source: /home/me/agent.yaml
`))

func TestValidate_VersionAndKind(t *testing.T) {
	method := validAgent(v1alpha.Prometheus)
	method.APIVersion = "v0.1"
	method.Kind = manifest.KindProject
	method.ManifestSource = "/home/me/agent.yaml"
	err := validate(method)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, method, err, 2,
		testutils.ExpectedError{
			Prop: "apiVersion",
			Code: validation.ErrorCodeEqualTo,
		},
		testutils.ExpectedError{
			Prop: "kind",
			Code: validation.ErrorCodeEqualTo,
		},
	)
}

func TestValidate_Metadata(t *testing.T) {
	agent := validAgent(v1alpha.Prometheus)
	agent.Metadata = Metadata{
		Name:        strings.Repeat("MY AGENT", 20),
		DisplayName: strings.Repeat("my-agent", 10),
		Project:     strings.Repeat("MY PROJECT", 20),
	}
	agent.ManifestSource = "/home/me/agent.yaml"
	err := validate(agent)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, agent, err, 5,
		testutils.ExpectedError{
			Prop: "metadata.name",
			Code: validation.ErrorCodeStringIsDNSSubdomain,
		},
		testutils.ExpectedError{
			Prop: "metadata.displayName",
			Code: validation.ErrorCodeStringLength,
		},
		testutils.ExpectedError{
			Prop: "metadata.project",
			Code: validation.ErrorCodeStringIsDNSSubdomain,
		},
	)
}

func TestValidate_Spec(t *testing.T) {
	t.Run("description is too long", func(t *testing.T) {
		agent := validAgent(v1alpha.Prometheus)
		agent.Spec.Description = strings.Repeat("A", 2000)
		err := validate(agent)
		testutils.AssertContainsErrors(t, agent, err, 1, testutils.ExpectedError{
			Prop: "spec.description",
			Code: validation.ErrorCodeStringDescription,
		})
	})
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
			agent.Spec.Prometheus = validAgentSpec(v1alpha.Prometheus).Prometheus
			err := validate(agent)
			testutils.AssertContainsErrors(t, agent, err, 1, testutils.ExpectedError{
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
			agent := validAgent(v1alpha.Prometheus)
			agent.Spec.ReleaseChannel = rc
			err := validate(agent)
			testutils.AssertNoError(t, agent, err)
		}
	})
	t.Run("invalid release channel", func(t *testing.T) {
		agent := validAgent(v1alpha.Prometheus)
		agent.Spec.ReleaseChannel = -1
		err := validate(agent)
		testutils.AssertContainsErrors(t, agent, err, 1, testutils.ExpectedError{
			Prop: "spec.releaseChannel",
			Code: validation.ErrorCodeOneOf,
		})
	})
}

func TestValidateSpec_QueryDelay(t *testing.T) {
	t.Run("required", func(t *testing.T) {
		agent := validAgent(v1alpha.Prometheus)
		agent.Spec.QueryDelay = &v1alpha.QueryDelay{}
		err := validate(agent)
		testutils.AssertContainsErrors(t, agent, err, 2,
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
	t.Run("valid units", func(t *testing.T) {
		for _, unit := range []v1alpha.DurationUnit{
			v1alpha.Minute,
			v1alpha.Second,
		} {
			agent := validAgent(v1alpha.Prometheus)
			agent.Spec.QueryDelay = &v1alpha.QueryDelay{Duration: v1alpha.Duration{
				Value: ptr(10),
				Unit:  unit,
			}}
			err := validate(agent)
			testutils.AssertNoError(t, agent, err)
		}
	})
	t.Run("invalid units", func(t *testing.T) {
		for _, unit := range []v1alpha.DurationUnit{
			v1alpha.Millisecond,
			v1alpha.Hour,
			"invalid",
		} {
			agent := validAgent(v1alpha.Prometheus)
			agent.Spec.QueryDelay = &v1alpha.QueryDelay{Duration: v1alpha.Duration{
				Value: ptr(10),
				Unit:  unit,
			}}
			err := validate(agent)
			testutils.AssertContainsErrors(t, agent, err, 1, testutils.ExpectedError{
				Prop: "spec.queryDelay.unit",
				Code: validation.ErrorCodeOneOf,
			})
		}
	})
	t.Run("delay larger than max query delay", func(t *testing.T) {
		agent := validAgent(v1alpha.Prometheus)
		agent.Spec.QueryDelay = &v1alpha.QueryDelay{Duration: v1alpha.Duration{
			Value: ptr(1441),
			Unit:  v1alpha.Minute,
		}}
		err := validate(agent)
		testutils.AssertContainsErrors(t, agent, err, 1, testutils.ExpectedError{
			Prop:    "spec.queryDelay",
			Message: "must be less than or equal to 1440m",
		})
	})
	t.Run("delay less than default", func(t *testing.T) {
		for _, typ := range v1alpha.DataSourceTypeValues() {
			t.Run(typ.String(), func(t *testing.T) {
				agent := validAgent(typ)
				defaultDelay := v1alpha.GetQueryDelayDefaults()[typ]
				agent.Spec.QueryDelay = &v1alpha.QueryDelay{Duration: v1alpha.Duration{
					Value: ptr(*defaultDelay.Value - 1),
					Unit:  defaultDelay.Unit,
				}}
				err := validate(agent)
				testutils.AssertContainsErrors(t, agent, err, 1, testutils.ExpectedError{
					Prop: "spec.queryDelay",
					Code: errCodeQueryDelayGreaterThanOrEqualToDefault,
				})
			})
		}
	})
}

func TestValidateSpec_HistoricalDataRetrieval(t *testing.T) {
	t.Run("valid units", func(t *testing.T) {
		for _, unit := range []v1alpha.HistoricalRetrievalDurationUnit{
			v1alpha.HRDMinute,
			v1alpha.HRDHour,
			v1alpha.HRDDay,
		} {
			agent := validAgent(v1alpha.Prometheus)
			agent.Spec.HistoricalDataRetrieval = &v1alpha.HistoricalDataRetrieval{
				MaxDuration:     v1alpha.HistoricalRetrievalDuration{Unit: unit, Value: ptr(0)},
				DefaultDuration: v1alpha.HistoricalRetrievalDuration{Unit: unit, Value: ptr(0)},
			}
			err := validate(agent)
			testutils.AssertNoError(t, agent, err)
		}
	})
	t.Run("required", func(t *testing.T) {
		agent := validAgent(v1alpha.Prometheus)
		agent.Spec.HistoricalDataRetrieval = &v1alpha.HistoricalDataRetrieval{}
		err := validate(agent)
		testutils.AssertContainsErrors(t, agent, err, 2,
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
		Duration    v1alpha.HistoricalRetrievalDuration
		Errors      []testutils.ExpectedError
		ErrorsCount int
	}{
		"required unit": {
			Duration:    v1alpha.HistoricalRetrievalDuration{Value: ptr(10)},
			ErrorsCount: 2,
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
			Duration:    v1alpha.HistoricalRetrievalDuration{Unit: v1alpha.HRDHour},
			ErrorsCount: 2,
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
			ErrorsCount: 2,
			Errors: []testutils.ExpectedError{
				{
					Prop: "spec.historicalDataRetrieval.maxDuration.value",
					Code: validation.ErrorCodeGreaterThanOrEqualTo,
				},
				{
					Prop: "spec.historicalDataRetrieval.defaultDuration.value",
					Code: validation.ErrorCodeGreaterThanOrEqualTo,
				},
			},
		},
		"value too large": {
			Duration: v1alpha.HistoricalRetrievalDuration{
				Value: ptr(43201),
				Unit:  v1alpha.HRDHour,
			},
			ErrorsCount: 3,
			Errors: []testutils.ExpectedError{
				{
					Prop: "spec.historicalDataRetrieval.maxDuration.value",
					Code: validation.ErrorCodeLessThanOrEqualTo,
				},
				{
					Prop: "spec.historicalDataRetrieval.defaultDuration.value",
					Code: validation.ErrorCodeLessThanOrEqualTo,
				},
			},
		},
		"invalid unit": {
			Duration: v1alpha.HistoricalRetrievalDuration{
				Value: ptr(200),
				Unit:  "invalid",
			},
			ErrorsCount: 2,
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
			agent := validAgent(v1alpha.Prometheus)
			agent.Spec.HistoricalDataRetrieval = &v1alpha.HistoricalDataRetrieval{
				MaxDuration:     test.Duration,
				DefaultDuration: test.Duration,
			}
			err := validate(agent)
			testutils.AssertContainsErrors(t, agent, err, test.ErrorsCount, test.Errors...)
		})
	}
	t.Run("valid units", func(t *testing.T) {
		for _, unit := range []v1alpha.HistoricalRetrievalDurationUnit{
			v1alpha.HRDMinute,
			v1alpha.HRDHour,
			v1alpha.HRDDay,
		} {
			agent := validAgent(v1alpha.Prometheus)
			agent.Spec.HistoricalDataRetrieval = &v1alpha.HistoricalDataRetrieval{
				MaxDuration:     v1alpha.HistoricalRetrievalDuration{Value: ptr(10), Unit: unit},
				DefaultDuration: v1alpha.HistoricalRetrievalDuration{Value: ptr(10), Unit: unit},
			}
			err := validate(agent)
			testutils.AssertNoError(t, agent, err)
		}
	})
	t.Run("data retrieval disabled for data source type", func(t *testing.T) {
		agent := validAgent(v1alpha.Generic)
		agent.Spec.HistoricalDataRetrieval = &v1alpha.HistoricalDataRetrieval{
			MaxDuration: v1alpha.HistoricalRetrievalDuration{
				Value: ptr(20),
				Unit:  v1alpha.HRDHour,
			},
			DefaultDuration: v1alpha.HistoricalRetrievalDuration{
				Value: ptr(10),
				Unit:  v1alpha.HRDHour,
			},
		}
		err := validate(agent)
		testutils.AssertContainsErrors(t, agent, err, 1, testutils.ExpectedError{
			Prop:    "spec.historicalDataRetrieval",
			Message: "historical data retrieval is not supported for Generic Agent",
		})
	})
	t.Run("data retrieval default larger than max", func(t *testing.T) {
		agent := validAgent(v1alpha.Prometheus)
		agent.Spec.HistoricalDataRetrieval = &v1alpha.HistoricalDataRetrieval{
			MaxDuration: v1alpha.HistoricalRetrievalDuration{
				Value: ptr(1),
				Unit:  v1alpha.HRDHour,
			},
			DefaultDuration: v1alpha.HistoricalRetrievalDuration{
				Value: ptr(2),
				Unit:  v1alpha.HRDHour,
			},
		}
		err := validate(agent)
		testutils.AssertContainsErrors(t, agent, err, 1, testutils.ExpectedError{
			Prop:    "spec.historicalDataRetrieval.defaultDuration",
			Message: "must be less than or equal to 'maxDuration' (1 Hour)",
		})
	})
	t.Run("data retrieval max greater than max allowed", func(t *testing.T) {
		for _, typ := range v1alpha.DataSourceTypeValues() {
			maxDuration, err := v1alpha.GetDataRetrievalMaxDuration(manifest.KindAgent, typ)
			// Skip unsupported types.
			if err != nil {
				continue
			}
			t.Run(typ.String(), func(t *testing.T) {
				agent := validAgent(typ)
				agent.Spec.HistoricalDataRetrieval = &v1alpha.HistoricalDataRetrieval{
					MaxDuration: v1alpha.HistoricalRetrievalDuration{
						Value: ptr(*maxDuration.Value + 1),
						Unit:  maxDuration.Unit,
					},
					DefaultDuration: v1alpha.HistoricalRetrievalDuration{
						Value: ptr(0),
						Unit:  maxDuration.Unit,
					},
				}
				objErr := validate(agent)
				testutils.AssertContainsErrors(t, agent, objErr, 1, testutils.ExpectedError{
					Prop: "spec.historicalDataRetrieval.maxDuration",
					Message: fmt.Sprintf("must be less than or equal to %d %s",
						*maxDuration.Value, maxDuration.Unit),
				})
			})
		}
	})
}

func TestValidateSpec_URLOnlyAgents(t *testing.T) {
	for propName, typ := range map[string]v1alpha.DataSourceType{
		"prometheus":  v1alpha.Prometheus,
		"appDynamics": v1alpha.AppDynamics,
		"splunk":      v1alpha.Splunk,
		"graphite":    v1alpha.Graphite,
		"opentsdb":    v1alpha.OpenTSDB,
		"grafanaLoki": v1alpha.GrafanaLoki,
		"sumoLogic":   v1alpha.SumoLogic,
		"instana":     v1alpha.Instana,
		"influxdb":    v1alpha.InfluxDB,
	} {
		t.Run(typ.String(), func(t *testing.T) {
			t.Run("passes", func(t *testing.T) {
				agent := validAgent(typ)
				err := validate(agent)
				testutils.AssertNoError(t, agent, err)
			})
			t.Run("required url", func(t *testing.T) {
				agent := validAgent(typ)
				setURLValue(t, &agent.Spec, typ.String(), "")
				err := validate(agent)
				testutils.AssertContainsErrors(t, agent, err, 1, testutils.ExpectedError{
					Prop: fmt.Sprintf("spec.%s.url", propName),
					Code: validation.ErrorCodeRequired,
				})
			})
			t.Run("invalid url", func(t *testing.T) {
				agent := validAgent(typ)
				setURLValue(t, &agent.Spec, typ.String(), "invalid")
				err := validate(agent)
				testutils.AssertContainsErrors(t, agent, err, 1, testutils.ExpectedError{
					Prop: fmt.Sprintf("spec.%s.url", propName),
					Code: validation.ErrorCodeStringURL,
				})
			})
		})
	}
}

func TestValidateSpec_EmptyConfigs(t *testing.T) {
	for _, typ := range []v1alpha.DataSourceType{
		v1alpha.ThousandEyes,
		v1alpha.BigQuery,
		v1alpha.CloudWatch,
		v1alpha.Pingdom,
		v1alpha.Redshift,
		v1alpha.GCM,
		v1alpha.Generic,
		v1alpha.Honeycomb,
	} {
		t.Run(typ.String(), func(t *testing.T) {
			agent := validAgent(typ)
			err := validate(agent)
			testutils.AssertNoError(t, agent, err)
		})
	}
}

func TestValidateSpec_Datadog(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		agent := validAgent(v1alpha.Datadog)
		err := validate(agent)
		testutils.AssertNoError(t, agent, err)
	})
	t.Run("required site", func(t *testing.T) {
		agent := validAgent(v1alpha.Datadog)
		agent.Spec.Datadog.Site = ""
		err := validate(agent)
		testutils.AssertContainsErrors(t, agent, err, 1, testutils.ExpectedError{
			Prop: "spec.datadog.site",
			Code: validation.ErrorCodeRequired,
		})
	})
	t.Run("invalid site", func(t *testing.T) {
		agent := validAgent(v1alpha.Datadog)
		agent.Spec.Datadog.Site = "invalid"
		err := validate(agent)
		testutils.AssertContainsErrors(t, agent, err, 1, testutils.ExpectedError{
			Prop: "spec.datadog.site",
			Code: validation.ErrorCodeOneOf,
		})
	})
}

func TestValidateSpec_NewRelic(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		agent := validAgent(v1alpha.NewRelic)
		err := validate(agent)
		testutils.AssertNoError(t, agent, err)
	})
	t.Run("required account id", func(t *testing.T) {
		agent := validAgent(v1alpha.NewRelic)
		agent.Spec.NewRelic.AccountID = 0
		err := validate(agent)
		testutils.AssertContainsErrors(t, agent, err, 1, testutils.ExpectedError{
			Prop: "spec.newRelic.accountId",
			Code: validation.ErrorCodeRequired,
		})
	})
	t.Run("invalid account id", func(t *testing.T) {
		agent := validAgent(v1alpha.NewRelic)
		agent.Spec.NewRelic.AccountID = -1
		err := validate(agent)
		testutils.AssertContainsErrors(t, agent, err, 1, testutils.ExpectedError{
			Prop: "spec.newRelic.accountId",
			Code: validation.ErrorCodeGreaterThanOrEqualTo,
		})
	})
}

func TestValidateSpec_Lightstep(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		agent := validAgent(v1alpha.Lightstep)
		agent.Spec.Lightstep.URL = ""
		err := validate(agent)
		testutils.AssertNoError(t, agent, err)
	})
	t.Run("required fields", func(t *testing.T) {
		agent := validAgent(v1alpha.Lightstep)
		agent.Spec.Lightstep.Organization = ""
		agent.Spec.Lightstep.Project = ""
		err := validate(agent)
		testutils.AssertContainsErrors(t, agent, err, 2,
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
	t.Run("invalid url", func(t *testing.T) {
		agent := validAgent(v1alpha.Lightstep)
		agent.Spec.Lightstep.URL = "h ttp"
		err := validate(agent)
		testutils.AssertContainsErrors(t, agent, err, 1, testutils.ExpectedError{
			Prop: "spec.lightstep.url",
			Code: validation.ErrorCodeURL,
		})
	})
	urlTests := map[string]struct {
		url     string
		isValid bool
	}{
		"valid .com": {
			url: "https://api.lightstep.com",
		},
		"valid .eu": {
			url: "https://api.eu.lightstep.com",
		},
	}
	t.Run("test url", func(t *testing.T) {
		for name, test := range urlTests {
			t.Run(name, func(t *testing.T) {
				agent := validAgent(v1alpha.Lightstep)
				agent.Spec.Lightstep.URL = test.url
				err := validate(agent)
				testutils.AssertNoError(t, agent, err)
			})
		}
	})
}

func TestValidateSpec_SplunkObservability(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		agent := validAgent(v1alpha.SplunkObservability)
		err := validate(agent)
		testutils.AssertNoError(t, agent, err)
	})
	t.Run("required fields", func(t *testing.T) {
		agent := validAgent(v1alpha.SplunkObservability)
		agent.Spec.SplunkObservability.Realm = ""
		err := validate(agent)
		testutils.AssertContainsErrors(t, agent, err, 1, testutils.ExpectedError{
			Prop: "spec.splunkObservability.realm",
			Code: validation.ErrorCodeRequired,
		})
	})
}

func TestValidateSpec_Dynatrace(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		agent := validAgent(v1alpha.Dynatrace)
		err := validate(agent)
		testutils.AssertNoError(t, agent, err)
	})
	t.Run("required url", func(t *testing.T) {
		agent := validAgent(v1alpha.Dynatrace)
		agent.Spec.Dynatrace.URL = ""
		err := validate(agent)
		testutils.AssertContainsErrors(t, agent, err, 1, testutils.ExpectedError{
			Prop: "spec.dynatrace.url",
			Code: validation.ErrorCodeRequired,
		})
	})
	t.Run("invalid url", func(t *testing.T) {
		agent := validAgent(v1alpha.Dynatrace)
		agent.Spec.Dynatrace.URL = "h ttp"
		err := validate(agent)
		testutils.AssertContainsErrors(t, agent, err, 1, testutils.ExpectedError{
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
		"SaaS with port explicit specified": {
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
			agent := validAgent(v1alpha.Dynatrace)
			agent.Spec.Dynatrace.URL = test.url
			err := validate(agent)
			if test.isValid {
				testutils.AssertNoError(t, agent, err)
			} else {
				testutils.AssertContainsErrors(t, agent, err, 1, testutils.ExpectedError{
					Prop: "spec.dynatrace.url",
					ContainsMessage: "Dynatrace SaaS URL (live.dynatrace.com)" +
						" requires https scheme and empty URL path",
				})
			}
		})
	}
}

func TestValidateSpec_AmazonPrometheus(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		agent := validAgent(v1alpha.AmazonPrometheus)
		agent.Spec.AmazonPrometheus.Region = strings.Repeat("l", 255)
		err := validate(agent)
		testutils.AssertNoError(t, agent, err)
	})
	t.Run("required fields", func(t *testing.T) {
		agent := validAgent(v1alpha.AmazonPrometheus)
		agent.Spec.AmazonPrometheus.URL = ""
		agent.Spec.AmazonPrometheus.Region = ""
		err := validate(agent)
		testutils.AssertContainsErrors(t, agent, err, 2,
			testutils.ExpectedError{
				Prop: "spec.amazonPrometheus.url",
				Code: validation.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.amazonPrometheus.region",
				Code: validation.ErrorCodeRequired,
			},
		)
	})
	t.Run("invalid fields", func(t *testing.T) {
		agent := validAgent(v1alpha.AmazonPrometheus)
		agent.Spec.AmazonPrometheus.URL = "invalid"
		agent.Spec.AmazonPrometheus.Region = strings.Repeat("l", 256)
		err := validate(agent)
		testutils.AssertContainsErrors(t, agent, err, 2,
			testutils.ExpectedError{
				Prop: "spec.amazonPrometheus.url",
				Code: validation.ErrorCodeStringURL,
			},
			testutils.ExpectedError{
				Prop: "spec.amazonPrometheus.region",
				Code: validation.ErrorCodeStringMaxLength,
			},
		)
	})
}

func TestValidateSpec_AzureMonitor(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		agent := validAgent(v1alpha.AzureMonitor)
		err := validate(agent)
		testutils.AssertNoError(t, agent, err)
	})
	t.Run("required tenantId", func(t *testing.T) {
		agent := validAgent(v1alpha.AzureMonitor)
		agent.Spec.AzureMonitor.TenantID = ""
		err := validate(agent)
		testutils.AssertContainsErrors(t, agent, err, 1, testutils.ExpectedError{
			Prop: "spec.azureMonitor.tenantId",
			Code: validation.ErrorCodeRequired,
		})
	})
	t.Run("invalid tenantId", func(t *testing.T) {
		agent := validAgent(v1alpha.AzureMonitor)
		agent.Spec.AzureMonitor.TenantID = "invalid"
		err := validate(agent)
		testutils.AssertContainsErrors(t, agent, err, 1, testutils.ExpectedError{
			Prop: "spec.azureMonitor.tenantId",
			Code: validation.ErrorCodeStringUUID,
		})
	})
}

func validAgent(typ v1alpha.DataSourceType) Agent {
	spec := validAgentSpec(typ)
	spec.Description = fmt.Sprintf("Example %s Agent", typ)
	spec.ReleaseChannel = v1alpha.ReleaseChannelStable
	return New(Metadata{
		Name:        strings.ToLower(typ.String()),
		DisplayName: typ.String() + " Agent",
		Project:     "default",
	}, spec)
}

func validAgentSpec(typ v1alpha.DataSourceType) Spec {
	specs := map[v1alpha.DataSourceType]Spec{
		v1alpha.Prometheus: {
			Prometheus: &PrometheusConfig{
				URL: "https://prometheus-service.monitoring:8080",
			},
		},
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
				URL:          "https://api.lightstep.com",
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
		v1alpha.LogicMonitor: {
			LogicMonitor: &LogicMonitorConfig{
				Account: "account",
			},
		},
		v1alpha.AzurePrometheus: {
			AzurePrometheus: &AzurePrometheusConfig{
				URL:      "https://prometheus-service.monitoring:8080",
				TenantID: "tenant_id",
			},
		},
	}

	return specs[typ]
}

// setURLValue is a help function which sets the value of 'URL' field of the given Agent config.
func setURLValue(t *testing.T, obj interface{}, fieldName, value string) {
	t.Helper()
	v := reflect.ValueOf(obj)
	v.Elem().
		FieldByName(fieldName).
		Elem().
		FieldByName("URL").
		SetString(value)
}

func ptr[T any](v T) *T { return &v }
