package parser

import (
	"embed"
	"encoding/json"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/internal/stringutils"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/alertmethod"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/alertpolicy"
)

//go:embed test_data
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
	UseStrictDecodingMode = true
	defer func() { UseStrictDecodingMode = false }()

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

func TestParseObject_NoErrorOnNonExistingKeys(t *testing.T) {
	filename := "project_with_non_existing_keys"
	UseStrictDecodingMode = false

	t.Run("json", func(t *testing.T) {
		jsonData, format := readParserTestFile(t, filename+".json")
		_, err := ParseObject(jsonData, manifest.KindProject, format)
		assert.NoError(t, err)
	})

	t.Run("yaml", func(t *testing.T) {
		yamlData, format := readParserTestFile(t, filename+".yaml")
		_, err := ParseObject(yamlData, manifest.KindProject, format)
		assert.NoError(t, err)
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
	assert.Equal(t, v1alpha.GenericObject{
		"apiVersion": "n9/v1alpha",
		"kind":       "Project",
		"metadata": map[string]interface{}{
			"name": "default",
			"fake": "fake",
		},
	}, jsonObject)
}

func TestParseAlertPolicy(t *testing.T) {
	testCases := map[string]struct {
		alertPolicy alertpolicy.AlertPolicy
		expected    string
	}{
		"passes, simple AlertMethodRef declared": {
			alertPolicy: validAlertPolicy(),
			expected:    "expected_alert_policy_with_alert_method_ref.json",
		},
		"passes, embed legacy alert method": {
			alertPolicy: withLegacyAlertMethodEmbedded(),
			expected:    "expected_alert_policy_with_legacy_alert_method_details.json",
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			marshalledJson, err := json.MarshalIndent(testCase.alertPolicy, "", "  ")
			jsonData, _ := readParserTestFile(t, testCase.expected)
			require.NoError(t, err)

			assert.Equal(t, string(marshalledJson), string(jsonData))
		})
	}
}

func Test_ParseObject_EnsureAllKindsAreParsed(t *testing.T) {
	for _, kind := range manifest.KindValues() {
		object := map[string]interface{}{
			"apiVersion": "n9/v1alpha",
			"kind":       kind.String(),
		}
		data, err := json.Marshal(object)
		require.NoError(t, err)
		_, err = ParseObject(data, kind, manifest.ObjectFormatJSON)
		assert.NotErrorIsf(t, err, manifest.ErrInvalidKind, "failed for %s", kind)
	}
}

// Reference: https://github.com/goccy/go-yaml/issues/455
func Test_ParseObject_DoubleQuotedJSONHandling(t *testing.T) {
	data, format := readParserTestFile(t, "cloudwatch_slo.yaml")
	_, err := ParseObject(data, manifest.KindSLO, format)
	require.NoError(t, err)
}

func readParserTestFile(t *testing.T, filename string) ([]byte, manifest.ObjectFormat) {
	t.Helper()
	data, err := parserTestData.ReadFile(path.Join("test_data", filename))
	require.NoError(t, err)
	format, err := manifest.ParseObjectFormat(filepath.Ext(filename)[1:])
	require.NoError(t, err)
	return []byte(stringutils.RemoveCR(string(data))), format
}

func validAlertPolicy() alertpolicy.AlertPolicy {
	return alertpolicy.AlertPolicy{
		APIVersion: manifest.VersionV1alpha,
		Kind:       manifest.KindAlertPolicy,
		Metadata: alertpolicy.Metadata{
			Name:        "this",
			DisplayName: "This",
			Project:     "default",
		},
		Spec: alertpolicy.Spec{
			Description:      "test",
			Severity:         "low",
			CoolDownDuration: "5m",
			Conditions:       nil,
			AlertMethods: []alertpolicy.AlertMethodRef{{
				Metadata: alertpolicy.AlertMethodRefMetadata{
					Name:    "this",
					Project: "default",
				},
			}},
		},
	}
}

func withLegacyAlertMethodEmbedded() alertpolicy.AlertPolicy {
	alertPolicy := validAlertPolicy()
	alertPolicy.Spec.AlertMethods[0].EmbedAlertMethodRef(alertmethod.AlertMethod{
		APIVersion: manifest.VersionV1alpha,
		Kind:       manifest.KindAlertMethod,
		Metadata: alertmethod.Metadata{
			Name:    "this",
			Project: "default",
		},
	})

	return alertPolicy
}
