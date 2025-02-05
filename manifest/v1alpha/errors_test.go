package v1alpha

import (
	"embed"
	"encoding/json"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/govy/pkg/govy"

	"github.com/nobl9/nobl9-go/internal/stringutils"
	"github.com/nobl9/nobl9-go/manifest"
)

//go:embed test_data/errors
var errorsTestData embed.FS

func TestObjectError(t *testing.T) {
	errs := govy.PropertyErrors{
		&govy.PropertyError{
			PropertyName:  "metadata.name",
			PropertyValue: "default",
			Errors:        []*govy.RuleError{{Message: "here's an error"}},
		},
		&govy.PropertyError{
			PropertyName:  "spec.description",
			PropertyValue: "some long description",
			Errors:        []*govy.RuleError{{Message: "here's another error"}},
		},
	}

	t.Run("non project scoped object", func(t *testing.T) {
		err := &ObjectError{
			Object: ObjectMetadata{
				Kind:   manifest.KindProject,
				Name:   "default",
				Source: "/home/me/project.yaml",
			},
			Errors: errs,
		}
		assert.EqualError(t, err, expectedErrorOutput(t, "object_error.txt"))
	})

	t.Run("project scoped object", func(t *testing.T) {
		err := &ObjectError{
			Object: ObjectMetadata{
				IsProjectScoped: true,
				Kind:            manifest.KindService,
				Name:            "my-service",
				Project:         "default",
				Source:          "/home/me/service.yaml",
			},
			Errors: errs,
		}
		assert.EqualError(t, err, expectedErrorOutput(t, "object_error_project_scoped.txt"))
	})
}

func TestObjectError_UnmarshalJSON(t *testing.T) {
	expected := ObjectError{
		Object: ObjectMetadata{
			Kind:            manifest.KindService,
			Name:            "test-service",
			Source:          "/home/me/service.yaml",
			IsProjectScoped: true,
			Project:         "default",
		},
		Errors: govy.PropertyErrors{
			{
				PropertyName:  "metadata.project",
				PropertyValue: "default",
				Errors:        []*govy.RuleError{{Message: "nested"}},
			},
			{
				PropertyName:  "metadata.name",
				PropertyValue: "my-project",
			},
		},
	}
	data, err := json.MarshalIndent(expected, "", " ")
	require.NoError(t, err)

	var actual ObjectError
	err = json.Unmarshal(data, &actual)
	require.NoError(t, err)

	assert.Equal(t, expected, actual)
}

func expectedErrorOutput(t *testing.T, name string) string {
	t.Helper()
	data, err := errorsTestData.ReadFile(path.Join("test_data", "errors", name))
	require.NoError(t, err)
	return stringutils.RemoveCR(string(data))
}
