package sdk

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	v1alphaAnnotation "github.com/nobl9/nobl9-go/manifest/v1alpha/annotation"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
)

//go:embed test_data/decode
var decodeTestData embed.FS

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
		{
			Input:              "multiline_double_quoted_description.yaml",
			ExpectedObjectsLen: 1,
			ExpectedNames:      []string{"test"},
			Format:             manifest.ObjectFormatYAML,
		},
		{
			Input:              "multiline_double_quoted_description_square_bracket_array.yaml",
			ExpectedObjectsLen: 1,
			ExpectedNames:      []string{"test"},
			Format:             manifest.ObjectFormatYAML,
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
	t.Run("scanner token size overflow", func(t *testing.T) {
		// Generate objects.
		objectsNum := 800
		expectedObjects := make([]manifest.Object, 0, objectsNum)
		for i := 0; i < objectsNum; i++ {
			expectedObjects = append(expectedObjects, project.New(
				project.Metadata{
					Name: fmt.Sprintf("%d", i),
				},
				project.Spec{},
			))
		}

		objectsData, err := yaml.Marshal(expectedObjects)
		// Ensure we're actually generating enough data to overflow the buffer default limit,
		// as defined in bufio package.
		require.Greater(t, len(objectsData), bufio.MaxScanTokenSize)
		require.NoError(t, err)

		// Write objects to file.
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "test.yaml")
		err = os.WriteFile(filename, objectsData, 0o600)
		require.NoError(t, err)

		// Read objects from file.
		data, err := os.ReadFile(filename)
		require.NoError(t, err)

		objects, err := DecodeObjects(data)
		require.NoError(t, err)
		assert.Len(t, expectedObjects, objectsNum)
		assert.IsType(t, project.Project{}, expectedObjects[0])

		objectNames := make([]string, 0, len(expectedObjects))
		for _, object := range expectedObjects {
			objectNames = append(objectNames, object.GetName())
		}
		for _, object := range objects {
			assert.Contains(t, objectNames, object.GetName())
		}
	})
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

func TestDecodeSingle_annotationRegionFailover(t *testing.T) {
	for _, category := range []string{"RegionFailover", "RegionRecovery"} {
		t.Run(category+" annotation keeps spec.regionFailover", func(t *testing.T) {
			// arrange
			input := fmt.Sprintf(`{
  "apiVersion": "n9/v1alpha",
  "kind": "Annotation",
  "metadata": {"name": "failover-123", "project": "default"},
  "spec": {
    "slo": "my-slo",
    "description": "Region failover of Direct cloudwatch-direct",
    "startTime": "2026-10-20T10:00:00Z",
    "category": %q,
    "regionFailover": {
      "dataSourceKind": "Direct",
      "dataSourceName": "cloudwatch-direct",
      "dataSourceProject": "infra",
      "primaryRegion": "us-east-1",
      "failoverRegion": "us-west-2",
      "eventId": "event-1"
    }
  },
  "status": {"updatedAt": "2026-10-20T10:00:05Z", "isSystem": true}
}`, category)

			// act
			annotation, err := DecodeObject[v1alphaAnnotation.Annotation]([]byte(input))
			require.NoError(t, err)
			var buf bytes.Buffer
			require.NoError(t, EncodeObject(annotation, &buf, manifest.ObjectFormatJSON))

			// assert
			assert.Equal(t, v1alphaAnnotation.Category(category), annotation.Spec.Category)
			assert.JSONEq(t, input, buf.String())
			assert.Empty(t, v1alphaAnnotation.GetValidatorWithoutCategoryRules().Validate(annotation))
		})
	}
}

func readInputFile(t *testing.T, name string) []byte {
	t.Helper()
	data, err := decodeTestData.ReadFile(path.Join("test_data", "decode", name))
	require.NoError(t, err)
	return data
}
