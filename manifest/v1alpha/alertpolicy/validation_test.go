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
			expectedCode:    errorCodeDurationGreaterThanOrEqual,
			expectedMessage: "duration must be equal or greater than 5m0s",
		},
		"fails, not greater or equal to 5m": {
			value:           "4m",
			expectedCode:    errorCodeDurationGreaterThanOrEqual,
			expectedMessage: "duration must be equal or greater than 5m0s",
		},
	}
	for name, valueAndExpectations := range tests {
		t.Run(name, func(t *testing.T) {
			alertPolicy := validAlertPolicy()
			alertPolicy.Spec.CoolDownDuration = valueAndExpectations.value
			err := validate(alertPolicy)
			testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
				Prop: "spec.coolDown",
				Code: valueAndExpectations.expectedCode,
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
		slo := validAlertPolicy()
		slo.Spec.Conditions = make([]AlertCondition, 0)
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.conditions",
			Code: validation.ErrorCodeSliceMinLength,
		})
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

	t.Run("passes", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.Conditions[0].AlertingWindow = "15m"
		err := validate(alertPolicy)
		testutils.AssertNoError(t, alertPolicy, err)
	})
	t.Run("passes, no value", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.Conditions[0].AlertingWindow = ""
		err := validate(alertPolicy)
		testutils.AssertNoError(t, alertPolicy, err)
	})

	tests := map[string]valueWithCodeExpect{
		"fails, wrong format": {
			value:           "1 hour",
			expectedCode:    errorCodeDuration,
			expectedMessage: `time: unknown unit " hour" in duration "1 hour"`,
		},
		"fails, negative": {
			value:           "-30m",
			expectedCode:    errorCodeDurationGreaterThanOrEqual,
			expectedMessage: "duration must be equal or greater than 0m0s",
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

	notMinutePrecisionTests := []string{
		"5m30s",
		"1h30s",
		"1h5m5s",
		"0.01h",
		"555s",
	}
	for _, value := range notMinutePrecisionTests {
		t.Run("fails, not minute precision", func(t *testing.T) {
			alertPolicy := validAlertPolicy()
			alertPolicy.Spec.Conditions[0].AlertingWindow = value
			err := validate(alertPolicy)
			testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
				Prop:    "spec.conditions[0].alertingWindow",
				Message: "duration must be defined with minute precision",
				Code:    errorCodeDurationFullMinutePrecision,
			})
		})
	}
	// other cases to cover
	//"555h": false,
	//"555d": false,
	//
	//// Invalid: Not supported unit
	//// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". (ref. time.ParseDuration)
	//"0.01y": false,
	//"0.5w":  false,
	//"1w":    false,
	//
	//// Invalid: Not a minute precision
	//"5m30s":  false,
	//"1h30s":  false,
	//"1h5m5s": false,
	//"0.01h":  false,
	//"555s":   false,
	//
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
