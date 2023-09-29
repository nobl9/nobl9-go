package validation

import (
	"reflect"

	"github.com/pkg/errors"
)

func Required[T any]() Rule[T] {
	return NewSingleRule(func(v T) error {
		if reflect.ValueOf(v).IsZero() {
			return errors.New("property is required but was empty")
		}
		return nil
	}).WithErrorCode(ErrorCodeRequired)
}