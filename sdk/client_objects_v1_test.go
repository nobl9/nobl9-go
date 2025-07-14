package sdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/nobl9/govy/pkg/govytest"
	"github.com/nobl9/govy/pkg/rules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	objectsV1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
)

func TestClient_Objects_V1_Get(t *testing.T) {
	responsePayload := []manifest.Object{
		v1alphaService.New(
			v1alphaService.Metadata{
				Name:    "service1",
				Project: "default",
			},
			v1alphaService.Spec{},
		),
		v1alphaService.New(
			v1alphaService.Metadata{
				Name:    "service2",
				Project: "default",
			},
			v1alphaService.Spec{},
		),
	}

	client, srv := prepareTestClient(t, endpointConfig{
		// Define endpoint response.
		Path: "api/get/service",
		ResponseFunc: func(t *testing.T, w http.ResponseWriter) {
			require.NoError(t, json.NewEncoder(w).Encode(responsePayload))
		},
		// Verify request parameters.
		TestRequestFunc: func(t *testing.T, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "non-default", r.Header.Get(HeaderProject))
			assert.Equal(t, url.Values{
				objectsV1.QueryKeyName:   {"service1", "service2"},
				objectsV1.QueryKeyLabels: {"team:green,team:purple"},
			}, r.URL.Query())
		},
	})
	defer srv.Close()

	// Run the API method.
	objects, err := client.Objects().V1().Get(
		context.Background(),
		manifest.KindService,
		http.Header{HeaderProject: []string{"non-default"}},
		url.Values{
			objectsV1.QueryKeyName:   {"service1", "service2"},
			objectsV1.QueryKeyLabels: {"team:green,team:purple"},
		},
	)
	// Verify response handling.
	require.NoError(t, err)
	require.Len(t, objects, 2)
	assert.Equal(t, responsePayload, objects)
}

func TestClient_GetObjects_NoObjectsInResponse(t *testing.T) {
	responsePayload := make([]manifest.Object, 0)

	client, srv := prepareTestClient(t, endpointConfig{
		// Define endpoint response.
		Path: "api/get/service",
		ResponseFunc: func(t *testing.T, w http.ResponseWriter) {
			require.NoError(t, json.NewEncoder(w).Encode(responsePayload))
		},
	})
	defer srv.Close()

	// Run the API method.
	objects, err := client.Objects().V1().Get(
		context.Background(),
		manifest.KindService,
		nil,
		nil,
	)
	// Verify response handling.
	require.NoError(t, err)
	require.Len(t, objects, 0)
}

func TestClient_Objects_V1_Apply(t *testing.T) {
	requestPayload := []manifest.Object{
		v1alphaService.New(
			v1alphaService.Metadata{
				Name:    "service1",
				Project: "default",
			},
			v1alphaService.Spec{},
		),
	}
	expected := addOrganization(requestPayload, "my-org")

	client, srv := prepareTestClient(t, endpointConfig{
		// Define endpoint response.
		Path: "api/apply",
		ResponseFunc: func(t *testing.T, w http.ResponseWriter) {
			w.WriteHeader(http.StatusOK)
		},
		// Verify request parameters.
		TestRequestFunc: func(t *testing.T, r *http.Request) {
			assert.Equal(t, http.MethodPut, r.Method)
			assert.Equal(t, url.Values{objectsV1.QueryKeyDryRun: {"true"}}, r.URL.Query())
			objects, err := ReadObjectsFromSources(context.Background(), NewObjectSourceReader(r.Body, ""))
			require.NoError(t, err)
			assert.Equal(t, expected, objects)
		},
	})
	defer srv.Close()

	// Run the API method.
	client.WithDryRun()
	err := client.Objects().V1().Apply(context.Background(), requestPayload)
	// Verify response handling.
	require.NoError(t, err)
}

func TestClient_Objects_V1_Delete(t *testing.T) {
	requestPayload := []manifest.Object{
		v1alphaService.New(
			v1alphaService.Metadata{
				Name:    "service1",
				Project: "default",
			},
			v1alphaService.Spec{},
		),
	}
	expected := addOrganization(requestPayload, "my-org")

	client, srv := prepareTestClient(t, endpointConfig{
		// Define endpoint response.
		Path: "api/delete",
		ResponseFunc: func(t *testing.T, w http.ResponseWriter) {
			w.WriteHeader(http.StatusOK)
		},
		// Verify request parameters.
		TestRequestFunc: func(t *testing.T, r *http.Request) {
			assert.Equal(t, http.MethodDelete, r.Method)
			assert.Equal(t, url.Values{objectsV1.QueryKeyDryRun: {"true"}}, r.URL.Query())
			objects, err := ReadObjectsFromSources(context.Background(), NewObjectSourceReader(r.Body, ""))
			require.NoError(t, err)
			assert.Equal(t, expected, objects)
		},
	})
	defer srv.Close()

	// Run the API method.
	client.WithDryRun()
	err := client.Objects().V1().Delete(context.Background(), requestPayload)
	// Verify response handling.
	require.NoError(t, err)
}

func TestClient_Objects_V1_DeleteByName(t *testing.T) {
	client, srv := prepareTestClient(t, endpointConfig{
		// Define endpoint response.
		Path: "api/delete/service",
		ResponseFunc: func(t *testing.T, w http.ResponseWriter) {
			w.WriteHeader(http.StatusOK)
		},
		// Verify request parameters.
		TestRequestFunc: func(t *testing.T, r *http.Request) {
			assert.Equal(t, http.MethodDelete, r.Method)
			assert.Equal(t, "my-project", r.Header.Get(HeaderProject))
			assert.Equal(t, url.Values{
				objectsV1.QueryKeyName:   {"service1", "service2"},
				objectsV1.QueryKeyDryRun: {"true"},
			}, r.URL.Query())
		},
	})
	defer srv.Close()

	// Run the API method.
	client.WithDryRun()
	err := client.Objects().V1().DeleteByName(
		context.Background(),
		manifest.KindService,
		"my-project",
		"service1",
		"service2",
	)
	// Verify response handling.
	require.NoError(t, err)
}

func TestClient_Objects_V1_MoveSLOs(t *testing.T) {
	client := Client{}

	err := client.Objects().V1().MoveSLOs(context.Background(), objectsV1.MoveSLOsRequest{
		SLONames:   []string{},
		NewProject: "bar",
		OldProject: "baz",
	})
	require.Error(t, err)
	govytest.AssertError(t, err, govytest.ExpectedRuleError{
		PropertyName: "sloNames",
		Code:         rules.ErrorCodeSliceMinLength,
	})
}
