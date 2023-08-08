package v1alpha

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetAlertPolicyDefaultSetDefaultLastsFor(t *testing.T) {
	for _, testCase := range []struct {
		desc                   string
		alertingWindow         string
		lastsFor               string
		expectedAlertingWindow string
		expectedLastsFor       string
	}{
		{
			desc:                   "when 'alertingWindow' is defined, 'lastsFor' default value should not be set",
			lastsFor:               "",
			alertingWindow:         "30m",
			expectedAlertingWindow: "30m",
			expectedLastsFor:       "",
		},
		{
			desc:                   "when 'alertingWindow' is not defined and 'lastsFor' is empty zero value should be set for `lastsFor`",
			lastsFor:               "",
			alertingWindow:         "",
			expectedAlertingWindow: "",
			expectedLastsFor:       DefaultAlertPolicyLastsForDuration,
		},
		{
			desc:                   "when 'alertingWindow' is not defined and 'lastsFor' is not empty do not change 'lastsFor'",
			lastsFor:               "30m",
			alertingWindow:         "",
			expectedAlertingWindow: "",
			expectedLastsFor:       "30m",
		},
	} {
		t.Run(testCase.desc, func(t *testing.T) {
			for _, measurement := range []Measurement{
				MeasurementBurnedBudget,
				MeasurementAverageBurnRate,
				MeasurementTimeToBurnBudget,
				MeasurementTimeToBurnEntireBudget,
			} {
				alertPolicy := &AlertPolicy{
					Spec: AlertPolicySpec{Conditions: []AlertCondition{
						0: {
							Measurement:      measurement.String(),
							Operator:         getAlertPolicyDefaultOperatorForMeasurement(measurement).String(),
							AlertingWindow:   testCase.alertingWindow,
							LastsForDuration: testCase.lastsFor,
						},
					}},
				}
				err := setAlertPolicyDefaults(alertPolicy)
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedLastsFor, alertPolicy.Spec.Conditions[0].LastsForDuration)
				assert.Equal(t, testCase.expectedAlertingWindow, alertPolicy.Spec.Conditions[0].AlertingWindow)
			}
		})
	}
}

func TestSetAlertPolicyDefaultSetDefaultOperator(t *testing.T) {
	for _, measurement := range []Measurement{
		MeasurementBurnedBudget,
		MeasurementAverageBurnRate,
		MeasurementTimeToBurnBudget,
		MeasurementTimeToBurnEntireBudget,
	} {
		t.Run(measurement.String(), func(t *testing.T) {
			alertPolicy := &AlertPolicy{
				Spec: AlertPolicySpec{Conditions: []AlertCondition{
					0: {Measurement: measurement.String()},
				}},
			}
			err := setAlertPolicyDefaults(alertPolicy)
			assert.NoError(t, err)
			assert.Equal(t,
				getAlertPolicyDefaultOperatorForMeasurement(measurement).String(),
				alertPolicy.Spec.Conditions[0].Operator,
			)
		})
	}
}

func TestDefaultOperatorForMeasurement(t *testing.T) {
	assert.Equal(t, GreaterThanEqual, getAlertPolicyDefaultOperatorForMeasurement(MeasurementBurnedBudget))
	assert.Equal(t, GreaterThanEqual, getAlertPolicyDefaultOperatorForMeasurement(MeasurementAverageBurnRate))
	assert.Equal(t, LessThanEqual, getAlertPolicyDefaultOperatorForMeasurement(MeasurementTimeToBurnEntireBudget))
	assert.Equal(t, LessThan, getAlertPolicyDefaultOperatorForMeasurement(MeasurementTimeToBurnBudget))
}
