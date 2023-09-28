package validation

type Mode uint8

// TODO implement these.
const (
	ModeFailFast Mode = iota + 1
	ModeCollectErrors
)

func RulesForObject(objectMetadata ObjectMetadata, validators ...func() *FieldError) ObjectRules {
	return ObjectRules{
		objectMetadata: objectMetadata,
		validators:     validators,
	}
}

type ObjectRules struct {
	objectMetadata ObjectMetadata
	validators     []func() *FieldError
	mode           Mode
}

func (r ObjectRules) WithMode(mode Mode) ObjectRules {
	r.mode = mode
	return r
}

func (r ObjectRules) Validate() *ObjectError {
	var errors []error
	for _, vf := range r.validators {
		if err := vf(); err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return &ObjectError{
			Object: r.objectMetadata,
			Errors: errors,
		}
	}
	return nil
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

func (r FieldRules[T]) Validate() *FieldError {
	for _, pred := range r.predicates {
		if pred != nil && !pred() {
			return nil
		}
	}
	fv := r.getter()
	var errors []error
	for i := range r.rules {
		if err := r.rules[i].Validate(fv); err != nil {
			if v, ok := err.(multiRuleError); ok {
				errors = append(errors, v...)
			} else {
				errors = append(errors, err)
			}
		}
	}
	if len(errors) > 0 {
		return &FieldError{
			FieldPath:  r.fieldPath,
			FieldValue: fv,
			Errors:     errors,
		}
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
