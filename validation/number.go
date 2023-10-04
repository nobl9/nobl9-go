package validation

import "fmt"

type number interface {
	uint | uint8 | uint16 | uint32 | uint64 | int | int8 | int16 | int32 | int64 | float32 | float64
}

func NumberEqual[T number](n T) SingleRule[T] {
	return numberCondition(cmpEqual, n)
}

func NumberGreaterThan[T number](n T) SingleRule[T] {
	return numberCondition(cmpGreaterThan, n)
}

func NumberGreaterThanOrEqual[T number](n T) SingleRule[T] {
	return numberCondition(cmpGreaterThanOrEqual, n)
}

func NumberLessThan[T number](n T) SingleRule[T] {
	return numberCondition(cmpLessThan, n)
}

func NumberLessThanOrEqual[T number](n T) SingleRule[T] {
	return numberCondition(cmpLessThanOrEqual, n)
}

func numberCondition[T number](op comparisonOperator, n T) SingleRule[T] {
	return func(v T) error {
		var passed bool
		switch op {
		case cmpEqual:
			passed = v == n
		case cmpGreaterThan:
			passed = v > n
		case cmpGreaterThanOrEqual:
			passed = v >= n
		case cmpLessThan:
			passed = v < n
		case cmpLessThanOrEqual:
			passed = v <= n
		}
		if !passed {
			return fmt.Errorf("%d should be %s %d", v, op.String(), n)
		}
		return nil
	}
}

type comparisonOperator uint8

const (
	cmpEqual comparisonOperator = iota
	cmpGreaterThan
	cmpGreaterThanOrEqual
	cmpLessThan
	cmpLessThanOrEqual
)

func (c comparisonOperator) String() string {
	switch c {
	case cmpEqual:
		return "equal to"
	case cmpGreaterThan:
		return "greater than"
	case cmpGreaterThanOrEqual:
		return "greater than or equal to"
	case cmpLessThan:
		return "less than"
	case cmpLessThanOrEqual:
		return "less than or equal to"
	default:
		return "unknown"
	}
}
