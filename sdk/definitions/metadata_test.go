package definitions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/sdk"
)

func TestMetadataAnnotations_AnnotateObject(t *testing.T) {
	t.Run("fill missing fields", func(t *testing.T) {
		result, err := MetadataAnnotations{
			Organization:   "my-org",
			Project:        "default",
			ManifestSource: "my-source",
		}.AnnotateObject(sdk.AnyJSONObj{"metadata": map[string]interface{}{}})
		require.NoError(t, err)
		expected := sdk.AnyJSONObj{"metadata": map[string]interface{}{
			"organization": "my-org",
			"project":      "default",
			"manifestSrc":  "my-source",
		}}
		assert.Equal(t, expected, result)
	})

	t.Run("don't fill fields if annotations are not provided", func(t *testing.T) {
		result, err := MetadataAnnotations{
			Organization:   "",
			Project:        "",
			ManifestSource: "",
		}.AnnotateObject(sdk.AnyJSONObj{"metadata": map[string]interface{}{}})
		require.NoError(t, err)
		expected := sdk.AnyJSONObj{"metadata": map[string]interface{}{}}
		assert.Equal(t, expected, result)
	})

	t.Run("don't fill fields if they are set already", func(t *testing.T) {
		result, err := MetadataAnnotations{
			Organization:   "different-org",
			Project:        "non-default",
			ManifestSource: "other-source",
		}.AnnotateObject(sdk.AnyJSONObj{"metadata": map[string]interface{}{
			"organization": "my-org",
			"project":      "default",
			"manifestSrc":  "my-source",
		}})
		require.NoError(t, err)
		expected := sdk.AnyJSONObj{"metadata": map[string]interface{}{
			"organization": "my-org",
			"project":      "default",
			"manifestSrc":  "my-source",
		}}
		assert.Equal(t, expected, result)
	})
}
