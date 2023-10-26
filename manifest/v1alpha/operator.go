package v1alpha

import (
	"fmt"
)

// Operator is allowed comparing method for labeling sli
type Operator int16

const (
	LessThanEqual Operator = iota + 1
	LessThan
	GreaterThanEqual
	GreaterThan
)

var operatorNames = map[string]Operator{
	"lte": LessThanEqual,
	"lt":  LessThan,
	"gte": GreaterThanEqual,
	"gt":  GreaterThan,
}

var operatorValues = map[Operator]string{
	LessThanEqual:    "lte",
	LessThan:         "lt",
	GreaterThanEqual: "gte",
	GreaterThan:      "gt",
}

func (m Operator) String() string {
	if s, ok := operatorValues[m]; ok {
		return s
	}
	return "Unknown"
}

// ParseOperator parses string to Operator
func ParseOperator(value string) (Operator, error) {
	if op, ok := operatorNames[value]; ok {
		return op, nil
	}
	return 0, fmt.Errorf("'%s' is not valid operator", value)
}

// OperatorNames returns a list of possible string values of Operator.
func OperatorNames() []string {
	names := make([]string, 0, len(operatorNames))
	for name := range operatorNames {
		names = append(names, name)
	}
	return names
}
