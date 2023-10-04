package validation

// Rule is the interface for all validation rules.
type Rule[T any] interface {
	Validate(v T) error
}

type SingleRule[T any] func(v T) error

func (r SingleRule[T]) Validate(v T) error { return r(v) }

// MultiRule allows defining Rule which aggregates multiple sub-rules.
type MultiRule[T any] []Rule[T]

func (r MultiRule[T]) Validate(v T) error {
	var mErr multiRuleError
	for i := range r {
		if err := r[i].Validate(v); err != nil {
			mErr = append(mErr, err)
		}
	}
	if len(mErr) > 0 {
		return mErr
	}
	return nil
}
