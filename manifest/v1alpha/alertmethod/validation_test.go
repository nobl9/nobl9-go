package alertmethod

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/validation"
)

//go:embed test_data/expected_metadata_error.txt
var expectedMetadataError string

func TestValidate_Metadata(t *testing.T) {
	alertMethod := validAlertMethod()
	alertMethod.Metadata = Metadata{
		Name:        strings.Repeat("MY ALERTMETHOD", 20),
		DisplayName: strings.Repeat("my-alertmethod", 10),
		Project:     strings.Repeat("MY PROJECT", 20),
	}
	alertMethod.ManifestSource = "/home/me/alertmethod.yaml"
	err := validate(alertMethod)
	require.Error(t, err)
	assert.Equal(t, strings.TrimSuffix(expectedMetadataError, "\n"), err.Error())
}

func TestValidate_Spec(t *testing.T) {
	t.Run("fails with too long description", func(t *testing.T) {
		alertMethod := validAlertMethod()
		alertMethod.Spec.Description = strings.Repeat("l", 1051)
		err := validate(alertMethod)
		testutils.AssertContainsErrors(t, alertMethod, err, 1, testutils.ExpectedError{
			Prop: "spec.description",
			Code: validation.ErrorCodeStringLength,
		})
	})
	t.Run("fails with empty spec", func(t *testing.T) {
		alertMethod := validAlertMethod()
		alertMethod.Spec = Spec{}
		err := validate(alertMethod)
		testutils.AssertContainsErrors(t, alertMethod, err, 1, testutils.ExpectedError{
			Prop:    "spec",
			Message: "exactly one alert method configuration is required",
		})
	})
	t.Run("fails with more than one method defined in spec", func(t *testing.T) {
		alertMethod := validAlertMethod()
		alertMethod.Spec = Spec{
			Slack: &SlackAlertMethod{
				URL: "https://example.com",
			},
			Teams: &TeamsAlertMethod{
				URL: "https://example.com",
			},
		}
		err := validate(alertMethod)
		testutils.AssertContainsErrors(t, alertMethod, err, 1, testutils.ExpectedError{
			Prop:    "spec",
			Message: "exactly one alert method configuration is required",
		})
	})
}

func ptr[T any](v T) *T { return &v }

func validAlertMethod() AlertMethod {
	return New(
		Metadata{
			Name:        "my-alertmethod",
			DisplayName: "my alertmethod",
			Project:     "default",
		},
		Spec{
			Slack: &SlackAlertMethod{
				URL: "https://example.com",
			},
		},
	)
}
