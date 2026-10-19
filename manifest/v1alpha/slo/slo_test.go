package slo

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestSpec_HasCompositeObjectives(t *testing.T) {
	t.Run("passes - composite SLO", func(t *testing.T) {
		slo := validCompositeSLO()

		assert.Equal(t, true, slo.Spec.HasCompositeObjectives())
	})

	t.Run("passes - normal SLO", func(t *testing.T) {
		slo := validSLO()

		assert.Equal(t, false, slo.Spec.HasCompositeObjectives())
	})
}

func TestAnomalyConfigNoData_TreatZeroAsNoDataSerialization(t *testing.T) {
	tests := map[string]struct {
		value           *bool
		expectedJSON    string
		expectedPresent bool
	}{
		"unset is omitted": {
			expectedJSON:    `{"alertMethods":[{"name":"my-name"}]}`,
			expectedPresent: false,
		},
		"true is preserved": {
			value:           ptr(true),
			expectedJSON:    `{"alertMethods":[{"name":"my-name"}],"treatZeroAsNoData":true}`,
			expectedPresent: true,
		},
		"false is preserved": {
			value:           ptr(false),
			expectedJSON:    `{"alertMethods":[{"name":"my-name"}],"treatZeroAsNoData":false}`,
			expectedPresent: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			noData := AnomalyConfigNoData{
				AlertMethods: []AnomalyConfigAlertMethod{
					{Name: "my-name"},
				},
				TreatZeroAsNoData: test.value,
			}

			data, err := json.Marshal(noData)
			require.NoError(t, err)
			assert.JSONEq(t, test.expectedJSON, string(data))

			var decoded map[string]any
			require.NoError(t, json.Unmarshal(data, &decoded))
			_, found := decoded["treatZeroAsNoData"]
			assert.Equal(t, test.expectedPresent, found)
		})
	}
}

func TestErrorDetails_ReasonSerialization(t *testing.T) {
	tests := map[string]struct {
		details      ErrorDetails
		expectedJSON string
	}{
		"empty reason is omitted": {
			details:      ErrorDetails{ValidationResult: "invalid"},
			expectedJSON: `{"message":null,"validationResult":"invalid","logTimestamp":null,"httpStatusCode":null,"query":""}`,
		},
		"invalidCredentials reason is preserved": {
			details: ErrorDetails{
				ValidationResult: "invalid",
				HTTPStatusCode:   ptr(401),
				Reason:           ErrorDetailsReasonInvalidCredentials,
			},
			expectedJSON: `{"message":null,"validationResult":"invalid","logTimestamp":null,"httpStatusCode":401,` +
				`"query":"","reason":"invalidCredentials"}`,
		},
		"accessDenied reason is preserved": {
			details: ErrorDetails{
				ValidationResult: "invalid",
				HTTPStatusCode:   ptr(403),
				Reason:           ErrorDetailsReasonAccessDenied,
			},
			expectedJSON: `{"message":null,"validationResult":"invalid","logTimestamp":null,"httpStatusCode":403,` +
				`"query":"","reason":"accessDenied"}`,
		},
		"invalidQuery reason is preserved": {
			details: ErrorDetails{ValidationResult: "invalid", HTTPStatusCode: ptr(400), Reason: ErrorDetailsReasonInvalidQuery},
			expectedJSON: `{"message":null,"validationResult":"invalid","logTimestamp":null,"httpStatusCode":400,` +
				`"query":"","reason":"invalidQuery"}`,
		},
		"dataSourceUnreachable reason is preserved": {
			details: ErrorDetails{
				ValidationResult: "invalid",
				HTTPStatusCode:   ptr(500),
				Reason:           ErrorDetailsReasonDataSourceUnreachable,
				Cause:            v1alpha.DataSourceUnreachableCauseHostNotFound,
			},
			expectedJSON: `{"message":null,"validationResult":"invalid","logTimestamp":null,"httpStatusCode":500,` +
				`"query":"","reason":"dataSourceUnreachable","cause":"hostNotFound"}`,
		},
		"dataSourceUnreachable reason with the certificate cause is preserved": {
			details: ErrorDetails{
				ValidationResult: "invalid",
				HTTPStatusCode:   ptr(500),
				Reason:           ErrorDetailsReasonDataSourceUnreachable,
				Cause:            v1alpha.DataSourceUnreachableCauseCertificate,
			},
			expectedJSON: `{"message":null,"validationResult":"invalid","logTimestamp":null,"httpStatusCode":500,` +
				`"query":"","reason":"dataSourceUnreachable","cause":"certificate"}`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := json.Marshal(test.details)
			require.NoError(t, err)
			assert.JSONEq(t, test.expectedJSON, string(data))

			var decoded ErrorDetails
			require.NoError(t, json.Unmarshal(data, &decoded))
			assert.Equal(t, test.details, decoded)
		})
	}

	t.Run("JSON without reason decodes to an empty reason", func(t *testing.T) {
		var decoded ErrorDetails
		require.NoError(t, json.Unmarshal([]byte(`{"validationResult":"invalid","httpStatusCode":400}`), &decoded))
		assert.Equal(t, ErrorDetails{ValidationResult: "invalid", HTTPStatusCode: ptr(400)}, decoded)
	})

	t.Run("struct without reason ignores it", func(t *testing.T) {
		type oldErrorDetails struct {
			ValidationResult string `json:"validationResult"`
			HTTPStatusCode   *int   `json:"httpStatusCode"`
		}
		var decoded oldErrorDetails
		data := []byte(`{"validationResult":"invalid","httpStatusCode":401,"reason":"invalidCredentials"}`)
		require.NoError(t, json.Unmarshal(data, &decoded))
		assert.Equal(t, oldErrorDetails{ValidationResult: "invalid", HTTPStatusCode: ptr(401)}, decoded)
	})
}
