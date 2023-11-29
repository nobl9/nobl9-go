package alertpolicy

import (
	_ "embed"
	"fmt"
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
			Conditions:       nil,
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
			Message: fmt.Sprintf(
				`severity must be set to one of the values: %s`,
				strings.Join([]string{
					v1alpha.SeverityLow.String(),
					v1alpha.SeverityMedium.String(),
					v1alpha.SeverityHigh.String(),
				}, ", ")),
			Code: errorCodeSeverity,
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
			expectedCode:    errorCodeDuration,
			expectedMessage: `time: unknown unit " hour" in duration "1 hour"`,
		},
		"fails, negative": {
			value:           "-1m",
			expectedCode:    errorCodeDurationNotNegative,
			expectedMessage: "duration '-1m' must be not negative value",
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
		},
	)
}

type valueWithCodeExpect struct {
	value           string
	expectedCode    string
	expectedMessage string
}
