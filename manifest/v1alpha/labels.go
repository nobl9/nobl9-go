package v1alpha

import "unicode/utf8"

type (
	Labels map[LabelKey][]LabelValue

	LabelKey   = string
	LabelValue = string
)

// TODO: change the definition to return an error instead of boolean
func (l Labels) Validate() bool {
	for key, values := range l {
		if !l.validateKey(key) {
			return false
		}
		if l.hasDuplicates(values) {
			return false
		}
		for _, value := range values {
			// Validate only if len(value) > 0, in case where we have only key labels,
			// there is always empty value string and this is not an error.
			if len(value) > 0 && !l.validateValue(value) {
				return false
			}
		}
	}
	return true
}

func (l Labels) validateKey(key LabelKey) bool {
	const maxLabelKeyLength = 63
	if len(key) > maxLabelKeyLength || len(key) < 1 {
		return false
	}

	if !labelKeyRegexp.MatchString(key) {
		return false
	}
	return !hasUpperCaseLettersRegexp.MatchString(key)
}

const (
	minLabelValueLength = 1
	maxLabelValueLength = 200
)

func (l Labels) validateValue(value LabelValue) bool {
	return utf8.RuneCountInString(value) >= minLabelValueLength && utf8.RuneCountInString(value) <= maxLabelValueLength
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
