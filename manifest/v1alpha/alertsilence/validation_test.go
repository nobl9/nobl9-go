package alertsilence

import (
	_ "embed"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest"
)

var validationMessageRegexp = regexp.MustCompile(strings.TrimSpace(`
(?s)Validation for AlertSilence '.*' in project '.*' has failed for the following fields:
.*
Manifest source: /home/me/alertsilence.yaml
`))

func TestValidate_VersionAndKind(t *testing.T) {
	silence := validAlertSilence()
	silence.APIVersion = "v0.1"
	silence.Kind = manifest.KindProject
	silence.ManifestSource = "/home/me/alertsilence.yaml"
	err := validate(silence)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, silence, err, 2,
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
	silence := validAlertSilence()
	silence.Metadata = Metadata{
		Name:    strings.Repeat("MY ALERTSILENCE", 20),
		Project: strings.Repeat("MY PROJECT", 20),
	}
	silence.ManifestSource = "/home/me/alertsilence.yaml"
	err := validate(silence)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, silence, err, 2,
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

func TestValidate_Metadata_Project(t *testing.T) {
	t.Run("fails, project required", func(t *testing.T) {
		alertSilence := validAlertSilence()
		alertSilence.Metadata.Project = ""
		err := validate(alertSilence)
		testutils.AssertContainsErrors(t, alertSilence, err, 1, testutils.ExpectedError{
			Prop: "metadata.project",
			Code: rules.ErrorCodeRequired,
		})
	})
}

func TestValidate_Spec_Description(t *testing.T) {
	alertSilence := validAlertSilence()
	alertSilence.Spec.Description = strings.Repeat("A", 2000)
	err := validate(alertSilence)
	testutils.AssertContainsErrors(t, alertSilence, err, 1, testutils.ExpectedError{
		Prop: "spec.description",
		Code: validationV1Alpha.ErrorCodeStringDescription,
	})
}

func TestValidate_Spec_Slo(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		alertSilence := validAlertSilence()
		alertSilence.Spec.SLO = "my-slo"
		err := validate(alertSilence)
		testutils.AssertNoError(t, alertSilence, err)
	})
	t.Run("fails", func(t *testing.T) {
		alertSilence := validAlertSilence()
		alertSilence.Spec.SLO = "MY SLO"
		err := validate(alertSilence)
		testutils.AssertContainsErrors(t, alertSilence, err, 1, testutils.ExpectedError{
			Prop: "spec.slo",
			Code: validationV1Alpha.ErrorCodeStringName,
		})
	})
	t.Run("fails, required", func(t *testing.T) {
		alertSilence := validAlertSilence()
		alertSilence.Spec.SLO = ""
		err := validate(alertSilence)
		testutils.AssertContainsErrors(t, alertSilence, err, 1, testutils.ExpectedError{
			Prop: "spec.slo",
			Code: rules.ErrorCodeRequired,
		})
	})
}

func TestValidate_Spec_AlertPolicy(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		alertSilence := validAlertSilence()
		alertSilence.Metadata.Project = "project"
		alertSilence.Spec.AlertPolicy = AlertPolicySource{
			Name:    "alert-policy-name",
			Project: "project",
		}
		err := validate(alertSilence)
		testutils.AssertNoError(t, alertSilence, err)
	})
	t.Run("passes, empty project", func(t *testing.T) {
		alertSilence := validAlertSilence()
		alertSilence.Spec.AlertPolicy = AlertPolicySource{
			Name:    "alert-policy-name",
			Project: "",
		}
		err := validate(alertSilence)
		testutils.AssertNoError(t, alertSilence, err)
	})
	t.Run("passes, consistent project with metadata", func(t *testing.T) {
		alertSilence := validAlertSilence()
		alertSilence.Metadata.Project = "default"
		alertSilence.Spec.AlertPolicy = AlertPolicySource{
			Name:    "alert-policy-name",
			Project: "default",
		}
		err := validate(alertSilence)
		testutils.AssertNoError(t, alertSilence, err)
	})
	t.Run("fails, invalid name", func(t *testing.T) {
		alertSilence := validAlertSilence()
		alertSilence.Spec.AlertPolicy.Name = "not valid NAME !!"
		err := validate(alertSilence)
		testutils.AssertContainsErrors(t, alertSilence, err, 1, testutils.ExpectedError{
			Prop: "spec.alertPolicy.name",
			Code: validationV1Alpha.ErrorCodeStringName,
		})
	})
	t.Run("fails, required", func(t *testing.T) {
		alertSilence := validAlertSilence()
		alertSilence.Spec.AlertPolicy.Name = ""
		err := validate(alertSilence)
		testutils.AssertContainsErrors(t, alertSilence, err, 1, testutils.ExpectedError{
			Prop: "spec.alertPolicy.name",
			Code: rules.ErrorCodeRequired,
		})
	})
	t.Run("fails, invalid project", func(t *testing.T) {
		alertSilence := validAlertSilence()
		alertSilence.Spec.AlertPolicy.Project = "not valid NAME !!"
		err := validate(alertSilence)
		testutils.AssertContainsErrors(t, alertSilence, err, 2, testutils.ExpectedError{
			Prop: "spec.alertPolicy.project",
			Code: validationV1Alpha.ErrorCodeStringName,
		})
	})
	t.Run("fails, inconsistent project", func(t *testing.T) {
		alertSilence := validAlertSilence()
		alertSilence.Metadata.Project = "project-1"
		alertSilence.Spec.AlertPolicy.Project = "project-2"
		err := validate(alertSilence)
		testutils.AssertContainsErrors(t, alertSilence, err, 1, testutils.ExpectedError{
			Prop: "spec.alertPolicy.project",
			Code: errorCodeInconsistentProject,
		})
	})
}

func TestValidate_Spec_Period(t *testing.T) {
	t.Run("fails, no endTime and duration provided", func(t *testing.T) {
		alertSilence := validAlertSilence()
		alertSilence.Spec.Period.EndTime = nil
		alertSilence.Spec.Period.Duration = ""
		err := validate(alertSilence)
		testutils.AssertContainsErrors(t, alertSilence, err, 1, testutils.ExpectedError{
			Prop: "spec.period",
			Code: rules.ErrorCodeMutuallyExclusive,
		})
	})
	t.Run("fails, both endTime and duration provided", func(t *testing.T) {
		alertSilence := validAlertSilence()
		alertSilence.Spec.Period.EndTime = ptr(
			time.Date(2023, 5, 11, 17, 10, 5, 0, time.UTC),
		)
		alertSilence.Spec.Period.Duration = "10m"
		err := validate(alertSilence)
		testutils.AssertContainsErrors(t, alertSilence, err, 1, testutils.ExpectedError{
			Prop: "spec.period",
			Code: rules.ErrorCodeMutuallyExclusive,
		})
	})
}

func TestValidate_Spec_Period_Duration(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		alertSilence := validAlertSilence()
		alertSilence.Spec.Period.StartTime = nil
		alertSilence.Spec.Period.EndTime = nil
		alertSilence.Spec.Period.Duration = "3m"
		err := validate(alertSilence)
		testutils.AssertNoError(t, alertSilence, err)
	})
	t.Run("passes, allowed empty", func(t *testing.T) {
		endTime := time.Now()
		alertSilence := validAlertSilence()
		alertSilence.Spec.Period.EndTime = &endTime
		alertSilence.Spec.Period.Duration = ""
		err := validate(alertSilence)
		testutils.AssertNoError(t, alertSilence, err)
	})
	t.Run("fails, invalid format", func(t *testing.T) {
		alertSilence := validAlertSilence()
		alertSilence.Spec.Period.StartTime = nil
		alertSilence.Spec.Period.EndTime = nil
		alertSilence.Spec.Period.Duration = "3 months"
		err := validate(alertSilence)
		testutils.AssertContainsErrors(t, alertSilence, err, 1, testutils.ExpectedError{
			Prop: "spec.period.duration",
			Code: govy.ErrorCodeTransform,
		})
	})
	t.Run("fails, invalid too small", func(t *testing.T) {
		alertSilence := validAlertSilence()
		alertSilence.Spec.Period.StartTime = nil
		alertSilence.Spec.Period.EndTime = nil
		alertSilence.Spec.Period.Duration = "0s"
		err := validate(alertSilence)
		testutils.AssertContainsErrors(t, alertSilence, err, 1, testutils.ExpectedError{
			Prop: "spec.period.duration",
			Code: rules.ErrorCodeGreaterThan,
		})
	})
}

func TestValidate_Spec_EndTime(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		alertSilence := validAlertSilence()
		alertSilence.Spec.Period.StartTime = nil
		alertSilence.Spec.Period.Duration = ""
		alertSilence.Spec.Period.EndTime = ptr(
			time.Date(2023, 5, 1, 17, 10, 5, 0, time.UTC),
		)
		err := validate(alertSilence)
		testutils.AssertNoError(t, alertSilence, err)
	})

	testsCases := map[string]Period{
		"fails, end time before start time": {
			StartTime: ptr(time.Date(2023, 5, 14, 17, 10, 5, 0, time.UTC)),
			EndTime:   ptr(time.Date(2023, 5, 11, 17, 10, 5, 0, time.UTC)),
		},
		"fails, end time equals start time": {
			StartTime: ptr(time.Date(2023, 5, 14, 17, 10, 5, 0, time.UTC)),
			EndTime:   ptr(time.Date(2023, 5, 14, 17, 10, 5, 0, time.UTC)),
		},
	}
	for name, testCase := range testsCases {
		t.Run(name, func(t *testing.T) {
			alertSilence := validAlertSilence()
			alertSilence.Spec.Period = testCase
			err := validate(alertSilence)
			testutils.AssertContainsErrors(t, alertSilence, err, 1, testutils.ExpectedError{
				Prop: "spec.period",
				Code: errorCodeEndTimeNotBeforeOrNotEqualStartTime,
			})
		})
	}
}

func validAlertSilence() AlertSilence {
	return New(
		Metadata{
			Name:    "alert-silence",
			Project: "default",
		},
		Spec{
			Description: "Example alert silence",
			SLO:         "existing-slo",
			AlertPolicy: AlertPolicySource{
				Name:    "alert-policy",
				Project: "default",
			},
			Period: Period{
				StartTime: ptr(time.Date(2023, 5, 1, 17, 10, 5, 0, time.UTC)),
				Duration:  "10m",
			},
		},
	)
}

func ptr[T any](v T) *T { return &v }
