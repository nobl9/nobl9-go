package v1alpha

import (
	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/validation"
)

// Measurement is allowed measurement types used for comparing values and triggering alerts
type Measurement int16

const (
	MeasurementBurnedBudget Measurement = iota + 1
	MeasurementAverageBurnRate
	MeasurementTimeToBurnBudget
	MeasurementTimeToBurnEntireBudget
)

const ErrorCodeMeasurement validation.ErrorCode = "measurement"

func getMeasurements() map[string]Measurement {
	return map[string]Measurement{
		"burnedBudget":           MeasurementBurnedBudget,
		"averageBurnRate":        MeasurementAverageBurnRate,
		"timeToBurnBudget":       MeasurementTimeToBurnBudget,
		"timeToBurnEntireBudget": MeasurementTimeToBurnEntireBudget,
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

// GetExpectedOperatorForMeasurement returns the operator that should be paired with a given measurement.
func GetExpectedOperatorForMeasurement(measurement Measurement) (Operator, error) {
	switch measurement {
	case MeasurementBurnedBudget:
		return GreaterThanEqual, nil
	case MeasurementAverageBurnRate:
		return GreaterThanEqual, nil
	case MeasurementTimeToBurnBudget:
		return LessThan, nil
	case MeasurementTimeToBurnEntireBudget:
		return LessThanEqual, nil
	default:
		return 0, errors.Errorf("unable to return expected operator for provided measurement: '%v'", measurement)
	}
}

func MeasurementValidation() validation.SingleRule[string] {
	return validation.OneOf(
		MeasurementBurnedBudget.String(),
		MeasurementAverageBurnRate.String(),
		MeasurementTimeToBurnBudget.String(),
		MeasurementTimeToBurnEntireBudget.String(),
	).WithErrorCode(ErrorCodeMeasurement)
}
