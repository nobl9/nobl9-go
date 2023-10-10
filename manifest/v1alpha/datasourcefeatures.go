package v1alpha

type FeatureName string

type DataSourceFeature struct {
	Supported           bool   `json:"supported"`
	MinimumAgentVersion string `json:"minimumAgentVersion"`
}

type DataSourceFeatures struct {
	Features map[FeatureName]DataSourceFeature `json:"features"`
}
