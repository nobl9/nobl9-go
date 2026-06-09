package v1alphaExamples

import (
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
)

func Project() []Example {
	return newExampleSlice(standardExample{
		Object: v1alphaProject.New(
			v1alphaProject.Metadata{
				Name:        "default",
				Labels:      exampleLabels(),
				Annotations: exampleMetadataAnnotations(),
			},
			v1alphaProject.Spec{
				Description: "Example Project",
			},
		),
	},
	)
}
