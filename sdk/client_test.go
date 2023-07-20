package sdk

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
)

func TestClient_GetObjects(t *testing.T) {
	responsePayload := []AnyJSONObj{
		{
			"apiVersion": "v1alpha",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"name":    "service1",
				"project": "default",
			},
		},
		{
			"apiVersion": "v1alpha",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"name":    "service2",
				"project": "default",
			},
		},
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
				QueryKeyName:         {"service1", "service2"},
				QueryKeyLabelsFilter: {"team:green,team:purple"},
			}, r.URL.Query())
		},
	})

	// Start and close the test server.
	srv.Start()
	defer srv.Close()

	// Run the API method.
	objects, err := client.GetObjects(
		context.Background(),
		"non-default",
		manifest.KindService,
		map[string][]string{"team": {"green", "purple"}},
		"service1", "service2",
	)
	// Verify response handling.
	require.NoError(t, err)
	require.Len(t, objects, 2)
	assert.Equal(t, responsePayload, objects)
}

func TestClient_GetObjects_GroupsEndpoint(t *testing.T) {
	calledTimes := 0
	client, srv := prepareTestClient(t, endpointConfig{
		// Define endpoint response.
		Path: "api/usrmgmt/groups",
		ResponseFunc: func(t *testing.T, w http.ResponseWriter) {
			require.NoError(t, json.NewEncoder(w).Encode([]AnyJSONObj{}))
		},
		// Verify request parameters.
		TestRequestFunc: func(t *testing.T, r *http.Request) {
			calledTimes++
		},
	})

	// Start and close the test server.
	srv.Start()
	defer srv.Close()

	// Run the API method.
	_, err := client.GetObjects(context.Background(), "", manifest.KindUserGroup, nil)
	// Verify response handling.
	require.NoError(t, err)
	assert.Equal(t, 1, calledTimes)
}

func TestClient_ApplyObjects(t *testing.T) {
	requestPayload := []AnyJSONObj{
		{
			"apiVersion": "v1alpha",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"name":    "service1",
				"project": "default",
			},
		},
	}

	client, srv := prepareTestClient(t, endpointConfig{
		// Define endpoint response.
		Path: "api/apply",
		ResponseFunc: func(t *testing.T, w http.ResponseWriter) {
			w.WriteHeader(http.StatusOK)
		},
		// Verify request parameters.
		TestRequestFunc: func(t *testing.T, r *http.Request) {
			assert.Equal(t, http.MethodPut, r.Method)
			assert.Equal(t, "", r.Header.Get(HeaderProject))
			assert.Equal(t, url.Values{QueryKeyDryRun: {"true"}}, r.URL.Query())
			var objects []AnyJSONObj
			require.NoError(t, json.NewDecoder(r.Body).Decode(&objects))
			assert.Equal(t, requestPayload, objects)
		},
	})

	// Start and close the test server.
	srv.Start()
	defer srv.Close()

	// Run the API method.
	err := client.ApplyObjects(context.Background(), requestPayload, true)
	// Verify response handling.
	require.NoError(t, err)
}

func TestClient_DeleteObjects(t *testing.T) {
	requestPayload := []AnyJSONObj{
		{
			"apiVersion": "v1alpha",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"name":    "service1",
				"project": "default",
			},
		},
	}

	client, srv := prepareTestClient(t, endpointConfig{
		// Define endpoint response.
		Path: "api/delete",
		ResponseFunc: func(t *testing.T, w http.ResponseWriter) {
			w.WriteHeader(http.StatusOK)
		},
		// Verify request parameters.
		TestRequestFunc: func(t *testing.T, r *http.Request) {
			assert.Equal(t, http.MethodDelete, r.Method)
			assert.Equal(t, "", r.Header.Get(HeaderProject))
			assert.Equal(t, url.Values{QueryKeyDryRun: {"true"}}, r.URL.Query())
		},
	})

	// Start and close the test server.
	srv.Start()
	defer srv.Close()

	// Run the API method.
	err := client.DeleteObjects(context.Background(), requestPayload, true)
	// Verify response handling.
	require.NoError(t, err)
}

func TestClient_DeleteObjectsByName(t *testing.T) {
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
				QueryKeyName:   {"service1", "service2"},
				QueryKeyDryRun: {"true"},
			}, r.URL.Query())
		},
	})

	// Start and close the test server.
	srv.Start()
	defer srv.Close()

	// Run the API method.
	err := client.DeleteObjectsByName(
		context.Background(),
		"my-project",
		manifest.KindService,
		true,
		"service1",
		"service2",
	)
	// Verify response handling.
	require.NoError(t, err)
}

func TestClient_GetAWSExternalID(t *testing.T) {
	client, srv := prepareTestClient(t, endpointConfig{
		// Define endpoint response.
		Path: "api/get/dataexport/aws-external-id",
		ResponseFunc: func(t *testing.T, w http.ResponseWriter) {
			require.NoError(t, json.NewEncoder(w).Encode(map[string]interface{}{"awsExternalID": "external-id"}))
		},
		// Verify request parameters.
		TestRequestFunc: func(t *testing.T, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "my-project", r.Header.Get(HeaderProject))
		},
	})

	// Start and close the test server.
	srv.Start()
	defer srv.Close()

	// Run the API method.
	externalID, err := client.GetAWSExternalID(context.Background(), "my-project")
	// Verify response handling.
	require.NoError(t, err)
	assert.Equal(t, "external-id", externalID)
}

func TestClient_GetAgentCredentials(t *testing.T) {
	responsePayload := M2MAppCredentials{
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
			assert.Equal(t, url.Values{QueryKeyName: {"my-agent"}}, r.URL.Query())
		},
	})

	// Start and close the test server.
	srv.Start()
	defer srv.Close()

	// Run the API method.
	objects, err := client.GetAgentCredentials(
		context.Background(),
		"agent-project",
		"my-agent",
	)
	// Verify response handling.
	require.NoError(t, err)
	assert.Equal(t, responsePayload, objects)
}

type endpointConfig struct {
	Path            string
	ResponseFunc    func(t *testing.T, w http.ResponseWriter)
	TestRequestFunc func(*testing.T, *http.Request)
}

func prepareTestClient(t *testing.T, endpoint endpointConfig) (client *Client, srv *httptest.Server) {
	t.Helper()
	urlScheme = "http"
	const (
		oktaAuthServer = "auseg9kiegWKEtJZC416"
		kid            = "my-kid"
		clientID       = "client-id"
		clientSecret   = "super-secret"
		organization   = "my-org"
		userAgent      = "sloctl"
	)
	// Declare the test server, we can provide the handler later on since it's not started yet.
	srv = httptest.NewUnstartedServer(nil)
	// Our server url will be our oktaOrgURL.
	oktaOrgURL := "http://" + srv.Listener.Addr().String()
	authServerURL, err := OktaAuthServerURL(oktaOrgURL, oktaAuthServer)
	require.NoError(t, err)

	// Create a signed token and use the generated public key to create JWK.
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Create a JSON Web Key with a key id matching the tokens' kid.
	JWK := jwk.NewRSAPublicKey()
	require.NoError(t, JWK.Set(jwk.KeyIDKey, kid))
	require.NoError(t, JWK.Set(jwk.AlgorithmKey, jwtSigningAlgorithm))
	require.NoError(t, JWK.FromRaw(&rsaKey.PublicKey))
	// Create a JWK Set and add a single JWK.
	jwks := jwk.NewSet()
	jwks.Add(JWK)

	// Prepare the token.
	claims := jwt.MapClaims{
		"iss": authServerURL.String(),
		"cid": clientID,
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Add(-time.Hour).Unix(),
		"nbf": time.Now().Add(-time.Hour).Unix(),
		"m2mProfile": map[string]interface{}{
			"environment":  authServerURL.Host, // We're using the same server to serve responses for all endpoints.
			"organization": organization,
			"user":         "test@nobl9.com",
		},
	}
	jwtToken := jwt.NewWithClaims(jwt.GetSigningMethod(jwtSigningAlgorithm.String()), claims)
	jwtToken.Header["kid"] = kid
	token, err := jwtToken.SignedString(rsaKey)
	require.NoError(t, err)

	// Define the handler for test server.
	srv.Config = &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path[1:] { // Trim leading '/'
		case OktaTokenEndpoint(authServerURL).Path:
			assert.Equal(t,
				// Basic base64(clientID:clientSecret)
				"Basic "+base64.StdEncoding.EncodeToString([]byte(clientID+":"+clientSecret)),
				r.Header.Get(HeaderAuthorization))
			require.NoError(t, json.NewEncoder(w).Encode(oktaTokenResponse{AccessToken: token}))
		case OktaKeysEndpoint(authServerURL).Path:
			require.NoError(t, json.NewEncoder(w).Encode(jwks))
		case endpoint.Path:
			// Headers we always require.
			assert.Equal(t, organization, r.Header.Get(HeaderOrganization))
			assert.Equal(t, userAgent, r.Header.Get(HeaderUserAgent))
			assert.Equal(t, "Bearer "+token, r.Header.Get(HeaderAuthorization))
			// Endpoint specific tests.
			endpoint.TestRequestFunc(t, r)
			// Record response.
			endpoint.ResponseFunc(t, w)
		default:
			t.Logf("unsupported path: %s", r.URL.Path)
			t.FailNow()
		}
	})}

	// Prepare our client.
	oktaURL, err := OktaAuthServerURL(oktaOrgURL, oktaAuthServer)
	require.NoError(t, err)
	client, err = NewClientBuilder(userAgent).
		WithDefaultCredentials(clientID, clientSecret).
		WithOktaAuthServerURL(oktaURL).
		Build()
	require.NoError(t, err)

	return client, srv
}
