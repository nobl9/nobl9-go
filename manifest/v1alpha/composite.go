package v1alpha

func isBurnRateSetForCompositeWithOccurrences(spec SLOSpec) bool {
	return !isBudgetingMethodOccurrences(spec) || spec.Composite.BurnRateCondition != nil
}

func isValidBudgetingMethodForCompositeWithBurnRate(spec SLOSpec) bool {
	return spec.Composite.BurnRateCondition == nil || isBudgetingMethodOccurrences(spec)
}

func isBudgetingMethodOccurrences(sloSpec SLOSpec) bool {
	return sloSpec.BudgetingMethod == BudgetingMethodOccurrences.String()
}
