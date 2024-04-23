package direct

import (
	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

//go:generate go run ../../../scripts/generate-object-impl.go Direct

func New(metadata Metadata, spec Spec) Direct {
	return Direct{
		APIVersion: manifest.VersionV1alpha,
		Kind:       manifest.KindDirect,
		Metadata:   metadata,
		Spec:       spec,
	}
}

// Direct struct which mapped one to one with kind: Direct yaml definition
type Direct struct {
	APIVersion manifest.Version `json:"apiVersion"`
	Kind       manifest.Kind    `json:"kind"`
	Metadata   Metadata         `json:"metadata"`
	Spec       Spec             `json:"spec"`
	Status     *Status          `json:"status,omitempty"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
}

type Metadata struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`
	Project     string `json:"project,omitempty"`
}

// Spec represents content of Spec typical for Direct Object
type Spec struct {
	Description             string                           `json:"description,omitempty"`
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
	LogicMonitor            *LogicMonitorConfig              `json:"logicMonitor,omitempty"`
	HistoricalDataRetrieval *v1alpha.HistoricalDataRetrieval `json:"historicalDataRetrieval,omitempty"`
	QueryDelay              *v1alpha.QueryDelay              `json:"queryDelay,omitempty"`
	// Interval, Timeout and Jitter are readonly and cannot be set via API
	Interval *v1alpha.Interval `json:"interval,omitempty"`
	Timeout  *v1alpha.Timeout  `json:"timeout,omitempty"`
	Jitter   *v1alpha.Jitter   `json:"jitter,omitempty"`
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
	v1alpha.LogicMonitor:        {},
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
	case spec.LogicMonitor != nil:
		return v1alpha.LogicMonitor, nil
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

// SplunkObservabilityConfig represents content of SplunkObservability Configuration typical for Direct Object.
type SplunkObservabilityConfig struct {
	Realm       string `json:"realm"`
	AccessToken string `json:"accessToken"`
}

// ThousandEyesConfig represents content of ThousandEyes Configuration typical for Direct Object.
type ThousandEyesConfig struct {
	OauthBearerToken string `json:"oauthBearerToken"`
}

// BigQueryConfig represents content of BigQuery configuration typical for Direct Object.
type BigQueryConfig struct {
	ServiceAccountKey string `json:"serviceAccountKey"`
}

// SplunkConfig represents content of Splunk Configuration typical for Direct Object.
type SplunkConfig struct {
	URL         string `json:"url"`
	AccessToken string `json:"accessToken"`
}

// CloudWatchConfig represents content of CloudWatch Configuration typical for Direct Object.
type CloudWatchConfig struct {
	// Deprecated: Access Keys are no longer supported. Switch to Cross Account IAM Roles.
	AccessKeyID string `json:"accessKeyID,omitempty"`
	// Deprecated: Access Keys are no longer supported. Switch to Cross Account IAM Roles.
	SecretAccessKey string `json:"secretAccessKey,omitempty"`
	RoleARN         string `json:"roleARN,omitempty"`
}

// PingdomConfig represents content of Pingdom Configuration typical for Direct Object.
type PingdomConfig struct {
	APIToken string `json:"apiToken"`
}

// RedshiftConfig represents content of Redshift configuration typical for Direct Object.
type RedshiftConfig struct {
	// Deprecated: Access Keys are no longer supported. Switch to Cross Account IAM Roles.
	AccessKeyID string `json:"accessKeyID,omitempty"`
	// Deprecated: Access Keys are no longer supported. Switch to Cross Account IAM Roles.
	SecretAccessKey string `json:"secretAccessKey,omitempty"`
	SecretARN       string `json:"secretARN"`
	RoleARN         string `json:"roleARN,omitempty"`
}

// SumoLogicConfig represents content of SumoLogic configuration typical for Direct Object.
type SumoLogicConfig struct {
	AccessID  string `json:"accessID"`
	AccessKey string `json:"accessKey"`
	URL       string `json:"url"`
}

// InstanaConfig represents content of Instana configuration typical for Direct Object.
type InstanaConfig struct {
	APIToken string `json:"apiToken"`
	URL      string `json:"url"`
}

// InfluxDBConfig represents content of InfluxDB configuration typical for Direct Object.
type InfluxDBConfig struct {
	URL            string `json:"url"`
	APIToken       string `json:"apiToken"`
	OrganizationID string `json:"organizationID"`
}

// GCMConfig represents content of GCM configuration typical for Direct Object.
type GCMConfig struct {
	ServiceAccountKey string `json:"serviceAccountKey"`
}

type LightstepConfig struct {
	Organization string `json:"organization"`
	Project      string `json:"project"`
	AppToken     string `json:"appToken"`
	URL          string `json:"url"`
}

// DynatraceConfig represents content of Dynatrace configuration typical for Direct Object.
type DynatraceConfig struct {
	URL            string `json:"url"`
	DynatraceToken string `json:"dynatraceToken"`
}

// AzureMonitorConfig represents content of AzureMonitor Configuration typical for Direct Object.
type AzureMonitorConfig struct {
	TenantID     string `json:"tenantId"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

// HoneycombConfig represents content of Honeycomb Configuration typical for Direct Object.
type HoneycombConfig struct {
	APIKey string `json:"apiKey"`
}

// LogicMonitorConfig represents content of LogicMonitor Configuration typical for Direct Object.
type LogicMonitorConfig struct {
	Account   string `json:"account"`
	AccessID  string `json:"accessId"`
	AccessKey string `json:"accessKey"`
}
