package validation

import (
	"fmt"

	"github.com/pkg/errors"
)

// RulesForEach creates a typed PropertyRules instance for a slice property
// which access is defined through getter function.
func RulesForEach[T, S any](getter PropertyGetter[[]T, S]) PropertyRulesForEach[T, S] {
	return PropertyRulesForEach[T, S]{getter: getter}
}

// PropertyRules is responsible for validating a single property.
type PropertyRulesForEach[T, S any] struct {
	name   string
	getter PropertyGetter[[]T, S]
	steps  []interface{}
}

func (r PropertyRulesForEach[T, S]) WithName(name string) PropertyRulesForEach[T, S] {
	r.name = name
	return r
}

type forEachElementError struct {
	PropValue interface{}
	Errors    []error
}

func (r PropertyRulesForEach[T, S]) Validate(st S) []error {
	var (
		allErrors          []error
		previousStepFailed bool
	)
	ruleErrors := make(map[int]forEachElementError, 0)
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
			errorEncounterd := false
			for i, element := range r.getter(st) {
				err := v.Validate(element)
				if err != nil {
					errs := ruleErrors[i].Errors
					errs = append(errs, err)
					ruleErrors[i] = forEachElementError{Errors: errs, PropValue: element}
					errorEncounterd = true
				}
			}
			previousStepFailed = errorEncounterd
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
	for i, element := range ruleErrors {
		allErrors = append(allErrors, NewPropertyError(
			fmt.Sprintf(sliceElementNameFmt, r.name, i),
			element.PropValue,
			element.Errors))
	}
	return allErrors
}

const sliceElementNameFmt = "%s[%d]"

func (r PropertyRulesForEach[T, S]) Rules(rules ...Rule[T]) PropertyRulesForEach[T, S] {
	r.steps = appendSteps(r.steps, rules)
	return r
}

func (r PropertyRulesForEach[T, S]) Include(rules ...Validator[T]) PropertyRulesForEach[T, S] {
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
