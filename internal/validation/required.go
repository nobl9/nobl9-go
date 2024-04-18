package validation

import (
	"reflect"

	"github.com/pkg/errors"
)

func Required[T any]() SingleRule[T] {
	msg := NewRequiredError().Message
	return NewSingleRule(func(v T) error {
		if isEmptyFunc(v) {
			return errors.New(msg)
		}
		return nil
	}).
		WithErrorCode(ErrorCodeRequired).
		WithDescription(msg)
}

// isEmptyFunc checks only the types which it makes sense for.
// It's hard to consider 0 an empty value for anything really.
func isEmptyFunc(v interface{}) bool {
	rv := reflect.ValueOf(v)
	return rv.Kind() == 0 || rv.IsZero()
}
