package validation

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		r := New[mockStruct](
			RulesFor[string](func(m mockStruct) string { return "test" }).
				WithName("test").
				Rules(NewSingleRule(func(v string) error { return nil })),
		)
		errs := r.Validate(mockStruct{})
		assert.Empty(t, errs)
	})

	t.Run("errors", func(t *testing.T) {
		err1 := errors.New("1")
		err2 := errors.New("2")
		r := New[mockStruct](
			RulesFor(func(m mockStruct) string { return "test" }).
				WithName("test").
				Rules(NewSingleRule(func(v string) error { return nil })),
			RulesFor(func(m mockStruct) string { return "name" }).
				WithName("test.name").
				Rules(NewSingleRule(func(v string) error { return err1 })),
			RulesFor(func(m mockStruct) string { return "display" }).
				WithName("test.display").
				Rules(NewSingleRule(func(v string) error { return err2 })),
		)
		errs := r.Validate(mockStruct{})
		require.Len(t, errs, 2)
		assert.Equal(t, []error{
			&PropertyError{
				PropertyName:  "test.name",
				PropertyValue: "name",
				Errors:        []string{err1.Error()},
			},
			&PropertyError{
				PropertyName:  "test.display",
				PropertyValue: "display",
				Errors:        []string{err2.Error()},
			},
		}, errs)
	})
}

func TestRulesFor(t *testing.T) {
	t.Run("no predicates, no error", func(t *testing.T) {
		r := RulesFor(func(m mockStruct) string { return "path" }).
			WithName("test.path").
			Rules(NewSingleRule(func(v string) error { return nil }))
		errs := r.Validate(mockStruct{})
		assert.Empty(t, errs)
	})

	t.Run("no predicates, validate", func(t *testing.T) {
		expectedErr := errors.New("ops!")
		r := RulesFor(func(m mockStruct) string { return "path" }).
			WithName("test.path").
			Rules(NewSingleRule(func(v string) error { return expectedErr }))
		errs := r.Validate(mockStruct{})
		require.Len(t, errs, 1)
		assert.Equal(t, PropertyError{
			PropertyName:  "test.path",
			PropertyValue: "path",
			Errors:        []string{expectedErr.Error()},
		}, *errs[0].(*PropertyError))
	})

	t.Run("predicate matches, don't validate", func(t *testing.T) {
		r := RulesFor(func(m mockStruct) string { return "value" }).
			WithName("test.path").
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
		r := RulesFor(func(m mockStruct) string { return "value" }).
			WithName("test.path").
			Rules(NewSingleRule(func(v string) error { return nil })).
			Rules(NewSingleRule(func(v string) error { return err1 })).
			Rules(NewSingleRule(func(v string) error { return nil })).
			Rules(NewSingleRule(func(v string) error { return err2 }))
		errs := r.Validate(mockStruct{})
		require.Len(t, errs, 1)
		assert.Equal(t, PropertyError{
			PropertyName:  "test.path",
			PropertyValue: "value",
			Errors:        []string{err1.Error(), err2.Error()},
		}, *errs[0].(*PropertyError))
	})
}

type mockStruct struct {
	Field string
}
