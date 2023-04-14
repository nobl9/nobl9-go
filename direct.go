package nobl9

import "encoding/json"

// Direct struct which mapped one to one with kind: Direct yaml definition
type Direct struct {
	ObjectHeader
	Spec   DirectSpec   `json:"spec"`
	Status DirectStatus `json:"status"`
}

// DirectStatus represents content of Status optional for Direct Object
type DirectStatus struct {
	DirectType string `json:"directType" example:"Datadog"`
}

// DirectSpec represents content of Spec typical for Direct Object
type DirectSpec struct {
	Description             string                           `json:"description,omitempty" example:"Datadog description"` //nolint:lll
	SourceOf                []string                         `json:"sourceOf" example:"Metrics,Services"`
	HistoricalDataRetrieval *HistoricalDataRetrieval         `json:"historicalDataRetrieval"`
	QueryDelay              *QueryDelayDuration              `json:"queryDelay"`
	AppDynamics             *AppDynamicsDirectConfig         `json:"appDynamics,omitempty"`
	BigQuery                *BigQueryDirectConfig            `json:"bigQuery,omitempty"`
	CloudWatch              *CloudWatchDirectConfig          `json:"cloudWatch,omitempty"`
	Datadog                 *DatadogDirectConfig             `json:"datadog,omitempty"`
	Dynatrace               *DynatraceDirectConfig           `json:"dynatrace,omitempty"`
	GCM                     *GCMDirectConfig                 `json:"gcm,omitempty"`
	InfluxDB                *InfluxDBDirectConfig            `json:"influxdb,omitempty"`
	Instana                 *InstanaDirectConfig             `json:"instana,omitempty"`
	Lightstep               *LightstepDirectConfig           `json:"lightstep,omitempty"`
	NewRelic                *NewRelicDirectConfig            `json:"newRelic,omitempty"`
	Pingdom                 *PingdomDirectConfig             `json:"pingdom,omitempty"`
	Redshift                *RedshiftDirectConfig            `json:"redshift,omitempty"`
	Splunk                  *SplunkDirectConfig              `json:"splunk,omitempty"`
	SplunkObservability     *SplunkObservabilityDirectConfig `json:"splunkObservability,omitempty"`
	SumoLogic               *SumoLogicDirectConfig           `json:"sumoLogic,omitempty"`
	ThousandEyes            *ThousandEyesDirectConfig        `json:"thousandEyes,omitempty"`
}

// AppDynamicsDirectConfig represents content of AppDynamics configuration typical for Direct Object
type AppDynamicsDirectConfig struct {
	URL          string `json:"url,omitempty" validate:"httpsURL" example:"https://{tenant}.saas.appdynamics.com"`
	AccountName  string `json:"accountName,omitempty" example:"accountID"`
	ClientID     string `json:"clientID,omitempty" example:"apiClientID@accountID"`
	ClientSecret string `json:"clientSecret,omitempty" example:"secret"`
	ClientName   string `json:"clientName,omitempty" example:"apiClientID"`
}

// BigQueryDirectConfig represents content of BigQuery configuration typical for Direct Object.
type BigQueryDirectConfig struct {
	ServiceAccountKey string `json:"serviceAccountKey,omitempty" example:"secret"`
}

// CloudWatchDirectConfig represents content of CloudWatch configuration typical for Direct Object.
type CloudWatchDirectConfig struct {
	AccessKeyID     string `json:"accessKeyID,omitempty" example:"secret"`
	SecretAccessKey string `json:"secretAccessKey,omitempty" example:"secret"`
}

// DatadogDirectConfig represents content of Datadog configuration typical for Direct Object.
type DatadogDirectConfig struct {
	Site           string `json:"site,omitempty" example:"eu"`
	APIKey         string `json:"apiKey,omitempty" example:"secret"`
	ApplicationKey string `json:"applicationKey,omitempty" example:"secret"`
}

// DynatraceDirectConfig represents content of Dynatrace configuration typical for Direct Object.
type DynatraceDirectConfig struct {
	URL            string `json:"url,omitempty" example:"https://{your-environment-id}.live.dynatrace.com or https://{your-domain}/e/{your-environment-id}"` //nolint: lll
	DynatraceToken string `json:"dynatraceToken,omitempty" example:"secret"`
}

// GCMDirectConfig represents content of GCM configuration typical for Direct Object.
type GCMDirectConfig struct {
	ServiceAccountKey string `json:"serviceAccountKey,omitempty" example:"secret"`
}

// InfluxDBDirectConfig represents content of InfluxDB configuration typical for Direct Object.
type InfluxDBDirectConfig struct {
	URL            string `json:"url,omitempty"`
	APIToken       string `json:"apiToken,omitempty" example:"secret"`
	OrganizationID string `json:"organizationID,omitempty" example:"secret"`
}

// InstanaDirectConfig represents content of Instana configuration typical for Direct Object.
type InstanaDirectConfig struct {
	URL      string `json:"url,omitempty"`
	APIToken string `json:"apiToken,omitempty" example:"secret"`
}

// LightstepDirectConfig represents content of Lightstep configuration typical for Direct Object.
type LightstepDirectConfig struct {
	Organization string `json:"organization,omitempty" example:"LightStep-Play"`
	Project      string `json:"project,omitempty" example:"play"`
	AppToken     string `json:"appToken,omitempty" example:"secret"`
}

// NewRelicDirectConfig represents content of NewRelic configuration typical for Direct Object.
type NewRelicDirectConfig struct {
	AccountID        json.Number `json:"accountId,omitempty"`
	InsightsQueryKey string      `json:"insightsQueryKey,omitempty" example:"secret"`
}

// PingdomDirectConfig represents content of Pingdom configuration typical for Direct Object.
type PingdomDirectConfig struct {
	APIToken string `json:"apiToken,omitempty" example:"secret"`
}

// RedshiftDirectConfig represents content of Redshift configuration typical for Direct Object.
type RedshiftDirectConfig struct {
	AccessKeyID     string `json:"accessKeyID,omitempty" example:"secret"`
	SecretAccessKey string `json:"secretAccessKey,omitempty" example:"secret"`
	SecretARN       string `json:"secretARN,omitempty"`
}

// SplunkDirectConfig represents content of Splunk configuration typical for Direct Object.
type SplunkDirectConfig struct {
	URL         string `json:"url,omitempty"`
	AccessToken string `json:"accessToken,omitempty" example:"secret"`
}

// SplunkObservabilityDirectConfig represents content of SplunkObservability configuration typical for Direct Object.
type SplunkObservabilityDirectConfig struct {
	Realm       string `json:"realm,omitempty" example:"us1"`
	AccessToken string `json:"accessToken,omitempty" example:"secret"`
}

// SumoLogicDirectConfig represents content of SumoLogic configuration typical for Direct Object.
type SumoLogicDirectConfig struct {
	URL       string `json:"url,omitempty"`
	AccessID  string `json:"accessID,omitempty" example:"secret"`
	AccessKey string `json:"accessKey,omitempty" example:"secret"`
}

// ThousandEyesDirectConfig represents content of ThousandEyes Configuration typical for Direct Object.
type ThousandEyesDirectConfig struct {
	OauthBearerToken string `json:"oauthBearerToken,omitempty" example:"secret"`
}

// genericToDirect converts ObjectGeneric to ObjectDirect.
func genericToDirect(o ObjectGeneric, onlyHeader bool) (Direct, error) {
	res := Direct{
		ObjectHeader: o.ObjectHeader,
	}
	if onlyHeader {
		return res, nil
	}
	var resSpec DirectSpec
	if err := json.Unmarshal(o.Spec, &resSpec); err != nil {
		err = EnhanceError(o, err)
		return res, err
	}
	res.Spec = resSpec
	return res, nil
}
