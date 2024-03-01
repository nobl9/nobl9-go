package sdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authDataV1 "github.com/nobl9/nobl9-go/sdk/endpoints/authdata/v1"
)

func TestClient_AuthData_V1_GetDataExportIAMRoleIDs(t *testing.T) {
	expectedData := authDataV1.IAMRoleIDs{
		ExternalID: "external-id",
	}

	client, srv := prepareTestClient(t, endpointConfig{
		// Define endpoint response.
		Path: "api/get/dataexport/aws-external-id",
		ResponseFunc: func(t *testing.T, w http.ResponseWriter) {
			require.NoError(t, json.NewEncoder(w).Encode(expectedData))
		},
		// Verify request parameters.
		TestRequestFunc: func(t *testing.T, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
		},
	})
	defer srv.Close()

	// Run the API method.
	response, err := client.AuthData().V1().GetDataExportIAMRoleIDs(context.Background())
	// Verify response handling.
	require.NoError(t, err)
	assert.Equal(t, expectedData, *response)
}

func TestClient_AuthData_V1_GetDirectIAMRoleIDs(t *testing.T) {
	expectedData := authDataV1.IAMRoleIDs{
		ExternalID: "N9-1AE8AC4A-33A909BC-2D0483BE-2874FCD1",
		AccountID:  "123456789012",
	}

	client, srv := prepareTestClient(t, endpointConfig{
		// Define endpoint response.
		Path: "api/data-sources/iam-role-auth-data/test-direct-name",
		ResponseFunc: func(t *testing.T, w http.ResponseWriter) {
			require.NoError(t, json.NewEncoder(w).Encode(expectedData))
		},
		// Verify request parameters.
		TestRequestFunc: func(t *testing.T, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
		},
	})
	defer srv.Close()

	// Run the API method.
	response, err := client.AuthData().V1().GetDirectIAMRoleIDs(context.Background(), "default", "test-direct-name")
	// Verify response handling.
	require.NoError(t, err)
	assert.Equal(t, expectedData, *response)
}

func TestClient_AuthData_V1_GetAgentCredentials(t *testing.T) {
	responsePayload := authDataV1.M2MAppCredentials{
		ClientID:     "agent-client-id",
		ClientSecret: "agent-client-secret",
	}

	client, srv := prepareTestClient(t, endpointConfig{
		// Define endpoint response.
		Path: "api/internal/agent/clientcreds",
		ResponseFunc: func(t *testing.T, w http.ResponseWriter) {
			require.NoError(t, json.NewEncoder(w).Encode(responsePayload))
		},
		// Verify request parameters.
		TestRequestFunc: func(t *testing.T, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "agent-project", r.Header.Get(HeaderProject))
			assert.Equal(t, url.Values{authDataV1.QueryKeyName: {"my-agent"}}, r.URL.Query())
		},
	})
	defer srv.Close()

	// Run the API method.
	objects, err := client.AuthData().V1().GetAgentCredentials(
		context.Background(),
		"agent-project",
		"my-agent",
	)
	// Verify response handling.
	require.NoError(t, err)
	assert.Equal(t, responsePayload, objects)
}
