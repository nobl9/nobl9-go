package sdk

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/sdk/definitions"
)

func TestClient_GetObjects(t *testing.T) {
	responsePayload := []manifest.Object{
		v1alpha.Service{
			APIVersion: v1alpha.APIVersion,
			Kind:       manifest.KindService,
			Metadata: v1alpha.ServiceMetadata{
				Name:    "service1",
				Project: "default",
			},
		},
		v1alpha.Service{
			APIVersion: v1alpha.APIVersion,
			Kind:       manifest.KindService,
			Metadata: v1alpha.ServiceMetadata{
				Name:    "service2",
				Project: "default",
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

func TestClient_GetObjects_NoObjectsInResponse(t *testing.T) {
	responsePayload := make([]manifest.Object, 0)

	client, srv := prepareTestClient(t, endpointConfig{
		// Define endpoint response.
		Path: "api/get/service",
		ResponseFunc: func(t *testing.T, w http.ResponseWriter) {
			require.NoError(t, json.NewEncoder(w).Encode(responsePayload))
		},
	})

	// Start and close the test server.
	srv.Start()
	defer srv.Close()

	// Run the API method.
	objects, err := client.GetObjects(
		context.Background(),
		ProjectsWildcard,
		manifest.KindService,
		nil,
		"service1",
	)
	// Verify response handling.
	require.NoError(t, err)
	require.Len(t, objects, 0)
}

func TestClient_GetObjects_UserGroupsEndpoint(t *testing.T) {
	responsePayload := []manifest.Object{
		v1alpha.UserGroup{
			APIVersion: v1alpha.APIVersion,
			Kind:       manifest.KindService,
			Metadata: v1alpha.UserGroupMetadata{
				Name: "service1",
			},
		},
	}

	calledTimes := 0
	client, srv := prepareTestClient(t, endpointConfig{
		// Define endpoint response.
		Path: "api/usrmgmt/groups",
		ResponseFunc: func(t *testing.T, w http.ResponseWriter) {
			require.NoError(t, json.NewEncoder(w).Encode(responsePayload))
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
	requestPayload := []manifest.Object{
		v1alpha.Service{
			APIVersion: v1alpha.APIVersion,
			Kind:       manifest.KindService,
			Metadata: v1alpha.ServiceMetadata{
				Name:    "service1",
				Project: "default",
			},
		},
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
			assert.Equal(t, "", r.Header.Get(HeaderProject))
			assert.Equal(t, url.Values{QueryKeyDryRun: {"true"}}, r.URL.Query())
			objects, err := definitions.ReadSources(context.Background(), definitions.NewReaderSource(r.Body, ""))
			require.NoError(t, err)
			assert.Equal(t, expected, objects)
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
	requestPayload := []manifest.Object{
		v1alpha.Service{
			APIVersion: v1alpha.APIVersion,
			Kind:       manifest.KindService,
			Metadata: v1alpha.ServiceMetadata{
				Name:    "service1",
				Project: "default",
			},
		},
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
			assert.Equal(t, "", r.Header.Get(HeaderProject))
			assert.Equal(t, url.Values{QueryKeyDryRun: {"true"}}, r.URL.Query())
			objects, err := definitions.ReadSources(context.Background(), definitions.NewReaderSource(r.Body, ""))
			require.NoError(t, err)
			assert.Equal(t, expected, objects)
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

func TestCreateRequest(t *testing.T) {
	client, srv := prepareTestClient(t, endpointConfig{})

	// Start and close the test server.
	srv.Start()
	defer srv.Close()

	t.Run("all parameters", func(t *testing.T) {
		values := url.Values{"name": []string{"this"}, "team": []string{"green"}}
		req, err := client.CreateRequest(
			context.Background(),
			http.MethodGet,
			"/test",
			"my-project",
			values,
			bytes.NewBufferString("foo"),
		)
		require.NoError(t, err)
		assert.Equal(t, "/api/test", req.URL.Path)
		assert.Equal(t, http.Header{
			HeaderOrganization: []string{"my-org"},
			HeaderProject:      []string{"my-project"},
			HeaderUserAgent:    []string{"sloctl"},
		}, req.Header)
		// If client.refreshAccessTokenOnce was not executed, the host wouldn't have been set.
		assert.Contains(t, srv.URL, req.URL.Host)
		assert.Equal(t, values, req.URL.Query())
		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		assert.Equal(t, "foo", string(body))
	})

	t.Run("no body or values", func(t *testing.T) {
		req, err := client.CreateRequest(
			context.Background(),
			http.MethodGet,
			"/test",
			"my-project",
			nil,
			nil,
		)
		require.NoError(t, err)
		assert.Empty(t, req.URL.Query())
		assert.Empty(t, req.Body)
	})

	t.Run("no project", func(t *testing.T) {
		req, err := client.CreateRequest(
			context.Background(),
			http.MethodGet,
			"/test",
			"",
			nil,
			nil,
		)
		require.NoError(t, err)
		assert.NotContains(t, req.Header, HeaderProject)
	})
}

func TestProcessResponseErrors(t *testing.T) {
	t.Parallel()
	c := Client{}

	t.Run("status code smaller than 300, no error", func(t *testing.T) {
		t.Parallel()
		for code := 200; code < 300; code++ {
			require.NoError(t, c.processResponseErrors(&http.Response{StatusCode: code}))
		}
	})

	t.Run("status code between 300 and 399", func(t *testing.T) {
		t.Parallel()
		for code := 300; code < 400; code++ {
			err := c.processResponseErrors(&http.Response{
				StatusCode: code,
				Body:       io.NopCloser(bytes.NewBufferString("error!"))})
			require.Error(t, err)
			require.EqualError(t, err, fmt.Sprintf("bad status code response: %d, body: error!", code))
		}
	})

	t.Run("user errors", func(t *testing.T) {
		t.Parallel()
		for code := 400; code < 500; code++ {
			err := c.processResponseErrors(&http.Response{
				StatusCode: code,
				Body:       io.NopCloser(bytes.NewBufferString("error!"))})
			require.Error(t, err)
			require.EqualError(t, err, "error!")
		}
	})

	t.Run("server errors", func(t *testing.T) {
		t.Parallel()
		for code := 500; code < 600; code++ {
			err := c.processResponseErrors(&http.Response{
				StatusCode: code,
				Header:     http.Header{HeaderTraceID: []string{"123"}},
				Body:       io.NopCloser(bytes.NewBufferString("error!"))})
			require.Error(t, err)
			require.EqualError(t,
				err,
				fmt.Sprintf("%s error message: error! error id: 123", http.StatusText(code)))
		}
	})
}

// TODO: Once the new tag is released, convert change the simple_module go.mod to point at concrete SDK version.
func TestDefaultUserAgent(t *testing.T) {
	getStderrFromExec := func(err error) string {
		if v, ok := err.(*exec.ExitError); ok {
			return string(v.Stderr)
		}
		return ""
	}
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "test-binary")
	// Build binary. This is the only way for debug package to work,
	// it needs to operate on a binary built from a module.
	_, err := exec.Command("go", "build", "-o", path, "./test_data/client/simple_module/main.go").Output()
	require.NoError(t, err, getStderrFromExec(err))
	// Execute the binary.
	out, err := exec.Command(path).Output()
	require.NoError(t, err, getStderrFromExec(err))
	assert.Contains(t, string(out), "sdk/(devel)")
}

type endpointConfig struct {
	Path            string
	ResponseFunc    func(t *testing.T, w http.ResponseWriter)
	TestRequestFunc func(*testing.T, *http.Request)
}

func addOrganization(objects []manifest.Object, org string) []manifest.Object {
	result := make([]manifest.Object, 0, len(objects))
	for _, obj := range objects {
		if objCtx, ok := obj.(v1alpha.ObjectContext); ok {
			result = append(result, objCtx.SetOrganization(org))
		}
	}
	return result
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
	oktaOrgURL := &url.URL{Scheme: "http", Host: srv.Listener.Addr().String()}
	authServerURL := oktaAuthServerURL(oktaOrgURL, oktaAuthServer)

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
		case oktaTokenEndpoint(authServerURL).Path:
			assert.Equal(t,
				// Basic base64(clientID:clientSecret)
				"Basic "+base64.StdEncoding.EncodeToString([]byte(clientID+":"+clientSecret)),
				r.Header.Get(HeaderAuthorization))
			require.NoError(t, json.NewEncoder(w).Encode(oktaTokenResponse{AccessToken: token}))
		case oktaKeysEndpoint(authServerURL).Path:
			require.NoError(t, json.NewEncoder(w).Encode(jwks))
		case endpoint.Path:
			// Headers we always require.
			assert.Equal(t, organization, r.Header.Get(HeaderOrganization))
			assert.Equal(t, userAgent, r.Header.Get(HeaderUserAgent))
			assert.Equal(t, "Bearer "+token, r.Header.Get(HeaderAuthorization))
			// Endpoint specific tests.
			if endpoint.TestRequestFunc != nil {
				endpoint.TestRequestFunc(t, r)
			}
			// Record response.
			endpoint.ResponseFunc(t, w)
		default:
			t.Logf("unsupported path: %s", r.URL.Path)
			t.FailNow()
		}
	})}

	// Prepare client.
	config, err := ReadConfig(
		ConfigOptionWithCredentials(clientID, clientSecret),
		ConfigOptionNoConfigFile())
	require.NoError(t, err)
	config.OktaOrgURL = oktaOrgURL
	config.OktaAuthServer = oktaAuthServer
	client, err = NewClientBuilder(config).WithUserAgent(userAgent).Build()
	require.NoError(t, err)

	return client, srv
}
