package nobl9

// genericToDirect converts ObjectGeneric to ObjectDirect
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

// Direct struct which mapped one to one with kind: Direct yaml definition
type Direct struct {
	ObjectHeader
	Spec   DirectSpec   `json:"spec"`
	Status DirectStatus `json:"status"`
}

// DirectSpec represents content of Spec typical for Direct Object
type DirectSpec struct {
	Description         string                           `json:"description,omitempty" example:"Datadog description"` //nolint:lll
	SourceOf            []string                         `json:"sourceOf" example:"Metrics,Services"`
	Datadog             *DatadogDirectConfig             `json:"datadog,omitempty"`
	NewRelic            *NewRelicDirectConfig            `json:"newRelic,omitempty"`
	AppDynamics         *AppDynamicsDirectConfig         `json:"appDynamics,omitempty"`
	SplunkObservability *SplunkObservabilityDirectConfig `json:"splunkObservability,omitempty"`
	ThousandEyes        *ThousandEyesDirectConfig        `json:"thousandEyes,omitempty"`
	BigQuery            *BigQueryDirectConfig            `json:"bigQuery,omitempty"`
}

// AppDynamicsDirectConfig represents content of AppDynamics Configuration typical for Direct Object
type AppDynamicsDirectConfig struct {
	URL          string `json:"url,omitempty"`
	ClientID     string `json:"clientID,omitempty" example:"apiClientID@accountID"`
	ClientSecret string `json:"clientSecret,omitempty" example:"secret"`
}

// DirectStatus represents content of Status optional for Direct Object
type DirectStatus struct {
	DirectType string `json:"directType" example:"Datadog"`
}
