package validation

import (
	"embed"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed test_data
var errorsTestData embed.FS

func TestMultiRuleError(t *testing.T) {
	err := multiRuleError{
		errors.New("this is just a test!"),
		errors.New("another error..."),
		errors.New("that is just fatal."),
	}
	assert.EqualError(t, err, expectedErrorOutput(t, "multi_error.txt"))
}

func TestFieldError(t *testing.T) {
	for typ, value := range map[string]interface{}{
		"string": "default",
		"slice":  []string{"this", "that"},
		"map":    map[string]string{"this": "that"},
		"struct": struct {
			This string `json:"this"`
			That string `json:"THAT"`
		}{This: "this", That: "that"},
	} {
		t.Run(typ, func(t *testing.T) {
			err := FieldError{
				FieldPath:  "metadata.name",
				FieldValue: value,
				Errors: []error{
					multiRuleError{
						errors.New("what a shame this happened"),
						errors.New("this is outrageous..."),
					},
					errors.New("here's another error"),
				},
			}
			assert.EqualError(t, err, expectedErrorOutput(t, fmt.Sprintf("field_error_%s.txt", typ)))
		})
	}
}

func TestObjectError(t *testing.T) {
	errs := []error{
		FieldError{
			FieldPath:  "metadata.name",
			FieldValue: "default",
			Errors:     []error{errors.New("here's an error")},
		},
		FieldError{
			FieldPath:  "spec.description",
			FieldValue: "some long description",
			Errors:     []error{errors.New("here's another error")},
		},
	}

	t.Run("non project scoped object", func(t *testing.T) {
		err := ObjectError{
			Object: ObjectMetadata{
				Kind:   "Project",
				Name:   "default",
				Source: "/home/me/project.yaml",
			},
			Errors: errs,
		}
		assert.EqualError(t, err, expectedErrorOutput(t, "object_error.txt"))
	})

	t.Run("project scoped object", func(t *testing.T) {
		err := ObjectError{
			Object: ObjectMetadata{
				IsProjectScoped: true,
				Kind:            "Service",
				Name:            "my-service",
				Project:         "default",
				Source:          "/home/me/service.yaml",
			},
			Errors: errs,
		}
		assert.EqualError(t, err, expectedErrorOutput(t, "object_error_project_scoped.txt"))
	})
}

func expectedErrorOutput(t *testing.T, name string) string {
	t.Helper()
	data, err := errorsTestData.ReadFile(filepath.Join("test_data", name))
	require.NoError(t, err)
	return string(data)
}
