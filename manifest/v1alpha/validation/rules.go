package validation

import (
	"github.com/nobl9/nobl9-go/manifest"
)

func RulesForObject(object manifest.Object, validators ...func() error) ObjectRules {
	return ObjectRules{
		object:     object,
		validators: validators,
	}
}

type ObjectRules struct {
	object     manifest.Object
	validators []func() error
}

func (v ObjectRules) Validate() error {
	for _, vf := range v.validators {
		if err := vf(); err != nil {
			// TODO: aggregate
			return err
		}
	}
	return nil
}

// RulesForField creates a typed FieldRules instance for the field which access is defined through getter function.
func RulesForField[T any](fieldPath string, getter func() T) FieldRules[T] {
	return FieldRules[T]{getter: getter}
}

// FieldRules is responsible for validating a single struct field.
type FieldRules[T any] struct {
	fieldPath  string
	getter     func() T
	rules      []Rule[T]
	predicates []func() bool
}

func (v FieldRules[T]) Validate() error {
	for _, pred := range v.predicates {
		if pred != nil && !pred() {
			return nil
		}
	}
	fv := v.getter()
	for i := range v.rules {
		if err := v.rules[i].Validate(fv); err != nil {
			// TODO aggregate.
			return err
		}
	}
	return nil
}

func (v FieldRules[T]) If(predicate func() bool) FieldRules[T] {
	v.predicates = append(v.predicates, predicate)
	return v
}

func (v FieldRules[T]) With(rules ...Rule[T]) FieldRules[T] {
	v.rules = append(v.rules, rules...)
	return v
}
