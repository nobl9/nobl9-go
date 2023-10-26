package validation

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPropertyRulesForEach(t *testing.T) {
	type mockStruct struct {
		Fields []string
	}

	t.Run("no predicates, no error", func(t *testing.T) {
		r := ForEach(func(m mockStruct) []string { return []string{"path"} }).
			WithName("test.path").
			RulesForEach(NewSingleRule(func(v string) error { return nil }))
		errs := r.Validate(mockStruct{})
		assert.Empty(t, errs)
	})

	t.Run("no predicates, validate", func(t *testing.T) {
		expectedErr := errors.New("ops!")
		r := ForEach(func(m mockStruct) []string { return []string{"path"} }).
			WithName("test.path").
			RulesForEach(NewSingleRule(func(v string) error { return expectedErr }))
		errs := r.Validate(mockStruct{})
		require.Len(t, errs, 1)
		assert.Equal(t, &PropertyError{
			PropertyName:  "test.path[0]",
			PropertyValue: "path",
			Errors:        []*RuleError{{Message: expectedErr.Error()}},
		}, errs[0])
	})

	t.Run("predicate matches, don't validate", func(t *testing.T) {
		r := ForEach(func(m mockStruct) []string { return []string{"value"} }).
			WithName("test.path").
			When(func(mockStruct) bool { return true }).
			When(func(mockStruct) bool { return true }).
			When(func(st mockStruct) bool { return len(st.Fields) == 0 }).
			RulesForEach(NewSingleRule(func(v string) error { return errors.New("ops!") }))
		errs := r.Validate(mockStruct{Fields: []string{"something"}})
		assert.Empty(t, errs)
	})

	t.Run("multiple rules and for each rules", func(t *testing.T) {
		err1 := errors.New("oh no!")
		err2 := errors.New("another error...")
		err3 := errors.New("rule error")
		err4 := errors.New("rule error again")
		r := ForEach(func(m mockStruct) []string { return m.Fields }).
			WithName("test.path").
			Rules(NewSingleRule(func(v []string) error { return err3 })).
			RulesForEach(
				NewSingleRule(func(v string) error { return err1 }),
				NewSingleRule(func(v string) error {
					return NewPropertyError("nested", "made-up", err2)
				}),
			).
			Rules(NewSingleRule(func(v []string) error {
				return NewPropertyError("nested", "nestedValue", err4)
			}))

		errs := r.Validate(mockStruct{Fields: []string{"1", "2"}})
		require.Len(t, errs, 6)
		assert.ElementsMatch(t, []*PropertyError{
			{
				PropertyName:  "test.path",
				PropertyValue: `["1","2"]`,
				Errors:        []*RuleError{{Message: err3.Error()}},
			},
			{
				PropertyName:  "test.path.nested",
				PropertyValue: "nestedValue",
				Errors:        []*RuleError{{Message: err4.Error()}},
			},
			{
				PropertyName:  "test.path[0]",
				PropertyValue: "1",
				Errors:        []*RuleError{{Message: err1.Error()}},
			},
			{
				PropertyName:  "test.path[1]",
				PropertyValue: "2",
				Errors:        []*RuleError{{Message: err1.Error()}},
			},
			{
				PropertyName:  "test.path[0].nested",
				PropertyValue: "made-up",
				Errors:        []*RuleError{{Message: err2.Error()}},
			},
			{
				PropertyName:  "test.path[1].nested",
				PropertyValue: "made-up",
				Errors:        []*RuleError{{Message: err2.Error()}},
			},
		}, errs)
	})

	t.Run("stop on error", func(t *testing.T) {
		expectedErr := errors.New("oh no!")
		r := ForEach(func(m mockStruct) []string { return []string{"value"} }).
			WithName("test.path").
			RulesForEach(NewSingleRule(func(v string) error { return expectedErr })).
			StopOnError().
			RulesForEach(NewSingleRule(func(v string) error { return errors.New("no") }))
		errs := r.Validate(mockStruct{})
		require.Len(t, errs, 1)
		assert.Equal(t, &PropertyError{
			PropertyName:  "test.path[0]",
			PropertyValue: "value",
			Errors:        []*RuleError{{Message: expectedErr.Error()}},
		}, errs[0])
	})

	t.Run("include for each validator", func(t *testing.T) {
		err1 := errors.New("oh no!")
		err2 := errors.New("included")
		err3 := errors.New("included again")
		r := ForEach(func(m mockStruct) []string { return m.Fields }).
			WithName("test.path").
			RulesForEach(NewSingleRule(func(v string) error { return err1 })).
			IncludeForEach(New[string](
				For(func(s string) string { return "nested" }).
					WithName("included").
					Rules(
						NewSingleRule(func(v string) error { return err2 }),
						NewSingleRule(func(v string) error { return err3 }),
					),
			))
		errs := r.Validate(mockStruct{Fields: []string{"value"}})
		require.Len(t, errs, 2)
		assert.ElementsMatch(t, []*PropertyError{
			{
				PropertyName:  "test.path[0]",
				PropertyValue: "value",
				Errors:        []*RuleError{{Message: err1.Error()}},
			},
			{
				PropertyName:  "test.path[0].included",
				PropertyValue: "nested",
				Errors: []*RuleError{
					{Message: err2.Error()},
					{Message: err3.Error()},
				},
			},
		}, errs)
	})
}
