package validation

import "fmt"

// ForMap creates a new [PropertyRulesForMap] instance for a map property
// which value is extracted through [PropertyGetter] function.
func ForMap[M ~map[K]V, K comparable, V, S any](getter PropertyGetter[M, S]) PropertyRulesForMap[M, K, V, S] {
	return PropertyRulesForMap[M, K, V, S]{getter: getter}
}

// PropertyRulesForMap is responsible for validating a single property.
type PropertyRulesForMap[M ~map[K]V, K comparable, V, S any] struct {
	name   string
	getter PropertyGetter[M, S]
	steps  []interface{}
}

// MapItem is a tuple container for map's key and value pair.
type MapItem[K comparable, V any] struct {
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
	valueErrors := make(map[K]mapItemError)
	keyErrors := make(map[K]mapItemError)
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
		case mapKeyRule[K], Rule[V], Rule[MapItem[K, V]]:
			errorEncountered := false
			m := r.getter(st)
			for key := range m {
				var (
					err        error
					isKeyError bool
				)
				switch stepValue := step.(type) {
				case mapKeyRule[K]:
					isKeyError = true
					err = stepValue.Validate(key)
				case Rule[V]:
					err = stepValue.Validate(m[key])
				case Rule[MapItem[K, V]]:
					err = stepValue.Validate(MapItem[K, V]{Key: key, Value: m[key]})
				}
				if err == nil {
					continue
				}
				errorEncountered = true
				switch ev := err.(type) {
				case *PropertyError:
					allErrors = append(allErrors, ev.PrependPropertyName(MapElementName(r.name, key)))
				default:
					if isKeyError {
						fErrs := keyErrors[key].Errors
						keyErrors[key] = mapItemError{Errors: append(fErrs, err), PropValue: key}
					} else {
						fErrs := valueErrors[key].Errors
						valueErrors[key] = mapItemError{Errors: append(fErrs, err), PropValue: m[key]}
					}
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
		case validatorI[K], validatorI[V], validatorI[MapItem[K, V]]:
			errorEncountered := false
			m := r.getter(st)
			for key := range m {
				var err *ValidatorError
				switch stepValue := step.(type) {
				case mapKeyValidator[K]:
					err = stepValue.Validate(key)
				case validatorI[V]:
					err = stepValue.Validate(m[key])
				case validatorI[MapItem[K, V]]:
					err = stepValue.Validate(MapItem[K, V]{Key: key, Value: m[key]})
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
	for key, item := range valueErrors {
		allErrors = append(allErrors, NewPropertyError(
			MapElementName(r.name, key),
			item.PropValue,
			item.Errors...))
	}
	for key, item := range keyErrors {
		propError := NewPropertyError(
			MapElementName(r.name, key),
			key,
			item.Errors...)
		propError.IsKeyError = true
		allErrors = append(allErrors, propError)
	}
	if len(allErrors) > 0 {
		return allErrors
	}
	return nil
}

type mapItemError struct {
	PropValue interface{}
	Errors    []error
}

func (r PropertyRulesForMap[M, K, V, S]) WithName(name string) PropertyRulesForMap[M, K, V, S] {
	r.name = name
	return r
}

// mapKeyRule wraps Rule for map keys in a custom type in order to discern between rules for keys and values.
// Otherwise, if key and value have the same type both Rule[K] and Rule[V] would match.
type mapKeyRule[K comparable] struct{ Rule[K] }

func (r PropertyRulesForMap[M, K, V, S]) RulesForKeys(rules ...Rule[K]) PropertyRulesForMap[M, K, V, S] {
	mapKeyRules := make([]mapKeyRule[K], 0, len(rules))
	for _, rule := range rules {
		mapKeyRules = append(mapKeyRules, mapKeyRule[K]{rule})
	}
	r.steps = appendSteps(r.steps, mapKeyRules)
	return r
}

func (r PropertyRulesForMap[M, K, V, S]) RulesForValues(rules ...Rule[V]) PropertyRulesForMap[M, K, V, S] {
	r.steps = appendSteps(r.steps, rules)
	return r
}

func (r PropertyRulesForMap[M, K, V, S]) RulesForItems(
	rules ...Rule[MapItem[K, V]],
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

// mapKeyValidator wraps Validator for map keys in a custom type in order
// to discern between validators for keys and values.
// Otherwise, if key and value have the same type both Validator[K] and Validator[V] would match.
type mapKeyValidator[K comparable] struct{ Validator[K] }

func (r PropertyRulesForMap[M, K, V, S]) IncludeForKeys(validators ...Validator[K]) PropertyRulesForMap[M, K, V, S] {
	mapKeyValidators := make([]mapKeyValidator[K], 0, len(validators))
	for _, validator := range validators {
		mapKeyValidators = append(mapKeyValidators, mapKeyValidator[K]{validator})
	}
	r.steps = appendSteps(r.steps, mapKeyValidators)
	return r
}

func (r PropertyRulesForMap[M, K, V, S]) IncludeForValues(rules ...Validator[V]) PropertyRulesForMap[M, K, V, S] {
	r.steps = appendSteps(r.steps, rules)
	return r
}

func (r PropertyRulesForMap[M, K, V, S]) IncludeForItems(
	rules ...Validator[MapItem[K, V]],
) PropertyRulesForMap[M, K, V, S] {
	r.steps = appendSteps(r.steps, rules)
	return r
}

func (r PropertyRulesForMap[M, K, V, S]) StopOnError() PropertyRulesForMap[M, K, V, S] {
	r.steps = append(r.steps, stopOnErrorStep(0))
	return r
}

func MapElementName(mapName, key any) string {
	if mapName == "" {
		return fmt.Sprintf("%v", key)
	}
	return fmt.Sprintf("%s.%v", mapName, key)
}
