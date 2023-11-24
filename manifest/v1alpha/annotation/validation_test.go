package annotation

import (
	_ "embed"
	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/validation"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/manifest"
)

//go:embed test_data/expected_error.txt
var expectedError string

func TestValidate_AllErrors(t *testing.T) {
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
			StartTime:     "",
			EndTime:       "",
		},
		ManifestSource: "/home/me/annotation.yaml",
	})
	assert.Equal(t, strings.TrimSuffix(expectedError, "\n"), err.Error())
}

func TestSpec(t *testing.T) {
	t.Run("date strings fields fails on comparison", func(t *testing.T) {
		tests := map[string]Spec{
			"EndTime equals StartTime, fails on comparison": {
				Slo:           "some-slo",
				ObjectiveName: "obj-name",
				Description:   "description",
				StartTime:     "2023-05-01T17:10:05Z",
				EndTime:       "2023-05-01T17:10:05Z",
			},
			"EndTime required as greater fails on comparison to StartTime": {
				Slo:           "some-slo",
				ObjectiveName: "obj-name",
				Description:   "description",
				StartTime:     "2023-05-05T17:10:05Z",
				EndTime:       "2023-05-01T17:10:05Z",
			},
		}
		for name, spec := range tests {
			t.Run(name, func(t *testing.T) {
				rb := New(Metadata{Name: "my-name"}, spec)
				err := validate(rb)
				testutils.AssertContainsErrors(t, rb, err, 1, testutils.ExpectedError{
					Prop: "spec",
					Code: validation.ErrorCodeDateStringGreater,
				})
			})
		}
	})
}
