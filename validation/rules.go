package validation

import (
	"reflect"
)

// For creates a typed PropertyRules instance for the property which access is defined through getter function.
func For[T, S any](getter PropertyGetter[T, S]) PropertyRules[T, S] {
	return PropertyRules[T, S]{getter: func(s S) (v T, isEmpty bool) { return getter(s), false }}
}

// ForPointer accepts a getter function returning a pointer and wraps its call in order to
// safely extract the value under the pointer or return a zero value for a give type T.
// If required is set to true, the nil pointer value will result in an error and the
// validation will not proceed.
func ForPointer[T, S any](getter PropertyGetter[*T, S]) PropertyRules[T, S] {
	return PropertyRules[T, S]{getter: func(s S) (indirect T, isEmpty bool) {
		ptr := getter(s)
		if ptr != nil {
			return *ptr, false
		}
		zv := *new(T)
		return zv, true
	}, isPointer: true}
}

// GetSelf is a convenience method for extracting 'self' property of a validated value.
func GetSelf[S any]() PropertyGetter[S, S] {
	return func(s S) S { return s }
}

type Predicate[S any] func(S) bool

type PropertyGetter[T, S any] func(S) T

type optionalPropertyGetter[T, S any] func(S) (v T, isEmpty bool)

// PropertyRules is responsible for validating a single property.
type PropertyRules[T, S any] struct {
	name      string
	getter    optionalPropertyGetter[T, S]
	steps     []interface{}
	required  bool
	omitempty bool
	isPointer bool
}

func (r PropertyRules[T, S]) Validate(st S) PropertyErrors {
	var (
		ruleErrors         []error
		allErrors          PropertyErrors
		previousStepFailed bool
	)
	propValue, isEmpty := r.getter(st)
	isEmpty = isEmpty || (!r.isPointer && isEmptyFunc(propValue))
	if r.required && isEmpty {
		return PropertyErrors{NewPropertyError(r.name, propValue, NewRequiredError())}
	}
	if isEmpty && (r.omitempty || r.isPointer) {
		return nil
	}
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
			err := v.Validate(propValue)
			if err != nil {
				switch ev := err.(type) {
				case *PropertyError:
					allErrors = append(allErrors, ev.PrependPropertyName(r.name))
				default:
					ruleErrors = append(ruleErrors, err)
				}
			}
			previousStepFailed = err != nil
		case validatorI[T]:
			err := v.Validate(propValue)
			if err != nil {
				for _, e := range err.Errors {
					allErrors = append(allErrors, e.PrependPropertyName(r.name))
				}
			}
			previousStepFailed = err != nil
		}
	}
	if len(ruleErrors) > 0 {
		allErrors = append(allErrors, NewPropertyError(r.name, propValue, ruleErrors...))
	}
	if len(allErrors) > 0 {
		return allErrors
	}
	return nil
}

func (r PropertyRules[T, S]) WithName(name string) PropertyRules[T, S] {
	r.name = name
	return r
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

func (r PropertyRules[T, S]) Required() PropertyRules[T, S] {
	r.required = true
	return r
}

func (r PropertyRules[T, S]) Omitempty() PropertyRules[T, S] {
	r.omitempty = true
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

// isEmptyFunc checks only the types which it makes sense for.
// It's hard to consider 0 an empty value for anything really.
func isEmptyFunc(v interface{}) bool {
	return reflect.ValueOf(v).IsZero()
}
