package v1alpha

import (
	"regexp"
	"unicode/utf8"

	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/internal/validation"
)

type (
	MetadataAnnotations map[annotationKey]annotationValue

	annotationKey   = string
	annotationValue = string
)

func ValidationRuleMetadataAnnotations() validation.SingleRule[MetadataAnnotations] {
	return validation.NewSingleRule(func(v MetadataAnnotations) error { return v.Validate() })
}

// Validate checks if the MetadataAnnotations keys and values are valid.
func (a MetadataAnnotations) Validate() error {
	for key, value := range a {
		if err := a.validateKey(key); err != nil {
			return err
		}
		if err := a.validateValue(key, value); err != nil {
			return err
		}
	}
	return nil
}

const (
	minAnnotationKeyLength   = 1
	maxAnnotationKeyLength   = 63
	minAnnotationValueLength = 1
	maxAnnotationValueLength = 1050
)

var (
	annotationKeyRegexp = regexp.MustCompile(`^\p{L}([_\-0-9\p{L}]*[0-9\p{L}])?$`)
)

func (a MetadataAnnotations) validateKey(key annotationKey) error {
	if len(key) > maxAnnotationKeyLength || len(key) < minAnnotationKeyLength {
		return errors.Errorf("annotation key '%s' length must be between %d and %d",
			key, minAnnotationKeyLength, maxAnnotationKeyLength)
	}
	if !annotationKeyRegexp.MatchString(key) {
		return errors.Errorf("annotation key '%s' does not match the regex: %s", key, annotationKeyRegexp.String())
	}
	if hasUpperCaseLettersRegexp.MatchString(key) {
		return errors.Errorf("annotation key '%s' must not have upper case letters", key)
	}
	return nil
}

func (a MetadataAnnotations) validateValue(key annotationKey, value annotationValue) error {
	if utf8.RuneCountInString(value) >= minAnnotationValueLength &&
		utf8.RuneCountInString(value) <= maxAnnotationValueLength {
		return nil
	}
	return errors.Errorf("annotation value '%s' length for key '%s' must be between %d and %d",
		value, key, minAnnotationValueLength, maxAnnotationValueLength)
}
