package v1alphaExamples

import (
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
)

func Service() []Example {
	return newExampleSlice(standardExample{
		Object: v1alphaService.New(
			v1alphaService.Metadata{
				Name:        "web",
				DisplayName: "Web Service",
				Project:     "default",
				Labels:      exampleLabels(),
				Annotations: exampleMetadataAnnotations(),
			},
			v1alphaService.Spec{
				Description: "Example web Service",
			},
		)},
	)
}
