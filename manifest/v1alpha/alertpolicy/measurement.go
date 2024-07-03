package alertpolicy

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

// Measurement is allowed measurement types used for comparing values and triggering alerts
type Measurement int16

const (
	MeasurementBurnedBudget Measurement = iota + 1
	MeasurementAverageBurnRate
	MeasurementTimeToBurnBudget
	MeasurementTimeToBurnEntireBudget
	MeasurementBudgetDrop
)

func getMeasurements() map[string]Measurement {
	return map[string]Measurement{
		"burnedBudget":           MeasurementBurnedBudget,
		"averageBurnRate":        MeasurementAverageBurnRate,
		"timeToBurnBudget":       MeasurementTimeToBurnBudget,
		"timeToBurnEntireBudget": MeasurementTimeToBurnEntireBudget,
		"budgetDrop":             MeasurementBudgetDrop,
	}
}

func (m Measurement) String() string {
	for key, val := range getMeasurements() {
		if val == m {
			return key
		}
	}
	//nolint: goconst
	return "Unknown"
}

// ParseMeasurement parses string to Measurement
func ParseMeasurement(value string) (Measurement, error) {
	result, ok := getMeasurements()[value]
	if !ok {
		return result, fmt.Errorf("'%s' is not valid measurement", value)
	}
	return result, nil
}

// GetDefaultOperatorForMeasurement returns the default operator when operator is undefined.
func GetDefaultOperatorForMeasurement(measurement Measurement) (v1alpha.Operator, error) {
	switch measurement {
	case MeasurementBurnedBudget:
		return v1alpha.GreaterThanEqual, nil
	case MeasurementAverageBurnRate:
		return v1alpha.GreaterThanEqual, nil
	case MeasurementTimeToBurnBudget:
		return v1alpha.LessThan, nil
	case MeasurementTimeToBurnEntireBudget:
		return v1alpha.LessThanEqual, nil
	case MeasurementBudgetDrop:
		return v1alpha.GreaterThanEqual, nil
	default:
		return 0, errors.Errorf("unable to return expected operator for provided measurement: '%v'", measurement)
	}
}

// getExpectedOperatorForMeasurement returns the operator that should be paired with a given measurement.
func getExpectedOperatorForMeasurement(measurement Measurement) (v1alpha.Operator, error) {
	switch measurement {
	case MeasurementAverageBurnRate, MeasurementTimeToBurnBudget,
		MeasurementTimeToBurnEntireBudget, MeasurementBudgetDrop:
		return GetDefaultOperatorForMeasurement(measurement)
	default:
		return 0, errors.Errorf("unable to return expected operator for provided measurement: '%v'", measurement)
	}
}

func measurementValidation() validation.SingleRule[string] {
	return validation.OneOf(
		MeasurementBurnedBudget.String(),
		MeasurementAverageBurnRate.String(),
		MeasurementTimeToBurnBudget.String(),
		MeasurementTimeToBurnEntireBudget.String(),
		MeasurementBudgetDrop.String(),
	)
}
