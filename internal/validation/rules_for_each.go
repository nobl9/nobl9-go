package validation

import (
	"fmt"
)

// ForEach creates a new [PropertyRulesForEach] instance for a slice property
// which value is extracted through [PropertyGetter] function.
func ForEach[T, S any](getter PropertyGetter[[]T, S]) PropertyRulesForEach[T, S] {
	return PropertyRulesForEach[T, S]{getter: getter}
}

// PropertyRulesForEach is responsible for validating a single property.
type PropertyRulesForEach[T, S any] struct {
	name   string
	getter PropertyGetter[[]T, S]
	steps  []interface{}
}

// Validate executes each of the rules sequentially and aggregates the encountered errors.
// nolint: prealloc, gocognit
func (r PropertyRulesForEach[T, S]) Validate(st S) PropertyErrors {
	var (
		allErrors          PropertyErrors
		sliceErrors        []error
		propValue          []T
		previousStepFailed bool
	)
	forEachErrors := make(map[int]forEachElementError)
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
				fErrs := forEachErrors[i].Errors
				switch ev := err.(type) {
				case *PropertyError:
					allErrors = append(allErrors, ev.PrependPropertyName(SliceElementName(r.name, i)))
				default:
					forEachErrors[i] = forEachElementError{Errors: append(fErrs, err), PropValue: element}
				}
			}
			previousStepFailed = errorEncountered
		case Rule[[]T]:
			propValue = r.getter(st)
			err := v.Validate(propValue)
			if err != nil {
				switch ev := err.(type) {
				case *PropertyError:
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
					allErrors = append(allErrors, e.PrependPropertyName(SliceElementName(r.name, i)))
				}
			}
			previousStepFailed = errorEncountered
		}
	}
	if len(sliceErrors) > 0 {
		allErrors = append(allErrors, NewPropertyError(r.name, propValue, sliceErrors...))
	}
	for i, element := range forEachErrors {
		allErrors = append(allErrors, NewPropertyError(
			SliceElementName(r.name, i),
			element.PropValue,
			element.Errors...))
	}
	if len(allErrors) > 0 {
		return allErrors
	}
	return nil
}

type forEachElementError struct {
	PropValue interface{}
	Errors    []error
}

func (r PropertyRulesForEach[T, S]) WithName(name string) PropertyRulesForEach[T, S] {
	r.name = name
	return r
}

func (r PropertyRulesForEach[T, S]) RulesForEach(rules ...Rule[T]) PropertyRulesForEach[T, S] {
	r.steps = appendSteps(r.steps, rules)
	return r
}

func (r PropertyRulesForEach[T, S]) Rules(rules ...Rule[[]T]) PropertyRulesForEach[T, S] {
	r.steps = appendSteps(r.steps, rules)
	return r
}

func (r PropertyRulesForEach[T, S]) When(predicate Predicate[S]) PropertyRulesForEach[T, S] {
	r.steps = append(r.steps, predicate)
	return r
}

func (r PropertyRulesForEach[T, S]) IncludeForEach(rules ...Validator[T]) PropertyRulesForEach[T, S] {
	r.steps = appendSteps(r.steps, rules)
	return r
}

func (r PropertyRulesForEach[T, S]) StopOnError() PropertyRulesForEach[T, S] {
	r.steps = append(r.steps, stopOnErrorStep(0))
	return r
}

func SliceElementName(sliceName string, index int) string {
	return fmt.Sprintf("%s[%d]", sliceName, index)
}
