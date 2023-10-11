package validation

import "github.com/pkg/errors"

// Rule is the interface for all validation rules.
type Rule[T any] interface {
	Validate(v T) error
}

func NewSingleRule[T any](validate func(v T) error) SingleRule[T] {
	return SingleRule[T]{validate: validate}
}

type SingleRule[T any] struct {
	validate  func(v T) error
	errorCode ErrorCode
}

func (r SingleRule[T]) Validate(v T) error {
	if err := r.validate(v); err != nil {
		return RuleError{Message: err.Error(), Code: r.errorCode}
	}
	return nil
}

func (r SingleRule[T]) WithErrorCode(code ErrorCode) SingleRule[T] {
	r.errorCode = code
	return r
}

func NewRuleSet[T any](rules ...Rule[T]) RuleSet[T] {
	return RuleSet[T]{rules: rules}
}

// RuleSet allows defining Rule which aggregates multiple sub-rules.
type RuleSet[T any] struct {
	rules     []Rule[T]
	errorCode ErrorCode
}

func (r RuleSet[T]) Validate(v T) error {
	var errs ruleSetError
	for i := range r.rules {
		if err := r.rules[i].Validate(v); err != nil {
			var ruleErr RuleError
			if errors.As(err, &ruleErr) {
				errs = append(errs, ruleErr.AddCode(r.errorCode))
			} else {
				errs = append(errs, RuleError{
					Message: err.Error(),
					Code:    r.errorCode,
				})
			}
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func (r RuleSet[T]) WithErrorCode(code ErrorCode) RuleSet[T] {
	r.errorCode = code
	return r
}
