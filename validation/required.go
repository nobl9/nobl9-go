package validation

import (
	"reflect"

	"github.com/pkg/errors"
)

func Required[T any]() Rule[T] {
	return NewSingleRule(func(v T) error {
		if isEmpty(v) {
			return errors.New("property is required but was empty")
		}
		return nil
	}).WithErrorCode(ErrorCodeRequired)
}

func isEmpty(v interface{}) bool {
	return reflect.ValueOf(v).IsZero()
}
