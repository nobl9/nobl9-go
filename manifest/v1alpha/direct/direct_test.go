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

func TestStatus_InvalidCredentialsSerialization(t *testing.T) {
	tests := map[string]struct {
		status       Status
		expectedJSON string
	}{
		"no invalid credentials is omitted": {
			status:       Status{DirectType: "Datadog"},
			expectedJSON: `{"directType":"Datadog"}`,
		},
		"degraded invalid credentials is preserved": {
			status: Status{
				DirectType: "Datadog",
				InvalidCredentials: &v1alpha.InvalidCredentials{
					State:          v1alpha.InvalidCredentialsStateDegraded,
					Since:          time.Date(2026, 10, 16, 9, 30, 0, 0, time.UTC),
					HTTPStatusCode: 401,
					AffectedSLOs:   1,
					TotalSLOs:      2,
				},
			},
			expectedJSON: `{"directType":"Datadog","invalidCredentials":{"state":"degraded",` +
				`"since":"2026-10-16T09:30:00Z","httpStatusCode":401,"affectedSLOs":1,"totalSLOs":2}}`,
		},
		"none state has only the state": {
			status: Status{
				DirectType: "Datadog",
				InvalidCredentials: &v1alpha.InvalidCredentials{
					State: v1alpha.InvalidCredentialsStateNone,
				},
			},
			expectedJSON: `{"directType":"Datadog",` +
				`"invalidCredentials":{"state":"none"}}`,
		},
		"verifying state has only the state": {
			status: Status{
				DirectType: "Datadog",
				InvalidCredentials: &v1alpha.InvalidCredentials{
					State: v1alpha.InvalidCredentialsStateVerifying,
				},
			},
			expectedJSON: `{"directType":"Datadog",` +
				`"invalidCredentials":{"state":"verifying"}}`,
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
		require.NoError(t, json.Unmarshal([]byte(`{"directType":"Datadog"}`), &decoded))
		assert.Equal(t, Status{DirectType: "Datadog"}, decoded)
	})

	t.Run("invalid credentials without state and counts decodes to zero values", func(t *testing.T) {
		var decoded Status
		data := []byte(`{"directType":"Datadog","invalidCredentials":{"since":"2026-10-16T09:30:00Z","httpStatusCode":401}}`)
		require.NoError(t, json.Unmarshal(data, &decoded))
		assert.Equal(t, Status{
			DirectType: "Datadog",
			InvalidCredentials: &v1alpha.InvalidCredentials{
				Since:          time.Date(2026, 10, 16, 9, 30, 0, 0, time.UTC),
				HTTPStatusCode: 401,
			},
		}, decoded)
	})

	t.Run("struct without invalid credentials ignores it", func(t *testing.T) {
		type oldStatus struct {
			DirectType string `json:"directType"`
		}
		var decoded oldStatus
		data := []byte(`{"directType":"Datadog",` +
			`"invalidCredentials":{"since":"2026-10-16T09:30:00Z","httpStatusCode":401}}`)
		require.NoError(t, json.Unmarshal(data, &decoded))
		assert.Equal(t, oldStatus{DirectType: "Datadog"}, decoded)
	})
}

func TestStatus_AccessDeniedSerialization(t *testing.T) {
	tests := map[string]struct {
		status       Status
		expectedJSON string
	}{
		"no access denied is omitted": {
			status:       Status{DirectType: "Datadog"},
			expectedJSON: `{"directType":"Datadog"}`,
		},
		"all access denied is preserved": {
			status: Status{
				DirectType: "Datadog",
				AccessDenied: &v1alpha.AccessDenied{
					State:          v1alpha.AccessDeniedStateAll,
					Since:          time.Date(2026, 10, 19, 9, 30, 0, 0, time.UTC),
					HTTPStatusCode: 403,
					AffectedSLOs:   2,
					TotalSLOs:      2,
				},
			},
			expectedJSON: `{"directType":"Datadog","accessDenied":{"state":"all",` +
				`"since":"2026-10-19T09:30:00Z","httpStatusCode":403,"affectedSLOs":2,"totalSLOs":2}}`,
		},
		"degraded access denied keeps the vendor status code": {
			status: Status{
				DirectType: "Datadog",
				AccessDenied: &v1alpha.AccessDenied{
					State:          v1alpha.AccessDeniedStateDegraded,
					Since:          time.Date(2026, 10, 19, 9, 30, 0, 0, time.UTC),
					HTTPStatusCode: 402,
					AffectedSLOs:   1,
					TotalSLOs:      3,
				},
			},
			expectedJSON: `{"directType":"Datadog","accessDenied":{"state":"degraded",` +
				`"since":"2026-10-19T09:30:00Z","httpStatusCode":402,"affectedSLOs":1,"totalSLOs":3}}`,
		},
		"none state has only the state": {
			status: Status{
				DirectType: "Datadog",
				AccessDenied: &v1alpha.AccessDenied{
					State: v1alpha.AccessDeniedStateNone,
				},
			},
			expectedJSON: `{"directType":"Datadog","accessDenied":{"state":"none"}}`,
		},
		"verifying state has only the state": {
			status: Status{
				DirectType: "Datadog",
				AccessDenied: &v1alpha.AccessDenied{
					State: v1alpha.AccessDeniedStateVerifying,
				},
			},
			expectedJSON: `{"directType":"Datadog","accessDenied":{"state":"verifying"}}`,
		},
		"access denied next to invalid credentials": {
			status: Status{
				DirectType: "Datadog",
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
			expectedJSON: `{"directType":"Datadog",` +
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
		require.NoError(t, json.Unmarshal([]byte(`{"directType":"Datadog"}`), &decoded))
		assert.Equal(t, Status{DirectType: "Datadog"}, decoded)
	})

	t.Run("struct without access denied ignores it", func(t *testing.T) {
		type oldStatus struct {
			DirectType string `json:"directType"`
		}
		var decoded oldStatus
		data := []byte(`{"directType":"Datadog",` +
			`"accessDenied":{"state":"all","since":"2026-10-19T09:30:00Z","httpStatusCode":403,"affectedSLOs":2,"totalSLOs":2}}`)
		require.NoError(t, json.Unmarshal(data, &decoded))
		assert.Equal(t, oldStatus{DirectType: "Datadog"}, decoded)
	})
}

func TestStatus_DataSourceUnreachableSerialization(t *testing.T) {
	tests := map[string]struct {
		status       Status
		expectedJSON string
	}{
		"no data source unreachable is omitted": {
			status:       Status{DirectType: "Prometheus"},
			expectedJSON: `{"directType":"Prometheus"}`,
		},
		"all data source unreachable is preserved": {
			status: Status{
				DirectType: "Prometheus",
				DataSourceUnreachable: &v1alpha.DataSourceUnreachable{
					State:        v1alpha.DataSourceUnreachableStateAll,
					Since:        time.Date(2026, 10, 18, 9, 30, 0, 0, time.UTC),
					AffectedSLOs: 2,
					TotalSLOs:    2,
					Cause:        v1alpha.DataSourceUnreachableCauseHostNotFound,
				},
			},
			expectedJSON: `{"directType":"Prometheus","dataSourceUnreachable":{"state":"all",` +
				`"since":"2026-10-18T09:30:00Z","affectedSLOs":2,"totalSLOs":2,"cause":"hostNotFound"}}`,
		},
		"degraded data source unreachable is preserved": {
			status: Status{
				DirectType: "Prometheus",
				DataSourceUnreachable: &v1alpha.DataSourceUnreachable{
					State:        v1alpha.DataSourceUnreachableStateDegraded,
					Since:        time.Date(2026, 10, 18, 9, 30, 0, 0, time.UTC),
					AffectedSLOs: 1,
					TotalSLOs:    3,
					Cause:        v1alpha.DataSourceUnreachableCauseNotDataSource,
				},
			},
			expectedJSON: `{"directType":"Prometheus","dataSourceUnreachable":{"state":"degraded",` +
				`"since":"2026-10-18T09:30:00Z","affectedSLOs":1,"totalSLOs":3,"cause":"notDataSource"}}`,
		},
		"certificate cause is preserved": {
			status: Status{
				DirectType: "Prometheus",
				DataSourceUnreachable: &v1alpha.DataSourceUnreachable{
					State:        v1alpha.DataSourceUnreachableStateAll,
					Since:        time.Date(2026, 10, 19, 9, 30, 0, 0, time.UTC),
					AffectedSLOs: 2,
					TotalSLOs:    2,
					Cause:        v1alpha.DataSourceUnreachableCauseCertificate,
				},
			},
			expectedJSON: `{"directType":"Prometheus","dataSourceUnreachable":{"state":"all",` +
				`"since":"2026-10-19T09:30:00Z","affectedSLOs":2,"totalSLOs":2,"cause":"certificate"}}`,
		},
		"none state has only the state": {
			status: Status{
				DirectType: "Prometheus",
				DataSourceUnreachable: &v1alpha.DataSourceUnreachable{
					State: v1alpha.DataSourceUnreachableStateNone,
				},
			},
			expectedJSON: `{"directType":"Prometheus","dataSourceUnreachable":{"state":"none"}}`,
		},
		"verifying state has only the state": {
			status: Status{
				DirectType: "Prometheus",
				DataSourceUnreachable: &v1alpha.DataSourceUnreachable{
					State: v1alpha.DataSourceUnreachableStateVerifying,
				},
			},
			expectedJSON: `{"directType":"Prometheus","dataSourceUnreachable":{"state":"verifying"}}`,
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
		require.NoError(t, json.Unmarshal([]byte(`{"directType":"Prometheus"}`), &decoded))
		assert.Equal(t, Status{DirectType: "Prometheus"}, decoded)
	})

	t.Run("struct without data source unreachable ignores it", func(t *testing.T) {
		type oldStatus struct {
			DirectType string `json:"directType"`
		}
		var decoded oldStatus
		data := []byte(`{"directType":"Prometheus",` +
			`"dataSourceUnreachable":{"state":"all","since":"2026-10-18T09:30:00Z","affectedSLOs":2,"totalSLOs":2,` +
			`"cause":"hostNotFound"}}`)
		require.NoError(t, json.Unmarshal(data, &decoded))
		assert.Equal(t, oldStatus{DirectType: "Prometheus"}, decoded)
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
