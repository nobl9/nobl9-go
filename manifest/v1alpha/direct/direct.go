package direct

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

//go:generate go run ../../scripts/generate-object-impl.go Direct,PublicDirect

func New(metadata Metadata, spec Spec) Direct {
	return Direct{
		APIVersion: v1alpha.APIVersion,
		Kind:       manifest.KindDirect,
		Metadata:   metadata,
		Spec:       spec,
	}
}

// Direct struct which mapped one to one with kind: Direct yaml definition
type Direct struct {
	APIVersion string        `json:"apiVersion"`
	Kind       manifest.Kind `json:"kind"`
	Metadata   Metadata      `json:"metadata"`
	Spec       Spec          `json:"spec"`
	Status     *Status       `json:"status,omitempty"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
}

type Metadata struct {
	Name        string         `json:"name" validate:"required,objectName"`
	DisplayName string         `json:"displayName,omitempty" validate:"omitempty,min=0,max=63"`
	Project     string         `json:"project,omitempty" validate:"objectName"`
	Labels      v1alpha.Labels `json:"labels,omitempty" validate:"omitempty,labels"`
}

// Spec represents content of Spec typical for Direct Object
type Spec struct {
	Description             string                           `json:"description,omitempty" validate:"description"`
	ReleaseChannel          v1alpha.ReleaseChannel           `json:"releaseChannel,omitempty"`
	LogCollectionEnabled    *bool                            `json:"logCollectionEnabled,omitempty"`
	Datadog                 *DatadogConfig                   `json:"datadog,omitempty"`
	NewRelic                *NewRelicConfig                  `json:"newRelic,omitempty"`
	AppDynamics             *AppDynamicsConfig               `json:"appDynamics,omitempty"`
	SplunkObservability     *SplunkObservabilityConfig       `json:"splunkObservability,omitempty"`
	ThousandEyes            *ThousandEyesConfig              `json:"thousandEyes,omitempty"`
	BigQuery                *BigQueryConfig                  `json:"bigQuery,omitempty"`
	Splunk                  *SplunkConfig                    `json:"splunk,omitempty"`
	CloudWatch              *CloudWatchConfig                `json:"cloudWatch,omitempty"`
	Pingdom                 *PingdomConfig                   `json:"pingdom,omitempty"`
	Redshift                *RedshiftConfig                  `json:"redshift,omitempty"`
	SumoLogic               *SumoLogicConfig                 `json:"sumoLogic,omitempty"`
	Instana                 *InstanaConfig                   `json:"instana,omitempty"`
	InfluxDB                *InfluxDBConfig                  `json:"influxdb,omitempty"`
	GCM                     *GCMConfig                       `json:"gcm,omitempty"`
	Lightstep               *LightstepConfig                 `json:"lightstep,omitempty"`
	Dynatrace               *DynatraceConfig                 `json:"dynatrace,omitempty"`
	AzureMonitor            *AzureMonitorConfig              `json:"azureMonitor,omitempty"`
	Honeycomb               *HoneycombConfig                 `json:"honeycomb,omitempty"`
	HistoricalDataRetrieval *v1alpha.HistoricalDataRetrieval `json:"historicalDataRetrieval,omitempty"`
	QueryDelay              *v1alpha.QueryDelay              `json:"queryDelay,omitempty"`
}

// Status represents content of Status optional for Direct Object
type Status struct {
	DirectType string `json:"directType"`
}

var validDirectTypes = map[v1alpha.DataSourceType]struct{}{
	v1alpha.Datadog:             {},
	v1alpha.NewRelic:            {},
	v1alpha.SplunkObservability: {},
	v1alpha.AppDynamics:         {},
	v1alpha.ThousandEyes:        {},
	v1alpha.BigQuery:            {},
	v1alpha.Splunk:              {},
	v1alpha.CloudWatch:          {},
	v1alpha.Pingdom:             {},
	v1alpha.Redshift:            {},
	v1alpha.SumoLogic:           {},
	v1alpha.Instana:             {},
	v1alpha.InfluxDB:            {},
	v1alpha.GCM:                 {},
	v1alpha.Lightstep:           {},
	v1alpha.Dynatrace:           {},
	v1alpha.AzureMonitor:        {},
	v1alpha.Honeycomb:           {},
}

func IsValidDirectType(directType v1alpha.DataSourceType) bool {
	_, isValid := validDirectTypes[directType]
	return isValid
}

func (spec Spec) GetType() (v1alpha.DataSourceType, error) {
	switch {
	case spec.Datadog != nil:
		return v1alpha.Datadog, nil
	case spec.NewRelic != nil:
		return v1alpha.NewRelic, nil
	case spec.SplunkObservability != nil:
		return v1alpha.SplunkObservability, nil
	case spec.AppDynamics != nil:
		return v1alpha.AppDynamics, nil
	case spec.ThousandEyes != nil:
		return v1alpha.ThousandEyes, nil
	case spec.BigQuery != nil:
		return v1alpha.BigQuery, nil
	case spec.Splunk != nil:
		return v1alpha.Splunk, nil
	case spec.CloudWatch != nil:
		return v1alpha.CloudWatch, nil
	case spec.Pingdom != nil:
		return v1alpha.Pingdom, nil
	case spec.Redshift != nil:
		return v1alpha.Redshift, nil
	case spec.SumoLogic != nil:
		return v1alpha.SumoLogic, nil
	case spec.Instana != nil:
		return v1alpha.Instana, nil
	case spec.InfluxDB != nil:
		return v1alpha.InfluxDB, nil
	case spec.GCM != nil:
		return v1alpha.GCM, nil
	case spec.Lightstep != nil:
		return v1alpha.Lightstep, nil
	case spec.Dynatrace != nil:
		return v1alpha.Dynatrace, nil
	case spec.AzureMonitor != nil:
		return v1alpha.AzureMonitor, nil
	case spec.Honeycomb != nil:
		return v1alpha.Honeycomb, nil
	}
	return 0, errors.New("BUG: unknown direct type")
}

// DatadogConfig represents content of Datadog Configuration typical for Direct Object.
type DatadogConfig struct {
	Site           string `json:"site"`
	APIKey         string `json:"apiKey"`
	ApplicationKey string `json:"applicationKey"`
}

// NewRelicConfig represents content of NewRelic Configuration typical for Direct Object.
type NewRelicConfig struct {
	AccountID        int    `json:"accountId"`
	InsightsQueryKey string `json:"insightsQueryKey"`
}

// AppDynamicsConfig represents content of AppDynamics Configuration typical for Direct Object.
type AppDynamicsConfig struct {
	URL          string `json:"url"`
	ClientID     string `json:"clientID"`
	ClientName   string `json:"clientName"`
	AccountName  string `json:"accountName"`
	ClientSecret string `json:"clientSecret"`
}

// GenerateMissingFields - this function is responsible for generating ClientID from AccountName and ClientName
// when provided with new, also it generates AccountName and ClientName for old already existing configs.
// TODO: what to do with this?
func (a *AppDynamicsConfig) GenerateMissingFields() {
	if a.AccountName != "" && a.ClientName != "" {
		a.ClientID = fmt.Sprintf("%s@%s", a.ClientName, a.AccountName)
	} else if a.ClientID != "" {
		at := strings.LastIndex(a.ClientID, "@")
		if at >= 0 {
			a.ClientName, a.AccountName = a.ClientID[:at], a.ClientID[at+1:]
		}
	}
}

// SplunkObservabilityConfig represents content of SplunkObservability Configuration typical for Direct Object.
type SplunkObservabilityConfig struct {
	Realm       string `json:"realm"`
	AccessToken string `json:"accessToken"`
}

type LightstepConfig struct {
	Organization string `json:"organization,omitempty" validate:"required" example:"LightStep-Play"`
	Project      string `json:"project,omitempty" validate:"required" example:"play"`
	AppToken     string `json:"appToken"`
}

// SplunkConfig represents content of Splunk Configuration typical for Direct Object.
type SplunkConfig struct {
	URL         string `json:"url,omitempty" validate:"httpsURL" example:"https://api.eu0.signalfx.com"`
	AccessToken string `json:"accessToken,omitempty"`
}

// ThousandEyesConfig represents content of ThousandEyes Configuration typical for Direct Object.
type ThousandEyesConfig struct {
	OauthBearerToken string `json:"oauthBearerToken,omitempty"`
}

// BigQueryConfig represents content of BigQuery configuration typical for Direct Object.
type BigQueryConfig struct {
	ServiceAccountKey string `json:"serviceAccountKey,omitempty"`
}

// GCMConfig represents content of GCM configuration typical for Direct Object.
type GCMConfig struct {
	ServiceAccountKey string `json:"serviceAccountKey,omitempty"`
}

// DynatraceConfig represents content of Dynatrace configuration typical for Direct Object.
type DynatraceConfig struct {
	URL            string `json:"url,omitempty" validate:"required,url,httpsURL" example:"https://{your-environment-id}.live.dynatrace.com or https://{your-domain}/e/{your-environment-id}"` //nolint: lll
	DynatraceToken string `json:"dynatraceToken,omitempty"`
}

// CloudWatchConfig represents content of CloudWatch Configuration typical for Direct Object.
type CloudWatchConfig struct {
	AccessKeyID     string `json:"accessKeyID,omitempty"`
	SecretAccessKey string `json:"secretAccessKey,omitempty"`
	RoleARN         string `json:"roleARN,omitempty" example:"arn:aws:iam::123456789012:role/SomeAccessRole"` //nolint: lll
}

// PingdomConfig represents content of Pingdom Configuration typical for Direct Object.
type PingdomConfig struct {
	APIToken string `json:"apiToken"`
}

// InstanaConfig represents content of Instana configuration typical for Direct Object.
type InstanaConfig struct {
	APIToken string `json:"apiToken"`
	URL      string `json:"url" validate:"required,url,httpsURL"`
}

// InfluxDBConfig represents content of InfluxDB configuration typical for Direct Object.
type InfluxDBConfig struct {
	URL            string `json:"url" validate:"required,url"`
	APIToken       string `json:"apiToken"`
	OrganizationID string `json:"organizationID"`
}

// RedshiftConfig represents content of Redshift configuration typical for Direct Object.
type RedshiftConfig struct {
	AccessKeyID     string `json:"accessKeyID,omitempty"`
	SecretAccessKey string `json:"secretAccessKey,omitempty"`
	SecretARN       string `json:"secretARN"`
	RoleARN         string `json:"roleARN,omitempty" example:"arn:aws:iam::123456789012:role/SomeAccessRole"` //nolint: lll
}

// SumoLogicConfig represents content of SumoLogic configuration typical for Direct Object.
type SumoLogicConfig struct {
	AccessID  string `json:"accessID"`
	AccessKey string `json:"accessKey"`
	URL       string `json:"url" validate:"required,url"`
}

// AzureMonitorConfig represents content of AzureMonitor Configuration typical for Direct Object.
type AzureMonitorConfig struct {
	TenantID     string `json:"tenantId" validate:"required,uuid_rfc4122" example:"abf988bf-86f1-41af-91ab-2d7cd011db46"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

// HoneycombConfig represents content of Honeycomb Configuration typical for Direct Object.
type HoneycombConfig struct {
	APIKey string `json:"apiKey,omitempty" example:"lwPoPt20Gmdi4dwTdW9dTR"`
}
