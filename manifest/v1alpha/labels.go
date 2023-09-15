package v1alpha

import (
	"regexp"
	"unicode/utf8"

	"github.com/pkg/errors"
)

type (
	Labels map[LabelKey][]LabelValue

	LabelKey   = string
	LabelValue = string
)

// Validate checks if the Labels keys and values are valid.
func (l Labels) Validate() error {
	for key, values := range l {
		if err := l.validateKey(key); err != nil {
			return err
		}
		if err := l.ensureValuesUniqueness(key, values); err != nil {
			return err
		}
		for _, value := range values {
			// Validate only if len(value) > 0, in case where we have only key labels,
			// there is always empty value string and this is not an error.
			if len(value) == 0 {
				continue
			}
			if err := l.validateValue(key, value); err != nil {
				return err
			}
		}
	}
	return nil
}

const (
	minLabelKeyLength   = 1
	maxLabelKeyLength   = 63
	minLabelValueLength = 1
	maxLabelValueLength = 200
)

var (
	labelKeyRegexp            = regexp.MustCompile(`^\p{L}([_\-0-9\p{L}]*[0-9\p{L}])?$`)
	hasUpperCaseLettersRegexp = regexp.MustCompile(`[A-Z]+`)
)

func (l Labels) validateKey(key LabelKey) error {
	if len(key) > maxLabelKeyLength || len(key) < minLabelKeyLength {
		return errors.Errorf("label key '%s' length must be between %d and %d",
			key, minLabelKeyLength, maxLabelKeyLength)
	}
	if !labelKeyRegexp.MatchString(key) {
		return errors.Errorf("label key '%s' does not match the regex: %s", key, labelKeyRegexp.String())
	}
	if hasUpperCaseLettersRegexp.MatchString(key) {
		return errors.Errorf("label key '%s' must not have upper case letters", key)
	}
	return nil
}

func (l Labels) validateValue(key LabelKey, value LabelValue) error {
	if utf8.RuneCountInString(value) >= minLabelValueLength &&
		utf8.RuneCountInString(value) <= maxLabelValueLength {
		return nil
	}
	return errors.Errorf("label value '%s' length for key '%s' must be between %d and %d",
		value, key, minLabelValueLength, maxLabelValueLength)
}

func (l Labels) ensureValuesUniqueness(key LabelKey, labelValues []LabelValue) error {
	uniqueValues := make(map[string]struct{})
	for _, value := range labelValues {
		if _, exists := uniqueValues[value]; exists {
			return errors.Errorf(
				"label value '%s' for key '%s' already exists, duplicates are not allowed", value, key)
		}
		uniqueValues[value] = struct{}{}
	}
	return nil
}
