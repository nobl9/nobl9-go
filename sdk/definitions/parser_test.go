package definitions

import (
	"embed"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

//go:embed test_data/parser
var parserTestData embed.FS

func TestDecode(t *testing.T) {
	for _, test := range []struct {
		Input              string
		ExpectedObjectsLen int
		ExpectedNames      []string
	}{
		{
			Input:              "list_of_objects.yaml",
			ExpectedObjectsLen: 2,
			ExpectedNames:      []string{"default0", "default1"},
		},
		{
			Input:              "multiple_documents.yaml",
			ExpectedObjectsLen: 3,
			ExpectedNames:      []string{"default0", "default1", "default2"},
		},
		{
			Input:              "single_document.yaml",
			ExpectedObjectsLen: 1,
			ExpectedNames:      []string{"default"},
		},
		{
			Input:              "single_document_with_document_separators.yaml",
			ExpectedObjectsLen: 1,
			ExpectedNames:      []string{"default"},
		},
		{
			Input:              "compacted_list_of_objects.json",
			ExpectedObjectsLen: 2,
			ExpectedNames:      []string{"default0", "default1"},
		},
		{
			Input:              "compacted_single_object.json",
			ExpectedObjectsLen: 1,
			ExpectedNames:      []string{"default"},
		},
		{
			Input:              "list_of_objects.json",
			ExpectedObjectsLen: 2,
			ExpectedNames:      []string{"default0", "default1"},
		},
		{
			Input:              "single_object.json",
			ExpectedObjectsLen: 1,
			ExpectedNames:      []string{"default"},
		},
	} {
		t.Run(test.Input, func(t *testing.T) {
			objects, err := Decode(readInputFile(t, test.Input))
			require.NoError(t, err)
			assert.Len(t, objects, test.ExpectedObjectsLen)
			assert.IsType(t, v1alpha.Project{}, objects[0])

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

}

func readInputFile(t *testing.T, name string) []byte {
	t.Helper()
	data, err := parserTestData.ReadFile(filepath.Join("test_data", "parser", name))
	require.NoError(t, err)
	return data
}
