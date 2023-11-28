package alertpolicy

import (
	_ "embed"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest"
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
			CoolDownDuration: "10s",
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
			Code: errorCodeSeverity,
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
			Description: "Example alertPolicy",
			Severity:    v1alpha.SeverityHigh.String(),
		},
	)
}
