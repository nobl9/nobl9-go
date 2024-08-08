package v1alphaExamples

import "github.com/nobl9/nobl9-go/manifest/v1alpha"

func MetadataAnnotations() []Example {
	return newExampleSlice(standardExample{
		Object: exampleMetadataAnnotations(),
	})
}

func exampleMetadataAnnotations() v1alpha.MetadataAnnotations {
	return v1alpha.MetadataAnnotations{
		"team":   "sales",
		"env":    "prod",
		"region": "us",
		"area":   "latency",
	}
}

func exampleCompositeMetadataAnnotations() v1alpha.MetadataAnnotations {
	return v1alpha.MetadataAnnotations{
		"team":   "ux",
		"env":    "prod",
		"region": "us",
		"area":   "user-experience",
	}
}
