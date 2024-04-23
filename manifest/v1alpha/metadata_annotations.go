package v1alpha

import (
	_ "embed"
	"regexp"

	"github.com/nobl9/nobl9-go/internal/validation"
)

type (
	MetadataAnnotations map[annotationKey]annotationValue

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

func MetadataAnnotationsValidationRules() validation.Validator[MetadataAnnotations] {
	return validation.New[MetadataAnnotations](
		validation.ForMap(validation.GetSelf[MetadataAnnotations]()).
			RulesForKeys(
				validation.StringLength(minAnnotationKeyLength, maxAnnotationKeyLength),
				validation.StringMatchRegexp(annotationKeyRegexp),
			).
			IncludeForValues(annotationValueValidator).
			WithExamples(metadataAnnotationsExamples),
	)
}

var annotationValueValidator = validation.New[annotationValue](
	validation.For(validation.GetSelf[string]()).
		Rules(
			validation.StringMaxLength(maxAnnotationValueLength),
		))
