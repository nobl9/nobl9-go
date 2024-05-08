package annotation

import (
	_ "embed"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest"
)

var validationMessageRegexp = regexp.MustCompile(strings.TrimSpace(`
(?s)Validation for Annotation '.*' in project '.*' has failed for the following fields:
.*
Manifest source: /home/me/annotation.yaml
`))

func TestValidate_VersionAndKind(t *testing.T) {
	annotation := validAnnotation()
	annotation.APIVersion = "v0.1"
	annotation.Kind = manifest.KindProject
	annotation.ManifestSource = "/home/me/annotation.yaml"
	err := validate(annotation)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, annotation, err, 2,
		testutils.ExpectedError{
			Prop: "apiVersion",
			Code: validation.ErrorCodeEqualTo,
		},
		testutils.ExpectedError{
			Prop: "kind",
			Code: validation.ErrorCodeEqualTo,
		},
	)
}

func TestValidate_Metadata(t *testing.T) {
	annotation := validAnnotation()
	annotation.Metadata = Metadata{
		Name:    strings.Repeat("MY ANNOTATION", 20),
		Project: strings.Repeat("MY ANNOTATION", 20),
	}
	annotation.ManifestSource = "/home/me/annotation.yaml"
	err := validate(annotation)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, annotation, err, 4,
		testutils.ExpectedError{
			Prop: "metadata.name",
			Code: validation.ErrorCodeStringIsDNSSubdomain,
		},
		testutils.ExpectedError{
			Prop: "metadata.project",
			Code: validation.ErrorCodeStringIsDNSSubdomain,
		},
	)
}

func TestValidate_Metadata_Project(t *testing.T) {
	t.Run("passes, no project", func(t *testing.T) {
		annotation := validAnnotation()
		annotation.Metadata.Project = ""
		err := validate(annotation)
		testutils.AssertNoError(t, annotation, err)
	})
}

func TestValidate_Spec_Description(t *testing.T) {
	t.Run("too long", func(t *testing.T) {
		annotation := validAnnotation()
		annotation.Spec.Description = strings.Repeat("A", 2000)
		err := validate(annotation)
		testutils.AssertContainsErrors(t, annotation, err, 1, testutils.ExpectedError{
			Prop: "spec.description",
			Code: validation.ErrorCodeStringLength,
		})
	})
	t.Run("required", func(t *testing.T) {
		annotation := validAnnotation()
		annotation.Spec.Description = ""
		err := validate(annotation)
		testutils.AssertContainsErrors(t, annotation, err, 1, testutils.ExpectedError{
			Prop: "spec.description",
			Code: validation.ErrorCodeRequired,
		})
	})
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
	t.Run("fails, required", func(t *testing.T) {
		annotation := validAnnotation()
		annotation.Spec.Slo = ""
		err := validate(annotation)
		testutils.AssertContainsErrors(t, annotation, err, 1, testutils.ExpectedError{
			Prop: "spec.slo",
			Code: validation.ErrorCodeRequired,
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
	tests := map[string]Spec{
		"passes, end time after start time": {
			StartTime: time.Date(2023, 5, 1, 17, 10, 5, 0, time.UTC),
			EndTime:   time.Date(2023, 5, 14, 17, 10, 5, 0, time.UTC),
		},
		"passes, end time equals start time": {
			StartTime: time.Date(2023, 5, 2, 17, 10, 5, 0, time.UTC),
			EndTime:   time.Date(2023, 5, 2, 17, 10, 5, 0, time.UTC),
		},
	}
	for name, spec := range tests {
		t.Run(name, func(t *testing.T) {
			annotation := validAnnotation()
			annotation.Spec.StartTime = spec.StartTime
			annotation.Spec.EndTime = spec.EndTime
			err := validate(annotation)
			testutils.AssertNoError(t, annotation, err)
		})
	}

	t.Run("fails, end time is before start time", func(t *testing.T) {
		annotation := validAnnotation()
		annotation.Spec.StartTime = time.Date(2023, 5, 5, 17, 10, 5, 0, time.UTC)
		annotation.Spec.EndTime = time.Date(2023, 5, 2, 17, 10, 5, 0, time.UTC)
		err := validate(annotation)
		testutils.AssertContainsErrors(t, annotation, err, 1, testutils.ExpectedError{
			Prop: "spec",
			Code: errorCodeEndTimeNotBeforeStartTime,
		})
	})
}

func TestSpec_Category(t *testing.T) {
	t.Run("passes, no category", func(t *testing.T) {
		annotation := validAnnotation()
		annotation.Spec.Category = ""
		err := validate(annotation)
		testutils.AssertNoError(t, annotation, err)
	})
	t.Run("fails, invalid category", func(t *testing.T) {
		annotation := validAnnotation()
		annotation.Spec.Category = "Invalid"
		err := validate(annotation)
		testutils.AssertContainsErrors(t, annotation, err, 1, testutils.ExpectedError{
			Prop: "spec",
			Code: errorCodeCategoryUserDefined,
		})
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
