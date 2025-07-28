package service

import (
	_ "embed"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/manifest/v1alphatest"
	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest"
)

var validationMessageRegexp = regexp.MustCompile(strings.TrimSpace(`
(?s)Validation for Service '.*' in project '.*' has failed for the following fields:
.*
Manifest source: /home/me/service.yaml
`))

func TestValidate_VersionAndKind(t *testing.T) {
	svc := validService()
	svc.APIVersion = "v0.1"
	svc.Kind = manifest.KindProject
	svc.ManifestSource = "/home/me/service.yaml"
	err := validate(svc)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, svc, err, 2,
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
	svc := validService()
	svc.Metadata = Metadata{
		Name:        strings.Repeat("MY SERVICE", 20),
		DisplayName: strings.Repeat("my-service", 20),
		Project:     strings.Repeat("MY PROJECT", 20),
	}
	svc.ManifestSource = "/home/me/service.yaml"
	err := validate(svc)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, svc, err, 3,
		testutils.ExpectedError{
			Prop: "metadata.name",
			Code: rules.ErrorCodeStringDNSLabel,
		},
		testutils.ExpectedError{
			Prop: "metadata.displayName",
			Code: rules.ErrorCodeStringLength,
		},
		testutils.ExpectedError{
			Prop: "metadata.project",
			Code: rules.ErrorCodeStringDNSLabel,
		},
	)
}

func TestValidate_Metadata_Labels(t *testing.T) {
	for name, test := range v1alphatest.GetLabelsTestCases[Service](t, "metadata.labels") {
		t.Run(name, func(t *testing.T) {
			svc := validService()
			svc.Metadata.Labels = test.Labels
			test.Test(t, svc, validate)
		})
	}
}

func TestValidate_Metadata_Annotations(t *testing.T) {
	for name, test := range v1alphatest.GetMetadataAnnotationsTestCases[Service](t, "metadata.annotations") {
		t.Run(name, func(t *testing.T) {
			svc := validService()
			svc.Metadata.Annotations = test.Annotations
			test.Test(t, svc, validate)
		})
	}
}

func TestValidate_Spec(t *testing.T) {
	t.Run("description too long", func(t *testing.T) {
		svc := validService()
		svc.Spec.Description = strings.Repeat("A", 2000)
		err := validate(svc)
		testutils.AssertContainsErrors(t, svc, err, 1, testutils.ExpectedError{
			Prop: "spec.description",
			Code: validationV1Alpha.ErrorCodeStringDescription,
		})
	})
}

func TestValidate_ReviewCycle(t *testing.T) {
	t.Run("valid reviewCycle", func(t *testing.T) {
		svc := validService()
		svc.Spec.ReviewCycle = &ReviewCycle{
			StartTime: "2023-01-01T00:00:00",
			TimeZone:  "UTC",
			RRule:     "FREQ=MONTHLY;INTERVAL=1",
		}
		err := validate(svc)
		testutils.AssertNoError(t, svc, err)
	})

	t.Run("invalid startTime - wrong format", func(t *testing.T) {
		svc := validService()
		svc.Spec.ReviewCycle = &ReviewCycle{
			StartTime: "2023-01-01",
			TimeZone:  "UTC",
			RRule:     "FREQ=MONTHLY;INTERVAL=1",
		}
		err := validate(svc)
		testutils.AssertContainsErrors(t, svc, err, 1, testutils.ExpectedError{
			Prop: "spec.reviewCycle.startTime",
			Code: "string_date_time",
		})
	})

	t.Run("empty startTime", func(t *testing.T) {
		svc := validService()
		svc.Spec.ReviewCycle = &ReviewCycle{
			StartTime: "",
			TimeZone:  "UTC",
			RRule:     "FREQ=MONTHLY;INTERVAL=1",
		}
		err := validate(svc)
		testutils.AssertContainsErrors(t, svc, err, 1, testutils.ExpectedError{
			Prop: "spec.reviewCycle.startTime",
			Code: rules.ErrorCodeRequired,
		})
	})

	t.Run("empty timeZone", func(t *testing.T) {
		svc := validService()
		svc.Spec.ReviewCycle = &ReviewCycle{
			StartTime: "2023-01-01T00:00:00",
			TimeZone:  "",
			RRule:     "FREQ=MONTHLY;INTERVAL=1",
		}
		err := validate(svc)
		testutils.AssertContainsErrors(t, svc, err, 1, testutils.ExpectedError{
			Prop: "spec.reviewCycle.timeZone",
			Code: rules.ErrorCodeRequired,
		})
	})

	t.Run("invalid timeZone", func(t *testing.T) {
		svc := validService()
		svc.Spec.ReviewCycle = &ReviewCycle{
			StartTime: "2023-01-01T00:00:00",
			TimeZone:  "Invalid/Timezone",
			RRule:     "FREQ=MONTHLY;INTERVAL=1",
		}
		err := validate(svc)
		testutils.AssertContainsErrors(t, svc, err, 1, testutils.ExpectedError{
			Prop: "spec.reviewCycle.timeZone",
			Code: "string_time_zone",
		})
	})

	t.Run("invalid rrule - missing FREQ", func(t *testing.T) {
		svc := validService()
		svc.Spec.ReviewCycle = &ReviewCycle{
			StartTime: "2023-01-01T00:00:00",
			TimeZone:  "UTC",
			RRule:     "INTERVAL=1",
		}
		err := validate(svc)
		testutils.AssertContainsErrors(t, svc, err, 1, testutils.ExpectedError{
			Prop: "spec.reviewCycle.rrule",
			Code: "transform",
		})
	})

	t.Run("empty rrule", func(t *testing.T) {
		svc := validService()
		svc.Spec.ReviewCycle = &ReviewCycle{
			StartTime: "2023-01-01T00:00:00",
			TimeZone:  "UTC",
			RRule:     "",
		}
		err := validate(svc)
		testutils.AssertContainsErrors(t, svc, err, 1, testutils.ExpectedError{
			Prop: "spec.reviewCycle.rrule",
			Code: rules.ErrorCodeRequired,
		})
	})

	t.Run("invalid rrule - hourly frequency not allowed", func(t *testing.T) {
		svc := validService()
		svc.Spec.ReviewCycle = &ReviewCycle{
			StartTime: "2023-01-01T00:00:00",
			TimeZone:  "UTC",
			RRule:     "FREQ=HOURLY;INTERVAL=1",
		}
		err := validate(svc)
		testutils.AssertContainsErrors(t, svc, err, 1, testutils.ExpectedError{
			Prop:    "spec.reviewCycle.rrule",
			Message: "rrule frequency must be at least DAILY",
		})
	})

	t.Run("invalid rrule - minutely frequency not allowed", func(t *testing.T) {
		svc := validService()
		svc.Spec.ReviewCycle = &ReviewCycle{
			StartTime: "2023-01-01T00:00:00",
			TimeZone:  "UTC",
			RRule:     "FREQ=MINUTELY;INTERVAL=60",
		}
		err := validate(svc)
		testutils.AssertContainsErrors(t, svc, err, 1, testutils.ExpectedError{
			Prop:    "spec.reviewCycle.rrule",
			Message: "rrule frequency must be at least DAILY",
		})
	})

	t.Run("valid rrule - daily frequency allowed", func(t *testing.T) {
		svc := validService()
		svc.Spec.ReviewCycle = &ReviewCycle{
			StartTime: "2023-01-01T00:00:00",
			TimeZone:  "UTC",
			RRule:     "FREQ=DAILY;INTERVAL=1",
		}
		err := validate(svc)
		testutils.AssertNoError(t, svc, err)
	})

	t.Run("valid rrule - weekly frequency allowed", func(t *testing.T) {
		svc := validService()
		svc.Spec.ReviewCycle = &ReviewCycle{
			StartTime: "2023-01-01T00:00:00",
			TimeZone:  "UTC",
			RRule:     "FREQ=WEEKLY;INTERVAL=1",
		}
		err := validate(svc)
		testutils.AssertNoError(t, svc, err)
	})

	t.Run("valid rrule - monthly frequency allowed", func(t *testing.T) {
		svc := validService()
		svc.Spec.ReviewCycle = &ReviewCycle{
			StartTime: "2023-01-01T00:00:00",
			TimeZone:  "UTC",
			RRule:     "FREQ=MONTHLY;INTERVAL=1",
		}
		err := validate(svc)
		testutils.AssertNoError(t, svc, err)
	})

	t.Run("valid rrule - yearly frequency allowed", func(t *testing.T) {
		svc := validService()
		svc.Spec.ReviewCycle = &ReviewCycle{
			StartTime: "2023-01-01T00:00:00",
			TimeZone:  "UTC",
			RRule:     "FREQ=YEARLY;INTERVAL=1",
		}
		err := validate(svc)
		testutils.AssertNoError(t, svc, err)
	})

	t.Run("valid rrule - single occurrence with any frequency allowed", func(t *testing.T) {
		svc := validService()
		svc.Spec.ReviewCycle = &ReviewCycle{
			StartTime: "2023-01-01T00:00:00",
			TimeZone:  "UTC",
			RRule:     "FREQ=HOURLY;COUNT=1",
		}
		err := validate(svc)
		testutils.AssertNoError(t, svc, err)
	})

	t.Run("nil reviewCycle should be valid", func(t *testing.T) {
		svc := validService()
		svc.Spec.ReviewCycle = nil
		err := validate(svc)
		testutils.AssertNoError(t, svc, err)
	})
}

func validService() Service {
	return New(
		Metadata{
			Name:    "service",
			Project: "default",
		},
		Spec{},
	)
}
