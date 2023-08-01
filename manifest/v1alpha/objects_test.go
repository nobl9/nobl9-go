package v1alpha_test

import (
	"embed"
	"fmt"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/sdk/definitions"
)

const testDataDir = "test_data"

//go:embed test_data
var testData embed.FS

//go:embed test_data/expected_error_conflicting_slo.txt
var expectedError string

func TestAPIObjects_Validate(t *testing.T) {
	var objects []manifest.Object
	for _, kind := range manifest.ApplicableKinds() {
		require.Contains(t,
			expectedError,
			kind.String(),
			"each applicable Kind must have a designated test file and appear in the expected error")

		data, err := testData.ReadFile(path.Join(testDataDir,
			fmt.Sprintf("conflicting_%s.yaml", kind.ToLower())))
		require.NoError(t, err)
		readObjects, err := definitions.Decode(data)
		require.NoError(t, err)
		objects = append(objects, readObjects...)
	}

	err := v1alpha.CheckObjectsUniqueness(objects)
	require.Error(t, err)
	// Trim any trailing newlines from the file and replace the other newlines with '; '
	// just to make the test file a bit easier to read and work with.
	expected := strings.Replace(strings.TrimSpace(expectedError), "\n", "; ", len(manifest.KindValues()))
	assert.EqualError(t, err, expected)
}
