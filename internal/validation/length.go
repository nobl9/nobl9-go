package validation

import (
	"unicode/utf8"

	"github.com/pkg/errors"
)

func StringLength(min, max int) SingleRule[string] {
	return NewSingleRule(
		func(v string) error { return lengthComparison(utf8.RuneCountInString(v), min, max) }).
		WithErrorCode(ErrorCodeStringLength)
}

func StringMinLength(min int) SingleRule[string] {
	return StringLength(min, noLengthBound).WithErrorCode(ErrorCodeStringMinLength)
}

func StringMaxLength(max int) SingleRule[string] {
	return StringLength(noLengthBound, max).WithErrorCode(ErrorCodeStringMaxLength)
}

func SliceLength[S ~[]E, E any](min, max int) SingleRule[S] {
	return NewSingleRule(func(v S) error { return lengthComparison(len(v), min, max) }).
		WithErrorCode(ErrorCodeSliceLength)
}

func SliceMinLength[S ~[]E, E any](min int) SingleRule[S] {
	return SliceLength[S](min, noLengthBound).WithErrorCode(ErrorCodeSliceMinLength)
}

func SliceMaxLength[S ~[]E, E any](max int) SingleRule[S] {
	return SliceLength[S](noLengthBound, max).WithErrorCode(ErrorCodeSliceMaxLength)
}

func MapLength[M ~map[K]V, K comparable, V any](min, max int) SingleRule[M] {
	return NewSingleRule(func(v M) error { return lengthComparison(len(v), min, max) }).
		WithErrorCode(ErrorCodeMapLength)
}

func MapMinLength[M ~map[K]V, K comparable, V any](min int) SingleRule[M] {
	return MapLength[M](min, noLengthBound).WithErrorCode(ErrorCodeMapMinLength)
}

func MapMaxLength[M ~map[K]V, K comparable, V any](max int) SingleRule[M] {
	return MapLength[M](noLengthBound, max).WithErrorCode(ErrorCodeMapMaxLength)
}

const noLengthBound = -1

func lengthComparison(length, min, max int) error {
	lowerBoundBreached := length < min
	upperBoundBreached := length > max
	if (lowerBoundBreached || upperBoundBreached) && min != noLengthBound && max != noLengthBound {
		return errors.Errorf("length must be between %d and %d", min, max)
	}
	if upperBoundBreached && min == noLengthBound && max != noLengthBound {
		return errors.Errorf("length must be %s %d", cmpLessThanOrEqual, max)
	}
	if lowerBoundBreached && max == noLengthBound && min != noLengthBound {
		return errors.Errorf("length must be %s %d", cmpGreaterThanOrEqual, min)
	}
	return nil
}
