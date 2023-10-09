package validation

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRulesForStruct(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		r := ForStruct[mockStruct](
			ForField[string]("test", func(m mockStruct) string { return "test" }).
				Rules(NewSingleRule(func(v string) error { return nil })),
		)
		errs := r.Validate(mockStruct{})
		assert.Empty(t, errs)
	})

	t.Run("errors", func(t *testing.T) {
		err1 := errors.New("1")
		err2 := errors.New("2")
		r := ForStruct[mockStruct](
			ForField("test", func(m mockStruct) string { return "test" }).
				Rules(NewSingleRule(func(v string) error { return nil })),
			ForField("test.name", func(m mockStruct) string { return "name" }).
				Rules(NewSingleRule(func(v string) error { return err1 })),
			ForField("test.display", func(m mockStruct) string { return "display" }).
				Rules(NewSingleRule(func(v string) error { return err2 })),
		)
		errs := r.Validate(mockStruct{})
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
		r := ForField("test.path", func(m mockStruct) string { return "path" }).
			Rules(NewSingleRule(func(v string) error { return nil }))
		errs := r.Validate(mockStruct{})
		assert.Empty(t, errs)
	})

	t.Run("no predicates, validate", func(t *testing.T) {
		expectedErr := errors.New("ops!")
		r := ForField("test.path", func(m mockStruct) string { return "path" }).
			Rules(NewSingleRule(func(v string) error { return expectedErr }))
		errs := r.Validate(mockStruct{})
		require.Len(t, errs, 1)
		assert.Equal(t, FieldError{
			FieldPath:  "test.path",
			FieldValue: "path",
			Errors:     []string{expectedErr.Error()},
		}, *errs[0].(*FieldError))
	})

	t.Run("predicate matches, don't validate", func(t *testing.T) {
		r := ForField("test.path", func(m mockStruct) string { return "value" }).
			When(func(mockStruct) bool { return true }).
			When(func(mockStruct) bool { return true }).
			When(func(st mockStruct) bool { return st.Field == "" }).
			Rules(NewSingleRule(func(v string) error { return errors.New("ops!") }))
		errs := r.Validate(mockStruct{Field: "something"})
		assert.Empty(t, errs)
	})

	t.Run("multiple rules", func(t *testing.T) {
		err1 := errors.New("oh no!")
		err2 := errors.New("ops!")
		r := ForField("test.path", func(m mockStruct) string { return "value" }).
			Rules(NewSingleRule(func(v string) error { return nil })).
			Rules(NewSingleRule(func(v string) error { return err1 })).
			Rules(NewSingleRule(func(v string) error { return nil })).
			Rules(NewSingleRule(func(v string) error { return err2 }))
		errs := r.Validate(mockStruct{})
		require.Len(t, errs, 1)
		assert.Equal(t, FieldError{
			FieldPath:  "test.path",
			FieldValue: "value",
			Errors:     []string{err1.Error(), err2.Error()},
		}, *errs[0].(*FieldError))
	})
}

type mockStruct struct {
	Field string
}
