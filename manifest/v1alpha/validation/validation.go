package validation

import (
	"github.com/pkg/errors"
)

// RulesFor creates a typed Validator instance for the field which access is defined through getter function.
func RulesFor[T any](getter func() T) Validator[T] {
	return Validator[T]{getter: getter}
}

// Rule is the interface for all validation rules.
type Rule[T any] interface {
	Validate(v T) error
}

// SingleRule represents a single validation Rule.
// The Error conveys the reason for the rules' failure and IsValid
// is the function which verifies if the rule passes or not.
type SingleRule[T any] struct {
	Error   string
	IsValid func(v T) bool
}

func (r SingleRule[T]) Validate(v T) error {
	if r.IsValid(v) {
		return nil
	}
	return errors.New(r.Error)
}

// MultiRule allows defining Rule which aggregates multiple sub-rules.
type MultiRule[T any] struct {
	Rules []Rule[T]
}

func (r MultiRule[T]) Validate(v T) error {
	for i := range r.Rules {
		if err := r.Rules[i].Validate(v); err != nil {
			// TODO: aggregate
			return err
		}
	}
	return nil
}

type Validator[T any] struct {
	getter     func() T
	rules      []Rule[T]
	predicates []func() bool
}

func (v Validator[T]) Validate() error {
	for _, pred := range v.predicates {
		if pred != nil && !pred() {
			return nil
		}
	}
	f := v.getter()
	for i := range v.rules {
		if err := v.rules[i].Validate(f); err != nil {
			// TODO aggregate.
			return err
		}
	}
	return nil
}

func (v Validator[T]) If(predicate func() bool) Validator[T] {
	v.predicates = append(v.predicates, predicate)
	return v
}

func (v Validator[T]) With(rules ...Rule[T]) Validator[T] {
	v.rules = append(v.rules, rules...)
	return v
}
