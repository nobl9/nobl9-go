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

func readParserTestFile(t *testing.T, filename string) ([]byte, manifest.ObjectFormat) {
	t.Helper()
	data, err := parserTestData.ReadFile(filepath.Join("test_data", "parser", filename))
	require.NoError(t, err)
	format, err := manifest.ParseObjectFormat(filepath.Ext(filename)[1:])
	require.NoError(t, err)
	return data, format
}
