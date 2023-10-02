package validation

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRulesForObject(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		r := RulesForObject(
			RulesForField[string]("test", func() string { return "test" }).
				With(SingleRule[string](func(v string) error { return nil })),
		)
		errs := r.Validate()
		assert.Empty(t, errs)
	})

	t.Run("errors", func(t *testing.T) {
		err1 := errors.New("1")
		err2 := errors.New("2")
		r := RulesForObject(
			RulesForField[string]("test", func() string { return "test" }).
				With(SingleRule[string](func(v string) error { return nil })),
			RulesForField[string]("test.name", func() string { return "name" }).
				With(SingleRule[string](func(v string) error { return err1 })),
			RulesForField[string]("test.display", func() string { return "display" }).
				With(SingleRule[string](func(v string) error { return err2 })),
		)
		errs := r.Validate()
		require.Len(t, errs, 2)
		assert.Equal(t, []error{
			&FieldError{
				FieldPath:  "test.name",
				FieldValue: "name",
				Errors:     []string{err1.Error()},
			},
			&FieldError{
				FieldPath:  "test.display",
				FieldValue: "display",
				Errors:     []string{err2.Error()},
			},
		}, errs)
	})
}

func TestRulesForField(t *testing.T) {
	t.Run("no predicates, no error", func(t *testing.T) {
		r := RulesForField[string]("test.path", func() string { return "path" }).
			With(SingleRule[string](func(v string) error { return nil }))
		err := r.Validate()
		assert.NoError(t, err)
	})

	t.Run("no predicates, validate", func(t *testing.T) {
		expectedErr := errors.New("ops!")
		r := RulesForField[string]("test.path", func() string { return "path" }).
			With(SingleRule[string](func(v string) error { return expectedErr }))
		err := r.Validate()
		require.Error(t, err)
		assert.Equal(t, FieldError{
			FieldPath:  "test.path",
			FieldValue: "path",
			Errors:     []string{expectedErr.Error()},
		}, *err.(*FieldError))
	})

	t.Run("predicate matches, don't validate", func(t *testing.T) {
		r := RulesForField[string]("test.path", func() string { return "value" }).
			If(func() bool { return true }).
			If(func() bool { return true }).
			If(func() bool { return false }).
			With(SingleRule[string](func(v string) error { return errors.New("ops!") }))
		err := r.Validate()
		assert.NoError(t, err)
	})

	t.Run("multiple rules", func(t *testing.T) {
		err1 := errors.New("oh no!")
		err2 := errors.New("ops!")
		r := RulesForField[string]("test.path", func() string { return "value" }).
			With(SingleRule[string](func(v string) error { return nil })).
			With(SingleRule[string](func(v string) error { return err1 })).
			With(SingleRule[string](func(v string) error { return nil })).
			With(SingleRule[string](func(v string) error { return err2 }))
		err := r.Validate()
		require.Error(t, err)
		assert.Equal(t, FieldError{
			FieldPath:  "test.path",
			FieldValue: "value",
			Errors:     []string{err1.Error(), err2.Error()},
		}, *err.(*FieldError))
	})
}
