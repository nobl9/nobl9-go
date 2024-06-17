package v1alphaExamples

import "github.com/nobl9/nobl9-go/manifest/v1alpha"

func MetadataAnnotations() v1alpha.MetadataAnnotations {
	return v1alpha.MetadataAnnotations{
		"team":   "sales",
		"env":    "prod",
		"region": "us",
		"area":   "latency",
	}
}
