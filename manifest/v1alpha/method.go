package v1alpha

import "fmt"

// BudgetingMethod indicates algorithm to calculate error budget
type BudgetingMethod int

const (
	// BudgetingMethodOccurrences method uses ratio of counts of good events and total count of event
	BudgetingMethodOccurrences BudgetingMethod = iota + 1
	// BudgetingMethodTimeslices method uses ratio of good time slices vs. total time slices in a budgeting period
	BudgetingMethodTimeslices
)

func getBudgetingMethodNames() map[string]BudgetingMethod {
	return map[string]BudgetingMethod{
		"Occurrences": BudgetingMethodOccurrences,
		"Timeslices":  BudgetingMethodTimeslices,
	}
}

func (m BudgetingMethod) String() string {
	for key, val := range getBudgetingMethodNames() {
		if val == m {
			return key
		}
	}
	return "Unknown"
}

func ParseBudgetingMethod(value string) (BudgetingMethod, error) {
	result, ok := getBudgetingMethodNames()[value]
	if !ok {
		return result, fmt.Errorf("'%s' is not valid budgeting method", value)
	}
	return result, nil
}
