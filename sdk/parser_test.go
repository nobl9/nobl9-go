package sdk

import (
	"embed"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
)

//go:embed test_data/parser
var parserTestData embed.FS

func TestDecode(t *testing.T) {
	for _, test := range []struct {
		Input              string
		ExpectedObjectsLen int
		ExpectedNames      []string
		Format             manifest.ObjectFormat
	}{
		{
			Input:              "list_of_objects.yaml",
			ExpectedObjectsLen: 2,
			ExpectedNames:      []string{"default0", "default1"},
			Format:             manifest.ObjectFormatYAML,
		},
		{
			Input:              "list_of_objects_with_whitespace.yaml",
			ExpectedObjectsLen: 2,
			ExpectedNames:      []string{"default0", "default1"},
			Format:             manifest.ObjectFormatYAML,
		},
		{
			Input:              "list_of_objects_with_comments.yaml",
			ExpectedObjectsLen: 2,
			ExpectedNames:      []string{"default0", "default1"},
			Format:             manifest.ObjectFormatYAML,
		},
		{
			Input:              "multiple_documents.yaml",
			ExpectedObjectsLen: 3,
			ExpectedNames:      []string{"default0", "default1", "default2"},
			Format:             manifest.ObjectFormatYAML,
		},
		{
			Input:              "single_document.yaml",
			ExpectedObjectsLen: 1,
			ExpectedNames:      []string{"default"},
			Format:             manifest.ObjectFormatYAML,
		},
		{
			Input:              "single_document_with_document_separators.yaml",
			ExpectedObjectsLen: 1,
			ExpectedNames:      []string{"default"},
			Format:             manifest.ObjectFormatYAML,
		},
		{
			Input:              "compacted_list_of_objects.json",
			ExpectedObjectsLen: 2,
			ExpectedNames:      []string{"default0", "default1"},
			Format:             manifest.ObjectFormatJSON,
		},
		{
			Input:              "compacted_single_object.json",
			ExpectedObjectsLen: 1,
			ExpectedNames:      []string{"default"},
			Format:             manifest.ObjectFormatJSON,
		},
		{
			Input:              "list_of_objects.json",
			ExpectedObjectsLen: 2,
			ExpectedNames:      []string{"default0", "default1"},
			Format:             manifest.ObjectFormatJSON,
		},
		{
			Input:              "list_of_objects_with_whitespace.json",
			ExpectedObjectsLen: 2,
			ExpectedNames:      []string{"default0", "default1"},
			Format:             manifest.ObjectFormatJSON,
		},
		{
			Input:              "single_object.json",
			ExpectedObjectsLen: 1,
			ExpectedNames:      []string{"default"},
			Format:             manifest.ObjectFormatJSON,
		},
	} {
		t.Run(test.Input, func(t *testing.T) {
			data := readInputFile(t, test.Input)

			isJSON := isJSONBuffer(data)
			switch test.Format {
			case manifest.ObjectFormatJSON:
				assert.True(t, isJSON, "expected the file contents to be interpreted as JSON")
			case manifest.ObjectFormatYAML:
				assert.False(t, isJSON, "expected the file contents to be interpreted as YAML")
			}

			objects, err := DecodeObjects(data)
			require.NoError(t, err)
			assert.Len(t, objects, test.ExpectedObjectsLen)
			assert.IsType(t, project.Project{}, objects[0])

			objectNames := make([]string, 0, len(objects))
			for _, object := range objects {
				objectNames = append(objectNames, object.GetName())
			}
			for _, name := range test.ExpectedNames {
				assert.Contains(t, objectNames, name)
			}
		})
	}
}

func TestDecodeSingle(t *testing.T) {
	t.Run("golden path", func(t *testing.T) {
		proj, err := DecodeObject[project.Project](readInputFile(t, "single_project.yaml"))
		require.NoError(t, err)
		assert.NotZero(t, proj)
		assert.Equal(t, "default", proj.GetName())
	})

	t.Run("multiple objects, return error", func(t *testing.T) {
		_, err := DecodeObject[project.Project](readInputFile(t, "two_projects.yaml"))
		require.Error(t, err)
		assert.EqualError(t, err, "unexpected number of objects: 2, expected exactly one")
	})

	t.Run("invalid type, return error", func(t *testing.T) {
		_, err := DecodeObject[v1alphaService.Service](readInputFile(t, "single_project.yaml"))
		require.Error(t, err)
		assert.EqualError(t, err, "object of type project.Project is not of type service.Service")
	})
}

func readInputFile(t *testing.T, name string) []byte {
	t.Helper()
	data, err := parserTestData.ReadFile(filepath.Join("test_data", "parser", name))
	require.NoError(t, err)
	return data
}
