package validation

import "github.com/pkg/errors"

// Rule is the interface for all validation rules.
type Rule[T any] interface {
	Validate(v T) error
}

// SingleRule represents a single validation Rule.
// The Message conveys the reason for the rules' failure and IsValid
// is the function which verifies if the rule passes or not.
type SingleRule[T any] struct {
	Message string
	IsValid func(v T) bool
}

func (r SingleRule[T]) Validate(v T) error {
	if r.IsValid(v) {
		return nil
	}
	return errors.New(r.Message)
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

type SingleRuleFunc[T any] func(v T) error

func (r SingleRuleFunc[T]) Validate(v T) error { return r(v) }
