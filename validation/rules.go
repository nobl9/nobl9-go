package validation

import "github.com/pkg/errors"

// RulesFor creates a typed PropertyRules instance for the property which access is defined through getter function.
func RulesFor[T, S any](getter PropertyGetter[T, S]) PropertyRules[T, S] {
	return PropertyRules[T, S]{getter: getter}
}

// GetSelf is a convenience method for extracting 'self' property of a validated value.
func GetSelf[S any]() PropertyGetter[S, S] {
	return func(s S) S { return s }
}

type Predicate[S any] func(S) bool

type PropertyGetter[T, S any] func(S) T

// PropertyRules is responsible for validating a single property.
type PropertyRules[T, S any] struct {
	name   string
	getter PropertyGetter[T, S]
	steps  []interface{}
}

func (r PropertyRules[T, S]) WithName(name string) PropertyRules[T, S] {
	r.name = name
	return r
}

func (r PropertyRules[T, S]) Validate(st S) []error {
	var (
		ruleErrors, allErrors []error
		propValue             T
		previousStepFailed    bool
	)
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
			propValue = r.getter(st)
			err := v.Validate(propValue)
			if err != nil {
				ruleErrors = append(ruleErrors, err)
			}
			previousStepFailed = err != nil
		case Rules[T]:
			structValue := r.getter(st)
			errs := v.Validate(structValue)
			for _, err := range errs {
				var fErr *PropertyError
				if ok := errors.As(err, &fErr); ok {
					fErr.PrependPropertyName(r.name)
				}
				allErrors = append(allErrors, err)
			}
			previousStepFailed = len(errs) > 0
		}
	}
	if len(ruleErrors) > 0 {
		allErrors = append(allErrors, NewPropertyError(r.name, propValue, ruleErrors))
	}
	return allErrors
}

func (r PropertyRules[T, S]) Rules(rules ...Rule[T]) PropertyRules[T, S] {
	r.steps = appendSteps(r.steps, rules)
	return r
}

func (r PropertyRules[T, S]) Include(rules ...Validator[T]) PropertyRules[T, S] {
	r.steps = appendSteps(r.steps, rules)
	return r
}

func (r PropertyRules[T, S]) When(predicates ...Predicate[S]) PropertyRules[T, S] {
	r.steps = appendSteps(r.steps, predicates)
	return r
}

type stopOnErrorStep uint8

func (r PropertyRules[T, S]) StopOnError() PropertyRules[T, S] {
	r.steps = append(r.steps, stopOnErrorStep(0))
	return r
}

func appendSteps[T any](slice []interface{}, steps []T) []interface{} {
	for _, step := range steps {
		slice = append(slice, step)
	}
	return slice
}
