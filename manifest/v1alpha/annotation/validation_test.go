package annotation

import (
	_ "embed"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/validation"
)

//go:embed test_data/expected_error.txt
var expectedError string

func TestValidate_AllErrors(t *testing.T) {
	err := validate(
		New(
			Metadata{
				Name:    strings.Repeat("MY ANNOTATION", 20),
				Project: strings.Repeat("MY ANNOTATION", 20),
			},
			Spec{
				Slo:           "existing-slo",
				ObjectiveName: "existing-slo-objective-1",
				Description:   "Example annotation",
				StartTime:     time.Date(2023, 5, 1, 17, 10, 5, 0, time.UTC),
				EndTime:       time.Date(2023, 5, 2, 17, 10, 5, 0, time.UTC),
			},
		),
	)
	assert.Equal(t, strings.TrimSuffix(expectedError, "\n"), err.Error())
}

func TestValidate_Spec_Slo(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		annotation := validAnnotation()
		annotation.Spec.Slo = "my-slo"
		err := validate(annotation)
		testutils.AssertNoError(t, annotation, err)
	})
	t.Run("fails", func(t *testing.T) {
		annotation := validAnnotation()
		annotation.Spec.Slo = "MY SLO"
		err := validate(annotation)
		testutils.AssertContainsErrors(t, annotation, err, 1, testutils.ExpectedError{
			Prop: "spec.slo",
			Code: validation.ErrorCodeStringIsDNSSubdomain,
		})
	})
}

func TestValidate_Spec_Objective(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		annotation := validAnnotation()
		annotation.Spec.ObjectiveName = "my-objective"
		err := validate(annotation)
		testutils.AssertNoError(t, annotation, err)
	})
	t.Run("fails", func(t *testing.T) {
		annotation := validAnnotation()
		annotation.Spec.ObjectiveName = "MY OBJECTIVE"
		err := validate(annotation)
		testutils.AssertContainsErrors(t, annotation, err, 1, testutils.ExpectedError{
			Prop: "spec.objectiveName",
			Code: validation.ErrorCodeStringIsDNSSubdomain,
		})
	})
}

func TestSpec_Time(t *testing.T) {
	t.Run("passes, end time after start time", func(t *testing.T) {
		annotation := validAnnotation()
		annotation.Spec.StartTime = time.Date(2023, 5, 1, 17, 10, 5, 0, time.UTC)
		annotation.Spec.EndTime = time.Date(2023, 5, 14, 17, 10, 5, 0, time.UTC)
		err := validate(annotation)
		testutils.AssertNoError(t, annotation, err)
	})

	t.Run("fails, end time is not after start time", func(t *testing.T) {
		tests := map[string]Spec{
			"end time equals start time": {
				Slo:           "my-slo",
				ObjectiveName: "my-obj",
				Description:   "my-annotation description",
				StartTime:     time.Date(2023, 5, 1, 17, 10, 5, 0, time.UTC),
				EndTime:       time.Date(2023, 5, 1, 17, 10, 5, 0, time.UTC),
			},
			"end time is before start time": {
				Slo:           "my-slo",
				ObjectiveName: "my-obj",
				Description:   "my-annotation description",
				StartTime:     time.Date(2023, 5, 5, 17, 10, 5, 0, time.UTC),
				EndTime:       time.Date(2023, 5, 2, 17, 10, 5, 0, time.UTC),
			},
		}
		for name, spec := range tests {
			t.Run(name, func(t *testing.T) {
				annotation := New(
					Metadata{Name: "my-name"},
					Spec{
						Slo:           "my-slo",
						ObjectiveName: "my-obj",
						Description:   "my-annotation description",
						StartTime:     spec.StartTime,
						EndTime:       spec.EndTime,
					},
				)
				err := validate(annotation)
				testutils.AssertContainsErrors(t, annotation, err, 1, testutils.ExpectedError{
					Prop: "spec",
					Code: errorCodeEndTimeAfterStartTime,
				})
			})
		}
	})
}

func validAnnotation() Annotation {
	return New(
		Metadata{
			Name:    "annotation",
			Project: "project",
		},
		Spec{
			Slo:           "existing-slo",
			ObjectiveName: "existing-slo-objective-1",
			Description:   "Example annotation",
			StartTime:     time.Date(2023, 5, 1, 17, 10, 5, 0, time.UTC),
			EndTime:       time.Date(2023, 5, 2, 17, 10, 5, 0, time.UTC),
		},
	)
}
