package validation

import "github.com/pkg/errors"

func PointerRequired[T any]() SingleRule[*T] {
	return func(v *T) error {
		if v == nil {
			return errors.New("field is required but was nil")
		}
		return nil
	}
}
