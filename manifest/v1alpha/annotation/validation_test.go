package annotation

import (
	_ "embed"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest"
)

//go:embed test_data/expected_error.txt
var expectedError string

func TestValidate_AllErrors(t *testing.T) {
	startTime, _ := time.Parse(time.RFC3339, "2023-05-01T17:10:05Z")
	endTime, _ := time.Parse(time.RFC3339, "2023-05-01T17:10:05Z")

	err := validate(Annotation{
		Kind: manifest.KindAnnotation,
		Metadata: Metadata{
			Name:    strings.Repeat("MY ANNOTATION", 20),
			Project: strings.Repeat("MY ANNOTATION", 20),
		},
		Spec: Spec{
			Slo:           strings.Repeat("l", 2000),
			ObjectiveName: strings.Repeat("l", 2000),
			Description:   strings.Repeat("l", 2000),
			StartTime:     startTime,
			EndTime:       endTime,
		},
		ManifestSource: "/home/me/annotation.yaml",
	})
	assert.Equal(t, strings.TrimSuffix(expectedError, "\n"), err.Error())
}

func TestSpec(t *testing.T) {
	t.Run("end time is after start time", func(t *testing.T) {
		startTime, _ := time.Parse(time.RFC3339, "2023-05-01T17:10:05Z")
		endTime, _ := time.Parse(time.RFC3339, "2023-05-01T13:10:05Z")

		annotation := New(
			Metadata{Name: "my-name"},
			Spec{
				Slo:           "my-slo",
				ObjectiveName: "my-obj",
				Description:   "my-annotation description",
				StartTime:     startTime,
				EndTime:       endTime,
			},
		)
		err := validate(annotation)
		testutils.AssertContainsErrors(t, annotation, err, 1, testutils.ExpectedError{
			Prop: "spec",
			Code: errorCodeEndTimeAfterStartTime,
		})
	})
}
