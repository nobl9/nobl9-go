package validation

import (
	"fmt"

	"github.com/pkg/errors"
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
func (r PropertyRulesForEach[T, S]) Validate(st S) []error {
	var (
		allErrors, sliceErrors []error
		propValue              []T
		previousStepFailed     bool
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
				if err != nil {
					errs := forEachErrors[i].Errors
					errs = append(errs, err)
					forEachErrors[i] = forEachElementError{Errors: errs, PropValue: element}
					errorEncountered = true
				}
			}
			previousStepFailed = errorEncountered
		case Rule[[]T]:
			propValue = r.getter(st)
			err := v.Validate(propValue)
			if err != nil {
				sliceErrors = append(sliceErrors, err)
			}
			previousStepFailed = err != nil
		case Rules[T]:
			for i, element := range r.getter(st) {
				errs := v.Validate(element)
				for _, err := range errs {
					var fErr *PropertyError
					if ok := errors.As(err, &fErr); ok {
						fErr.PrependPropertyName(fmt.Sprintf(sliceElementNameFmt, r.name, i))
					}
					allErrors = append(allErrors, err)
				}
				previousStepFailed = len(errs) > 0
			}
		}
	}
	if len(sliceErrors) > 0 {
		allErrors = append(allErrors, NewPropertyError(r.name, propValue, sliceErrors))
	}
	for i, element := range forEachErrors {
		allErrors = append(allErrors, NewPropertyError(
			fmt.Sprintf(sliceElementNameFmt, r.name, i),
			element.PropValue,
			element.Errors))
	}
	return allErrors
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
