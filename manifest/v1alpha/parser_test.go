package v1alpha

import (
	"embed"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
)

//go:embed test_data/parser
var parserTestData embed.FS

func TestParseObject(t *testing.T) {
	for name, kind := range map[string]manifest.Kind{
		"cloudwatch_agent": manifest.KindAgent,
		"redshift_agent":   manifest.KindAgent,
	} {
		t.Run(strings.ReplaceAll(name, "_", " "), func(t *testing.T) {
			jsonData, format := readParserTestFile(t, name+".json")
			jsonObject, err := ParseObject(jsonData, kind, format)
			require.NoError(t, err)

			yamlData, format := readParserTestFile(t, name+".yaml")
			yamlObject, err := ParseObject(yamlData, kind, format)
			require.NoError(t, err)

			assert.Equal(t, jsonObject, yamlObject)
		})
	}
}

func TestParseObject_ErrorOnNonExistingKeys(t *testing.T) {
	filename := "project_with_non_existing_keys"

	t.Run("json", func(t *testing.T) {
		jsonData, format := readParserTestFile(t, filename+".json")
		_, err := ParseObject(jsonData, manifest.KindProject, format)
		require.Error(t, err)
		assert.ErrorContains(t, err, "horsepower")
	})

	t.Run("yaml", func(t *testing.T) {
		yamlData, format := readParserTestFile(t, filename+".yaml")
		_, err := ParseObject(yamlData, manifest.KindProject, format)
		require.Error(t, err)
		assert.ErrorContains(t, err, "horsepower")
	})
}

func TestParseObjectUsingGenericObject(t *testing.T) {
	UseGenericObjects = true
	defer func() { UseGenericObjects = false }()

	jsonData, format := readParserTestFile(t, "generic_project.json")
	jsonObject, err := ParseObject(jsonData, manifest.KindProject, format)
	require.NoError(t, err)

	yamlData, format := readParserTestFile(t, "generic_project.json")
	yamlObject, err := ParseObject(yamlData, manifest.KindProject, format)
	require.NoError(t, err)

	assert.Equal(t, jsonObject, yamlObject)
	assert.Equal(t, GenericObject{
		"apiVersion": "n9/v1alpha",
		"kind":       "Project",
		"metadata": map[string]interface{}{
			"name": "default",
			"fake": "fake",
		},
	}, jsonObject)
}

func readParserTestFile(t *testing.T, filename string) ([]byte, manifest.ObjectFormat) {
	t.Helper()
	data, err := parserTestData.ReadFile(filepath.Join("test_data", "parser", filename))
	require.NoError(t, err)
	format, err := manifest.ParseObjectFormat(filepath.Ext(filename)[1:])
	require.NoError(t, err)
	return data, format
}
