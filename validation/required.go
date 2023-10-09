package validation

import (
	"reflect"

	"github.com/pkg/errors"
)

func Required[T any]() SingleRule[T] {
	return NewSingleRule(func(v T) error {
		if reflect.ValueOf(v).IsZero() {
			return errors.New("field is required but was empty")
		}
		return nil
	})
}
