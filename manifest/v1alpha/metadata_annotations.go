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

var keyValidator = validation.New[string](
	validation.For(func(key string) string { return key }).
		WithName("key").
		Required().
		Rules(
			validation.StringLength(minAnnotationKeyLength, maxAnnotationKeyLength),
			validation.StringMatchRegexp(annotationKeyRegexp),
			validation.StringDenyRegexp(hasUpperCaseLettersRegexp),
		))

var valueValidator = validation.New[string](
	validation.For(func(key string) string { return key }).
		WithName("value").
		Required().
		Rules(
			validation.StringLength(minAnnotationValueLength, maxAnnotationValueLength),
		))

func ValidationRuleMetadataAnnotations() validation.SingleRule[MetadataAnnotations] {
	return validation.NewSingleRule(func(a MetadataAnnotations) error {
		for key, value := range a {
			if err := keyValidator.Validate(key); err != nil {
				return err
			}
			if err := valueValidator.Validate(value); err != nil {
				return err
			}
		}
		return nil
	})
}
