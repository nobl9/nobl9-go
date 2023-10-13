package validation

import (
	"unicode/utf8"

	"github.com/pkg/errors"
)

func StringLength(min, max int) SingleRule[string] {
	return NewSingleRule(func(v string) error {
		rc := utf8.RuneCountInString(v)
		if rc < min || rc > max {
			return lengthError(min, max)
		}
		return nil
	}).WithErrorCode(ErrorCodeStringLength)
}

func SliceLength[S ~[]E, E any](min, max int) SingleRule[S] {
	return NewSingleRule(func(v S) error {
		if len(v) < min || len(v) > max {
			return lengthError(min, max)
		}
		return nil
	}).WithErrorCode(ErrorCodeSliceLength)
}

func MapLength[M ~map[K]V, K comparable, V any](min, max int) SingleRule[M] {
	return NewSingleRule(func(v M) error {
		if len(v) < min || len(v) > max {
			return lengthError(min, max)
		}
		return nil
	}).WithErrorCode(ErrorCodeMapLength)
}

func lengthError(min, max int) error {
	return errors.Errorf("length must be between %d and %d", min, max)
}
