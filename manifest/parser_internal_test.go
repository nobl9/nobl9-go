package manifest

import (
	"embed"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed test_data/parser
var parserTestData embed.FS

func Test(t *testing.T) {
	for _, test := range []struct {
		Input              string
		ExpectedObjectsLen int
		ExpectedNames      []string
		Format             ObjectFormat
	}{
		{
			Input:              "list_of_objects.yaml",
			ExpectedObjectsLen: 2,
			ExpectedNames:      []string{"default0", "default1"},
			Format:             ObjectFormatYAML,
		},
		{
			Input:              "list_of_objects_with_whitespace.yaml",
			ExpectedObjectsLen: 2,
			ExpectedNames:      []string{"default0", "default1"},
			Format:             ObjectFormatYAML,
		},
		{
			Input:              "list_of_objects_with_comments.yaml",
			ExpectedObjectsLen: 2,
			ExpectedNames:      []string{"default0", "default1"},
			Format:             ObjectFormatYAML,
		},
		{
			Input:              "multiple_documents.yaml",
			ExpectedObjectsLen: 3,
			ExpectedNames:      []string{"default0", "default1", "default2"},
			Format:             ObjectFormatYAML,
		},
		{
			Input:              "single_document.yaml",
			ExpectedObjectsLen: 1,
			ExpectedNames:      []string{"default"},
			Format:             ObjectFormatYAML,
		},
		{
			Input:              "single_document_with_document_separators.yaml",
			ExpectedObjectsLen: 1,
			ExpectedNames:      []string{"default"},
			Format:             ObjectFormatYAML,
		},
		{
			Input:              "compacted_list_of_objects.json",
			ExpectedObjectsLen: 2,
			ExpectedNames:      []string{"default0", "default1"},
			Format:             ObjectFormatJSON,
		},
		{
			Input:              "compacted_single_object.json",
			ExpectedObjectsLen: 1,
			ExpectedNames:      []string{"default"},
			Format:             ObjectFormatJSON,
		},
		{
			Input:              "list_of_objects.json",
			ExpectedObjectsLen: 2,
			ExpectedNames:      []string{"default0", "default1"},
			Format:             ObjectFormatJSON,
		},
		{
			Input:              "list_of_objects_with_whitespace.json",
			ExpectedObjectsLen: 2,
			ExpectedNames:      []string{"default0", "default1"},
			Format:             ObjectFormatJSON,
		},
		{
			Input:              "single_object.json",
			ExpectedObjectsLen: 1,
			ExpectedNames:      []string{"default"},
			Format:             ObjectFormatJSON,
		},
	} {
		t.Run(test.Input, func(t *testing.T) {
			data := readInputFile(t, test.Input)

			isJSON := isJSONBuffer(data)
			switch test.Format {
			case ObjectFormatJSON:
				assert.True(t, isJSON, "expected the file contents to be interpreted as JSON")
			case ObjectFormatYAML:
				assert.False(t, isJSON, "expected the file contents to be interpreted as YAML")
			}
		})
	}
}

func readInputFile(t *testing.T, name string) []byte {
	t.Helper()
	data, err := parserTestData.ReadFile(filepath.Join("test_data", "parser", name))
	require.NoError(t, err)
	return data
}
