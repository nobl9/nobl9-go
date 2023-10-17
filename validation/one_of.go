package validation

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

func OneOf[T comparable](values ...T) SingleRule[T] {
	return NewSingleRule(func(v T) error {
		for i := range values {
			if v == values[i] {
				return nil
			}
		}
		return errors.New("must be one of " + prettyStringList(values...))
	}).WithErrorCode(ErrorCodeOneOf)
}

func prettyStringList[T any](values ...T) string {
	b := strings.Builder{}
	b.Grow(2 + len(values))
	b.WriteString("[")
	for i := range values {
		b.WriteString(fmt.Sprint(values[i]))
		if i != len(values)-1 {
			b.WriteString(", ")
		}
	}
	b.WriteString("]")
	return b.String()
}
