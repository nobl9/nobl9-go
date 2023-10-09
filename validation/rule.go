package validation

// Rule is the interface for all validation rules.
type Rule[T any] interface {
	CascadeModeGetter
	Validate(v T) error
}

func NewSingleRule[T any](validate func(v T) error) SingleRule[T] {
	return SingleRule[T]{validate: validate}
}

type SingleRule[T any] struct {
	validate    func(v T) error
	cascadeMode CascadeMode
}

func (r SingleRule[T]) Validate(v T) error { return r.validate(v) }

func (r SingleRule[T]) CascadeMode(mode CascadeMode) SingleRule[T] {
	r.cascadeMode = mode
	return r
}

func (r SingleRule[T]) GetCascadeMode() CascadeMode {
	return r.cascadeMode
}

func NewMultiRule[T any](rules ...Rule[T]) MultiRule[T] {
	return MultiRule[T]{rules: rules}
}

// MultiRule allows defining Rule which aggregates multiple sub-rules.
type MultiRule[T any] struct {
	rules       []Rule[T]
	cascadeMode CascadeMode
}

func (r MultiRule[T]) CascadeMode(mode CascadeMode) MultiRule[T] {
	r.cascadeMode = mode
	return r
}

func (r MultiRule[T]) GetCascadeMode() CascadeMode {
	return r.cascadeMode
}

func (r MultiRule[T]) Validate(v T) error {
	var mErr multiRuleError
	for i := range r.rules {
		if err := r.rules[i].Validate(v); err != nil {
			mErr = append(mErr, err)
			if r.rules[i].GetCascadeMode() == CascadeModeStop {
				break
			}
		}
	}
	if len(mErr) > 0 {
		return mErr
	}
	return nil
}
