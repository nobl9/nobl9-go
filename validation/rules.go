package validation

type fieldRules[S any] interface {
	Validate(s S) error
}

func RulesForStruct[S any](rules ...fieldRules[S]) StructRules[S] {
	return StructRules[S]{fieldRules: rules}
}

type StructRules[S any] struct {
	fieldRules []fieldRules[S]
}

func (r StructRules[S]) Validate(st S) []error {
	var errors []error
	for _, field := range r.fieldRules {
		if err := field.Validate(st); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

// RulesForField creates a typed FieldRules instance for the field which access is defined through getter function.
func RulesForField[T, S any](fieldPath string, getter func(S) T) FieldRules[T, S] {
	return FieldRules[T, S]{fieldPath: fieldPath, getter: getter}
}

// FieldRules is responsible for validating a single struct field.
type FieldRules[T, S any] struct {
	fieldPath  string
	getter     func(S) T
	rules      []Rule[T]
	predicates []func() bool
}

func (r FieldRules[T, S]) Validate(st S) error {
	for _, pred := range r.predicates {
		if pred != nil && !pred() {
			return nil
		}
	}
	fv := r.getter(st)
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

func (r FieldRules[T, S]) If(predicate func() bool) FieldRules[T, S] {
	r.predicates = append(r.predicates, predicate)
	return r
}

func (r FieldRules[T, S]) With(rules ...Rule[T]) FieldRules[T, S] {
	r.rules = append(r.rules, rules...)
	return r
}
