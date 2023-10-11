package validation

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidator(t *testing.T) {
	type mockStruct struct {
		Field string
	}

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
				Errors:        []RuleError{{Message: err1.Error()}},
			},
			&PropertyError{
				PropertyName:  "test.display",
				PropertyValue: "display",
				Errors:        []RuleError{{Message: err2.Error()}},
			},
		}, errs)
	})
}
