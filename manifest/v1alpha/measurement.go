package v1alpha

import "fmt"

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
