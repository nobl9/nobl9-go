package sdk

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	objectsV2 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v2"
)

func TestClient_Objects_V2_Apply(t *testing.T) {
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
			assert.Equal(t, url.Values{objectsV2.QueryKeyDryRun: {"true"}}, r.URL.Query())
			objects, err := ReadObjectsFromSources(context.Background(), NewObjectSourceReader(r.Body, ""))
			require.NoError(t, err)
			assert.Equal(t, expected, objects)
		},
	})
	defer srv.Close()

	// Run the API method.
	client.WithDryRun()
	err := client.Objects().V2().Apply(context.Background(), objectsV2.ApplyRequest{Objects: requestPayload})
	// Verify response handling.
	require.NoError(t, err)
}

func TestClient_Objects_V2_Delete(t *testing.T) {
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
			assert.Equal(t,
				url.Values{
					objectsV2.QueryKeyDryRun:        {"true"},
					objectsV2.QueryKeyCascadeDelete: {"false"},
				},
				r.URL.Query())
			objects, err := ReadObjectsFromSources(context.Background(), NewObjectSourceReader(r.Body, ""))
			require.NoError(t, err)
			assert.Equal(t, expected, objects)
		},
	})
	defer srv.Close()

	// Run the API method.
	client.WithDryRun()
	err := client.Objects().V2().Delete(context.Background(), objectsV2.DeleteRequest{Objects: requestPayload})
	// Verify response handling.
	require.NoError(t, err)
}

func TestClient_Objects_V2_DeleteByName(t *testing.T) {
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
				objectsV2.QueryKeyName:          {"service1", "service2"},
				objectsV2.QueryKeyDryRun:        {"true"},
				objectsV2.QueryKeyCascadeDelete: {"true"},
			}, r.URL.Query())
		},
	})
	defer srv.Close()

	// Run the API method.
	client.WithDryRun()
	err := client.Objects().V2().DeleteByName(
		context.Background(),
		objectsV2.DeleteByNameRequest{
			Kind:    manifest.KindService,
			Project: "my-project",
			Names:   []string{"service1", "service2"},
			Cascade: true,
		},
	)
	// Verify response handling.
	require.NoError(t, err)
}
