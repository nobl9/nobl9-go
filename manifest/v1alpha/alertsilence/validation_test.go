package alertsilence

import (
	_ "embed"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/validation"
)

//go:embed test_data/expected_error.txt
var expectedError string

func TestValidate_Metadata(t *testing.T) {
	periodStart := time.Date(2023, 5, 1, 17, 10, 5, 0, time.UTC)

	err := validate(AlertSilence{
		Kind: manifest.KindAlertSilence,
		Metadata: Metadata{
			Name:    strings.Repeat("MY ALERTSILENCE", 20),
			Project: strings.Repeat("MY PROJECT", 20),
		},
		Spec: Spec{
			Slo: "slo-name",
			AlertPolicy: AlertPolicySource{
				Name:    "my-alert-policy",
				Project: "default",
			},
			Description: strings.Repeat("l", 2000),
			Period: Period{
				StartTime: &periodStart,
				Duration:  "10m",
			},
		},
		ManifestSource: "/home/me/alertsilence.yaml",
	})
	assert.Equal(t, strings.TrimSuffix(expectedError, "\n"), err.Error())
}

func TestValidate_Metadata_Project(t *testing.T) {
	t.Run("passes, no project", func(t *testing.T) {
		alertSilence := validAlertSilence()
		alertSilence.Metadata.Project = ""
		err := validate(alertSilence)
		testutils.AssertNoError(t, alertSilence, err)
	})
}

func TestValidate_Spec_Slo(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		alertSilence := validAlertSilence()
		alertSilence.Spec.Slo = "my-slo"
		err := validate(alertSilence)
		testutils.AssertNoError(t, alertSilence, err)
	})
	t.Run("fails", func(t *testing.T) {
		alertSilence := validAlertSilence()
		alertSilence.Spec.Slo = "MY SLO"
		err := validate(alertSilence)
		testutils.AssertContainsErrors(t, alertSilence, err, 1, testutils.ExpectedError{
			Prop: "spec.slo",
			Code: validation.ErrorCodeStringIsDNSSubdomain,
		})
	})
	t.Run("fails, required", func(t *testing.T) {
		alertSilence := validAlertSilence()
		alertSilence.Spec.Slo = ""
		err := validate(alertSilence)
		testutils.AssertContainsErrors(t, alertSilence, err, 1, testutils.ExpectedError{
			Prop: "spec.slo",
			Code: validation.ErrorCodeRequired,
		})
	})
}

func validAlertSilence() AlertSilence {
	periodStart := time.Date(2023, 5, 1, 17, 10, 5, 0, time.UTC)

	return New(
		Metadata{
			Name:    "alert-silence",
			Project: "default",
		},
		Spec{
			Description: "Example alert silence",
			Slo:         "existing-slo",
			AlertPolicy: AlertPolicySource{
				Name:    "alert-policy",
				Project: "default",
			},
			Period: Period{
				StartTime: &periodStart,
				Duration:  "10m",
			},
		},
	)
}
