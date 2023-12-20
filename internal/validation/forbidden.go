package validation

import "github.com/pkg/errors"

func Forbidden[T any]() SingleRule[T] {
	return NewSingleRule(func(v T) error {
		if isEmptyFunc(v) {
			return nil
		}
		return errors.New("property is forbidden")
	}).WithErrorCode(ErrorCodeForbidden)
}
