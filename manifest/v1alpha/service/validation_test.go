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
			StartDate: "2023-01-01T00:00:00Z",
			RRule:     "FREQ=MONTHLY;INTERVAL=1",
		}
		err := validate(svc)
		testutils.AssertNoError(t, svc, err)
	})

	t.Run("invalid startDate - not RFC3339", func(t *testing.T) {
		svc := validService()
		svc.Spec.ReviewCycle = &ReviewCycle{
			StartDate: "2023-01-01",
			RRule:     "FREQ=MONTHLY;INTERVAL=1",
		}
		err := validate(svc)
		testutils.AssertContainsErrors(t, svc, err, 1, testutils.ExpectedError{
			Prop: "spec.reviewCycle.startDate",
			Code: "invalid_rfc3339",
		})
	})

	t.Run("empty startDate", func(t *testing.T) {
		svc := validService()
		svc.Spec.ReviewCycle = &ReviewCycle{
			StartDate: "",
			RRule:     "FREQ=MONTHLY;INTERVAL=1",
		}
		err := validate(svc)
		testutils.AssertContainsErrors(t, svc, err, 1, testutils.ExpectedError{
			Prop: "spec.reviewCycle.startDate",
			Code: rules.ErrorCodeRequired,
		})
	})

	t.Run("invalid rrule - missing FREQ", func(t *testing.T) {
		svc := validService()
		svc.Spec.ReviewCycle = &ReviewCycle{
			StartDate: "2023-01-01T00:00:00Z",
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
			StartDate: "2023-01-01T00:00:00Z",
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
			StartDate: "2023-01-01T00:00:00Z",
			RRule:     "FREQ=HOURLY;INTERVAL=1",
		}
		err := validate(svc)
		testutils.AssertContainsErrors(t, svc, err, 1, testutils.ExpectedError{
			Prop:    "spec.reviewCycle.rrule",
			Message: "frequency must be daily, weekly, monthly, or yearly",
		})
	})

	t.Run("invalid rrule - minutely frequency not allowed", func(t *testing.T) {
		svc := validService()
		svc.Spec.ReviewCycle = &ReviewCycle{
			StartDate: "2023-01-01T00:00:00Z",
			RRule:     "FREQ=MINUTELY;INTERVAL=60",
		}
		err := validate(svc)
		testutils.AssertContainsErrors(t, svc, err, 1, testutils.ExpectedError{
			Prop:    "spec.reviewCycle.rrule",
			Message: "frequency must be daily, weekly, monthly, or yearly",
		})
	})

	t.Run("valid rrule - daily frequency allowed", func(t *testing.T) {
		svc := validService()
		svc.Spec.ReviewCycle = &ReviewCycle{
			StartDate: "2023-01-01T00:00:00Z",
			RRule:     "FREQ=DAILY;INTERVAL=1",
		}
		err := validate(svc)
		testutils.AssertNoError(t, svc, err)
	})

	t.Run("valid rrule - weekly frequency allowed", func(t *testing.T) {
		svc := validService()
		svc.Spec.ReviewCycle = &ReviewCycle{
			StartDate: "2023-01-01T00:00:00Z",
			RRule:     "FREQ=WEEKLY;INTERVAL=1",
		}
		err := validate(svc)
		testutils.AssertNoError(t, svc, err)
	})

	t.Run("valid rrule - monthly frequency allowed", func(t *testing.T) {
		svc := validService()
		svc.Spec.ReviewCycle = &ReviewCycle{
			StartDate: "2023-01-01T00:00:00Z",
			RRule:     "FREQ=MONTHLY;INTERVAL=1",
		}
		err := validate(svc)
		testutils.AssertNoError(t, svc, err)
	})

	t.Run("valid rrule - yearly frequency allowed", func(t *testing.T) {
		svc := validService()
		svc.Spec.ReviewCycle = &ReviewCycle{
			StartDate: "2023-01-01T00:00:00Z",
			RRule:     "FREQ=YEARLY;INTERVAL=1",
		}
		err := validate(svc)
		testutils.AssertNoError(t, svc, err)
	})

	t.Run("valid rrule - single occurrence with any frequency allowed", func(t *testing.T) {
		svc := validService()
		svc.Spec.ReviewCycle = &ReviewCycle{
			StartDate: "2023-01-01T00:00:00Z",
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
