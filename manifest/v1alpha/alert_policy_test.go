package v1alpha

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlertingWindowValidation(t *testing.T) {
	for testCase, isValid := range map[string]bool{
		"5m":             true,
		"1h":             true,
		"72h":            true,
		"1h30m":          true,
		"1h1m60s":        true,
		"300s":           true,
		"300000ms":       true,
		"300000000000ns": true,
		"30000000000ns":  false,
		"3m":             false,
		"30s":            false,
		"90s":            false,
		"120s":           false,
		"5m30s":          false,
		"1h5m5s":         false,
		"555s":           false,
	} {
		condition := AlertCondition{
			Measurement:    MeasurementAverageBurnRate.String(),
			Value:          1.0,
			AlertingWindow: testCase,
		}

		t.Run(testCase, func(t *testing.T) {
			err := NewValidator().Check(condition)
			if isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
