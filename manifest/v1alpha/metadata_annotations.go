package v1alpha

import (
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

var (
	annotationKeyRegexp = regexp.MustCompile(`^\p{Ll}([_\-0-9\p{Ll}]*[0-9\p{Ll}])?$`)
)

func MetadataAnnotationsValidationRules() validation.Validator[MetadataAnnotations] {
	return validation.New[MetadataAnnotations](
		validation.ForMap(validation.GetSelf[MetadataAnnotations]()).
			RulesForKeys(
				validation.StringLength(minAnnotationKeyLength, maxAnnotationKeyLength),
				validation.StringMatchRegexp(annotationKeyRegexp),
			).
			IncludeForValues(annotationValueValidator),
	)
}

var annotationValueValidator = validation.New[annotationValue](
	//validation.For(func(value annotationValue) string { return value }).
	validation.For(validation.GetSelf[string]()).
		Rules(
			validation.StringMaxLength(maxAnnotationValueLength),
		))
