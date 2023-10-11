package validation

import (
	"fmt"

	"github.com/pkg/errors"
	"golang.org/x/exp/constraints"
)

func EqualTo[T comparable](compared T) Rule[T] {
	return NewSingleRule(func(v T) error {
		if v != compared {
			return errors.Errorf(comparisonFmt, v, cmpEqualTo, compared)
		}
		return nil
	}).WithErrorCode(ErrorCodeEqualTo)
}

func NotEqualTo[T comparable](compared T) Rule[T] {
	return NewSingleRule(func(v T) error {
		if v == compared {
			return errors.Errorf(comparisonFmt, v, cmpNotEqualTo, compared)
		}
		return nil
	}).WithErrorCode(ErrorCodeNotEqualTo)
}

func GreaterThan[T constraints.Ordered](n T) Rule[T] {
	return NewSingleRule(orderedComparisonRule(cmpGreaterThan, n)).
		WithErrorCode(ErrorCodeGreaterThan)
}

func GreaterThanOrEqualTo[T constraints.Ordered](n T) Rule[T] {
	return NewSingleRule(orderedComparisonRule(cmpGreaterThanOrEqual, n)).
		WithErrorCode(ErrorCodeGreaterThanOrEqualTo)
}

func LessThan[T constraints.Ordered](n T) Rule[T] {
	return NewSingleRule(orderedComparisonRule(cmpLessThan, n)).
		WithErrorCode(ErrorCodeLessThan)
}

func LessThanOrEqualTo[T constraints.Ordered](n T) Rule[T] {
	return NewSingleRule(orderedComparisonRule(cmpLessThanOrEqual, n)).
		WithErrorCode(ErrorCodeLessThanOrEqualTo)
}

var comparisonFmt = "%v should be %s %v"

func orderedComparisonRule[T constraints.Ordered](op comparisonOperator, compared T) func(T) error {
	return func(v T) error {
		var passed bool
		//nolint: exhaustive
		switch op {
		case cmpGreaterThan:
			passed = v > compared
		case cmpGreaterThanOrEqual:
			passed = v >= compared
		case cmpLessThan:
			passed = v < compared
		case cmpLessThanOrEqual:
			passed = v <= compared
		}
		if !passed {
			return fmt.Errorf(comparisonFmt, v, op, compared)
		}
		return nil
	}
}

type comparisonOperator uint8

const (
	cmpEqualTo comparisonOperator = iota
	cmpNotEqualTo
	cmpGreaterThan
	cmpGreaterThanOrEqual
	cmpLessThan
	cmpLessThanOrEqual
)

func (c comparisonOperator) String() string {
	//exhaustive: enforce
	switch c {
	case cmpEqualTo:
		return "equal to"
	case cmpNotEqualTo:
		return "not equal to"
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
