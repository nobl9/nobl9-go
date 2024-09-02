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
	minAnnotationKeyLength   = 1
	maxAnnotationKeyLength   = 63
	maxAnnotationValueLength = 1050
)

//go:embed metadata_annotations_examples.yaml
var metadataAnnotationsExamples string

var annotationKeyRegexp = regexp.MustCompile(`^\p{Ll}([_\-0-9\p{Ll}]*[0-9\p{Ll}])?$`)

func MetadataAnnotationsValidationRules() govy.Validator[MetadataAnnotations] {
	return govy.New(
		govy.ForMap(govy.GetSelf[MetadataAnnotations]()).
			RulesForKeys(
				rules.StringLength(minAnnotationKeyLength, maxAnnotationKeyLength),
				rules.StringMatchRegexp(annotationKeyRegexp),
			).
			IncludeForValues(annotationValueValidator).
			WithExamples(metadataAnnotationsExamples),
	)
}

var annotationValueValidator = govy.New(
	govy.For(govy.GetSelf[string]()).
		Rules(
			rules.StringMaxLength(maxAnnotationValueLength),
		))
