package manifest

import (
	"context"
	"path/filepath"
	"testing"

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
	objects, err := sdk.ReadObjects(context.Background(), filepath.Join(moduleRoot, "manifest/**/example*.yaml"))
	require.NoError(t, err)
	assert.Greater(t, len(objects), 0, "no object examples found")
	errs := manifest.Validate(objects)
	assert.Empty(t, errs)
}
