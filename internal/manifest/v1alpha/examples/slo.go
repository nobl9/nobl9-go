package v1alphaExamples

import v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"

func SLO() v1alphaSLO.SLO {
	return v1alphaSLO.New(
		v1alphaSLO.Metadata{
			Name:        "prometheus",
			Project:     "default",
			Labels:      Labels(),
			Annotations: MetadataAnnotations(),
		},
		v1alphaSLO.Spec{
			Description: "Example SLO",
		})
}
