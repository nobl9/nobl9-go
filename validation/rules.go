package validation

import "github.com/pkg/errors"

type Rules[S any] interface {
	CascadeModeGetter
	Validate(v S) []error
}

func New[S any](rules ...Rules[S]) Validator[S] {
	return Validator[S]{rules: rules}
}

type Validator[S any] struct {
	rules       []Rules[S]
	cascadeMode CascadeMode
}

func (v Validator[S]) CascadeMode(mode CascadeMode) Validator[S] {
	v.cascadeMode = mode
	return v
}

func (v Validator[S]) GetCascadeMode() CascadeMode {
	return v.cascadeMode
}

func (v Validator[S]) Validate(st S) []error {
	var allErrors []error
	for _, rule := range v.rules {
		if errs := rule.Validate(st); len(errs) > 0 {
			allErrors = append(allErrors, errs...)
		}
	}
	return allErrors
}

// RulesFor creates a typed PropertyRules instance for the property which access is defined through getter function.
func RulesFor[T, S any](getter PropertyGetter[S, T]) PropertyRules[T, S] {
	return PropertyRules[T, S]{getter: getter}
}

// GetSelf is a convenience method for extracting 'self' property of a validated value.
func GetSelf[S any]() PropertyGetter[S, S] {
	return func(s S) S { return s }
}

type Predicate[S any] func(S) bool

type PropertyGetter[S, T any] func(S) T

// PropertyRules is responsible for validating a single property.
type PropertyRules[T, S any] struct {
	name        string
	getter      PropertyGetter[S, T]
	steps       []interface{}
	cascadeMode CascadeMode
}

func (r PropertyRules[T, S]) WithName(name string) PropertyRules[T, S] {
	r.name = name
	return r
}

func (r PropertyRules[T, S]) Validate(st S) []error {
	var (
		ruleErrors, allErrors []error
		propValue             T
	)
loop:
	for _, step := range r.steps {
		switch v := step.(type) {
		case Predicate[S]:
			if !v(st) {
				break loop
			}
		// Same as Rule[S] as for GetSelf we'd get the same type on T and S.
		case Rule[T]:
			propValue = r.getter(st)
			if err := v.Validate(propValue); err != nil {
				ruleErrors = append(ruleErrors, err)
			}
		case Rules[T]:
			structValue := r.getter(st)
			errs := v.Validate(structValue)
			for _, err := range errs {
				var fErr *PropertyError
				if ok := errors.As(err, &fErr); ok {
					fErr.PrependPropertyPath(r.name)
				}
				allErrors = append(allErrors, err)
			}
		}
		if (len(ruleErrors) > 0 || len(allErrors) > 0) && r.cascadeMode == CascadeModeStop {
			break
		}
	}
	if len(ruleErrors) > 0 {
		allErrors = append(allErrors, NewPropertyError(r.name, propValue, ruleErrors))
	}
	return allErrors
}

func (r PropertyRules[T, S]) Rules(rules ...Rule[T]) PropertyRules[T, S] {
	for _, rule := range rules {
		r.steps = append(r.steps, rule)
	}
	return r
}

func (r PropertyRules[T, S]) Self(rules ...Rule[S]) PropertyRules[T, S] {
	for _, rule := range rules {
		r.steps = append(r.steps, rule)
	}
	return r
}

func (r PropertyRules[T, S]) Include(rules ...Validator[T]) PropertyRules[T, S] {
	for _, rule := range rules {
		r.steps = append(r.steps, rule)
	}
	return r
}

func (r PropertyRules[T, S]) When(predicates ...Predicate[S]) PropertyRules[T, S] {
	for _, predicate := range predicates {
		r.steps = append(r.steps, predicate)
	}
	return r
}

func (r PropertyRules[T, S]) CascadeMode(mode CascadeMode) PropertyRules[T, S] {
	r.cascadeMode = mode
	return r
}

func (r PropertyRules[T, S]) GetCascadeMode() CascadeMode {
	return r.cascadeMode
}
