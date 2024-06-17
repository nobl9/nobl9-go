package v1alphaExamples

import (
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
)

func Service() v1alphaService.Service {
	return v1alphaService.New(
		v1alphaService.Metadata{
			Name:        "prometheus",
			Project:     "default",
			Labels:      Labels(),
			Annotations: MetadataAnnotations(),
		},
		v1alphaService.Spec{
			Description: "Example Service",
		},
	)
}
