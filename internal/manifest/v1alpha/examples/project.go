package v1alphaExamples

import v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"

func Project() v1alphaProject.Project {
	return v1alphaProject.New(
		v1alphaProject.Metadata{
			Name:        "default",
			Labels:      Labels(),
			Annotations: MetadataAnnotations(),
		},
		v1alphaProject.Spec{
			Description: "Example Project",
		},
	)
}
