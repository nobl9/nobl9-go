package validation

import (
	"fmt"
)

// ForEach creates a typed PropertyRules instance for a slice property
// which access is defined through getter function.
func ForEach[T, S any](getter PropertyGetter[[]T, S]) PropertyRulesForEach[T, S] {
	return PropertyRulesForEach[T, S]{getter: getter}
}

// PropertyRulesForEach is responsible for validating a single property.
type PropertyRulesForEach[T, S any] struct {
	name   string
	getter PropertyGetter[[]T, S]
	steps  []interface{}
}

// Validate executes each of the steps sequentially and aggregates the encountered errors.
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
				forEachErrors[i] = forEachElementError{Errors: append(fErrs, err), PropValue: element}
			}
			previousStepFailed = errorEncountered
		case Rule[[]T]:
			propValue = r.getter(st)
			err := v.Validate(propValue)
			if err != nil {
				sliceErrors = append(sliceErrors, err)
			}
			previousStepFailed = err != nil
		case Validator[T]:
			errorEncountered := false
			for i, element := range r.getter(st) {
				err := v.Validate(element)
				if err == nil {
					continue
				}
				errorEncountered = true
				for _, e := range err.Errors {
					e.PrependPropertyName(fmt.Sprintf(sliceElementNameFmt, r.name, i))
					allErrors = append(allErrors, e)
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
			fmt.Sprintf(sliceElementNameFmt, r.name, i),
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

const sliceElementNameFmt = "%s[%d]"

func (r PropertyRulesForEach[T, S]) WithName(name string) PropertyRulesForEach[T, S] {
	r.name = name
	return r
}

func (r PropertyRulesForEach[T, S]) RulesForEach(rules ...Rule[T]) PropertyRulesForEach[T, S] {
	r.steps = appendSteps(r.steps, rules)
	return r
}

type RulesForEachContainer[T, S any] struct {
	prop        PropertyRulesForEach[T, S]
	predicate   Predicate[S]
	stopOnError bool
}

func (r RulesForEachContainer[T, S]) When(predicate Predicate[S]) RulesForEachContainer[T, S] {
	r.predicate = predicate
	return r
}

func (r RulesForEachContainer[T, S]) StopOnError() RulesForEachContainer[T, S] {
	r.stopOnError = true
	return r
}

func (r PropertyRulesForEach[T, S]) With(rules ...Rule[[]T]) PropertyRulesForEach[T, S] {
	r.steps = appendSteps(r.steps, rules)
	return r
}

func (r PropertyRulesForEach[T, S]) Rules(rules ...Rule[[]T]) PropertyRulesForEach[T, S] {
	r.steps = appendSteps(r.steps, rules)
	return r
}

func (r PropertyRulesForEach[T, S]) IncludeForEach(rules ...Validator[T]) PropertyRulesForEach[T, S] {
	r.steps = appendSteps(r.steps, rules)
	return r
}

func (r PropertyRulesForEach[T, S]) When(predicates ...Predicate[S]) PropertyRulesForEach[T, S] {
	r.steps = appendSteps(r.steps, predicates)
	return r
}

func (r PropertyRulesForEach[T, S]) StopOnError() PropertyRulesForEach[T, S] {
	r.steps = append(r.steps, stopOnErrorStep(0))
	return r
}
