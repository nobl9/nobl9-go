package validation

// For creates a new [PropertyRules] instance for the property
// which value is extracted through [PropertyGetter] function.
func For[T, S any](getter PropertyGetter[T, S]) PropertyRules[T, S] {
	return PropertyRules[T, S]{getter: func(s S) (v T, err error) { return getter(s), nil }}
}

// ForPointer accepts a getter function returning a pointer and wraps its call in order to
// safely extract the value under the pointer or return a zero value for a give type T.
// If required is set to true, the nil pointer value will result in an error and the
// validation will not proceed.
func ForPointer[T, S any](getter PropertyGetter[*T, S]) PropertyRules[T, S] {
	return PropertyRules[T, S]{getter: func(s S) (indirect T, err error) {
		ptr := getter(s)
		if ptr != nil {
			return *ptr, nil
		}
		zv := *new(T)
		return zv, emptyErr{}
	}, isPointer: true}
}

// Transform transforms value from one type to another.
// Value returned by [PropertyGetter] is transformed through [Transformer] function.
// If [Transformer] returns an error, the validation will not proceed and transformation error will be reported.
// [Transformer] is only called if [PropertyGetter] returns a non-zero value.
func Transform[T, N, S any](getter PropertyGetter[T, S], transform Transformer[T, N]) PropertyRules[N, S] {
	return PropertyRules[N, S]{
		transformGetter: func(s S) (transformed N, original any, err error) {
			v := getter(s)
			if err != nil {
				return transformed, nil, err
			}
			if isEmptyFunc(v) {
				return transformed, nil, emptyErr{}
			}
			transformed, err = transform(v)
			if err != nil {
				return transformed, v, NewRuleError(err.Error(), ErrorCodeTransform)
			}
			return transformed, v, nil
		},
	}
}

// GetSelf is a convenience method for extracting 'self' property of a validated value.
func GetSelf[S any]() PropertyGetter[S, S] {
	return func(s S) S { return s }
}

type Transformer[T, N any] func(T) (N, error)

type Predicate[S any] func(S) bool

type PropertyGetter[T, S any] func(S) T

type internalPropertyGetter[T, S any] func(S) (v T, err error)
type internalTransformPropertyGetter[T, S any] func(S) (transformed T, original any, err error)
type emptyErr struct{}

func (emptyErr) Error() string { return "" }

// PropertyRules is responsible for validating a single property.
type PropertyRules[T, S any] struct {
	name            string
	getter          internalPropertyGetter[T, S]
	transformGetter internalTransformPropertyGetter[T, S]
	steps           []interface{}
	required        bool
	omitEmpty       bool
	hideValue       bool
	isPointer       bool
}

// Validate validates the property value using provided rules.
// nolint: gocognit
func (r PropertyRules[T, S]) Validate(st S) PropertyErrors {
	var (
		ruleErrors         []error
		allErrors          PropertyErrors
		previousStepFailed bool
	)
	propValue, skip, err := r.getValue(st)
	if err != nil {
		if r.hideValue {
			err = err.HideValue()
		}
		return err
	}
	if skip {
		return nil
	}
loop:
	for _, step := range r.steps {
		switch v := step.(type) {
		case stopOnErrorStep:
			if previousStepFailed {
				break loop
			}
		case Predicate[S]:
			if !v(st) {
				break loop
			}
		// Same as Rule[S] as for GetSelf we'd get the same type on T and S.
		case Rule[T]:
			err := v.Validate(propValue)
			if err != nil {
				switch ev := err.(type) {
				case *PropertyError:
					allErrors = append(allErrors, ev.PrependPropertyName(r.name))
				default:
					ruleErrors = append(ruleErrors, err)
				}
			}
			previousStepFailed = err != nil
		case validatorI[T]:
			err := v.Validate(propValue)
			if err != nil {
				for _, e := range err.Errors {
					allErrors = append(allErrors, e.PrependPropertyName(r.name))
				}
			}
			previousStepFailed = err != nil
		}
	}
	if len(ruleErrors) > 0 {
		allErrors = append(allErrors, NewPropertyError(r.name, propValue, ruleErrors...))
	}
	if len(allErrors) > 0 {
		if r.hideValue {
			allErrors = allErrors.HideValue()
		}
		return allErrors
	}
	return nil
}

func (r PropertyRules[T, S]) WithName(name string) PropertyRules[T, S] {
	r.name = name
	return r
}

func (r PropertyRules[T, S]) Rules(rules ...Rule[T]) PropertyRules[T, S] {
	r.steps = appendSteps(r.steps, rules)
	return r
}

func (r PropertyRules[T, S]) Include(rules ...Validator[T]) PropertyRules[T, S] {
	r.steps = appendSteps(r.steps, rules)
	return r
}

func (r PropertyRules[T, S]) When(predicates ...Predicate[S]) PropertyRules[T, S] {
	r.steps = appendSteps(r.steps, predicates)
	return r
}

func (r PropertyRules[T, S]) Required() PropertyRules[T, S] {
	r.required = true
	return r
}

func (r PropertyRules[T, S]) OmitEmpty() PropertyRules[T, S] {
	r.omitEmpty = true
	return r
}

func (r PropertyRules[T, S]) HideValue() PropertyRules[T, S] {
	r.hideValue = true
	return r
}

type stopOnErrorStep uint8

func (r PropertyRules[T, S]) StopOnError() PropertyRules[T, S] {
	r.steps = append(r.steps, stopOnErrorStep(0))
	return r
}

func appendSteps[T any](slice []interface{}, steps []T) []interface{} {
	for _, step := range steps {
		slice = append(slice, step)
	}
	return slice
}

func (r PropertyRules[T, S]) getValue(st S) (v T, skip bool, errs PropertyErrors) {
	var (
		err           error
		originalValue any
	)
	if r.transformGetter != nil {
		v, originalValue, err = r.transformGetter(st)
	} else {
		v, err = r.getter(st)
	}
	_, isEmptyError := err.(emptyErr)
	// Any error other than [emptyErr] is considered critical, we don't proceed with validation.
	if err != nil && !isEmptyError {
		// If the value was transformed, we need to set the property value to the original, pre-transformed one.
		var propValue interface{}
		if HasErrorCode(err, ErrorCodeTransform) {
			propValue = originalValue
		} else {
			propValue = v
		}
		return v, false, PropertyErrors{NewPropertyError(r.name, propValue, err)}
	}
	isEmpty := isEmptyError || (!r.isPointer && isEmptyFunc(v))
	if r.required && isEmpty {
		return v, false, PropertyErrors{NewPropertyError(r.name, nil, NewRequiredError())}
	}
	if isEmpty && (r.omitEmpty || r.isPointer) {
		return v, true, nil
	}
	return v, false, nil
}
