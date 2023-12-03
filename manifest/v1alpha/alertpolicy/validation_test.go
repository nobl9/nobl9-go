package alertpolicy

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

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
			Severity:         v1alpha.SeverityHigh.String(),
			CoolDownDuration: "5m",
			Conditions:       []AlertCondition{validAlertCondition()},
			AlertMethods:     nil,
		},
		ManifestSource: "/home/me/alertpolicy.yaml",
	})
	assert.Equal(t, strings.TrimSuffix(expectedError, "\n"), err.Error())
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
		alertPolicy.Spec.Severity = v1alpha.SeverityHigh.String()
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
			Code: v1alpha.ErrorCodeSeverity,
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

	tests := map[string]valueWithCodeExpect{
		"fails, wrong format": {
			value:           "1 hour",
			expectedCode:    validation.ErrorCodeTransform,
			expectedMessage: `time: unknown unit " hour" in duration "1 hour"`,
		},
		"fails, negative": {
			value:           "-10m",
			expectedCode:    validation.ErrorCodeGreaterThanOrEqualTo,
			expectedMessage: `should be greater than or equal to '5m0s'`,
		},
		"fails, not greater or equal to 5m": {
			value:           "4m",
			expectedCode:    validation.ErrorCodeGreaterThanOrEqualTo,
			expectedMessage: `should be greater than or equal to '5m0s'`,
		},
	}
	for name, valueAndExpectations := range tests {
		t.Run(name, func(t *testing.T) {
			alertPolicy := validAlertPolicy()
			alertPolicy.Spec.CoolDownDuration = valueAndExpectations.value
			err := validate(alertPolicy)
			testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
				Prop:    "spec.coolDown",
				Code:    valueAndExpectations.expectedCode,
				Message: valueAndExpectations.expectedMessage,
			})
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
	t.Run("passes", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.Conditions[0].Measurement = v1alpha.MeasurementBurnedBudget.String()
		err := validate(alertPolicy)
		testutils.AssertNoError(t, alertPolicy, err)
	})
	t.Run("fails, required", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.Conditions[0].Measurement = ""
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
		testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
			Prop: "spec.conditions[0].measurement",
			Code: v1alpha.ErrorCodeMeasurement,
		})
	})
}

func TestValidate_Spec_Condition_Value(t *testing.T) {
	tests := map[string]interface{}{
		"passes, float": 0.97,
		"passes, int":   2,
	}
	for name, value := range tests {
		t.Run(name, func(t *testing.T) {
			alertPolicy := validAlertPolicy()
			alertPolicy.Spec.Conditions[0].Value = value
			err := validate(alertPolicy)
			testutils.AssertNoError(t, alertPolicy, err)
		})
	}

	t.Run("fails, required", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.Conditions[0].Value = ""
		err := validate(alertPolicy)
		testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
			Prop: "spec.conditions[0].value",
			Code: validation.ErrorCodeRequired,
		})
	})
}

func TestValidate_Spec_Condition_AlertingWindow(t *testing.T) {
	validValues := []string{
		"",
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

	tests := map[string]valueWithCodeExpect{
		"fails, wrong format": {
			value:           "1 hour",
			expectedCode:    validation.ErrorCodeTransform,
			expectedMessage: `time: unknown unit " hour" in duration "1 hour"`,
		},
	}
	for name, valueAndExpectations := range tests {
		t.Run(name, func(t *testing.T) {
			alertPolicy := validAlertPolicy()
			alertPolicy.Spec.Conditions[0].AlertingWindow = valueAndExpectations.value
			err := validate(alertPolicy)
			testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
				Prop:    "spec.conditions[0].alertingWindow",
				Message: valueAndExpectations.expectedMessage,
				Code:    valueAndExpectations.expectedCode,
			})
		})
	}

	failParseUnitTests := map[string]valuesWithCodeExpect{
		"fails, cannot parse days in format": {
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
	for name, testCase := range failParseUnitTests {
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

}

func validAlertPolicy() AlertPolicy {
	return New(
		Metadata{
			Name:    "alert-policy",
			Project: "project",
		},
		Spec{
			Description:      "Example alertPolicy",
			Severity:         v1alpha.SeverityHigh.String(),
			CoolDownDuration: "5m",
			Conditions:       []AlertCondition{validAlertCondition()}},
	)
}

func validAlertCondition() AlertCondition {
	return AlertCondition{
		Measurement:      v1alpha.MeasurementBurnedBudget.String(),
		Value:            "0.97",
		AlertingWindow:   "",
		LastsForDuration: "",
		Operator:         "",
	}
}

type valueWithCodeExpect struct {
	value           string
	expectedCode    string
	expectedMessage string
}

type valuesWithCodeExpect struct {
	values          []string
	expectedCode    string
	expectedMessage string
}
