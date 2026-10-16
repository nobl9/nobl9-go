package direct

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

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

func TestStatus_CredentialsErrorSerialization(t *testing.T) {
	tests := map[string]struct {
		status       Status
		expectedJSON string
	}{
		"no credentials error is omitted": {
			status:       Status{DirectType: "Datadog"},
			expectedJSON: `{"directType":"Datadog"}`,
		},
		"credentials error is preserved": {
			status: Status{
				DirectType: "Datadog",
				CredentialsError: &v1alpha.CredentialsError{
					Since:          time.Date(2026, 10, 16, 9, 30, 0, 0, time.UTC),
					HTTPStatusCode: 401,
				},
			},
			expectedJSON: `{"directType":"Datadog",` +
				`"credentialsError":{"since":"2026-10-16T09:30:00Z","httpStatusCode":401}}`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := json.Marshal(test.status)
			require.NoError(t, err)
			assert.JSONEq(t, test.expectedJSON, string(data))

			var decoded Status
			require.NoError(t, json.Unmarshal(data, &decoded))
			assert.Equal(t, test.status, decoded)
		})
	}

	t.Run("JSON without credentials error decodes to nil", func(t *testing.T) {
		var decoded Status
		require.NoError(t, json.Unmarshal([]byte(`{"directType":"Datadog"}`), &decoded))
		assert.Equal(t, Status{DirectType: "Datadog"}, decoded)
	})

	t.Run("struct without credentials error ignores it", func(t *testing.T) {
		type oldStatus struct {
			DirectType string `json:"directType"`
		}
		var decoded oldStatus
		data := []byte(`{"directType":"Datadog",` +
			`"credentialsError":{"since":"2026-10-16T09:30:00Z","httpStatusCode":403}}`)
		require.NoError(t, json.Unmarshal(data, &decoded))
		assert.Equal(t, oldStatus{DirectType: "Datadog"}, decoded)
	})
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
