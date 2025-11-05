package alertpolicy

import (
	_ "embed"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/manifest/v1alphatest"
	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest"
)

var validationMessageRegexp = regexp.MustCompile(strings.TrimSpace(`
(?s)Validation for AlertPolicy '.*' in project '.*' has failed for the following fields:
.*
Manifest source: /home/me/alertpolicy.yaml
`))

func TestValidate_VersionAndKind(t *testing.T) {
	policy := validAlertPolicy()
	policy.APIVersion = "v0.1"
	policy.Kind = manifest.KindProject
	policy.ManifestSource = "/home/me/alertpolicy.yaml"
	err := validate(policy)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, policy, err, 2,
		testutils.ExpectedError{
			Prop: "apiVersion",
			Code: rules.ErrorCodeEqualTo,
		},
		testutils.ExpectedError{
			Prop: "kind",
			Code: rules.ErrorCodeEqualTo,
		},
	)
}

func TestValidate_Metadata(t *testing.T) {
	policy := validAlertPolicy()
	policy.Metadata = Metadata{
		Name:    strings.Repeat("MY ALERTPOLICY", 20),
		Project: strings.Repeat("MY ALERTPOLICY", 20),
	}
	policy.ManifestSource = "/home/me/alertpolicy.yaml"
	err := validate(policy)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, policy, err, 2,
		testutils.ExpectedError{
			Prop: "metadata.name",
			Code: validationV1Alpha.ErrorCodeStringName,
		},
		testutils.ExpectedError{
			Prop: "metadata.project",
			Code: validationV1Alpha.ErrorCodeStringName,
		},
	)
}

func TestValidate_Metadata_Labels(t *testing.T) {
	for name, test := range v1alphatest.GetLabelsTestCases[AlertPolicy](t, "metadata.labels") {
		t.Run(name, func(t *testing.T) {
			svc := validAlertPolicy()
			svc.Metadata.Labels = test.Labels
			test.Test(t, svc, validate)
		})
	}
}

func TestValidate_Metadata_Annotations(t *testing.T) {
	for name, test := range v1alphatest.GetMetadataAnnotationsTestCases[AlertPolicy](t, "metadata.annotations") {
		t.Run(name, func(t *testing.T) {
			svc := validAlertPolicy()
			svc.Metadata.Annotations = test.Annotations
			test.Test(t, svc, validate)
		})
	}
}

func TestValidate_Metadata_Project(t *testing.T) {
	t.Run("fails, project required", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Metadata.Project = ""
		err := validate(alertPolicy)
		testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
			Prop: "metadata.project",
			Code: rules.ErrorCodeRequired,
		})
	})
}

func TestValidate_Spec_Description(t *testing.T) {
	alertPolicy := validAlertPolicy()
	alertPolicy.Spec.Description = strings.Repeat("A", 2000)
	err := validate(alertPolicy)
	testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
		Prop: "spec.description",
		Code: validationV1Alpha.ErrorCodeStringDescription,
	})
}

func TestValidate_Spec_Severity(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		for _, severity := range getSeverityLevels() {
			alertPolicy := validAlertPolicy()
			alertPolicy.Spec.Severity = severity.String()
			err := validate(alertPolicy)
			testutils.AssertNoError(t, alertPolicy, err)
		}
	})
	t.Run("fails, required", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.Severity = ""
		err := validate(alertPolicy)
		testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
			Prop: "spec.severity",
			Code: rules.ErrorCodeRequired,
		})
	})
	t.Run("fails, invalid", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.Severity = "Highest"
		err := validate(alertPolicy)
		testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
			Prop: "spec.severity",
			Code: rules.ErrorCodeOneOf,
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
			values:       []string{"1 hour"},
			expectedCode: govy.ErrorCodeTransform,
		},
		"fails, value too small": {
			values:          []string{"60s", "4m"},
			expectedCode:    rules.ErrorCodeGreaterThanOrEqualTo,
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
			Code: rules.ErrorCodeSliceMinLength,
		})
	})
}

func TestValidate_Spec_Condition(t *testing.T) {
	t.Run("passes, alertingWindows and lastsFor can be empty", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.Conditions[0].AlertingWindow = ""
		alertPolicy.Spec.Conditions[0].LastsForDuration = ""
		err := validate(alertPolicy)
		testutils.AssertNoError(t, alertPolicy, err)
	})

	t.Run("fails, only alertingWindow or lastsFor can be defined", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.Conditions[0].AlertingWindow = "6m"
		alertPolicy.Spec.Conditions[0].LastsForDuration = "10m"
		err := validate(alertPolicy)
		testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
			Prop: "spec.conditions[0]",
			Code: rules.ErrorCodeMutuallyExclusive,
		})
	})
}

func TestValidate_Spec_Condition_Measurement(t *testing.T) {
	t.Run("passes, with alertingWindow defined", func(t *testing.T) {
		for _, testCase := range []AlertCondition{
			{
				Measurement:    MeasurementAverageBurnRate.String(),
				Value:          10.3,
				AlertingWindow: "10m",
			},
			{
				Measurement:    MeasurementBudgetDrop.String(),
				Value:          0.1,
				AlertingWindow: "5m",
			},
			{
				Measurement:    MeasurementTimeToBurnBudget.String(),
				Value:          "1h",
				AlertingWindow: "10m",
			},
			{
				Measurement:    MeasurementTimeToBurnEntireBudget.String(),
				Value:          "1h",
				AlertingWindow: "10m",
			},
		} {
			alertPolicy := validAlertPolicy()
			alertPolicy.Spec.Conditions = []AlertCondition{testCase}
			err := validate(alertPolicy)
			testutils.AssertNoError(t, alertPolicy, err)
		}
	})
	testCases := map[string][]AlertCondition{
		"passes, lastsFor is defined": {
			{
				Measurement:      MeasurementTimeToBurnEntireBudget.String(),
				Value:            "10m",
				AlertingWindow:   "",
				LastsForDuration: "8m",
			},
			{
				Measurement:      MeasurementTimeToBurnBudget.String(),
				Value:            "10m",
				AlertingWindow:   "",
				LastsForDuration: "8m",
			},
		},
		"passes, lastsFor defined with numeric value": {
			{
				Measurement:      MeasurementBurnedBudget.String(),
				Value:            0.97,
				AlertingWindow:   "",
				LastsForDuration: "8m",
			},
			{
				Measurement:      MeasurementAverageBurnRate.String(),
				Value:            0.97,
				AlertingWindow:   "",
				LastsForDuration: "8m",
			},
		},
		"passes, lastsFor defined, numeric value allowed for burnedBudget": {
			{
				Measurement:      MeasurementBurnedBudget.String(),
				Value:            0.97,
				AlertingWindow:   "",
				LastsForDuration: "8m",
			},
		},
	}
	for name, alertConditionCase := range testCases {
		t.Run(name, func(t *testing.T) {
			for _, condition := range alertConditionCase {
				alertPolicy := validAlertPolicy()
				alertPolicy.Spec.Conditions[0] = condition
				err := validate(alertPolicy)
				testutils.AssertNoError(t, alertPolicy, err)
			}
		})
	}
	t.Run("fails, required", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.Conditions[0].Measurement = ""
		alertPolicy.Spec.Conditions[0].AlertingWindow = ""
		err := validate(alertPolicy)
		testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
			Prop: "spec.conditions[0].measurement",
			Code: rules.ErrorCodeRequired,
		})
	})
	t.Run("fails, invalid measurement", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.Conditions[0].Measurement = "Unknown"
		err := validate(alertPolicy)
		testutils.AssertContainsErrors(t, alertPolicy, err, 2, testutils.ExpectedError{
			Prop: "spec.conditions[0].measurement",
			Code: rules.ErrorCodeOneOf,
		})
	})
	t.Run("fails, alertingWindow is defined", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.Conditions[0].Measurement = MeasurementBurnedBudget.String()
		alertPolicy.Spec.Conditions[0].Value = 0.1
		err := validate(alertPolicy)
		testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
			Prop: "spec.conditions[0].measurement",
			ContainsMessage: fmt.Sprintf(
				`must be equal to one of '%s' when 'alertingWindow' is defined`,
				strings.Join(alertingWindowSupportedMeasurements(), ","),
			),
			Code: errorCodeMeasurementWithAlertingWindow,
		})
	})
	t.Run("fails, lastsFor is defined and alertingWindow is missing", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.Conditions[0].Measurement = MeasurementBudgetDrop.String()
		alertPolicy.Spec.Conditions[0].Value = 0.1
		alertPolicy.Spec.Conditions[0].LastsForDuration = "5m"
		alertPolicy.Spec.Conditions[0].AlertingWindow = ""
		err := validate(alertPolicy)
		testutils.AssertContainsErrors(t, alertPolicy, err, 2,
			testutils.ExpectedError{
				Prop: "spec.conditions[0].measurement",
				ContainsMessage: fmt.Sprintf(
					`must be equal to one of '%s' when 'lastsFor' is defined`,
					strings.Join(lastsForSupportedMeasurements(), ","),
				),
				Code: errorCodeMeasurementWithLastsFor,
			},
			testutils.ExpectedError{
				Prop: "spec.conditions[0].measurement",
				ContainsMessage: fmt.Sprintf(
					`alerting window is required for measurement '%s'`,
					MeasurementBudgetDrop.String(),
				),
				Code: rules.ErrorCodeRequired,
			},
		)
	})
}

func TestValidate_Spec_Condition_WithLastsFor_Value(t *testing.T) {
	passWithLastsForTests := map[string]measurementDetermined{
		"passes, valid duration when measurement is timeToBurnBudget or timeToBurnEntireBudget and " +
			"lastsFor is defined": {
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
		"passes, valid float, measurement is burnedBudget or averageBurnRate and lastsFor is defined": {
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
		"passes, allows empty values, measurement is burnedBudget or averageBurnRate": {
			values: []interface{}{"", 0.0},
			measurements: []Measurement{
				MeasurementAverageBurnRate,
				MeasurementBurnedBudget,
			},
		},
	}
	for name, testCase := range passWithLastsForTests {
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

	testCasesWithLastsFor := map[string]measurementDetermined{
		"fails, greater than 0 when measurement with lastsFor is timeToBurnBudget or timeToBurnEntireBudget": {
			values: []interface{}{
				"-1ms",
				"-15s",
				"-1h",
			},
			measurements: []Measurement{
				MeasurementTimeToBurnBudget,
				MeasurementTimeToBurnEntireBudget,
			},

			expectedCode:    rules.ErrorCodeGreaterThan,
			expectedMessage: "should be greater than '0s'",
		},
		"fails, unexpected format when measurement with lastsFor is averageBurnRate or burnedBudget": {
			values: []interface{}{
				"1.0",
				"1.9",
				"100",
			},
			measurements: []Measurement{
				MeasurementAverageBurnRate,
				MeasurementBurnedBudget,
			},
			expectedCode:    govy.ErrorCodeTransform,
			expectedMessage: "float64 expected, got ",
		},
		"fails, unexpected format when measurement with lastsFor is burnedBudget or averageBurnRate": {
			values: []interface{}{
				"-1ms",
				"s892k",
				"100",
			},
			measurements: []Measurement{
				MeasurementAverageBurnRate,
				MeasurementBurnedBudget,
			},
			expectedCode:    govy.ErrorCodeTransform,
			expectedMessage: "float64 expected, got ",
		},
	}
	for name, testCase := range testCasesWithLastsFor {
		t.Run(name, func(t *testing.T) {
			for _, value := range testCase.values {
				for _, measurement := range testCase.measurements {
					alertPolicy := validAlertPolicy()
					alertPolicy.Spec.Conditions[0].Measurement = measurement.String()
					alertPolicy.Spec.Conditions[0].Value = value
					alertPolicy.Spec.Conditions[0].LastsForDuration = "5m"
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

func TestValidate_Spec_Condition_WithAlertingWindow_Value(t *testing.T) {
	passWithAlertingWindowTests := map[string]measurementDetermined{
		"passes, valid duration when measurement is timeToBurnBudget or timeToBurnEntireBudget and " +
			"alertingWindow is defined": {
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
		"passes, valid float, measurement is burnedBudget or averageBurnRate and alertingWindow is defined": {
			values: []interface{}{
				0.000000020,
				0.97,
				2.00,
				157.00,
			},
			measurements: []Measurement{
				MeasurementAverageBurnRate,
				MeasurementBudgetDrop,
			},
		},
		"passes, allows empty values, measurement is burnedBudget, averageBurnRate or budgetDrop": {
			values: []interface{}{"", 0.0},
			measurements: []Measurement{
				MeasurementAverageBurnRate,
				MeasurementBudgetDrop,
			},
		},
	}
	for name, testCase := range passWithAlertingWindowTests {
		t.Run(name, func(t *testing.T) {
			for _, value := range testCase.values {
				for _, measurement := range testCase.measurements {
					alertPolicy := validAlertPolicy()
					alertPolicy.Spec.Conditions[0].Measurement = measurement.String()
					alertPolicy.Spec.Conditions[0].Value = value
					alertPolicy.Spec.Conditions[0].LastsForDuration = ""
					alertPolicy.Spec.Conditions[0].AlertingWindow = "10m"
					err := validate(alertPolicy)
					testutils.AssertNoError(t, alertPolicy, err)
				}
			}
		})
	}

	testCasesWithAlertingWindow := map[string]measurementDetermined{
		"fails, greater than 0 when measurement with alertingWindow is timeToBurnBudget or timeToBurnEntireBudget": {
			values: []interface{}{
				"-1ms",
				"-15s",
				"-1h",
			},
			measurements: []Measurement{
				MeasurementTimeToBurnBudget,
				MeasurementTimeToBurnEntireBudget,
			},

			expectedCode:    rules.ErrorCodeGreaterThan,
			expectedMessage: "should be greater than '0s'",
		},
		"fails, unexpected format when measurement with alertingWindow is averageBurnRate or budgetDrop": {
			values: []interface{}{
				"1.0",
				"1.9",
				"100",
			},
			measurements: []Measurement{
				MeasurementAverageBurnRate,
				MeasurementBudgetDrop,
			},
			expectedCode:    govy.ErrorCodeTransform,
			expectedMessage: "float64 expected, got ",
		},
		"fails, unexpected format when measurement with alertingWindow is budgetDrop or averageBurnRate": {
			values: []interface{}{
				"-1ms",
				"s892k",
				"100",
			},
			measurements: []Measurement{
				MeasurementAverageBurnRate,
				MeasurementBudgetDrop,
			},
			expectedCode:    govy.ErrorCodeTransform,
			expectedMessage: "float64 expected, got ",
		},
	}
	for name, testCase := range testCasesWithAlertingWindow {
		t.Run(name, func(t *testing.T) {
			for _, value := range testCase.values {
				for _, measurement := range testCase.measurements {
					alertPolicy := validAlertPolicy()
					alertPolicy.Spec.Conditions[0].Measurement = measurement.String()
					alertPolicy.Spec.Conditions[0].Value = value
					alertPolicy.Spec.Conditions[0].LastsForDuration = ""
					alertPolicy.Spec.Conditions[0].AlertingWindow = "10m"
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
	testCases := map[string]valuesWithCodeExpect{
		"fails, wrong format": {
			values:       []string{"1 hour"},
			expectedCode: govy.ErrorCodeTransform,
		},
		"fails, cannot parse unit format": {
			values: []string{
				"555d",
				"0.01y",
				"0.5w",
				"1w",
			},
			expectedCode: govy.ErrorCodeTransform,
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			for _, value := range testCase.values {
				alertPolicy := validAlertPolicy()
				alertPolicy.Spec.Conditions[0].AlertingWindow = value
				err := validate(alertPolicy)
				testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
					Prop: "spec.conditions[0].alertingWindow",
					Code: testCase.expectedCode,
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
				"360001ms",
			},
			expectedCode:    rules.ErrorCodeDurationPrecision,
			expectedMessage: "duration must be defined with 1m0s precision",
		},
		"fails, too long": {
			values: []string{
				"555h",
				"168h1m0s",
			},
			expectedCode:    rules.ErrorCodeLessThanOrEqualTo,
			expectedMessage: `should be less than or equal to '168h0m0s'`,
		},
		"fails, too short": {
			values: []string{
				"4m",
				"60000ms",
			},
			expectedCode:    rules.ErrorCodeGreaterThanOrEqualTo,
			expectedMessage: `should be greater than or equal to '5m0s'`,
		},
		"fails, zero value": {
			values: []string{
				"0",
				"0ms",
				"0s",
				"0m",
			},
			expectedCode:    govy.ErrorCodeTransform,
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
			values:       []string{"1 hour"},
			expectedCode: govy.ErrorCodeTransform,
		},
		"fails, wrong unit in format": {
			values:       []string{"365d"},
			expectedCode: govy.ErrorCodeTransform,
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
					Prop: "spec.conditions[0].lastsFor",
					Code: testCase.expectedCode,
				})
			}
		})
	}
}

func TestValidate_Spec_Condition_Operator(t *testing.T) {
	allValidOpts := []string{"gt", "lt", "lte", "gte", ""}

	t.Run("empty operator or only specific operator for measurement", func(t *testing.T) {
		testCases := []AlertCondition{
			// based on lasts for
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
				Measurement:      MeasurementAverageBurnRate.String(),
				LastsForDuration: "5m",
				Value:            30.0,
			},
			// based on alerting window
			{
				Measurement:    MeasurementTimeToBurnEntireBudget.String(),
				AlertingWindow: "10m",
				Value:          "30m",
			},
			{
				Measurement:    MeasurementTimeToBurnBudget.String(),
				AlertingWindow: "10m",
				Value:          "30m",
			},
			{
				Measurement:    MeasurementAverageBurnRate.String(),
				AlertingWindow: "5m",
				Value:          30.0,
			},
			{
				Measurement:    MeasurementBudgetDrop.String(),
				AlertingWindow: "5m",
				Value:          0.1,
			},
		}
		for _, alertCondition := range testCases {
			measurement, err := ParseMeasurement(alertCondition.Measurement)
			assert.NoError(t, err)
			expectedOperator, err := getExpectedOperatorForMeasurement(measurement)
			assert.NoError(t, err)

			allowedOps := []string{expectedOperator.String(), ""}
			for _, op := range allValidOpts {
				alertPolicy := validAlertPolicy()
				alertCondition.Operator = op
				alertPolicy.Spec.Conditions[0] = alertCondition
				err := validate(alertPolicy)
				if slices.Contains(allowedOps, op) {
					testutils.AssertNoError(t, alertPolicy, err)
				} else {
					testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
						Prop: "spec.conditions[0].op",
						Message: fmt.Sprintf(
							`measurement '%s' determines operator must be defined with '%s' or left empty`,
							measurement.String(), expectedOperator,
						),
					})
				}
			}
		}
	})

	t.Run("empty operator or any operator for measurement", func(t *testing.T) {
		testCases := []AlertCondition{
			{
				Measurement:      MeasurementBurnedBudget.String(),
				LastsForDuration: "10m",
				Value:            0.3,
			},
		}
		for _, alertCondition := range testCases {
			for _, op := range allValidOpts {
				alertPolicy := validAlertPolicy()
				alertCondition.Operator = op
				alertPolicy.Spec.Conditions[0] = alertCondition
				err := validate(alertPolicy)
				testutils.AssertNoError(t, alertPolicy, err)
			}
		}
	})

	t.Run("fails, invalid operator", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.Conditions[0].Operator = "noop"
		err := validate(alertPolicy)
		testutils.AssertContainsErrors(t, alertPolicy, err, 1, testutils.ExpectedError{
			Prop:    "spec.conditions[0].op",
			Message: "'noop' is not valid operator",
		})
	})
}

func TestValidate_Spec_AlertMethodsRefMetadata(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.AlertMethods = []AlertMethodRef{
			{
				Metadata: AlertMethodRefMetadata{
					Name:    "my-alert-method",
					Project: "my-project",
				},
			},
			{
				Metadata: AlertMethodRefMetadata{
					Name:    "my-alert-method",
					Project: "my-project",
				},
			},
			{
				Metadata: AlertMethodRefMetadata{
					Name: "my-alert-method-2",
				},
			},
		}
		err := validate(alertPolicy)
		testutils.AssertNoError(t, alertPolicy, err)
	})
	t.Run("fails, invalid name", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.AlertMethods = []AlertMethodRef{
			{
				Metadata: AlertMethodRefMetadata{
					Name: strings.Repeat("MY AlertMethodName", 20),
				},
			},
		}
		err := validate(alertPolicy)
		testutils.AssertContainsErrors(t, alertPolicy, err, 1,
			testutils.ExpectedError{
				Prop: "spec.alertMethods[0].metadata.name",
				Code: validationV1Alpha.ErrorCodeStringName,
			},
		)
	})
	t.Run("fails, invalid project", func(t *testing.T) {
		alertPolicy := validAlertPolicy()
		alertPolicy.Spec.AlertMethods = []AlertMethodRef{
			{
				Metadata: AlertMethodRefMetadata{
					Name:    "alert-method-name",
					Project: strings.Repeat("MY AlertMethodName", 20),
				},
			},
		}
		err := validate(alertPolicy)
		testutils.AssertContainsErrors(t, alertPolicy, err, 1,
			testutils.ExpectedError{
				Prop: "spec.alertMethods[0].metadata.project",
				Code: validationV1Alpha.ErrorCodeStringName,
			},
		)
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
			Conditions:       []AlertCondition{validAlertCondition()},
		},
	)
}

func validAlertCondition() AlertCondition {
	return AlertCondition{
		Measurement:    MeasurementAverageBurnRate.String(),
		Value:          0.97,
		AlertingWindow: "10m",
	}
}

type valuesWithCodeExpect struct {
	values          []string
	expectedCode    govy.ErrorCode
	expectedMessage string
}

type measurementDetermined struct {
	values          []interface{}
	measurements    []Measurement
	expectedCode    govy.ErrorCode
	expectedMessage string
}
