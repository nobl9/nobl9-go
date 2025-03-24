package v1alpha

import (
	_ "embed"
	"regexp"

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

const (
	minAnnotationKeyLength = 1
	// Subdomain + separator + qualified name.
	// This way we're keeping it roughly compatible with OpenSLO.
	maxAnnotationKeyLength   = 253 + 1 + 63
	maxAnnotationValueLength = 1050
)

//go:embed metadata_annotations_examples.yaml
var metadataAnnotationsExamples string

var annotationKeyRegexp = regexp.MustCompile(`^\p{L}([_\-0-9\p{L}]*[0-9\p{L}])?$`)

func MetadataAnnotationsValidationRules() govy.Validator[MetadataAnnotations] {
	return govy.New[MetadataAnnotations](
		govy.ForMap(govy.GetSelf[MetadataAnnotations]()).
			RulesForKeys(
				rules.StringLength(minAnnotationKeyLength, maxAnnotationKeyLength),
				rules.StringMatchRegexp(annotationKeyRegexp),
			).
			RulesForValues(rules.StringMaxLength(maxAnnotationValueLength)).
			WithExamples(metadataAnnotationsExamples),
	)
}
