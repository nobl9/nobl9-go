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
	minAnnotationValueLength = 1
	maxAnnotationValueLength = 1050
)

var (
	annotationKeyRegexp = regexp.MustCompile(`^\p{L}([_\-0-9\p{L}]*[0-9\p{L}])?$`)
)

var valueValidator = validation.New[string](
	validation.For(func(key string) string { return key }).
		Required().
		Rules(
			validation.StringLength(minAnnotationValueLength, maxAnnotationValueLength),
		))

func MetadataAnnotationsValidationRules() validation.Validator[MetadataAnnotations] {
	return validation.New[MetadataAnnotations](
		validation.ForMap(validation.GetSelf[MetadataAnnotations]()).
			RulesForKeys(validation.StringLength(minAnnotationKeyLength, maxAnnotationKeyLength),
				validation.StringMatchRegexp(annotationKeyRegexp),
				validation.StringDenyRegexp(hasUpperCaseLettersRegexp),
			).
			IncludeForValues(valueValidator),
	)
}
