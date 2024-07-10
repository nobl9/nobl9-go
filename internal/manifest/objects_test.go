package manifest

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/internal/pathutils"
	"github.com/nobl9/nobl9-go/manifest"
	v1alphaParser "github.com/nobl9/nobl9-go/manifest/v1alpha/parser"
	"github.com/nobl9/nobl9-go/sdk"
)

func TestMain(m *testing.M) {
	v1alphaParser.UseStrictDecodingMode = true
	m.Run()
}

func TestObjectExamples(t *testing.T) {
	moduleRoot := pathutils.FindModuleRoot()
	objects, err := sdk.ReadObjects(context.Background(),
		filepath.Join(moduleRoot, "manifest/**/example*.yaml"),
		filepath.Join(moduleRoot, "manifest/**/examples/*.yaml"),
	)
	require.NoError(t, err)
	assert.NotEmpty(t, objects, "no object examples found")
	for i := range objects {
		err = objects[i].Validate()
		require.NoError(t, err)
		// Make sure YAML and JSON are interoperable.
		yamlData, err := yaml.Marshal(objects[i])
		require.NoError(t, err)
		jsonData, err := yaml.YAMLToJSON(yamlData)
		assert.NoError(t, err)
		object, err := sdk.DecodeObject[manifest.Object](jsonData)
		assert.NoError(t, err)
		err = object.Validate()
		require.NoError(t, err)
	}
}
