package alertpolicy

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
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

// ParseMeasurement parses string to Measurement
func ParseMeasurement(value string) (Measurement, error) {
	result, ok := getMeasurements()[value]
	if !ok {
		return result, fmt.Errorf("'%s' is not valid measurement", value)
	}
	return result, nil
}

// GetExpectedOperatorForMeasurement returns the operator that should be paired with a given measurement.
func GetExpectedOperatorForMeasurement(measurement Measurement) (v1alpha.Operator, error) {
	switch measurement {
	case MeasurementBurnedBudget:
		return v1alpha.GreaterThanEqual, nil
	case MeasurementAverageBurnRate:
		return v1alpha.GreaterThanEqual, nil
	case MeasurementTimeToBurnBudget:
		return v1alpha.LessThan, nil
	case MeasurementTimeToBurnEntireBudget:
		return v1alpha.LessThanEqual, nil
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
	)
}
