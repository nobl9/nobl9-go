package v1alpha

import (
	"embed"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/nobl9/nobl9-go/manifest"
)

const testDataDir = "test_data"

//go:embed test_data
var testData embed.FS

//go:embed test_data/expected_error_conflicting_slo.txt
var expectedError string

func TestAPIObjects_Validate(t *testing.T) {
	objects := APIObjects{}
	for _, kind := range manifest.ApplicableKinds() {
		require.Contains(t,
			expectedError,
			kind.String(),
			"each applicable Kind must have a designated test file and appear in the expected error")

		data, err := testData.ReadFile(path.Join(testDataDir,
			fmt.Sprintf("conflicting_%s.yaml", kind.ToLower())))
		require.NoError(t, err)

		var decodedYAML []map[string]interface{}
		err = yaml.Unmarshal(data, &decodedYAML)
		require.NoError(t, err)

		rawJSON, err := json.Marshal(decodedYAML)
		require.NoError(t, err)

		var genericObjects []manifest.ObjectGeneric
		err = json.Unmarshal(rawJSON, &genericObjects)
		require.NoError(t, err)
		require.Greater(t, len(genericObjects), 0)

		for _, object := range genericObjects {
			// So that we can skip the Agent's constraints which allows only one to be applied (at the time being).
			if object.Kind == manifest.KindAgent {
				var agent Agent
				agent, err = genericToAgent(object, NewValidator(), false)
				require.NoError(t, err)
				objects.Agents = append(objects.Agents, agent)
				continue
			}
			err = Parse(object, &objects, false)
			require.NoError(t, err)
		}
	}

	err := objects.Validate()
	require.Error(t, err)
	// Trim any trailing newlines from the file and replace the other newlines with '; '
	// just to make the test file a bit easier to read and work with.
	expected := strings.Replace(strings.TrimSpace(expectedError), "\n", "; ", len(manifest.KindValues()))
	assert.EqualError(t, err, expected)
}
