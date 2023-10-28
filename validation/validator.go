package validation

type validatorI[S any] interface {
	Validate(s S) *ValidatorError
}

type propertyRulesI[S any] interface {
	Validate(s S) PropertyErrors
}

func New[S any](props ...propertyRulesI[S]) Validator[S] {
	return Validator[S]{props: props}
}

type Validator[S any] struct {
	props      []propertyRulesI[S]
	name       string
	predicates []Predicate[S]
}

func (v Validator[S]) WithName(name string) Validator[S] {
	v.name = name
	return v
}

func (v Validator[S]) When(predicates ...Predicate[S]) Validator[S] {
	v.predicates = append(v.predicates, predicates...)
	return v
}

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
