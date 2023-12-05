package alertpolicy

import (
	_ "embed"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

//go:embed test_data/expected_error.txt
var expectedError string

func TestValidate_AllErrors(t *testing.T) {
	err := validate(AlertPolicy{
		Kind: manifest.KindAlertPolicy,
		Metadata: Metadata{
			Name:    strings.Repeat("MY ALERTPOLICY", 20),
			Project: strings.Repeat("MY ALERTPOLICY", 20),
		},
		Spec: Spec{
			Description:      strings.Repeat("l", 2000),
			Severity:         SeverityHigh.String(),
			CoolDownDuration: "5m",
			Conditions:       []AlertCondition{validAlertCondition()},
			AlertMethods:     nil,
		},
		ManifestSource: "/home/me/alertpolicy.yaml",
	})
	assert.Equal(t, strings.TrimSuffix(expectedError, "\n"), err.Error())
}

func TestValidate_Metadata_Labels(t *testing.T) {
	t.Run("passes, no labels", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Metadata.Labels = nil
		err := validate(alertPolicy)
		testutils.AssertNoError(t, alertPolicy, err)
	})
	t.Run("passes, valid label", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Metadata.Labels = v1alpha.Labels{
			"label-key": []string{"label-1", "label-2"},
		}
		err := validate(alertPolicy)
		testutils.AssertNoError(t, alertPolicy, err)
	})
	t.Run("fails, invalid label", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Metadata.Labels = v1alpha.Labels{
			"L O L": []string{"dip", "dip"},
		}
		err := validate(alertPolicy)
		testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
			Prop:    "metadata.labels",
			Message: "label key 'L O L' does not match the regex: ^\\p{L}([_\\-0-9\\p{L}]*[0-9\\p{L}])?$",
		})
	})
}

func TestValidate_Metadata_Project(t *testing.T) {
	t.Run("passes, no project", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Metadata.Project = ""
		err := validate(alertPolicy)
		testutils.AssertNoError(t, alertPolicy, err)
	})
}

func TestValidate_Spec_Severity(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.Severity = SeverityHigh.String()
		err := validate(alertPolicy)
		testutils.AssertNoError(t, alertPolicy, err)
	})
	t.Run("fails, required", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.Severity = ""
		err := validate(alertPolicy)
		testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
			Prop: "spec.severity",
			Code: validation.ErrorCodeRequired,
		})
	})
	t.Run("fails", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.Severity = "Highest"
		err := validate(alertPolicy)
		testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
			Prop: "spec.severity",
			Code: ErrorCodeSeverity,
		})
	})
}

func TestValidate_Spec_CoolDownDuration(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.CoolDownDuration = "5m"
		err := validate(alertPolicy)
		testutils.AssertNoError(t, alertPolicy, err)
	})
	t.Run("passes, no value", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.CoolDownDuration = ""
		err := validate(alertPolicy)
		testutils.AssertNoError(t, alertPolicy, err)
	})

	tests := map[string]valuesWithCodeExpect{
		"fails, wrong format": {
			values:          []string{"1 hour"},
			expectedCode:    validation.ErrorCodeTransform,
			expectedMessage: `time: unknown unit " hour" in duration "1 hour"`,
		},
		"fails, not greater or equal to 5m": {
			values:          []string{"-10m", "4m"},
			expectedCode:    validation.ErrorCodeGreaterThanOrEqualTo,
			expectedMessage: `should be greater than or equal to '5m0s'`,
		},
	}
	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			for _, value := range testCase.values {
				alertPolicy := validAlertPolicy()
				alertPolicy.Spec.CoolDownDuration = value
				err := validate(alertPolicy)
				testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
					Prop:    "spec.coolDown",
					Code:    testCase.expectedCode,
					Message: testCase.expectedMessage,
				})
			}
		})
	}
}

func TestValidate_Spec_Conditions(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.Conditions = []AlertCondition{validAlertCondition()}
		err := validate(alertPolicy)
		testutils.AssertNoError(t, alertPolicy, err)
	})
	t.Run("fails, too few conditions", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.Conditions = make([]AlertCondition, 0)
		err := validate(alertPolicy)
		testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
			Prop: "spec.conditions",
			Code: validation.ErrorCodeSliceMinLength,
		})
	})
}

func TestValidate_Spec_Condition(t *testing.T) {
	t.Run("fields mutual exclusion", func(t *testing.T) {
		tests := map[string]AlertCondition{
			"fails, both alertingWindow and lastsFor": {
				AlertingWindow:   "6m",
				LastsForDuration: "16m",
			},
			"fails, no alertingWindow and no lastsFor": {
				AlertingWindow:   "",
				LastsForDuration: "",
			},
		}
		for name, testCase := range tests {
			t.Run(name, func(t *testing.T) {
				alertPolicy := validAlertPolicy()
				alertPolicy.Spec.Conditions[0].AlertingWindow = testCase.AlertingWindow
				alertPolicy.Spec.Conditions[0].LastsForDuration = testCase.LastsForDuration
				err := validate(alertPolicy)
				testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
					Prop: "spec.conditions[0]",
					Code: validation.ErrorCodeMutuallyExclusive,
				})
			})
		}
	})
}

func TestValidate_Spec_Condition_Measurement(t *testing.T) {
	t.Run("passes, with alertingWindow defined", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.Conditions[0].Measurement = MeasurementAverageBurnRate.String()
		err := validate(alertPolicy)
		testutils.AssertNoError(t, alertPolicy, err)
	})
	t.Run("passes, with lastsFor defined", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.Conditions[0].Measurement = MeasurementBurnedBudget.String()
		alertPolicy.Spec.Conditions[0].AlertingWindow = ""
		alertPolicy.Spec.Conditions[0].LastsForDuration = "8m"
		err := validate(alertPolicy)
		testutils.AssertNoError(t, alertPolicy, err)
	})
	t.Run("fails, required", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.Conditions[0].Measurement = ""
		alertPolicy.Spec.Conditions[0].AlertingWindow = ""
		alertPolicy.Spec.Conditions[0].LastsForDuration = "8m"
		err := validate(alertPolicy)
		testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
			Prop: "spec.conditions[0].measurement",
			Code: validation.ErrorCodeRequired,
		})
	})
	t.Run("fails", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.Conditions[0].Measurement = "Unknown"
		err := validate(alertPolicy)
		testutils.AssertContainsErrors(t, alertPolicy, err, 2, testutils.ExpectedError{
			Prop: "spec.conditions[0].measurement",
			Code: ErrorCodeMeasurement,
		})
	})
	failTests := map[string]measurementDetermined{
		"fails, alertingWindow is defined": {
			measurements: []Measurement{
				MeasurementTimeToBurnBudget,
				MeasurementTimeToBurnEntireBudget,
				MeasurementBurnedBudget,
			},
			expectedCode: errorCodeMeasurementWithAlertingWindow,
			expectedMessage: fmt.Sprintf(
				`measurement must be set to '%s' when alertingWindow is defined`,
				MeasurementAverageBurnRate.String(),
			),
		},
	}
	for name, testCase := range failTests {
		t.Run(name, func(t *testing.T) {
			for _, value := range testCase.values {
				for _, measurement := range testCase.measurements {
					alertPolicy := validAlertPolicy()
					alertPolicy.Spec.Conditions[0].Measurement = measurement.String()
					alertPolicy.Spec.Conditions[0].Value = value
					err := validate(alertPolicy)
					testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
						Prop:            "spec.conditions[0]",
						ContainsMessage: testCase.expectedMessage,
						Code:            testCase.expectedCode,
					})
				}
			}
		})
	}
}

func TestValidate_Spec_Condition_Value(t *testing.T) {
	passTests := map[string]measurementDetermined{
		"passes, valid duration when measurement is timeToBurnBudget or timeToBurnEntireBudget and lastsFor is defined": {
			values: []interface{}{
				"1ms",
				"15s",
				"15m",
				"1h",
			},
			measurements: []Measurement{
				MeasurementTimeToBurnBudget,
				MeasurementTimeToBurnEntireBudget,
			},
		},
		"passes, valid float measurement is burnedBudget or averageBurnRate and lastsFor is defined": {
			values: []interface{}{
				0.000000020,
				0.97,
				2.00,
				157.00,
			},
			measurements: []Measurement{
				MeasurementAverageBurnRate,
				MeasurementBurnedBudget,
			},
		},
	}
	for name, testCase := range passTests {
		t.Run(name, func(t *testing.T) {
			for _, value := range testCase.values {
				for _, measurement := range testCase.measurements {
					alertPolicy := validAlertPolicy()
					alertPolicy.Spec.Conditions[0].Measurement = measurement.String()
					alertPolicy.Spec.Conditions[0].Value = value
					alertPolicy.Spec.Conditions[0].LastsForDuration = "5m"
					alertPolicy.Spec.Conditions[0].AlertingWindow = ""
					err := validate(alertPolicy)
					testutils.AssertNoError(t, alertPolicy, err)
				}
			}
		})
	}
	t.Run("fails, required", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.Conditions[0].Value = ""
		alertPolicy.Spec.Conditions[0].Measurement = MeasurementAverageBurnRate.String()
		err := validate(alertPolicy)
		testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
			Prop: "spec.conditions[0].value",
			Code: validation.ErrorCodeRequired,
		})
	})

	failTests := map[string]measurementDetermined{
		"fails, greater than 0 when measurement is timeToBurnBudget or timeToBurnEntireBudget": {
			values: []interface{}{
				"-1ms",
				"-15s",
				"-1h",
			},
			measurements: []Measurement{
				MeasurementTimeToBurnBudget,
				MeasurementTimeToBurnEntireBudget,
			},
			expectedCode:    validation.ErrorCodeGreaterThan,
			expectedMessage: "should be greater than '0s'",
		},
		"fails, unexpected format when measurement is timeToBurnBudget or timeToBurnEntireBudget": {
			values: []interface{}{
				"1.0",
				"1.9",
				"100",
			},
			measurements: []Measurement{
				MeasurementAverageBurnRate,
				MeasurementBurnedBudget,
			},
			expectedCode:    validation.ErrorCodeTransform,
			expectedMessage: "must be valid float64",
		},
		"fails, unexpected format when measurement is burnedBudget or averageBurnRate": {
			values: []interface{}{
				"-1ms",
				"s892k",
				"100",
			},
			measurements: []Measurement{
				MeasurementAverageBurnRate,
				MeasurementBurnedBudget,
			},
			expectedCode:    validation.ErrorCodeTransform,
			expectedMessage: "must be valid float64",
		},
	}
	for name, testCase := range failTests {
		t.Run(name, func(t *testing.T) {
			for _, value := range testCase.values {
				for _, measurement := range testCase.measurements {
					alertPolicy := validAlertPolicy()
					alertPolicy.Spec.Conditions[0].Measurement = measurement.String()
					alertPolicy.Spec.Conditions[0].Value = value
					alertPolicy.Spec.Conditions[0].LastsForDuration = "8m"
					alertPolicy.Spec.Conditions[0].AlertingWindow = ""
					err := validate(alertPolicy)
					testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
						Prop:            "spec.conditions[0].value",
						ContainsMessage: testCase.expectedMessage,
						Code:            testCase.expectedCode,
					})
				}
			}
		})
	}
}

func TestValidate_Spec_Condition_AlertingWindow(t *testing.T) {
	validValues := []string{
		"5m",
		"1h",
		"72h",
		"1h30m",
		"1h1m60s",
		"300s",
		"0.1h",
		"300000ms",
		"300000000000ns",
	}
	for _, value := range validValues {
		t.Run("passes", func(t *testing.T) {
			alertPolicy := validAlertPolicy()
			alertPolicy.Spec.Conditions[0].AlertingWindow = value
			err := validate(alertPolicy)
			testutils.AssertNoError(t, alertPolicy, err)
		})
	}
	t.Run("passes, empty with lastsFor defined", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.Conditions[0].AlertingWindow = ""
		alertPolicy.Spec.Conditions[0].LastsForDuration = "5m"
		err := validate(alertPolicy)
		testutils.AssertNoError(t, alertPolicy, err)
	})

	tests := map[string]valuesWithCodeExpect{
		"fails, wrong format": {
			values:          []string{"1 hour"},
			expectedCode:    validation.ErrorCodeTransform,
			expectedMessage: `time: unknown unit " hour" in duration "1 hour"`,
		},
		"fails, cannot parse unit format": {
			values: []string{
				"555d",
				"0.01y",
				"0.5w",
				"1w",
			},
			expectedCode:    validation.ErrorCodeTransform,
			expectedMessage: `time: unknown unit`,
		},
	}
	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			for _, value := range testCase.values {
				alertPolicy := validAlertPolicy()
				alertPolicy.Spec.Conditions[0].AlertingWindow = value
				err := validate(alertPolicy)
				testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
					Prop:            "spec.conditions[0].alertingWindow",
					ContainsMessage: testCase.expectedMessage,
					Code:            testCase.expectedCode,
				})
			}
		})
	}

	failTests := map[string]valuesWithCodeExpect{
		"fails, not minute precision": {
			values: []string{
				"5m30s",
				"1h30s",
				"1h5m5s",
				"0.21h",
				"555s",
			},
			expectedCode:    errorCodeDurationFullMinutePrecision,
			expectedMessage: "duration must be defined with minute precision",
		},
		"fails, too long": {
			values: []string{
				"555h",
				"168h1m0s",
			},
			expectedCode:    validation.ErrorCodeLessThanOrEqualTo,
			expectedMessage: `should be less than or equal to '168h0m0s'`,
		},
		"fails, too short": {
			values: []string{
				"-10m",
				"-168h",
			},
			expectedCode:    validation.ErrorCodeGreaterThanOrEqualTo,
			expectedMessage: `should be greater than or equal to '5m0s'`,
		},
	}
	for name, testCase := range failTests {
		t.Run(name, func(t *testing.T) {
			for _, value := range testCase.values {
				alertPolicy := validAlertPolicy()
				alertPolicy.Spec.Conditions[0].AlertingWindow = value
				err := validate(alertPolicy)
				testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
					Prop:    "spec.conditions[0].alertingWindow",
					Message: testCase.expectedMessage,
					Code:    testCase.expectedCode,
				})
			}
		})
	}
}

func TestValidate_Spec_Condition_LastsForDuration(t *testing.T) {
	validValues := []string{
		"0",
		"5m",
		"1h",
		"72h",
		"1h20m",
		"1h1m35s",
		"300s",
		"0.1h",
		"300000ms",
		"300000000000ns",
		"1546h",
	}
	for _, value := range validValues {
		t.Run("passes", func(t *testing.T) {
			alertPolicy := validAlertPolicy()
			alertPolicy.Spec.Conditions[0].AlertingWindow = ""
			alertPolicy.Spec.Conditions[0].LastsForDuration = value
			err := validate(alertPolicy)
			testutils.AssertNoError(t, alertPolicy, err)
		})
	}
	t.Run("passes, empty with alertingWindow defined", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.Conditions[0].AlertingWindow = "5m"
		alertPolicy.Spec.Conditions[0].LastsForDuration = ""
		err := validate(alertPolicy)
		testutils.AssertNoError(t, alertPolicy, err)
	})

	tests := map[string]valuesWithCodeExpect{
		"fails, wrong format": {
			values:          []string{"1 hour"},
			expectedCode:    validation.ErrorCodeTransform,
			expectedMessage: `time: unknown unit " hour" in duration "1 hour"`,
		},
		"fails, wrong unit in format": {
			values:          []string{"365d"},
			expectedCode:    validation.ErrorCodeTransform,
			expectedMessage: `time: unknown unit "d" in duration "365d"`,
		},
		"fails, too short": {
			values: []string{
				"-10m",
				"-168h",
			},
			expectedCode:    validation.ErrorCodeGreaterThanOrEqualTo,
			expectedMessage: `should be greater than or equal to '0s'`,
		},
	}
	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			for _, value := range testCase.values {
				alertPolicy := validAlertPolicy()
				alertPolicy.Spec.Conditions[0].AlertingWindow = ""
				alertPolicy.Spec.Conditions[0].LastsForDuration = value
				err := validate(alertPolicy)
				testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
					Prop:    "spec.conditions[0].lastsFor",
					Message: testCase.expectedMessage,
					Code:    testCase.expectedCode,
				})
			}
		})
	}

	failTests := map[string]valuesWithCodeExpect{
		"fails, too short": {
			values: []string{
				"-10m",
				"-168h",
			},
			expectedCode:    validation.ErrorCodeGreaterThanOrEqualTo,
			expectedMessage: `should be greater than or equal to '0s'`,
		},
	}
	for name, testCase := range failTests {
		t.Run(name, func(t *testing.T) {
			for _, value := range testCase.values {
				alertPolicy := validAlertPolicy()
				alertPolicy.Spec.Conditions[0].AlertingWindow = ""
				alertPolicy.Spec.Conditions[0].LastsForDuration = value
				err := validate(alertPolicy)
				testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
					Prop:    "spec.conditions[0].lastsFor",
					Message: testCase.expectedMessage,
					Code:    testCase.expectedCode,
				})
			}
		})
	}
}

func TestValidate_Spec_Condition_Operator(t *testing.T) {
	const emptyOperator = ""
	allValidOps := []string{"gt", "lt", "lte", "gte", ""}

	testCases := []AlertCondition{
		{
			Measurement:      MeasurementTimeToBurnEntireBudget.String(),
			LastsForDuration: "10m",
			Value:            "30m",
		},
		{
			Measurement:      MeasurementTimeToBurnBudget.String(),
			LastsForDuration: "10m",
			Value:            "30m",
		},
		{
			Measurement:    MeasurementAverageBurnRate.String(),
			Value:          30.0,
			AlertingWindow: "5m",
		},
		{
			Measurement:      MeasurementAverageBurnRate.String(),
			Value:            30.0,
			LastsForDuration: "5m",
		},
	}

	for _, alertCondition := range testCases {
		t.Run("operator with a reference to Measurement", func(t *testing.T) {
			measurement, _ := ParseMeasurement(alertCondition.Measurement)
			expectedOperator, err := GetExpectedOperatorForMeasurement(measurement)
			assert.NoError(t, err)

			allowedOps := []string{expectedOperator.String(), emptyOperator}
			for _, op := range allValidOps {
				alertPolicy := validAlertPolicy()
				alertCondition.Operator = op
				alertPolicy.Spec.Conditions[0] = alertCondition
				err := validate(alertPolicy)
				if slices.Contains(allowedOps, op) {
					testutils.AssertNoError(t, alertPolicy, err)
				} else {
					testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
						Prop: "spec.conditions[0].operator",
						Message: fmt.Sprintf(
							`measurement '%s' determines operator must be defined with '%s' or left empty`,
							measurement.String(), expectedOperator,
						),
						Code: errorCodeOperatorAppropriateOperatorRegardingMeasurement,
					})
				}
			}
		})
	}
	t.Run("fails, invalid operator", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.Conditions[0].Operator = "noop"
		err := validate(alertPolicy)
		testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
			Prop:    "spec.conditions[0].operator",
			Message: "'noop' is not valid operator",
			Code:    errorCodeOperatorAppropriateOperatorRegardingMeasurement,
		})
	})
}

func validAlertPolicy() AlertPolicy {
	return New(
		Metadata{
			Name:    "alert-policy",
			Project: "project",
		},
		Spec{
			Description:      "Example alertPolicy",
			Severity:         SeverityHigh.String(),
			CoolDownDuration: "5m",
			Conditions:       []AlertCondition{validAlertCondition()}},
	)
}

func validAlertCondition() AlertCondition {
	return AlertCondition{
		Measurement:      MeasurementAverageBurnRate.String(),
		Value:            0.97,
		AlertingWindow:   "10m",
		LastsForDuration: "",
		Operator:         "",
	}
}

type valuesWithCodeExpect struct {
	values          []string
	expectedCode    string
	expectedMessage string
}

type measurementDetermined struct {
	values          []interface{}
	measurements    []Measurement
	expectedCode    string
	expectedMessage string
}
