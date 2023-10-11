package v1alpha

type FeatureName string

type DataSourceFeature struct {
	Supported           bool   `json:"supported"`
	MinimumAgentVersion string `json:"minimumAgentVersion,omitempty"`
	ReleaseChannel      string `json:"releaseChannel,omitempty"`
}

type DataSourceFeatures map[FeatureName]DataSourceFeature

func (f DataSourceFeatures) appendFeature(name string, feature DataSourceFeature) {
	f[FeatureName(name)] = feature
}
