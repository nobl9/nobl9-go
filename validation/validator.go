package validation

type validatorI[S any] interface {
	Validate(s S) *ValidatorError
}

type propertyRulesI[S any] interface {
	Validate(s S) PropertyErrors
}

// New creates a new [Validator] aggregating the provided property rules.
func New[S any](props ...propertyRulesI[S]) Validator[S] {
	return Validator[S]{props: props}
}

// Validator is the top level validation entity.
// It serves as an aggregator for [PropertyRules].
type Validator[S any] struct {
	props      []propertyRulesI[S]
	name       string
	predicates []Predicate[S]
}

// WithName when a rule fails will pass the provided name to [ValidatorError.WithName].
func (v Validator[S]) WithName(name string) Validator[S] {
	v.name = name
	return v
}

// When defines accepts predicates which will be evaluated BEFORE [Validator] validates ANY rules.
func (v Validator[S]) When(predicates ...Predicate[S]) Validator[S] {
	v.predicates = append(v.predicates, predicates...)
	return v
}

// Validate will first evaluate predicates before validating any rules.
// If any predicate does not pass the validation won't be executed (returns nil).
// All errors returned by property rules will be aggregated and wrapped in [ValidatorError].
func (v Validator[S]) Validate(st S) *ValidatorError {
	for _, predicate := range v.predicates {
		if !predicate(st) {
			return nil
		}
	}
	var allErrors PropertyErrors
	for _, rules := range v.props {
		if errs := rules.Validate(st); len(errs) > 0 {
			allErrors = append(allErrors, errs...)
		}
	}
	if len(allErrors) != 0 {
		return NewValidatorError(allErrors).WithName(v.name)
	}
	return nil
}
