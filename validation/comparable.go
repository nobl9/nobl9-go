package validation

import (
	"fmt"

	"github.com/pkg/errors"
	"golang.org/x/exp/constraints"
)

func EqualTo[T comparable](compared T) SingleRule[T] {
	return NewSingleRule(func(v T) error {
		if v != compared {
			return errors.Errorf(comparisonFmt, v, cmpEqualTo, compared)
		}
		return nil
	})
}

func NotEqualTo[T comparable](compared T) SingleRule[T] {
	return NewSingleRule(func(v T) error {
		if v == compared {
			return errors.Errorf(comparisonFmt, v, cmpNotEqualTo, compared)
		}
		return nil
	})
}

func GreaterThan[T constraints.Ordered](n T) SingleRule[T] {
	return orderedComparisonRule(cmpGreaterThan, n)
}

func GreaterThanOrEqualTo[T constraints.Ordered](n T) SingleRule[T] {
	return orderedComparisonRule(cmpGreaterThanOrEqual, n)
}

func LessThan[T constraints.Ordered](n T) SingleRule[T] {
	return orderedComparisonRule(cmpLessThan, n)
}

func LessThanOrEqualTo[T constraints.Ordered](n T) SingleRule[T] {
	return orderedComparisonRule(cmpLessThanOrEqual, n)
}

var comparisonFmt = "%v should be %s %v"

func orderedComparisonRule[T constraints.Ordered](op comparisonOperator, compared T) SingleRule[T] {
	return NewSingleRule(func(v T) error {
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
	})
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
