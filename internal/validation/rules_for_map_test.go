package validation

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPropertyRulesForMap(t *testing.T) {
	type mockStruct struct {
		Map map[string]string
	}

	t.Run("no predicates, no error", func(t *testing.T) {
		baseRules := ForMap(func(m mockStruct) map[string]string { return map[string]string{"key": "value"} }).
			WithName("test.path")
		for _, r := range []PropertyRulesForMap[map[string]string, string, string, mockStruct]{
			baseRules.RulesForKeys(NewSingleRule(func(v string) error { return nil })),
			baseRules.RulesForValues(NewSingleRule(func(v string) error { return nil })),
			baseRules.RulesForItems(NewSingleRule(func(v MapItem[string, string]) error { return nil })),
		} {
			errs := r.Validate(mockStruct{})
			assert.Nil(t, errs)
		}
	})

	t.Run("no predicates, validate", func(t *testing.T) {
		expectedErr := errors.New("ops!")
		baseRules := ForMap(func(m mockStruct) map[string]string { return map[string]string{"key": "value"} }).
			WithName("test.path")
		for _, r := range []PropertyRulesForMap[map[string]string, string, string, mockStruct]{
			baseRules.RulesForKeys(NewSingleRule(func(v string) error { return expectedErr })),
			baseRules.RulesForValues(NewSingleRule(func(v string) error { return expectedErr })),
			baseRules.RulesForItems(NewSingleRule(func(v MapItem[string, string]) error { return expectedErr })),
		} {
			errs := r.Validate(mockStruct{})
			require.Len(t, errs, 1)
			assert.Equal(t, &PropertyError{
				PropertyName:  "test.path.key",
				PropertyValue: "key",
				Errors:        []*RuleError{{Message: expectedErr.Error()}},
			}, errs[0])
		}
	})

	t.Run("predicate matches, don't validate", func(t *testing.T) {
		baseRules := ForMap(func(m mockStruct) map[string]string { return map[string]string{"key": "value"} }).
			WithName("test.path").
			When(func(mockStruct) bool { return true }).
			When(func(mockStruct) bool { return true }).
			When(func(st mockStruct) bool { return len(st.Map) == 0 })
		for _, r := range []PropertyRulesForMap[map[string]string, string, string, mockStruct]{
			baseRules.RulesForKeys(NewSingleRule(func(v string) error { return errors.New("ops!") })),
			baseRules.RulesForValues(NewSingleRule(func(v string) error { return errors.New("ops!") })),
			baseRules.RulesForItems(NewSingleRule(func(v MapItem[string, string]) error { return errors.New("ops!") })),
		} {
			errs := r.Validate(mockStruct{Map: map[string]string{"different": "map"}})
			assert.Nil(t, errs)
		}
	})

	t.Run("multiple rules for keys, values and items", func(t *testing.T) {
		err1 := errors.New("rule error")
		err2 := errors.New("key error")
		err3 := errors.New("nested key error")
		err4 := errors.New("value error")
		err5 := errors.New("nested value error")
		r := ForMap(func(m mockStruct) map[string]string { return m.Map }).
			WithName("test.path").
			Rules(NewSingleRule(func(v map[string]string) error { return err1 })).
			RulesForKeys(
				NewSingleRule(func(v string) error { return err2 }),
				NewSingleRule(func(v string) error {
					return NewPropertyError("nested", "nestedKey", err3)
				}),
			).
			RulesForValues(
				NewSingleRule(func(v string) error { return err4 }),
				NewSingleRule(func(v string) error {
					return NewPropertyError("nested", "nestedValue", err5)
				}),
			)
		//RulesForItems(
		//	NewSingleRule(func(v MapItem[string, string]) error { return err1 }),
		//).
		//Rules(NewSingleRule(func(v map[string]string) error {
		//	return NewPropertyError("nested", "nestedValue", err4)
		//}))

		errs := r.Validate(mockStruct{Map: map[string]string{
			"key1": "value1",
			"key2": "value2",
		}})
		require.Len(t, errs, 7)
		assert.ElementsMatch(t, []*PropertyError{
			{
				PropertyName:  "test.path",
				PropertyValue: `{"key1":"value1","key2":"value2"}`,
				Errors:        []*RuleError{{Message: err1.Error()}},
			},
			{
				PropertyName:  "test.path.key1",
				PropertyValue: "key1",
				Errors:        []*RuleError{{Message: err2.Error()}},
			},
			{
				PropertyName:  "test.path.key2",
				PropertyValue: "key2",
				Errors:        []*RuleError{{Message: err2.Error()}},
			},
			{
				PropertyName:  "test.path.key1.nested",
				PropertyValue: "nestedKey",
				Errors:        []*RuleError{{Message: err3.Error()}},
			},
			{
				PropertyName:  "test.path.key2.nested",
				PropertyValue: "nestedKey",
				Errors:        []*RuleError{{Message: err3.Error()}},
			},
			{
				PropertyName:  "test.path.key1",
				PropertyValue: "value1",
				Errors:        []*RuleError{{Message: err4.Error()}},
			},
			{
				PropertyName:  "test.path.key2",
				PropertyValue: "value2",
				Errors:        []*RuleError{{Message: err4.Error()}},
			},
			{
				PropertyName:  "test.path.key1.nested",
				PropertyValue: "nestedValue",
				Errors:        []*RuleError{{Message: err5.Error()}},
			},
			{
				PropertyName:  "test.path.key2.nested",
				PropertyValue: "nestedValue",
				Errors:        []*RuleError{{Message: err5.Error()}},
			},
		}, errs)
	})

	//t.Run("stop on error", func(t *testing.T) {
	//	expectedErr := errors.New("oh no!")
	//	r := ForEach(func(m mockStruct) []string { return []string{"value"} }).
	//		WithName("test.path").
	//		RulesForEach(NewSingleRule(func(v string) error { return expectedErr })).
	//		StopOnError().
	//		RulesForEach(NewSingleRule(func(v string) error { return errors.New("no") }))
	//	errs := r.Validate(mockStruct{})
	//	require.Len(t, errs, 1)
	//	assert.Equal(t, &PropertyError{
	//		PropertyName:  "test.path[0]",
	//		PropertyValue: "value",
	//		Errors:        []*RuleError{{Message: expectedErr.Error()}},
	//	}, errs[0])
	//})
	//
	//t.Run("include for keys validator", func(t *testing.T) {
	//	err1 := errors.New("oh no!")
	//	err2 := errors.New("included")
	//	err3 := errors.New("included again")
	//	r := ForMap(func(m mockStruct) []string { return m.Map }).
	//		WithName("test.path").
	//		Rules(NewSingleRule(func(v string) error { return err1 })).
	//		IncludeForKeys(New[string](
	//			For(func(s string) string { return "nested" }).
	//				WithName("included").
	//				Rules(
	//					NewSingleRule(func(v string) error { return err2 }),
	//					NewSingleRule(func(v string) error { return err3 }),
	//				),
	//		))
	//	errs := r.Validate(mockStruct{Map: map[string]string{"key": "value"}})
	//	require.Len(t, errs, 2)
	//	assert.ElementsMatch(t, []*PropertyError{
	//		{
	//			PropertyName:  "test.path[0]",
	//			PropertyValue: "value",
	//			Errors:        []*RuleError{{Message: err1.Error()}},
	//		},
	//		{
	//			PropertyName:  "test.path[0].included",
	//			PropertyValue: "nested",
	//			Errors: []*RuleError{
	//				{Message: err2.Error()},
	//				{Message: err3.Error()},
	//			},
	//		},
	//	}, errs)
	//})
}
