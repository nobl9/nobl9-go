package v1alpha

import (
	_ "embed"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// MetadataAnnotations are non-identifiable key-value pairs that can be attached to
// SLOs, services, projects, and alert policies.
// Metadata annotations are used for descriptive purposes only.
type MetadataAnnotations map[annotationKey]annotationValue
type (
	annotationKey   = string
	annotationValue = string
)

const maxAnnotationValueLength = 1050

//go:embed metadata_annotations_examples.yaml
var metadataAnnotationsExamples string

func MetadataAnnotationsValidationRules() govy.Validator[MetadataAnnotations] {
	return govy.New[MetadataAnnotations](
		govy.ForMap(govy.GetSelf[MetadataAnnotations]()).
			RulesForKeys(rules.StringKubernetesQualifiedName()).
			RulesForValues(rules.StringMaxLength(maxAnnotationValueLength)).
			WithExamples(metadataAnnotationsExamples),
	)
}
