package direct

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestDirect_Spec_GetType(t *testing.T) {
	for direct := range validDirectTypes {
		t.Run(direct.String(), func(t *testing.T) {
			spec := Spec{}
			directTypeStr := direct.String()
			if direct == v1alpha.GCM {
				directTypeStr = "GCM"
			}
			setZeroValue(t, &spec, directTypeStr)
			typ, err := spec.GetType()
			require.NoError(t, err)
			assert.Equal(t, typ, direct)
		})
	}
}

func TestDynatraceConfig_JSONFields(t *testing.T) {
	data, err := json.Marshal(DynatraceConfig{
		URL:               "https://example.live.dynatrace.com",
		DynatraceToken:    "token",
		OAuthClientID:     "client-id",
		OAuthClientSecret: "client-secret",
		AccountURN:        "urn:dtaccount:example",
		OAuthScopes:       "storage:buckets:read storage:logs:read",
	})
	require.NoError(t, err)

	assert.Contains(t, string(data), `"accountUrn":"urn:dtaccount:example"`)
	assert.NotContains(t, string(data), "accountURN")
	assert.NotContains(t, string(data), "dqlUrl")
}

// setZeroValue sets a zero value of a pointer field in a struct using reflection.
func setZeroValue(t *testing.T, obj interface{}, fieldName string) {
	t.Helper()
	objValue := reflect.ValueOf(obj).Elem()
	fieldValue := objValue.FieldByName(fieldName)
	if !fieldValue.IsValid() || !fieldValue.CanSet() {
		t.Fatalf("cannot set value for field: %s", fieldName)
	}
	fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
}
