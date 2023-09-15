package v1alpha

import (
	"unicode/utf8"

	"github.com/pkg/errors"
)

type (
	Labels map[LabelKey][]LabelValue

	LabelKey   = string
	LabelValue = string
)

func (l Labels) Validate() error {
	for key, values := range l {
		if err := l.validateKey(key); err != nil {
			return err
		}
		if l.hasDuplicates(values) {
			return errors.New("TODO")
		}
		for _, value := range values {
			// Validate only if len(value) > 0, in case where we have only key labels,
			// there is always empty value string and this is not an error.
			if len(value) == 0 {
				continue
			}
			if err := l.validateValue(value); err != nil {
				return err
			}
		}
	}
	return nil
}

func (l Labels) validateKey(key LabelKey) error {
	const maxLabelKeyLength = 63
	if len(key) > maxLabelKeyLength || len(key) < 1 {
		return errors.New("TODO")
	}

	if !labelKeyRegexp.MatchString(key) {
		return errors.New("TODO")
	}
	if hasUpperCaseLettersRegexp.MatchString(key) {
		return errors.New("TODO")
	}
	return nil
}

const (
	minLabelValueLength = 1
	maxLabelValueLength = 200
)

func (l Labels) validateValue(value LabelValue) error {
	if utf8.RuneCountInString(value) >= minLabelValueLength &&
		utf8.RuneCountInString(value) <= maxLabelValueLength {
		return nil
	}
	return errors.New("TODO")
}

func (l Labels) hasDuplicates(labelValues []LabelValue) bool {
	duplicateFrequency := make(map[string]int)

	for _, value := range labelValues {
		_, exist := duplicateFrequency[value]

		if exist {
			duplicateFrequency[value]++
		} else {
			duplicateFrequency[value] = 1
		}
		if duplicateFrequency[value] > 1 {
			return true
		}
	}
	return false
}
