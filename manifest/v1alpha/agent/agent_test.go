package agent

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestAgent_Spec_GetType(t *testing.T) {
	for _, agent := range v1alpha.DataSourceTypeValues() {
		t.Run(agent.String(), func(t *testing.T) {
			spec := Spec{}
			agentTypeStr := agent.String()
			if agent == v1alpha.GCM {
				agentTypeStr = "GCM"
			}
			setZeroValue(t, &spec, agentTypeStr)
			typ, err := spec.GetType()
			require.NoError(t, err)
			assert.Equal(t, typ.String(), agent.String())
		})
	}
}

func TestStatus_InvalidCredentialsSerialization(t *testing.T) {
	tests := map[string]struct {
		status       Status
		expectedJSON string
	}{
		"no invalid credentials is omitted": {
			status:       Status{AgentType: "Datadog"},
			expectedJSON: `{"agentType":"Datadog","environment":{}}`,
		},
		"degraded invalid credentials is preserved": {
			status: Status{
				AgentType: "Datadog",
				InvalidCredentials: &v1alpha.InvalidCredentials{
					State:          v1alpha.InvalidCredentialsStateDegraded,
					Since:          time.Date(2026, 10, 16, 9, 30, 0, 0, time.UTC),
					HTTPStatusCode: 403,
					AffectedSLOs:   1,
					TotalSLOs:      2,
				},
			},
			expectedJSON: `{"agentType":"Datadog","environment":{},"invalidCredentials":{"state":"degraded",` +
				`"since":"2026-10-16T09:30:00Z","httpStatusCode":403,"affectedSLOs":1,"totalSLOs":2}}`,
		},
		"none state has no since and no HTTP status code": {
			status: Status{
				AgentType: "Datadog",
				InvalidCredentials: &v1alpha.InvalidCredentials{
					State:     v1alpha.InvalidCredentialsStateNone,
					TotalSLOs: 2,
				},
			},
			expectedJSON: `{"agentType":"Datadog","environment":{},` +
				`"invalidCredentials":{"state":"none","affectedSLOs":0,"totalSLOs":2}}`,
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

	t.Run("JSON without invalid credentials decodes to nil", func(t *testing.T) {
		var decoded Status
		require.NoError(t, json.Unmarshal([]byte(`{"agentType":"Datadog","agentVersion":"1.0.0"}`), &decoded))
		assert.Equal(t, Status{AgentType: "Datadog", AgentVersion: "1.0.0"}, decoded)
	})

	t.Run("invalid credentials without state and counts decodes to zero values", func(t *testing.T) {
		var decoded Status
		data := []byte(`{"agentType":"Datadog","invalidCredentials":{"since":"2026-10-16T09:30:00Z","httpStatusCode":403}}`)
		require.NoError(t, json.Unmarshal(data, &decoded))
		assert.Equal(t, Status{
			AgentType: "Datadog",
			InvalidCredentials: &v1alpha.InvalidCredentials{
				Since:          time.Date(2026, 10, 16, 9, 30, 0, 0, time.UTC),
				HTTPStatusCode: 403,
			},
		}, decoded)
	})

	t.Run("struct without invalid credentials ignores it", func(t *testing.T) {
		type oldStatus struct {
			AgentType    string `json:"agentType"`
			AgentVersion string `json:"agentVersion,omitempty"`
		}
		var decoded oldStatus
		data := []byte(`{"agentType":"Datadog","agentVersion":"1.0.0",` +
			`"invalidCredentials":{"since":"2026-10-16T09:30:00Z","httpStatusCode":401}}`)
		require.NoError(t, json.Unmarshal(data, &decoded))
		assert.Equal(t, oldStatus{AgentType: "Datadog", AgentVersion: "1.0.0"}, decoded)
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
