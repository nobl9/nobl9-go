package validation

import "github.com/pkg/errors"

type Rules[S any] interface {
	CascadeModeGetter
	Validate(v S) []error
}

func ForStruct[S any](rules ...Rules[S]) StructRules[S] {
	return StructRules[S]{rules: rules}
}

type StructRules[S any] struct {
	rules       []Rules[S]
	cascadeMode CascadeMode
}

func (r StructRules[S]) CascadeMode(mode CascadeMode) StructRules[S] {
	r.cascadeMode = mode
	return r
}

func (r StructRules[S]) GetCascadeMode() CascadeMode {
	return r.cascadeMode
}

func (r StructRules[S]) Validate(st S) []error {
	var allErrors []error
	for _, field := range r.rules {
		if errs := field.Validate(st); len(errs) > 0 {
			allErrors = append(allErrors, errs...)
		}
	}
	return allErrors
}

// ForField creates a typed FieldRules instance for the field which access is defined through getter function.
func ForField[T, S any](fieldPath string, getter func(S) T) FieldRules[T, S] {
	return FieldRules[T, S]{fieldPath: fieldPath, getter: getter}
}

// ForSelf creates a typed FieldRules instance for the field which access is defined through getter function.
func ForSelf[S any]() FieldRules[S, S] {
	return FieldRules[S, S]{getter: func(s S) S { return s }}
}

type Predicate[S any] func(S) bool

// FieldRules is responsible for validating a single struct field.
type FieldRules[T, S any] struct {
	fieldPath   string
	getter      func(S) T
	operations  []interface{}
	cascadeMode CascadeMode
}

func (r FieldRules[T, S]) Validate(st S) []error {
	var (
		ruleErrors, allErrors []error
		fieldValue            T
	)
loop:
	for _, op := range r.operations {
		switch v := op.(type) {
		case Predicate[S]:
			if !v(st) {
				break loop
			}
		// Same as Rule[S] as for ForSelf we'd get the same type on T and S.
		case Rule[T]:
			fieldValue = r.getter(st)
			if err := v.Validate(fieldValue); err != nil {
				ruleErrors = append(ruleErrors, err)
			}
		case Rules[T]:
			structValue := r.getter(st)
			errs := v.Validate(structValue)
			for _, err := range errs {
				var fErr *FieldError
				if ok := errors.As(err, &fErr); ok {
					fErr.PrependFieldPath(r.fieldPath)
				}
				allErrors = append(allErrors, err)
			}
		}
		if (len(ruleErrors) > 0 || len(allErrors) > 0) && r.cascadeMode == CascadeModeStop {
			break
		}
	}
	if len(ruleErrors) > 0 {
		allErrors = append(allErrors, NewFieldError(r.fieldPath, fieldValue, ruleErrors))
	}
	return allErrors
}

func (r FieldRules[T, S]) Rules(rules ...Rule[T]) FieldRules[T, S] {
	for _, rule := range rules {
		r.operations = append(r.operations, rule)
	}
	return r
}

func (r FieldRules[T, S]) Self(rules ...Rule[S]) FieldRules[T, S] {
	for _, rule := range rules {
		r.operations = append(r.operations, rule)
	}
	return r
}

func (r FieldRules[T, S]) Include(rules ...StructRules[T]) FieldRules[T, S] {
	for _, rule := range rules {
		r.operations = append(r.operations, rule)
	}
	return r
}

func (r FieldRules[T, S]) When(predicates ...Predicate[S]) FieldRules[T, S] {
	for _, predicate := range predicates {
		r.operations = append(r.operations, predicate)
	}
	return r
}

func (r FieldRules[T, S]) CascadeMode(mode CascadeMode) FieldRules[T, S] {
	r.cascadeMode = mode
	return r
}

func (r FieldRules[T, S]) GetCascadeMode() CascadeMode {
	return r.cascadeMode
}
