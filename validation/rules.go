package validation

type fieldRules interface {
	Validate() error
}

func RulesForStruct(rules ...fieldRules) StructRules {
	return StructRules{fieldRules: rules}
}

type StructRules struct {
	fieldRules []fieldRules
}

func (s StructRules) Validate() []error {
	var errors []error
	for _, field := range s.fieldRules {
		if err := field.Validate(); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

// RulesForField creates a typed FieldRules instance for the field which access is defined through getter function.
func RulesForField[T any](fieldPath string, getter func() T) FieldRules[T] {
	return FieldRules[T]{fieldPath: fieldPath, getter: getter}
}

// FieldRules is responsible for validating a single struct field.
type FieldRules[T any] struct {
	fieldPath  string
	getter     func() T
	rules      []Rule[T]
	predicates []func() bool
}

func (r FieldRules[T]) Validate() error {
	for _, pred := range r.predicates {
		if pred != nil && !pred() {
			return nil
		}
	}
	fv := r.getter()
	var errors []error
	for i := range r.rules {
		if err := r.rules[i].Validate(fv); err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return NewFieldError(r.fieldPath, fv, errors)
	}
	return nil
}

func (r FieldRules[T]) If(predicate func() bool) FieldRules[T] {
	r.predicates = append(r.predicates, predicate)
	return r
}

func (r FieldRules[T]) With(rules ...Rule[T]) FieldRules[T] {
	r.rules = append(r.rules, rules...)
	return r
}
