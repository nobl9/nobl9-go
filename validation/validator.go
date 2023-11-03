package validation

type Rules[S any] interface {
	Validate(v S) []error
}

func New[S any](rules ...Rules[S]) Validator[S] {
	return Validator[S]{rules: rules}
}

type Validator[S any] struct {
	rules []Rules[S]
}

func (v Validator[S]) Validate(st S) []error {
	var allErrors []error
	for _, rule := range v.rules {
		if errs := rule.Validate(st); len(errs) > 0 {
			allErrors = append(allErrors, errs...)
		}
	}
	return allErrors
}
