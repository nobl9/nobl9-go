package direct

import (
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/internal/manifest/v1alphatest"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

var validationMessageRegexp = regexp.MustCompile(strings.TrimSpace(`
(?s)Validation for Direct '.*' in project '.*' has failed for the following fields:
.*
Manifest source: /home/me/direct.yaml
`))

func TestValidate_VersionAndKind(t *testing.T) {
	direct := validDirect(v1alpha.CloudWatch)
	direct.APIVersion = "v0.1"
	direct.Kind = manifest.KindProject
	direct.ManifestSource = "/home/me/direct.yaml"
	err := validate(direct)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, direct, err, 2,
		testutils.ExpectedError{
			Prop: "apiVersion",
			Code: rules.ErrorCodeEqualTo,
		},
		testutils.ExpectedError{
			Prop: "kind",
			Code: rules.ErrorCodeEqualTo,
		},
	)
}

func TestValidate_Metadata(t *testing.T) {
	direct := validDirect(v1alpha.Datadog)
	direct.Metadata = Metadata{
		Name:        strings.Repeat("MY DIRECT", 20),
		DisplayName: strings.Repeat("my-direct", 29),
		Project:     strings.Repeat("MY PROJECT", 20),
	}
	direct.ManifestSource = "/home/me/direct.yaml"
	err := validate(direct)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, direct, err, 3,
		testutils.ExpectedError{
			Prop: "metadata.name",
			Code: validationV1Alpha.ErrorCodeStringName,
		},
		testutils.ExpectedError{
			Prop: "metadata.displayName",
			Code: rules.ErrorCodeStringMaxLength,
		},
		testutils.ExpectedError{
			Prop: "metadata.project",
			Code: validationV1Alpha.ErrorCodeStringName,
		},
	)
}

func TestValidate_Metadata_Annotations(t *testing.T) {
	for name, test := range v1alphatest.GetMetadataAnnotationsTestCases[Direct](t, "metadata.annotations") {
		t.Run(name, func(t *testing.T) {
			svc := validDirect(v1alpha.Datadog)
			svc.Metadata.Annotations = test.Annotations
			test.Test(t, svc, validate)
		})
	}
}

func TestValidateSpec(t *testing.T) {
	t.Run("description too long", func(t *testing.T) {
		direct := validDirect(v1alpha.Datadog)
		direct.Spec.Description = strings.Repeat("a", 2000)
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.description",
			Code: validationV1Alpha.ErrorCodeStringDescription,
		})
	})
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
			direct.Spec.Datadog = validDirectSpec(v1alpha.Datadog).Datadog
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
			Code: rules.ErrorCodeOneOf,
		})
	})
	t.Run("data source type using supported release channel", func(t *testing.T) {
		direct := validDirect(v1alpha.Datadog)
		direct.Spec.ReleaseChannel = v1alpha.ReleaseChannelBeta
		err := validate(direct)
		testutils.AssertNoError(t, direct, err)
	})
	t.Run("data source type using unsupported release channel", func(t *testing.T) {
		direct := validDirect(v1alpha.Datadog)
		direct.Spec.ReleaseChannel = v1alpha.ReleaseChannelAlpha
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.releaseChannel",
			Code: errCodeUnsupportedReleaseChannel,
		})
	})
	t.Run("data source type using unsupported release channel", func(t *testing.T) {
		direct := validDirect(v1alpha.SplunkObservability)
		direct.Spec.ReleaseChannel = v1alpha.ReleaseChannelStable
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.releaseChannel",
			Code: errCodeUnsupportedReleaseChannel,
		})
	})
	t.Run("alpha enabled for Honeycomb", func(t *testing.T) {
		direct := validDirect(v1alpha.Honeycomb)
		direct.Spec.ReleaseChannel = v1alpha.ReleaseChannelAlpha
		err := validate(direct)
		testutils.AssertNoError(t, direct, err)
	})
}

func TestValidateSpec_QueryDelay(t *testing.T) {
	t.Run("required", func(t *testing.T) {
		direct := validDirect(v1alpha.BigQuery)
		direct.Spec.QueryDelay = &v1alpha.QueryDelay{}
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 2,
			testutils.ExpectedError{
				Prop: "spec.queryDelay.value",
				Code: rules.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.queryDelay.unit",
				Code: rules.ErrorCodeRequired,
			},
		)
	})
	t.Run("valid units", func(t *testing.T) {
		for _, unit := range []v1alpha.DurationUnit{
			v1alpha.Minute,
			v1alpha.Second,
		} {
			direct := validDirect(v1alpha.BigQuery)
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
			direct := validDirect(v1alpha.BigQuery)
			direct.Spec.QueryDelay = &v1alpha.QueryDelay{Duration: v1alpha.Duration{
				Value: ptr(10),
				Unit:  unit,
			}}
			err := validate(direct)
			testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
				Prop: "spec.queryDelay.unit",
				Code: rules.ErrorCodeOneOf,
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
			Prop: "spec.queryDelay",
			Code: errCodeQueryDelayOutOfBounds,
		})
	})

	t.Run("delay larger than data source max query delay", func(t *testing.T) {
		direct := validDirect(v1alpha.SplunkObservability)
		direct.Spec.QueryDelay = &v1alpha.QueryDelay{Duration: v1alpha.Duration{
			Value: ptr(16),
			Unit:  v1alpha.Minute,
		}}
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.queryDelay",
			Code: errCodeQueryDelayOutOfBounds,
		})
	})
	t.Run("delay less than default", func(t *testing.T) {
		for typ := range validDirectTypes {
			t.Run(typ.String(), func(t *testing.T) {
				direct := validDirect(typ)
				defaultDelay := v1alpha.GetQueryDelayDefaults()[typ]
				direct.Spec.QueryDelay = &v1alpha.QueryDelay{Duration: v1alpha.Duration{
					Value: ptr(*defaultDelay.Value - 1),
					Unit:  defaultDelay.Unit,
				}}
				err := validate(direct)
				testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
					Prop: "spec.queryDelay",
					Code: errCodeQueryDelayOutOfBounds,
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
			direct := validDirect(v1alpha.Datadog)
			direct.Spec.HistoricalDataRetrieval = &v1alpha.HistoricalDataRetrieval{
				MaxDuration:     v1alpha.HistoricalRetrievalDuration{Unit: unit, Value: ptr(0)},
				DefaultDuration: v1alpha.HistoricalRetrievalDuration{Unit: unit, Value: ptr(0)},
			}
			err := validate(direct)
			testutils.AssertNoError(t, direct, err)
		}
	})
	t.Run("required", func(t *testing.T) {
		direct := validDirect(v1alpha.Datadog)
		direct.Spec.HistoricalDataRetrieval = &v1alpha.HistoricalDataRetrieval{}
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 2,
			testutils.ExpectedError{
				Prop: "spec.historicalDataRetrieval.maxDuration",
				Code: rules.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.historicalDataRetrieval.defaultDuration",
				Code: rules.ErrorCodeRequired,
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
					Code: rules.ErrorCodeRequired,
				},
				{
					Prop: "spec.historicalDataRetrieval.defaultDuration.unit",
					Code: rules.ErrorCodeRequired,
				},
			},
		},
		"required value": {
			Duration:    v1alpha.HistoricalRetrievalDuration{Unit: v1alpha.HRDHour},
			ErrorsCount: 2,
			Errors: []testutils.ExpectedError{
				{
					Prop: "spec.historicalDataRetrieval.maxDuration.value",
					Code: rules.ErrorCodeRequired,
				},
				{
					Prop: "spec.historicalDataRetrieval.defaultDuration.value",
					Code: rules.ErrorCodeRequired,
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
					Code: rules.ErrorCodeGreaterThanOrEqualTo,
				},
				{
					Prop: "spec.historicalDataRetrieval.defaultDuration.value",
					Code: rules.ErrorCodeGreaterThanOrEqualTo,
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
					Code: rules.ErrorCodeLessThanOrEqualTo,
				},
				{
					Prop: "spec.historicalDataRetrieval.defaultDuration.value",
					Code: rules.ErrorCodeLessThanOrEqualTo,
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
					Code: rules.ErrorCodeOneOf,
				},
				{
					Prop: "spec.historicalDataRetrieval.defaultDuration.unit",
					Code: rules.ErrorCodeOneOf,
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
			testutils.AssertContainsErrors(t, direct, err, test.ErrorsCount, test.Errors...)
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
		direct := validDirect(v1alpha.Instana)
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
			Message: "historical data retrieval is not supported for Instana Direct",
		})
	})
	t.Run("data retrieval default, defaultDuration larger than max",
		func(t *testing.T) {
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
				Message: "must be less than or equal to 'maxDuration' (1 Hour)",
			})
		})
	t.Run("data retrieval max greater than max allowed", func(t *testing.T) {
		for typ := range validDirectTypes {
			maxDuration, err := v1alpha.GetDataRetrievalMaxDuration(manifest.KindDirect, typ)
			// Skip unsupported types.
			if err != nil {
				continue
			}
			t.Run(typ.String(), func(t *testing.T) {
				direct := validDirect(typ)
				direct.Spec.HistoricalDataRetrieval = &v1alpha.HistoricalDataRetrieval{
					MaxDuration: v1alpha.HistoricalRetrievalDuration{
						Value: ptr(*maxDuration.Value + 1),
						Unit:  maxDuration.Unit,
					},
					DefaultDuration: v1alpha.HistoricalRetrievalDuration{
						Value: ptr(0),
						Unit:  maxDuration.Unit,
					},
				}
				objErr := validate(direct)
				testutils.AssertContainsErrors(t, direct, objErr, 1, testutils.ExpectedError{
					Prop: "spec.historicalDataRetrieval.maxDuration",
					Message: fmt.Sprintf("must be less than or equal to %d %s",
						*maxDuration.Value, maxDuration.Unit),
				})
			})
		}
	})
}

func TestValidateSpec_Datadog(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		for name, direct := range map[string]Direct{
			"with secrets": validDirect(v1alpha.Datadog),
			"empty secrets": func() Direct {
				d := validDirect(v1alpha.Datadog)
				d.Spec.Datadog.ApplicationKey = ""
				d.Spec.Datadog.APIKey = ""
				return d
			}(),
			"hidden secrets": func() Direct {
				d := validDirect(v1alpha.Datadog)
				d.Spec.Datadog.ApplicationKey = v1alpha.HiddenValue
				d.Spec.Datadog.APIKey = v1alpha.HiddenValue
				return d
			}(),
		} {
			t.Run(name, func(t *testing.T) {
				err := validate(direct)
				testutils.AssertNoError(t, direct, err)
			})
		}
	})
	t.Run("required site", func(t *testing.T) {
		direct := validDirect(v1alpha.Datadog)
		direct.Spec.Datadog.Site = ""
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.datadog.site",
			Code: rules.ErrorCodeRequired,
		})
	})
	t.Run("invalid site", func(t *testing.T) {
		direct := validDirect(v1alpha.Datadog)
		direct.Spec.Datadog.Site = "invalid"
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.datadog.site",
			Code: rules.ErrorCodeOneOf,
		})
	})
}

func TestValidateSpec_NewRelic(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		for name, direct := range map[string]Direct{
			"with secrets": validDirect(v1alpha.NewRelic),
			"empty secrets": func() Direct {
				d := validDirect(v1alpha.NewRelic)
				d.Spec.NewRelic.InsightsQueryKey = ""
				return d
			}(),
			"hidden secrets": func() Direct {
				d := validDirect(v1alpha.NewRelic)
				d.Spec.NewRelic.InsightsQueryKey = v1alpha.HiddenValue
				return d
			}(),
		} {
			t.Run(name, func(t *testing.T) {
				err := validate(direct)
				testutils.AssertNoError(t, direct, err)
			})
		}
	})
	t.Run("required account id", func(t *testing.T) {
		direct := validDirect(v1alpha.NewRelic)
		direct.Spec.NewRelic.AccountID = 0
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.newRelic.accountId",
			Code: rules.ErrorCodeRequired,
		})
	})
	t.Run("invalid account id", func(t *testing.T) {
		direct := validDirect(v1alpha.NewRelic)
		direct.Spec.NewRelic.AccountID = -1
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.newRelic.accountId",
			Code: rules.ErrorCodeGreaterThanOrEqualTo,
		})
	})
	t.Run("invalid insights key", func(t *testing.T) {
		direct := validDirect(v1alpha.NewRelic)
		direct.Spec.NewRelic.InsightsQueryKey = "123"
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.newRelic.insightsQueryKey",
			Code: rules.ErrorCodeStringStartsWith,
		})
	})
}

func TestValidateSpec_AppDynamics(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		direct := validDirect(v1alpha.AppDynamics)
		err := validate(direct)
		testutils.AssertNoError(t, direct, err)
	})
	t.Run("required fields", func(t *testing.T) {
		direct := validDirect(v1alpha.AppDynamics)
		direct.Spec.AppDynamics.URL = ""
		direct.Spec.AppDynamics.ClientName = ""
		direct.Spec.AppDynamics.AccountName = ""
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 3,
			testutils.ExpectedError{
				Prop: "spec.appDynamics.url",
				Code: rules.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.appDynamics.clientName",
				Code: rules.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.appDynamics.accountName",
				Code: rules.ErrorCodeRequired,
			},
		)
	})
	t.Run("invalid url", func(t *testing.T) {
		direct := validDirect(v1alpha.AppDynamics)
		direct.Spec.AppDynamics.URL = "nobl9.com"
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.appDynamics.url",
			Code: rules.ErrorCodeURL,
		})
	})
	t.Run("url must be https", func(t *testing.T) {
		direct := validDirect(v1alpha.AppDynamics)
		direct.Spec.AppDynamics.URL = "http://nobl9.com"
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.appDynamics.url",
			Code: errorCodeHTTPSSchemeRequired,
		})
	})
}

func TestValidateSpec_BigQuery(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		for name, direct := range map[string]Direct{
			"with secrets": validDirect(v1alpha.BigQuery),
			"empty secrets": func() Direct {
				d := validDirect(v1alpha.BigQuery)
				d.Spec.BigQuery.ServiceAccountKey = ""
				return d
			}(),
			"hidden secrets": func() Direct {
				d := validDirect(v1alpha.BigQuery)
				d.Spec.BigQuery.ServiceAccountKey = v1alpha.HiddenValue
				return d
			}(),
		} {
			t.Run(name, func(t *testing.T) {
				err := validate(direct)
				testutils.AssertNoError(t, direct, err)
			})
		}
	})
	t.Run("serviceAccountKey must be a valid JSON", func(t *testing.T) {
		direct := validDirect(v1alpha.BigQuery)
		direct.Spec.BigQuery.ServiceAccountKey = "{["
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.bigQuery.serviceAccountKey",
			Code: rules.ErrorCodeStringJSON,
		})
	})
}

func TestValidateSpec_SplunkObservability(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		direct := validDirect(v1alpha.SplunkObservability)
		direct.Spec.ReleaseChannel = v1alpha.ReleaseChannelAlpha
		err := validate(direct)
		testutils.AssertNoError(t, direct, err)
	})
	t.Run("required realm", func(t *testing.T) {
		direct := validDirect(v1alpha.SplunkObservability)
		direct.Spec.ReleaseChannel = v1alpha.ReleaseChannelAlpha
		direct.Spec.SplunkObservability.Realm = ""
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.splunkObservability.realm",
			Code: rules.ErrorCodeRequired,
		})
	})
}

func TestValidateSpec_Splunk(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		direct := validDirect(v1alpha.Splunk)
		err := validate(direct)
		testutils.AssertNoError(t, direct, err)
	})
	t.Run("required url", func(t *testing.T) {
		direct := validDirect(v1alpha.Splunk)
		direct.Spec.Splunk.URL = ""
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.splunk.url",
			Code: rules.ErrorCodeRequired,
		})
	})
	t.Run("invalid url", func(t *testing.T) {
		direct := validDirect(v1alpha.Splunk)
		direct.Spec.Splunk.URL = "nobl9.com"
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.splunk.url",
			Code: rules.ErrorCodeURL,
		})
	})
	t.Run("url must be https", func(t *testing.T) {
		direct := validDirect(v1alpha.Splunk)
		direct.Spec.Splunk.URL = "http://nobl9.com"
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.splunk.url",
			Code: errorCodeHTTPSSchemeRequired,
		})
	})
}

func TestValidateSpec_Redshift(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		direct := validDirect(v1alpha.Redshift)
		err := validate(direct)
		testutils.AssertNoError(t, direct, err)
	})
	t.Run("required secretARN", func(t *testing.T) {
		direct := validDirect(v1alpha.Redshift)
		direct.Spec.Redshift.SecretARN = ""
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.redshift.secretARN",
			Code: rules.ErrorCodeRequired,
		})
	})
}

func TestValidateSpec_SumoLogic(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		direct := validDirect(v1alpha.SumoLogic)
		err := validate(direct)
		testutils.AssertNoError(t, direct, err)
	})
	t.Run("required url", func(t *testing.T) {
		direct := validDirect(v1alpha.SumoLogic)
		direct.Spec.SumoLogic.URL = ""
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.sumoLogic.url",
			Code: rules.ErrorCodeRequired,
		})
	})
	t.Run("invalid url", func(t *testing.T) {
		direct := validDirect(v1alpha.SumoLogic)
		direct.Spec.SumoLogic.URL = "nobl9.com"
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.sumoLogic.url",
			Code: rules.ErrorCodeURL,
		})
	})
	t.Run("url must be https", func(t *testing.T) {
		direct := validDirect(v1alpha.SumoLogic)
		direct.Spec.SumoLogic.URL = "http://nobl9.com"
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.sumoLogic.url",
			Code: errorCodeHTTPSSchemeRequired,
		})
	})
}

func TestValidateSpec_Instana(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		direct := validDirect(v1alpha.Instana)
		err := validate(direct)
		testutils.AssertNoError(t, direct, err)
	})
	t.Run("required url", func(t *testing.T) {
		direct := validDirect(v1alpha.Instana)
		direct.Spec.Instana.URL = ""
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.instana.url",
			Code: rules.ErrorCodeRequired,
		})
	})
	t.Run("invalid url", func(t *testing.T) {
		direct := validDirect(v1alpha.Instana)
		direct.Spec.Instana.URL = "nobl9.com"
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.instana.url",
			Code: rules.ErrorCodeURL,
		})
	})
	t.Run("url must be https", func(t *testing.T) {
		direct := validDirect(v1alpha.Instana)
		direct.Spec.Instana.URL = "http://nobl9.com"
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.instana.url",
			Code: errorCodeHTTPSSchemeRequired,
		})
	})
}

func TestValidateSpec_InfluxDB(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		direct := validDirect(v1alpha.InfluxDB)
		err := validate(direct)
		testutils.AssertNoError(t, direct, err)
	})
	t.Run("required url", func(t *testing.T) {
		direct := validDirect(v1alpha.InfluxDB)
		direct.Spec.InfluxDB.URL = ""
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.influxdb.url",
			Code: rules.ErrorCodeRequired,
		})
	})
	t.Run("invalid url", func(t *testing.T) {
		direct := validDirect(v1alpha.InfluxDB)
		direct.Spec.InfluxDB.URL = "nobl9.com"
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.influxdb.url",
			Code: rules.ErrorCodeURL,
		})
	})
	t.Run("url must be https", func(t *testing.T) {
		direct := validDirect(v1alpha.InfluxDB)
		direct.Spec.InfluxDB.URL = "http://nobl9.com"
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.influxdb.url",
			Code: errorCodeHTTPSSchemeRequired,
		})
	})
}

func TestValidateSpec_GCM(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		for name, direct := range map[string]Direct{
			"with secrets": validDirect(v1alpha.GCM),
			"empty secrets": func() Direct {
				d := validDirect(v1alpha.GCM)
				d.Spec.GCM.ServiceAccountKey = ""
				return d
			}(),
			"hidden secrets": func() Direct {
				d := validDirect(v1alpha.GCM)
				d.Spec.GCM.ServiceAccountKey = v1alpha.HiddenValue
				return d
			}(),
		} {
			t.Run(name, func(t *testing.T) {
				err := validate(direct)
				testutils.AssertNoError(t, direct, err)
			})
		}
	})
	t.Run("serviceAccountKey must be a valid JSON", func(t *testing.T) {
		direct := validDirect(v1alpha.GCM)
		direct.Spec.GCM.ServiceAccountKey = "{["
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.gcm.serviceAccountKey",
			Code: rules.ErrorCodeStringJSON,
		})
	})
}

func TestValidateSpec_Lightstep(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		direct := validDirect(v1alpha.Lightstep)
		direct.Spec.Lightstep.URL = ""
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
				Code: rules.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.lightstep.project",
				Code: rules.ErrorCodeRequired,
			},
		)
	})
	t.Run("invalid url", func(t *testing.T) {
		direct := validDirect(v1alpha.Lightstep)
		direct.Spec.Lightstep.URL = "h ttp"
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.lightstep.url",
			Code: rules.ErrorCodeURL,
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
				direct := validDirect(v1alpha.Lightstep)
				direct.Spec.Lightstep.URL = test.url
				err := validate(direct)
				testutils.AssertNoError(t, direct, err)
			})
		}
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
			Code: rules.ErrorCodeRequired,
		})
	})
	t.Run("invalid url", func(t *testing.T) {
		direct := validDirect(v1alpha.Dynatrace)
		direct.Spec.Dynatrace.URL = "h ttp"
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.dynatrace.url",
			Code: rules.ErrorCodeURL,
		})
	})
	t.Run("url must be https", func(t *testing.T) {
		direct := validDirect(v1alpha.Dynatrace)
		direct.Spec.Dynatrace.URL = "http://nobl9.com"
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.dynatrace.url",
			Code: errorCodeHTTPSSchemeRequired,
		})
	})
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
			Code: rules.ErrorCodeRequired,
		})
	})
	t.Run("invalid tenantId", func(t *testing.T) {
		direct := validDirect(v1alpha.AzureMonitor)
		direct.Spec.AzureMonitor.TenantID = "invalid"
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.azureMonitor.tenantId",
			Code: rules.ErrorCodeStringUUID,
		})
	})
}

func TestValidateSpec_LogicMonitor(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		direct := validDirect(v1alpha.LogicMonitor)
		err := validate(direct)
		testutils.AssertNoError(t, direct, err)
	})
	t.Run("required account", func(t *testing.T) {
		direct := validDirect(v1alpha.LogicMonitor)
		direct.Spec.LogicMonitor.Account = ""
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.logicMonitor.account",
			Code: rules.ErrorCodeRequired,
		})
	})
}

func TestValidateSpec_AzurePrometheus(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		direct := validDirect(v1alpha.AzurePrometheus)
		err := validate(direct)
		testutils.AssertNoError(t, direct, err)
	})
	t.Run("invalid tenantId", func(t *testing.T) {
		direct := validDirect(v1alpha.AzurePrometheus)
		direct.Spec.AzurePrometheus.TenantID = "invalid"
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.azurePrometheus.tenantId",
			Code: rules.ErrorCodeStringUUID,
		})
	})
	t.Run("required fields", func(t *testing.T) {
		direct := validDirect(v1alpha.AzurePrometheus)
		direct.Spec.AzurePrometheus.URL = ""
		direct.Spec.AzurePrometheus.TenantID = ""
		direct.Spec.AzurePrometheus.ClientID = ""
		direct.Spec.AzurePrometheus.ClientSecret = ""
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 2,
			testutils.ExpectedError{
				Prop: "spec.azurePrometheus.url",
				Code: rules.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.azurePrometheus.tenantId",
				Code: rules.ErrorCodeRequired,
			},
		)
	})
	t.Run("invalid fields", func(t *testing.T) {
		direct := validDirect(v1alpha.AzurePrometheus)
		direct.Spec.AzurePrometheus.URL = "invalid"
		direct.Spec.AzurePrometheus.TenantID = strings.Repeat("l", 256)
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 2,
			testutils.ExpectedError{
				Prop: "spec.azurePrometheus.url",
				Code: rules.ErrorCodeURL,
			},
			testutils.ExpectedError{
				Prop: "spec.azurePrometheus.tenantId",
				Code: rules.ErrorCodeStringUUID,
			},
		)
	})

	t.Run("url must be https", func(t *testing.T) {
		direct := validDirect(v1alpha.AzurePrometheus)
		direct.Spec.AzurePrometheus.URL = "http://nobl9.com"
		err := validate(direct)
		testutils.AssertContainsErrors(t, direct, err, 1, testutils.ExpectedError{
			Prop: "spec.azurePrometheus.url",
			Code: errorCodeHTTPSSchemeRequired,
		})
	})
}

func validDirect(typ v1alpha.DataSourceType) Direct {
	spec := validDirectSpec(typ)
	spec.Description = fmt.Sprintf("Example %s direct", typ)
	spec.ReleaseChannel = v1alpha.ReleaseChannelStable
	return New(Metadata{
		Name:        strings.ToLower(typ.String()),
		DisplayName: typ.String() + " Direct",
		Project:     "default",
	}, spec)
}

func validDirectSpec(typ v1alpha.DataSourceType) Spec {
	specs := map[v1alpha.DataSourceType]Spec{
		v1alpha.Datadog: {
			Datadog: &DatadogConfig{
				Site:           "datadoghq.com",
				APIKey:         "secret",
				ApplicationKey: "secret",
			},
		},
		v1alpha.NewRelic: {
			NewRelic: &NewRelicConfig{
				AccountID:        123,
				InsightsQueryKey: "NRIQ-123",
			},
		},
		v1alpha.AppDynamics: {
			AppDynamics: &AppDynamicsConfig{
				URL:         "https://nobl9.saas.appdynamics.com",
				ClientName:  "client-name",
				AccountName: "account-name",
			},
		},
		v1alpha.Splunk: {
			Splunk: &SplunkConfig{
				URL:         "https://localhost:8089/servicesNS/admin/",
				AccessToken: "secret",
			},
		},
		v1alpha.SplunkObservability: {
			SplunkObservability: &SplunkObservabilityConfig{
				Realm: "us-1",
			},
		},
		v1alpha.ThousandEyes: {
			ThousandEyes: &ThousandEyesConfig{
				OauthBearerToken: "secret",
			},
		},
		v1alpha.BigQuery: {
			BigQuery: &BigQueryConfig{
				ServiceAccountKey: `{"secret": "key"}`,
			},
		},
		v1alpha.CloudWatch: {
			CloudWatch: &CloudWatchConfig{
				RoleARN: "arn:partition:service:region:account-id:resource-id",
			},
		},
		v1alpha.Pingdom: {
			Pingdom: &PingdomConfig{
				APIToken: "secret",
			},
		},
		v1alpha.Redshift: {
			Redshift: &RedshiftConfig{
				SecretARN: "secret",
				RoleARN:   "arn:partition:service:region:account-id:resource-id",
			},
		},
		v1alpha.SumoLogic: {
			SumoLogic: &SumoLogicConfig{
				AccessID:  "secret",
				AccessKey: "secret",
				URL:       "https://sumologic-service.monitoring:443",
			},
		},
		v1alpha.Instana: {
			Instana: &InstanaConfig{
				APIToken: "secret",
				URL:      "https://instana-service.monitoring:443",
			},
		},
		v1alpha.InfluxDB: {
			InfluxDB: &InfluxDBConfig{
				URL:            "https://influxdb-service.monitoring:8086",
				APIToken:       "secret",
				OrganizationID: "secret",
			},
		},
		v1alpha.GCM: {
			GCM: &GCMConfig{
				ServiceAccountKey: `{"secret": "key"}`,
			},
		},
		v1alpha.Lightstep: {
			Lightstep: &LightstepConfig{
				Organization: "LightStep-Play",
				Project:      "play",
				URL:          "https://api.lightstep.com",
			},
		},
		v1alpha.Dynatrace: {
			Dynatrace: &DynatraceConfig{
				URL:            "https://rxh70845.live.dynatrace.com/",
				DynatraceToken: "secret",
			},
		},
		v1alpha.AzureMonitor: {
			AzureMonitor: &AzureMonitorConfig{
				TenantID:     "abf988bf-86f1-41af-91ab-2d7cd011db46",
				ClientID:     "secret",
				ClientSecret: "secret",
			},
		},
		v1alpha.Honeycomb: {
			Honeycomb: &HoneycombConfig{
				APIKey: "secret",
			},
		},
		v1alpha.LogicMonitor: {
			LogicMonitor: &LogicMonitorConfig{
				Account:   "account",
				AccessID:  "secret",
				AccessKey: "secret",
			},
		},
		v1alpha.AzurePrometheus: {
			AzurePrometheus: &AzurePrometheusConfig{
				URL:          "https://prometheus-service.monitoring:8080",
				TenantID:     "abf988bf-86f1-41af-91ab-2d7cd011db46",
				ClientID:     "secret",
				ClientSecret: "secret",
			},
		},
	}

	return specs[typ]
}

func ptr[T any](v T) *T { return &v }
