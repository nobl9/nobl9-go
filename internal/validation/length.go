package validation

import (
	"fmt"
	"unicode/utf8"

	"github.com/pkg/errors"
)

func StringLength(min, max int) SingleRule[string] {
	msg := fmt.Sprintf("length must be between %d and %d", min, max)
	return NewSingleRule(func(v string) error {
		length := utf8.RuneCountInString(v)
		if length < min || length > max {
			return errors.New(msg)
		}
		return nil
	}).
		WithErrorCode(ErrorCodeStringLength).
		WithDescription(msg)
}

func StringMinLength(min int) SingleRule[string] {
	msg := fmt.Sprintf("length must be %s %d", cmpGreaterThanOrEqual, min)
	return NewSingleRule(func(v string) error {
		length := utf8.RuneCountInString(v)
		if length < min {
			return errors.New(msg)
		}
		return nil
	}).
		WithErrorCode(ErrorCodeStringMinLength).
		WithDescription(msg)
}

func StringMaxLength(max int) SingleRule[string] {
	msg := fmt.Sprintf("length must be %s %d", cmpLessThanOrEqual, max)
	return NewSingleRule(func(v string) error {
		length := utf8.RuneCountInString(v)
		if length > max {
			return errors.New(msg)
		}
		return nil
	}).
		WithErrorCode(ErrorCodeStringMaxLength).
		WithDescription(msg)
}

func SliceLength[S ~[]E, E any](min, max int) SingleRule[S] {
	msg := fmt.Sprintf("length must be between %d and %d", min, max)
	return NewSingleRule(func(v S) error {
		length := len(v)
		if length < min || length > max {
			return errors.New(msg)
		}
		return nil
	}).
		WithErrorCode(ErrorCodeSliceLength).
		WithDescription(msg)
}

func SliceMinLength[S ~[]E, E any](min int) SingleRule[S] {
	msg := fmt.Sprintf("length must be %s %d", cmpGreaterThanOrEqual, min)
	return NewSingleRule(func(v S) error {
		length := len(v)
		if length < min {
			return errors.New(msg)
		}
		return nil
	}).
		WithErrorCode(ErrorCodeSliceMinLength).
		WithDescription(msg)
}

func SliceMaxLength[S ~[]E, E any](max int) SingleRule[S] {
	msg := fmt.Sprintf("length must be %s %d", cmpLessThanOrEqual, max)
	return NewSingleRule(func(v S) error {
		length := len(v)
		if length > max {
			return errors.New(msg)
		}
		return nil
	}).
		WithErrorCode(ErrorCodeSliceMaxLength).
		WithDescription(msg)
}

func MapLength[M ~map[K]V, K comparable, V any](min, max int) SingleRule[M] {
	msg := fmt.Sprintf("length must be between %d and %d", min, max)
	return NewSingleRule(func(v M) error {
		length := len(v)
		if length < min || length > max {
			return errors.New(msg)
		}
		return nil
	}).
		WithErrorCode(ErrorCodeMapLength).
		WithDescription(msg)
}

func MapMinLength[M ~map[K]V, K comparable, V any](min int) SingleRule[M] {
	msg := fmt.Sprintf("length must be %s %d", cmpGreaterThanOrEqual, min)
	return NewSingleRule(func(v M) error {
		length := len(v)
		if length < min {
			return errors.New(msg)
		}
		return nil
	}).
		WithErrorCode(ErrorCodeMapMinLength).
		WithDescription(msg)
}

func MapMaxLength[M ~map[K]V, K comparable, V any](max int) SingleRule[M] {
	msg := fmt.Sprintf("length must be %s %d", cmpLessThanOrEqual, max)
	return NewSingleRule(func(v M) error {
		length := len(v)
		if length > max {
			return errors.New(msg)
		}
		return nil
	}).
		WithErrorCode(ErrorCodeMapMaxLength).
		WithDescription(msg)
}
