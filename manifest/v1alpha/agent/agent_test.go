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
					HTTPStatusCode: 401,
					AffectedSLOs:   1,
					TotalSLOs:      2,
				},
			},
			expectedJSON: `{"agentType":"Datadog","environment":{},"invalidCredentials":{"state":"degraded",` +
				`"since":"2026-10-16T09:30:00Z","httpStatusCode":401,"affectedSLOs":1,"totalSLOs":2}}`,
		},
		"none state has only the state": {
			status: Status{
				AgentType: "Datadog",
				InvalidCredentials: &v1alpha.InvalidCredentials{
					State: v1alpha.InvalidCredentialsStateNone,
				},
			},
			expectedJSON: `{"agentType":"Datadog","environment":{},` +
				`"invalidCredentials":{"state":"none"}}`,
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
		data := []byte(`{"agentType":"Datadog","invalidCredentials":{"since":"2026-10-16T09:30:00Z","httpStatusCode":401}}`)
		require.NoError(t, json.Unmarshal(data, &decoded))
		assert.Equal(t, Status{
			AgentType: "Datadog",
			InvalidCredentials: &v1alpha.InvalidCredentials{
				Since:          time.Date(2026, 10, 16, 9, 30, 0, 0, time.UTC),
				HTTPStatusCode: 401,
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

func TestStatus_AccessDeniedSerialization(t *testing.T) {
	tests := map[string]struct {
		status       Status
		expectedJSON string
	}{
		"no access denied is omitted": {
			status:       Status{AgentType: "Datadog"},
			expectedJSON: `{"agentType":"Datadog","environment":{}}`,
		},
		"all access denied is preserved": {
			status: Status{
				AgentType: "Datadog",
				AccessDenied: &v1alpha.AccessDenied{
					State:          v1alpha.AccessDeniedStateAll,
					Since:          time.Date(2026, 10, 19, 9, 30, 0, 0, time.UTC),
					HTTPStatusCode: 403,
					AffectedSLOs:   2,
					TotalSLOs:      2,
				},
			},
			expectedJSON: `{"agentType":"Datadog","environment":{},"accessDenied":{"state":"all",` +
				`"since":"2026-10-19T09:30:00Z","httpStatusCode":403,"affectedSLOs":2,"totalSLOs":2}}`,
		},
		"degraded access denied keeps the vendor status code": {
			status: Status{
				AgentType: "Datadog",
				AccessDenied: &v1alpha.AccessDenied{
					State:          v1alpha.AccessDeniedStateDegraded,
					Since:          time.Date(2026, 10, 19, 9, 30, 0, 0, time.UTC),
					HTTPStatusCode: 402,
					AffectedSLOs:   1,
					TotalSLOs:      3,
				},
			},
			expectedJSON: `{"agentType":"Datadog","environment":{},"accessDenied":{"state":"degraded",` +
				`"since":"2026-10-19T09:30:00Z","httpStatusCode":402,"affectedSLOs":1,"totalSLOs":3}}`,
		},
		"none state has only the state": {
			status: Status{
				AgentType: "Datadog",
				AccessDenied: &v1alpha.AccessDenied{
					State: v1alpha.AccessDeniedStateNone,
				},
			},
			expectedJSON: `{"agentType":"Datadog","environment":{},"accessDenied":{"state":"none"}}`,
		},
		"access denied next to invalid credentials": {
			status: Status{
				AgentType: "Datadog",
				InvalidCredentials: &v1alpha.InvalidCredentials{
					State:          v1alpha.InvalidCredentialsStateDegraded,
					Since:          time.Date(2026, 10, 19, 9, 0, 0, 0, time.UTC),
					HTTPStatusCode: 401,
					AffectedSLOs:   1,
					TotalSLOs:      3,
				},
				AccessDenied: &v1alpha.AccessDenied{
					State:          v1alpha.AccessDeniedStateDegraded,
					Since:          time.Date(2026, 10, 19, 9, 30, 0, 0, time.UTC),
					HTTPStatusCode: 403,
					AffectedSLOs:   2,
					TotalSLOs:      3,
				},
			},
			expectedJSON: `{"agentType":"Datadog","environment":{},` +
				`"invalidCredentials":{"state":"degraded","since":"2026-10-19T09:00:00Z","httpStatusCode":401,` +
				`"affectedSLOs":1,"totalSLOs":3},` +
				`"accessDenied":{"state":"degraded","since":"2026-10-19T09:30:00Z","httpStatusCode":403,` +
				`"affectedSLOs":2,"totalSLOs":3}}`,
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

	t.Run("JSON without access denied decodes to nil", func(t *testing.T) {
		var decoded Status
		require.NoError(t, json.Unmarshal([]byte(`{"agentType":"Datadog"}`), &decoded))
		assert.Equal(t, Status{AgentType: "Datadog"}, decoded)
	})

	t.Run("struct without access denied ignores it", func(t *testing.T) {
		type oldStatus struct {
			AgentType string `json:"agentType"`
		}
		var decoded oldStatus
		data := []byte(`{"agentType":"Datadog",` +
			`"accessDenied":{"state":"all","since":"2026-10-19T09:30:00Z","httpStatusCode":403,"affectedSLOs":2,"totalSLOs":2}}`)
		require.NoError(t, json.Unmarshal(data, &decoded))
		assert.Equal(t, oldStatus{AgentType: "Datadog"}, decoded)
	})
}

func TestStatus_DataSourceUnreachableSerialization(t *testing.T) {
	tests := map[string]struct {
		status       Status
		expectedJSON string
	}{
		"no data source unreachable is omitted": {
			status:       Status{AgentType: "Prometheus"},
			expectedJSON: `{"agentType":"Prometheus","environment":{}}`,
		},
		"all data source unreachable is preserved": {
			status: Status{
				AgentType: "Prometheus",
				DataSourceUnreachable: &v1alpha.DataSourceUnreachable{
					State:        v1alpha.DataSourceUnreachableStateAll,
					Since:        time.Date(2026, 10, 18, 9, 30, 0, 0, time.UTC),
					AffectedSLOs: 2,
					TotalSLOs:    2,
				},
			},
			expectedJSON: `{"agentType":"Prometheus","environment":{},"dataSourceUnreachable":{"state":"all",` +
				`"since":"2026-10-18T09:30:00Z","affectedSLOs":2,"totalSLOs":2}}`,
		},
		"degraded data source unreachable is preserved": {
			status: Status{
				AgentType: "Prometheus",
				DataSourceUnreachable: &v1alpha.DataSourceUnreachable{
					State:        v1alpha.DataSourceUnreachableStateDegraded,
					Since:        time.Date(2026, 10, 18, 9, 30, 0, 0, time.UTC),
					AffectedSLOs: 1,
					TotalSLOs:    3,
				},
			},
			expectedJSON: `{"agentType":"Prometheus","environment":{},"dataSourceUnreachable":{"state":"degraded",` +
				`"since":"2026-10-18T09:30:00Z","affectedSLOs":1,"totalSLOs":3}}`,
		},
		"none state has only the state": {
			status: Status{
				AgentType: "Prometheus",
				DataSourceUnreachable: &v1alpha.DataSourceUnreachable{
					State: v1alpha.DataSourceUnreachableStateNone,
				},
			},
			expectedJSON: `{"agentType":"Prometheus","environment":{},"dataSourceUnreachable":{"state":"none"}}`,
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

	t.Run("JSON without data source unreachable decodes to nil", func(t *testing.T) {
		var decoded Status
		require.NoError(t, json.Unmarshal([]byte(`{"agentType":"Prometheus"}`), &decoded))
		assert.Equal(t, Status{AgentType: "Prometheus"}, decoded)
	})

	t.Run("struct without data source unreachable ignores it", func(t *testing.T) {
		type oldStatus struct {
			AgentType string `json:"agentType"`
		}
		var decoded oldStatus
		data := []byte(`{"agentType":"Prometheus",` +
			`"dataSourceUnreachable":{"state":"all","since":"2026-10-18T09:30:00Z","affectedSLOs":2,"totalSLOs":2}}`)
		require.NoError(t, json.Unmarshal(data, &decoded))
		assert.Equal(t, oldStatus{AgentType: "Prometheus"}, decoded)
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
