package v1alpha

import "fmt"

// Operator is allowed comparing method for labeling sli
type Operator int16

const (
	LessThanEqual Operator = iota + 1
	LessThan
	GreaterThanEqual
	GreaterThan
)

func getOperators() map[string]Operator {
	return map[string]Operator{
		"gt":  GreaterThan,
		"gte": GreaterThanEqual,
		"lt":  LessThan,
		"lte": LessThanEqual,
	}
}

func (m Operator) String() string {
	for key, val := range getOperators() {
		if val == m {
			return key
		}
	}
	return "Unknown"
}

// ParseOperator parses string to Operator
func ParseOperator(value string) (Operator, error) {
	result, ok := getOperators()[value]
	if !ok {
		return result, fmt.Errorf("'%s' is not valid operator", value)
	}
	return result, nil
}
