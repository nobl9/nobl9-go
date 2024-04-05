package validation

import (
	"fmt"
)

// ForSlice creates a new [PropertyRulesForSlice] instance for a slice property
// which value is extracted through [PropertyGetter] function.
func ForSlice[T, S any](getter PropertyGetter[[]T, S]) PropertyRulesForSlice[T, S] {
	return PropertyRulesForSlice[T, S]{getter: getter}
}

// PropertyRulesForSlice is responsible for validating a single property.
type PropertyRulesForSlice[T, S any] struct {
	name   string
	getter PropertyGetter[[]T, S]
	steps  []interface{}
}

// Validate executes each of the rules sequentially and aggregates the encountered errors.
// nolint: prealloc, gocognit
func (r PropertyRulesForSlice[T, S]) Validate(st S) PropertyErrors {
	var (
		allErrors          PropertyErrors
		sliceErrors        []error
		propValue          []T
		previousStepFailed bool
	)
	sliceElementErrors := make(map[int]sliceElementError)
loop:
	for _, step := range r.steps {
		switch v := step.(type) {
		case stopOnErrorStep:
			if previousStepFailed {
				break loop
			}
		case Predicate[S]:
			if !v(st) {
				break loop
			}
		// Same as Rule[S] as for GetSelf we'd get the same type on T and S.
		case Rule[T]:
			errorEncountered := false
			for i, element := range r.getter(st) {
				err := v.Validate(element)
				if err == nil {
					continue
				}
				errorEncountered = true
				switch ev := err.(type) {
				case *PropertyError:
					ev.IsSliceElementError = true
					allErrors = append(allErrors, ev.PrependPropertyName(SliceElementName(r.name, i)))
				default:
					fErrs := sliceElementErrors[i].Errors
					sliceElementErrors[i] = sliceElementError{Errors: append(fErrs, err), PropValue: element}
				}
			}
			previousStepFailed = errorEncountered
		case Rule[[]T]:
			propValue = r.getter(st)
			err := v.Validate(propValue)
			if err != nil {
				switch ev := err.(type) {
				case *PropertyError:
					ev.IsSliceElementError = true
					allErrors = append(allErrors, ev.PrependPropertyName(r.name))
				default:
					sliceErrors = append(sliceErrors, err)
				}
			}
			previousStepFailed = err != nil
		case validatorI[T]:
			errorEncountered := false
			for i, element := range r.getter(st) {
				err := v.Validate(element)
				if err == nil {
					continue
				}
				errorEncountered = true
				for _, e := range err.Errors {
					e.IsSliceElementError = true
					allErrors = append(allErrors, e.PrependPropertyName(SliceElementName(r.name, i)))
				}
			}
			previousStepFailed = errorEncountered
		}
	}
	if len(sliceErrors) > 0 {
		allErrors = append(allErrors, NewPropertyError(r.name, propValue, sliceErrors...))
	}
	for i, element := range sliceElementErrors {
		pErr := NewPropertyError(
			SliceElementName(r.name, i),
			element.PropValue,
			element.Errors...)
		pErr.IsSliceElementError = true
		allErrors = append(allErrors, pErr)
	}
	if len(allErrors) > 0 {
		return allErrors
	}
	return nil
}

type sliceElementError struct {
	PropValue interface{}
	Errors    []error
}

func (r PropertyRulesForSlice[T, S]) WithName(name string) PropertyRulesForSlice[T, S] {
	r.name = name
	return r
}

func (r PropertyRulesForSlice[T, S]) RulesForEach(rules ...Rule[T]) PropertyRulesForSlice[T, S] {
	r.steps = appendSteps(r.steps, rules)
	return r
}

func (r PropertyRulesForSlice[T, S]) Rules(rules ...Rule[[]T]) PropertyRulesForSlice[T, S] {
	r.steps = appendSteps(r.steps, rules)
	return r
}

func (r PropertyRulesForSlice[T, S]) When(predicate Predicate[S]) PropertyRulesForSlice[T, S] {
	r.steps = append(r.steps, predicate)
	return r
}

func (r PropertyRulesForSlice[T, S]) IncludeForEach(rules ...Validator[T]) PropertyRulesForSlice[T, S] {
	r.steps = appendSteps(r.steps, rules)
	return r
}

func (r PropertyRulesForSlice[T, S]) StopOnError() PropertyRulesForSlice[T, S] {
	r.steps = append(r.steps, stopOnErrorStep(0))
	return r
}

func SliceElementName(sliceName string, index int) string {
	if sliceName == "" {
		return fmt.Sprintf("[%d]", index)
	}
	return fmt.Sprintf("%s[%d]", sliceName, index)
}
