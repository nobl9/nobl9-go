package v1alphaExamples

import (
	"fmt"
	"slices"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaDirect "github.com/nobl9/nobl9-go/manifest/v1alpha/direct"
	"github.com/nobl9/nobl9-go/sdk"
)

type directExample struct {
	standardExample
	typ v1alpha.DataSourceType
}

func (d directExample) GetDataSourceType() v1alpha.DataSourceType {
	return d.typ
}

func Direct() []Example {
	types := v1alpha.DataSourceTypeValues()
	examples := make([]Example, 0, len(types))
	for _, typ := range types {
		if !v1alphaDirect.IsValidDirectType(typ) {
			continue
		}
		example := directExample{
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

var betaChannelDirects = []v1alpha.DataSourceType{
	v1alpha.AzureMonitor,
	v1alpha.Honeycomb,
	v1alpha.LogicMonitor,
	v1alpha.GoogleCloudMonitoring,
	v1alpha.AzurePrometheus,
}

func (d directExample) Generate() v1alphaDirect.Direct {
	titleName := dataSourceTypePrettyName(d.typ)
	direct := v1alphaDirect.New(
		v1alphaDirect.Metadata{
			Name:        d.Variant,
			DisplayName: titleName + " Direct",
			Project:     sdk.DefaultProject,
		},
		v1alphaDirect.Spec{
			Description:    fmt.Sprintf("Example %s Direct", titleName),
			ReleaseChannel: v1alpha.ReleaseChannelStable,
		},
	)
	direct = d.generateVariant(direct)
	typ, _ := direct.Spec.GetType()
	if maxDuration, err := v1alpha.GetDataRetrievalMaxDuration(manifest.KindDirect, typ); err == nil {
		direct.Spec.HistoricalDataRetrieval = &v1alpha.HistoricalDataRetrieval{
			MaxDuration: maxDuration,
			DefaultDuration: v1alpha.HistoricalRetrievalDuration{
				Value: ptr(*maxDuration.Value / 2),
				Unit:  maxDuration.Unit,
			},
		}
	}
	defaultQueryDelay := v1alpha.GetQueryDelayDefaults()[typ]
	direct.Spec.QueryDelay = &v1alpha.QueryDelay{
		Duration: v1alpha.Duration{
			Value: ptr(*defaultQueryDelay.Value + 1),
			Unit:  defaultQueryDelay.Unit,
		},
	}
	if slices.Contains(betaChannelDirects, typ) {
		direct.Spec.ReleaseChannel = v1alpha.ReleaseChannelBeta
	} else {
		direct.Spec.ReleaseChannel = v1alpha.ReleaseChannelStable
	}
	return direct
}

func (d directExample) generateVariant(direct v1alphaDirect.Direct) v1alphaDirect.Direct {
	switch d.typ {
	case v1alpha.AppDynamics:
		direct.Spec.AppDynamics = &v1alphaDirect.AppDynamicsConfig{
			URL:          "https://my-org.saas.appdynamics.com",
			ClientName:   "prod-direct",
			AccountName:  "my-account",
			ClientSecret: "[secret]",
		}
	case v1alpha.AzureMonitor:
		direct.Spec.ReleaseChannel = v1alpha.ReleaseChannelBeta
		direct.Spec.AzureMonitor = &v1alphaDirect.AzureMonitorConfig{
			TenantID:     "5cdecca3-c2c5-4072-89dd-5555faf05202",
			ClientID:     "70747025-9367-41a5-98f1-59b18b5793c3",
			ClientSecret: "[secret]",
		}
	case v1alpha.BigQuery:
		direct.Spec.BigQuery = &v1alphaDirect.BigQueryConfig{
			ServiceAccountKey: gcloudServiceAccountKey,
		}
	case v1alpha.CloudWatch:
		direct.Spec.ReleaseChannel = v1alpha.ReleaseChannelBeta
		direct.Spec.CloudWatch = &v1alphaDirect.CloudWatchConfig{
			RoleARN: "arn:aws:iam::123456578901:role/awsCrossAccountProdCloudwatch-prod-app",
		}
	case v1alpha.Datadog:
		direct.Spec.Datadog = &v1alphaDirect.DatadogConfig{
			Site:           "com",
			APIKey:         "[secret]",
			ApplicationKey: "[secret]",
		}
	case v1alpha.Dynatrace:
		direct.Spec.Dynatrace = &v1alphaDirect.DynatraceConfig{
			URL:            "https://zvf10945.live.dynatrace.com/",
			DynatraceToken: "[secret]",
		}
	case v1alpha.GCM:
		direct.Spec.GCM = &v1alphaDirect.GCMConfig{
			ServiceAccountKey: gcloudServiceAccountKey,
		}
	case v1alpha.Honeycomb:
		direct.Spec.ReleaseChannel = v1alpha.ReleaseChannelBeta
		direct.Spec.Honeycomb = &v1alphaDirect.HoneycombConfig{
			APIKey: "[secret]",
		}
	case v1alpha.InfluxDB:
		direct.Spec.InfluxDB = &v1alphaDirect.InfluxDBConfig{
			URL:            "https://us-west-2-2.aws.cloud2.influxdata.com",
			APIToken:       "[secret]",
			OrganizationID: "my-org",
		}
	case v1alpha.Instana:
		direct.Spec.Instana = &v1alphaDirect.InstanaConfig{
			APIToken: "[secret]",
			URL:      "https://orange-my-org12.instana.io",
		}
	case v1alpha.Lightstep:
		direct.Spec.Lightstep = &v1alphaDirect.LightstepConfig{
			Organization: "MyOrg",
			Project:      "prod-app",
			AppToken:     "[secret]",
			URL:          "https://api.lightstep.com",
		}
	case v1alpha.LogicMonitor:
		direct.Spec.LogicMonitor = &v1alphaDirect.LogicMonitorConfig{
			Account:   "my-account-name",
			AccessID:  "9xA2BssShK21ld9LoOYu",
			AccessKey: "[secret]",
		}
	case v1alpha.NewRelic:
		direct.Spec.NewRelic = &v1alphaDirect.NewRelicConfig{
			AccountID:        1234567,
			InsightsQueryKey: "NRIQ-2f66237213814496669180ba",
		}
	case v1alpha.Pingdom:
		direct.Spec.Pingdom = &v1alphaDirect.PingdomConfig{
			APIToken: "[secret]",
		}
	case v1alpha.Redshift:
		direct.Spec.Redshift = &v1alphaDirect.RedshiftConfig{
			SecretARN: "arn:aws:secretsmanager:eu-central-1:123456578901:secret:prod-redshift-db-user",
			RoleARN:   "arn:aws:iam::123456578901:role/awsCrossAccountProdRedshift-prod-app",
		}
	case v1alpha.Splunk:
		direct.Spec.Splunk = &v1alphaDirect.SplunkConfig{
			URL:         "https://splunk.my-org.com/services",
			AccessToken: "[secret]",
		}
	case v1alpha.SplunkObservability:
		direct.Spec.SplunkObservability = &v1alphaDirect.SplunkObservabilityConfig{
			Realm:       "us1",
			AccessToken: "[secret]",
		}
	case v1alpha.SumoLogic:
		direct.Spec.SumoLogic = &v1alphaDirect.SumoLogicConfig{
			AccessID:  "wzeulXAULylic8",
			AccessKey: "[secret]",
			URL:       "https://service.sumologic.com",
		}
	case v1alpha.ThousandEyes:
		direct.Spec.ThousandEyes = &v1alphaDirect.ThousandEyesConfig{
			OauthBearerToken: "[secret]",
		}
	case v1alpha.AzurePrometheus:
		direct.Spec.AzurePrometheus = &v1alphaDirect.AzurePrometheusConfig{
			URL:          "https://prod-app.azuremonitor.com",
			TenantID:     "5cdecca3-c2c5-4072-89dd-5555faf05202",
			ClientID:     "70747025-9367-41a5-98f1-59b18b5793c3",
			ClientSecret: "[secret]",
		}
	default:
		panic(fmt.Sprintf("unexpected v1alpha.DataSourceType: %#v", d.typ))
	}
	return direct
}

// #nosec G101
const gcloudServiceAccountKey = `{
  "type": "service_account",
  "project_id": "prod-app",
  "private_key_id": "669180ba44964eddba9e2f6623721381",
  "private_key": "-----BEGIN PRIVATE KEY-----\nSECRET_KEY_GOES_HERE\n-----END PRIVATE KEY-----\n",
  "client_email": "nobl9@nobl9.iam.gserviceaccount.com",
  "client_id": "eddba9e2f66237213812",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/nobl9%40nobl9.iam.gserviceaccount.com"
}`
