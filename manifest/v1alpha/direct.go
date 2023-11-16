package v1alpha

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/manifest"
)

//go:generate go run ../../scripts/generate-object-impl.go Direct,PublicDirect

// Direct struct which mapped one to one with kind: Direct yaml definition
type Direct struct {
	APIVersion string         `json:"apiVersion"`
	Kind       manifest.Kind  `json:"kind"`
	Metadata   DirectMetadata `json:"metadata"`
	Spec       DirectSpec     `json:"spec"`
	Status     *DirectStatus  `json:"status,omitempty"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
}

// PublicDirect struct which mapped one to one with kind: Direct yaml definition without secrets
type PublicDirect struct {
	APIVersion string           `json:"apiVersion"`
	Kind       manifest.Kind    `json:"kind"`
	Metadata   DirectMetadata   `json:"metadata"`
	Spec       PublicDirectSpec `json:"spec"`
	Status     *DirectStatus    `json:"status,omitempty"`

	ManifestSource string `json:"manifestSrc,omitempty"`
}

type DirectMetadata struct {
	Name        string `json:"name" validate:"required,objectName"`
	DisplayName string `json:"displayName,omitempty" validate:"omitempty,min=0,max=63"`
	Project     string `json:"project,omitempty" validate:"objectName"`
	Labels      Labels `json:"labels,omitempty" validate:"omitempty,labels"`
}

// DirectSpec represents content of Spec typical for Direct Object
type DirectSpec struct {
	Description             string                           `json:"description,omitempty" validate:"description" example:"Datadog description"` //nolint:lll
	SourceOf                []string                         `json:"sourceOf" example:"Metrics,Services"`
	ReleaseChannel          ReleaseChannel                   `json:"releaseChannel,omitempty" example:"beta,stable"`
	Datadog                 *DatadogDirectConfig             `json:"datadog,omitempty"`
	LogCollectionEnabled    *bool                            `json:"logCollectionEnabled,omitempty"`
	NewRelic                *NewRelicDirectConfig            `json:"newRelic,omitempty"`
	AppDynamics             *AppDynamicsDirectConfig         `json:"appDynamics,omitempty"`
	SplunkObservability     *SplunkObservabilityDirectConfig `json:"splunkObservability,omitempty"`
	ThousandEyes            *ThousandEyesDirectConfig        `json:"thousandEyes,omitempty"`
	BigQuery                *BigQueryDirectConfig            `json:"bigQuery,omitempty"`
	Splunk                  *SplunkDirectConfig              `json:"splunk,omitempty"`
	CloudWatch              *CloudWatchDirectConfig          `json:"cloudWatch,omitempty"`
	Pingdom                 *PingdomDirectConfig             `json:"pingdom,omitempty"`
	Redshift                *RedshiftDirectConfig            `json:"redshift,omitempty"`
	SumoLogic               *SumoLogicDirectConfig           `json:"sumoLogic,omitempty"`
	Instana                 *InstanaDirectConfig             `json:"instana,omitempty"`
	InfluxDB                *InfluxDBDirectConfig            `json:"influxdb,omitempty"`
	GCM                     *GCMDirectConfig                 `json:"gcm,omitempty"`
	Lightstep               *LightstepDirectConfig           `json:"lightstep,omitempty"`
	Dynatrace               *DynatraceDirectConfig           `json:"dynatrace,omitempty"`
	AzureMonitor            *AzureMonitorDirectConfig        `json:"azureMonitor,omitempty"`
	Honeycomb               *HoneycombDirectConfig           `json:"honeycomb,omitempty"`
	HistoricalDataRetrieval *HistoricalDataRetrieval         `json:"historicalDataRetrieval,omitempty"`
	QueryDelay              *QueryDelay                      `json:"queryDelay,omitempty"`
}

var allDirectTypes = map[string]struct{}{
	Datadog.String():             {},
	NewRelic.String():            {},
	SplunkObservability.String(): {},
	AppDynamics.String():         {},
	ThousandEyes.String():        {},
	BigQuery.String():            {},
	Splunk.String():              {},
	CloudWatch.String():          {},
	Pingdom.String():             {},
	Redshift.String():            {},
	SumoLogic.String():           {},
	Instana.String():             {},
	InfluxDB.String():            {},
	GCM.String():                 {},
	Lightstep.String():           {},
	Dynatrace.String():           {},
	AzureMonitor.String():        {},
	Honeycomb.String():           {},
}

func IsValidDirectType(directType string) bool {
	_, isValid := allDirectTypes[directType]
	return isValid
}

func (spec DirectSpec) GetType() (string, error) {
	switch {
	case spec.Datadog != nil:
		return Datadog.String(), nil
	case spec.NewRelic != nil:
		return NewRelic.String(), nil
	case spec.SplunkObservability != nil:
		return SplunkObservability.String(), nil
	case spec.AppDynamics != nil:
		return AppDynamics.String(), nil
	case spec.ThousandEyes != nil:
		return ThousandEyes.String(), nil
	case spec.BigQuery != nil:
		return BigQuery.String(), nil
	case spec.Splunk != nil:
		return Splunk.String(), nil
	case spec.CloudWatch != nil:
		return CloudWatch.String(), nil
	case spec.Pingdom != nil:
		return Pingdom.String(), nil
	case spec.Redshift != nil:
		return Redshift.String(), nil
	case spec.SumoLogic != nil:
		return SumoLogic.String(), nil
	case spec.Instana != nil:
		return Instana.String(), nil
	case spec.InfluxDB != nil:
		return InfluxDB.String(), nil
	case spec.GCM != nil:
		return GCM.String(), nil
	case spec.Lightstep != nil:
		return Lightstep.String(), nil
	case spec.Dynatrace != nil:
		return Dynatrace.String(), nil
	case spec.AzureMonitor != nil:
		return AzureMonitor.String(), nil
	case spec.Honeycomb != nil:
		return Honeycomb.String(), nil
	}
	return "", errors.New("unknown direct type")
}

// PublicDirectSpec represents content of Spec typical for Direct Object without secrets
type PublicDirectSpec struct {
	Description             string                                 `json:"description,omitempty" validate:"description" example:"Datadog description"` //nolint:lll
	SourceOf                []string                               `json:"sourceOf" example:"Metrics,Services"`
	ReleaseChannel          string                                 `json:"releaseChannel,omitempty" example:"beta,stable"`
	LogCollectionEnabled    bool                                   `json:"logCollectionEnabled,omitempty"`
	Datadog                 *PublicDatadogDirectConfig             `json:"datadog,omitempty"`
	NewRelic                *PublicNewRelicDirectConfig            `json:"newRelic,omitempty"`
	SplunkObservability     *PublicSplunkObservabilityDirectConfig `json:"splunkObservability,omitempty"`
	AppDynamics             *PublicAppDynamicsDirectConfig         `json:"appDynamics,omitempty"`
	ThousandEyes            *PublicThousandEyesDirectConfig        `json:"thousandEyes,omitempty"`
	BigQuery                *PublicBigQueryDirectConfig            `json:"bigQuery,omitempty"`
	Splunk                  *PublicSplunkDirectConfig              `json:"splunk,omitempty"`
	CloudWatch              *PublicCloudWatchDirectConfig          `json:"cloudWatch,omitempty"`
	Pingdom                 *PublicPingdomDirectConfig             `json:"pingdom,omitempty"`
	Redshift                *PublicRedshiftDirectConfig            `json:"redshift,omitempty"`
	SumoLogic               *PublicSumoLogicDirectConfig           `json:"sumoLogic,omitempty"`
	Instana                 *PublicInstanaDirectConfig             `json:"instana,omitempty"`
	InfluxDB                *PublicInfluxDBDirectConfig            `json:"influxdb,omitempty"`
	GCM                     *PublicGCMDirectConfig                 `json:"gcm,omitempty"`
	Lightstep               *PublicLightstepDirectConfig           `json:"lightstep,omitempty"`
	Dynatrace               *PublicDynatraceDirectConfig           `json:"dynatrace,omitempty"`
	AzureMonitor            *PublicAzureMonitorDirectConfig        `json:"azureMonitor,omitempty"`
	Honeycomb               *PublicHoneycombDirectConfig           `json:"honeycomb,omitempty"`
	HistoricalDataRetrieval *HistoricalDataRetrieval               `json:"historicalDataRetrieval,omitempty"`
	QueryDelay              *QueryDelay                            `json:"queryDelay,omitempty"`
}

// DirectStatus represents content of Status optional for Direct Object
type DirectStatus struct {
	DirectType string `json:"directType" example:"Datadog"`
}

// DatadogDirectConfig represents content of Datadog Configuration typical for Direct Object.
type DatadogDirectConfig struct {
	Site           string `json:"site,omitempty" validate:"site" example:"eu,us3.datadoghq.com"`
	APIKey         string `json:"apiKey" example:"secret"`
	ApplicationKey string `json:"applicationKey" example:"secret"`
}

// PublicDatadogDirectConfig represents content of Datadog Configuration typical for Direct Object without secrets.
type PublicDatadogDirectConfig struct {
	Site                 string `json:"site,omitempty" example:"eu,us3.datadoghq.com"`
	HiddenAPIKey         string `json:"apiKey" example:"[hidden]"`
	HiddenApplicationKey string `json:"applicationKey" example:"[hidden]"`
}

// NewRelicDirectConfig represents content of NewRelic Configuration typical for Direct Object.
type NewRelicDirectConfig struct {
	AccountID        int    `json:"accountId" validate:"required" example:"123654"`
	InsightsQueryKey string `json:"insightsQueryKey" validate:"newRelicApiKey" example:"secret"`
}

// PublicNewRelicDirectConfig represents content of NewRelic Configuration typical for Direct Object without secrets.
type PublicNewRelicDirectConfig struct {
	AccountID              int    `json:"accountId,omitempty" example:"123654"`
	HiddenInsightsQueryKey string `json:"insightsQueryKey" example:"[hidden]"`
}

// PublicAppDynamicsDirectConfig represents public content of AppDynamics Configuration
// typical for Direct Object without secrets.
type PublicAppDynamicsDirectConfig struct {
	URL                string `json:"url,omitempty" example:"https://nobl9.saas.appdynamics.com"`
	ClientID           string `json:"clientID,omitempty" example:"apiClientID@accountID"`
	ClientName         string `json:"clientName,omitempty" example:"apiClientID"`
	AccountName        string `json:"accountName,omitempty" example:"accountID"`
	HiddenClientSecret string `json:"clientSecret,omitempty" example:"[hidden]"`
}

// GenerateMissingFields checks if there is no ClientName and AccountName
//
//	then separates ClientID into ClientName and AccountName.
func (a *PublicAppDynamicsDirectConfig) GenerateMissingFields() {
	if a.ClientName == "" && a.AccountName == "" {
		at := strings.LastIndex(a.ClientID, "@")
		if at >= 0 {
			a.ClientName, a.AccountName = a.ClientID[:at], a.ClientID[at+1:]
		}
	}
}

// AppDynamicsDirectConfig represents content of AppDynamics Configuration typical for Direct Object.
type AppDynamicsDirectConfig struct {
	URL          string `json:"url,omitempty" validate:"httpsURL" example:"https://nobl9.saas.appdynamics.com"`
	ClientID     string `json:"clientID,omitempty" example:"apiClientID@accountID"`
	ClientName   string `json:"clientName,omitempty" example:"apiClientID"`
	AccountName  string `json:"accountName,omitempty" example:"accountID"`
	ClientSecret string `json:"clientSecret,omitempty" example:"secret"`
}

// GenerateMissingFields - this function is responsible for generating ClientID from AccountName and ClientName
// when provided with new, also it generates AccountName and ClientName for old already existing configs.
func (a *AppDynamicsDirectConfig) GenerateMissingFields() {
	if a.AccountName != "" && a.ClientName != "" {
		a.ClientID = fmt.Sprintf("%s@%s", a.ClientName, a.AccountName)
	} else if a.ClientID != "" {
		at := strings.LastIndex(a.ClientID, "@")
		if at >= 0 {
			a.ClientName, a.AccountName = a.ClientID[:at], a.ClientID[at+1:]
		}
	}
}

// SplunkDirectConfig represents content of Splunk Configuration typical for Direct Object.
type SplunkDirectConfig struct {
	URL         string `json:"url,omitempty" validate:"httpsURL" example:"https://api.eu0.signalfx.com"`
	AccessToken string `json:"accessToken,omitempty"`
}

// PublicSplunkDirectConfig represents content of Splunk Configuration typical for Direct Object.
type PublicSplunkDirectConfig struct {
	URL               string `json:"url,omitempty" example:"https://api.eu0.signalfx.com"`
	HiddenAccessToken string `json:"accessToken,omitempty"`
}

type LightstepDirectConfig struct {
	Organization string `json:"organization,omitempty" validate:"required" example:"LightStep-Play"`
	Project      string `json:"project,omitempty" validate:"required" example:"play"`
	AppToken     string `json:"appToken"`
}

type PublicLightstepDirectConfig struct {
	Organization   string `json:"organization,omitempty" example:"LightStep-Play"`
	Project        string `json:"project,omitempty" example:"play"`
	HiddenAppToken string `json:"appToken"`
}

// SplunkObservabilityDirectConfig represents content of SplunkObservability Configuration typical for Direct Object.
type SplunkObservabilityDirectConfig struct {
	Realm       string `json:"realm,omitempty" validate:"required" example:"us1"`
	AccessToken string `json:"accessToken,omitempty"`
}

// PublicSplunkObservabilityDirectConfig represents content of SplunkObservability
// Configuration typical for Direct Object.
type PublicSplunkObservabilityDirectConfig struct {
	Realm             string `json:"realm,omitempty" example:"us1"`
	HiddenAccessToken string `json:"accessToken,omitempty"`
}

// ThousandEyesDirectConfig represents content of ThousandEyes Configuration typical for Direct Object.
type ThousandEyesDirectConfig struct {
	OauthBearerToken string `json:"oauthBearerToken,omitempty"`
}

// PublicThousandEyesDirectConfig content of ThousandEyes
// Configuration typical for Direct Object
type PublicThousandEyesDirectConfig struct {
	HiddenOauthBearerToken string `json:"oauthBearerToken,omitempty"`
}

// BigQueryDirectConfig represents content of BigQuery configuration typical for Direct Object.
type BigQueryDirectConfig struct {
	ServiceAccountKey string `json:"serviceAccountKey,omitempty"`
}

// PublicBigQueryDirectConfig represents content of BigQuery configuration typical for Direct Object without secrets.
type PublicBigQueryDirectConfig struct {
	HiddenServiceAccountKey string `json:"serviceAccountKey,omitempty"`
}

// GCMDirectConfig represents content of GCM configuration typical for Direct Object.
type GCMDirectConfig struct {
	ServiceAccountKey string `json:"serviceAccountKey,omitempty"`
}

// PublicGCMDirectConfig represents content of GCM configuration typical for Direct Object without secrets.
type PublicGCMDirectConfig struct {
	HiddenServiceAccountKey string `json:"serviceAccountKey,omitempty"`
}

// DynatraceDirectConfig represents content of Dynatrace configuration typical for Direct Object.
type DynatraceDirectConfig struct {
	URL            string `json:"url,omitempty" validate:"required,url,httpsURL" example:"https://{your-environment-id}.live.dynatrace.com or https://{your-domain}/e/{your-environment-id}"` //nolint: lll
	DynatraceToken string `json:"dynatraceToken,omitempty"`
}

// PublicDynatraceDirectConfig represents content of Dynatrace configuration typical for Direct Object without secrets.
type PublicDynatraceDirectConfig struct {
	URL                  string `json:"url,omitempty" validate:"required,url,httpsURL" example:"https://{your-environment-id}.live.dynatrace.com or https://{your-domain}/e/{your-environment-id}"` //nolint: lll
	HiddenDynatraceToken string `json:"dynatraceToken,omitempty"`
}

// CloudWatchDirectConfig represents content of CloudWatch Configuration typical for Direct Object.
type CloudWatchDirectConfig struct {
	AccessKeyID     string `json:"accessKeyID,omitempty"`
	SecretAccessKey string `json:"secretAccessKey,omitempty"`
	RoleARN         string `json:"roleARN,omitempty" example:"arn:aws:iam::123456789012:role/SomeAccessRole"` //nolint: lll
}

// PublicCloudWatchDirectConfig represents content of CloudWatch Configuration typical for Direct Object
// without secrets.
type PublicCloudWatchDirectConfig struct {
	HiddenAccessKeyID     string `json:"accessKeyID,omitempty"`
	HiddenSecretAccessKey string `json:"secretAccessKey,omitempty"`
	HiddenRoleARN         string `json:"roleARN,omitempty"`
}

// PingdomDirectConfig represents content of Pingdom Configuration typical for Direct Object.
type PingdomDirectConfig struct {
	APIToken string `json:"apiToken"`
}

type PublicPingdomDirectConfig struct {
	HiddenAPIToken string `json:"apiToken"`
}

// InstanaDirectConfig represents content of Instana configuration typical for Direct Object.
type InstanaDirectConfig struct {
	APIToken string `json:"apiToken"`
	URL      string `json:"url" validate:"required,url,httpsURL"`
}

// PublicInstanaDirectConfig represents content of Instana configuration typical for Direct Object without secrets.
type PublicInstanaDirectConfig struct {
	HiddenAPIToken string `json:"apiToken"`
	URL            string `json:"url"`
}

// InfluxDBDirectConfig represents content of InfluxDB configuration typical for Direct Object.
type InfluxDBDirectConfig struct {
	URL            string `json:"url" validate:"required,url"`
	APIToken       string `json:"apiToken"`
	OrganizationID string `json:"organizationID"`
}

// PublicInfluxDBDirectConfig represents content of InfluxDB configuration typical for Direct Object without secrets.
type PublicInfluxDBDirectConfig struct {
	URL                  string `json:"url"`
	HiddenAPIToken       string `json:"apiToken"`
	HiddenOrganizationID string `json:"organizationID"`
}

// RedshiftDirectConfig represents content of Redshift configuration typical for Direct Object.
type RedshiftDirectConfig struct {
	AccessKeyID     string `json:"accessKeyID,omitempty"`
	SecretAccessKey string `json:"secretAccessKey,omitempty"`
	SecretARN       string `json:"secretARN"`
	RoleARN         string `json:"roleARN,omitempty" example:"arn:aws:iam::123456789012:role/SomeAccessRole"` //nolint: lll
}

// PublicRedshiftDirectConfig represents content of Redshift configuration typical for Direct Object without secrets.
type PublicRedshiftDirectConfig struct {
	HiddenAccessKeyID     string `json:"accessKeyID,omitempty"`
	HiddenSecretAccessKey string `json:"secretAccessKey,omitempty"`
	SecretARN             string `json:"secretARN"`
	HiddenRoleARN         string `json:"roleARN,omitempty"`
}

// SumoLogicDirectConfig represents content of SumoLogic configuration typical for Direct Object.
type SumoLogicDirectConfig struct {
	AccessID  string `json:"accessID"`
	AccessKey string `json:"accessKey"`
	URL       string `json:"url" validate:"required,url"`
}

// PublicSumoLogicDirectConfig represents content of SumoLogic configuration typical for Direct Object without secrets.
type PublicSumoLogicDirectConfig struct {
	HiddenAccessID  string `json:"accessID"`
	HiddenAccessKey string `json:"accessKey"`
	URL             string `json:"url"`
}

// AzureMonitorDirectConfig represents content of AzureMonitor Configuration typical for Direct Object.
type AzureMonitorDirectConfig struct {
	TenantID     string `json:"tenantId" validate:"required,uuid_rfc4122" example:"abf988bf-86f1-41af-91ab-2d7cd011db46"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

// PublicAzureMonitorDirectConfig represents content of AzureMonitor Configuration
// typical for Direct Object without secrets.
type PublicAzureMonitorDirectConfig struct {
	TenantID           string `json:"tenantId" validate:"required,uuid_rfc4122" example:"abf988bf-86f1-41af-91ab-2d7cd011db46"` //nolint: lll
	HiddenClientID     string `json:"clientId"`
	HiddenClientSecret string `json:"clientSecret"`
}

// HoneycombDirectConfig represents content of Honeycomb Configuration typical for Direct Object.
type HoneycombDirectConfig struct {
	APIKey string `json:"apiKey,omitempty" example:"lwPoPt20Gmdi4dwTdW9dTR"`
}

// PublicHoneycombDirectConfig represents content of Honeycomb Configuration typical for Direct Object without secrets.
type PublicHoneycombDirectConfig struct {
	HiddenAPIKey string `json:"apiKey,omitempty"`
}

// AWSIAMRoleAuthExternalIDs struct which is used for exposing AWS IAM role auth data
type AWSIAMRoleAuthExternalIDs struct {
	ExternalID string `json:"externalID"`
	AccountID  string `json:"accountID"`
}
