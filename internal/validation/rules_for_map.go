package validation

import "fmt"

func ForMap[M ~map[K]V, K comparable, V, S any](getter PropertyGetter[M, S]) PropertyRulesForMap[M, K, V, S] {
	return PropertyRulesForMap[M, K, V, S]{getter: getter}
}

type PropertyRulesForMap[M ~map[K]V, K comparable, V, S any] struct {
	name   string
	getter PropertyGetter[M, S]
	steps  []interface{}
}

type MapKeyValue[K comparable, V any] struct {
	Key   K
	Value V
}

// Validate executes each of the rules sequentially and aggregates the encountered errors.
// nolint: prealloc, gocognit
func (r PropertyRulesForMap[M, K, V, S]) Validate(st S) PropertyErrors {
	var (
		allErrors          PropertyErrors
		mapErrors          []error
		propValue          M
		previousStepFailed bool
	)
	forEachErrors := make(map[K]forEachElementError)
loop:
	for _, step := range r.steps {
		switch stepValue := step.(type) {
		case stopOnErrorStep:
			if previousStepFailed {
				break loop
			}
		case Predicate[S]:
			if !stepValue(st) {
				break loop
			}
		case Rule[K], Rule[V], Rule[MapKeyValue[K, V]]:
			errorEncountered := false
			m := r.getter(st)
			for key := range m {
				var err error
				switch stepValue := step.(type) {
				case validatorI[K]:
					err = stepValue.Validate(key)
				case validatorI[V]:
					err = stepValue.Validate(m[key])
				case validatorI[MapKeyValue[K, V]]:
					err = stepValue.Validate(MapKeyValue[K, V]{Key: key, Value: m[key]})
				}
				if err == nil {
					continue
				}
				errorEncountered = true
				fErrs := forEachErrors[key].Errors
				switch ev := err.(type) {
				case *PropertyError:
					allErrors = append(allErrors, ev.PrependPropertyName(MapElementName(r.name, key)))
				default:
					forEachErrors[key] = forEachElementError{Errors: append(fErrs, err), PropValue: key}
				}
			}
			previousStepFailed = errorEncountered
		case Rule[M]:
			propValue = r.getter(st)
			err := stepValue.Validate(propValue)
			if err != nil {
				switch ev := err.(type) {
				case *PropertyError:
					allErrors = append(allErrors, ev.PrependPropertyName(r.name))
				default:
					mapErrors = append(mapErrors, err)
				}
			}
			previousStepFailed = err != nil
		case validatorI[K], validatorI[V], validatorI[MapKeyValue[K, V]]:
			errorEncountered := false
			m := r.getter(st)
			for key := range m {
				var err *ValidatorError
				switch stepValue := step.(type) {
				case validatorI[K]:
					err = stepValue.Validate(key)
				case validatorI[V]:
					err = stepValue.Validate(m[key])
				case validatorI[MapKeyValue[K, V]]:
					err = stepValue.Validate(MapKeyValue[K, V]{Key: key, Value: m[key]})
				}
				if err == nil {
					continue
				}
				errorEncountered = true
				for _, e := range err.Errors {
					allErrors = append(allErrors, e.PrependPropertyName(MapElementName(r.name, key)))
				}
			}
			previousStepFailed = errorEncountered
		}
	}
	if len(mapErrors) > 0 {
		allErrors = append(allErrors, NewPropertyError(r.name, propValue, mapErrors...))
	}
	for key, element := range forEachErrors {
		allErrors = append(allErrors, NewPropertyError(
			MapElementName(r.name, key),
			element.PropValue,
			element.Errors...))
	}
	if len(allErrors) > 0 {
		return allErrors
	}
	return nil
}

func (r PropertyRulesForMap[M, K, V, S]) WithName(name string) PropertyRulesForMap[M, K, V, S] {
	r.name = name
	return r
}

func (r PropertyRulesForMap[M, K, V, S]) RulesForKeys(rules ...Rule[K]) PropertyRulesForMap[M, K, V, S] {
	r.steps = appendSteps(r.steps, rules)
	return r
}

func (r PropertyRulesForMap[M, K, V, S]) RulesForValues(rules ...Rule[V]) PropertyRulesForMap[M, K, V, S] {
	r.steps = appendSteps(r.steps, rules)
	return r
}

func (r PropertyRulesForMap[M, K, V, S]) RulesForItems(
	rules ...Rule[MapKeyValue[K, V]],
) PropertyRulesForMap[M, K, V, S] {
	r.steps = appendSteps(r.steps, rules)
	return r
}

func (r PropertyRulesForMap[M, K, V, S]) Rules(rules ...Rule[M]) PropertyRulesForMap[M, K, V, S] {
	r.steps = appendSteps(r.steps, rules)
	return r
}

func (r PropertyRulesForMap[M, K, V, S]) When(predicate Predicate[S]) PropertyRulesForMap[M, K, V, S] {
	r.steps = append(r.steps, predicate)
	return r
}

func (r PropertyRulesForMap[M, K, V, S]) IncludeForKeys(rules ...Validator[K]) PropertyRulesForMap[M, K, V, S] {
	r.steps = appendSteps(r.steps, rules)
	return r
}

func (r PropertyRulesForMap[M, K, V, S]) IncludeForValues(rules ...Validator[V]) PropertyRulesForMap[M, K, V, S] {
	r.steps = appendSteps(r.steps, rules)
	return r
}

func (r PropertyRulesForMap[M, K, V, S]) IncludeForItems(
	rules ...Validator[MapKeyValue[K, V]],
) PropertyRulesForMap[M, K, V, S] {
	r.steps = appendSteps(r.steps, rules)
	return r
}

func (r PropertyRulesForMap[M, K, V, S]) StopOnError() PropertyRulesForMap[M, K, V, S] {
	r.steps = append(r.steps, stopOnErrorStep(0))
	return r
}

func MapElementName(mapName, key any) string {
	return fmt.Sprintf("%s.%v", mapName, key)
}
