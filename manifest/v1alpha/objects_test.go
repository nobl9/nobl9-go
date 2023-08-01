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

		var genericObjects []ObjectGeneric
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

func TestSetAlertPolicyDefaults(t *testing.T) {
	for _, testCase := range []struct {
		desc string
		in   AlertPolicy
		out  AlertPolicy
	}{
		{
			desc: "when alertingWindow is defined, lastsFor default value should not be set",
			in: AlertPolicy{
				Spec: AlertPolicySpec{
					Conditions: []AlertCondition{
						{
							AlertingWindow: "30m",
						},
					},
				},
			},
			out: AlertPolicy{
				Spec: AlertPolicySpec{
					Conditions: []AlertCondition{
						{
							AlertingWindow: "30m",
						},
					},
				},
			},
		},
		{
			desc: "when alertingWindow is not defined and lastsFor is empty zero value should be set",
			in: AlertPolicy{
				Spec: AlertPolicySpec{
					Conditions: []AlertCondition{
						{},
					},
				},
			},
			out: AlertPolicy{
				Spec: AlertPolicySpec{
					Conditions: []AlertCondition{
						{
							LastsForDuration: "0m",
						},
					},
				},
			},
		},
		{
			desc: "when alertingWindow is not defined and lastsFor is not empty do not change lastsFor",
			in: AlertPolicy{
				Spec: AlertPolicySpec{
					Conditions: []AlertCondition{
						{
							LastsForDuration: "1h",
						},
					},
				},
			},
			out: AlertPolicy{
				Spec: AlertPolicySpec{
					Conditions: []AlertCondition{
						{
							LastsForDuration: "1h",
						},
					},
				},
			},
		},
	} {
		t.Run(testCase.desc, func(t *testing.T) {
			setAlertPolicyDefaults(&testCase.in)
			assert.Equal(t, testCase.out, testCase.in)
		})
	}
}

func setAlertPolicyDefaults(policy *AlertPolicy) {
	for i, condition := range policy.Spec.Conditions {
		if condition.AlertingWindow == "" && condition.LastsForDuration == "" {
			policy.Spec.Conditions[i].LastsForDuration = DefaultAlertPolicyLastsForDuration
		}
	}
}
