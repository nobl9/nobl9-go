package sdk

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestClient_CreateRequest(t *testing.T) {
	client, srv := prepareTestClient(t, endpointConfig{})
	defer srv.Close()

	t.Run("all parameters", func(t *testing.T) {
		values := url.Values{"name": []string{"this"}, "team": []string{"green"}}
		req, err := client.CreateRequest(
			context.Background(),
			http.MethodGet,
			"/test",
			http.Header{HeaderProject: []string{"my-project"}},
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
			http.Header{HeaderProject: []string{"my-project"}},
			nil,
			nil,
		)
		require.NoError(t, err)
		assert.Empty(t, req.URL.Query())
		assert.Empty(t, req.Body)
		assert.Equal(t, "my-project", req.Header.Get(HeaderProject))
	})

	t.Run("no project header, use default", func(t *testing.T) {
		req, err := client.CreateRequest(
			context.Background(),
			http.MethodGet,
			"/test",
			nil,
			nil,
			nil,
		)
		require.NoError(t, err)
		assert.Equal(t, client.Config.Project, req.Header.Get(HeaderProject))
	})
}

func TestDefaultUserAgent(t *testing.T) {
	getStderrFromExec := func(err error) string {
		if v, ok := err.(*exec.ExitError); ok {
			return string(v.Stderr)
		}
		return ""
	}
	tempDir := t.TempDir()
	binName := "test-binary"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	path := filepath.Join(tempDir, binName)
	// Build binary. This is the only way for debug package to work,
	// it needs to operate on a binary built from a module.
	_, err := exec.Command(
		"go",
		"build",
		"-o", path,
		filepath.FromSlash("./test_data/client/simple_module/main.go"),
	).Output()
	require.NoError(t, err, getStderrFromExec(err))
	// Execute the binary.
	out, err := exec.Command(path).Output()
	require.NoError(t, err, getStderrFromExec(err))
	assert.Contains(t, string(out), "sdk/(devel)")
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
	oktaOrgURL := &url.URL{Scheme: "http", Host: srv.Listener.Addr().String()}
	authServerURL := oktaAuthServerURL(oktaOrgURL, oktaAuthServer)

	// Create a signed token and use the generated public key to create JWK.
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Create a JSON Web Key with a key id matching the tokens' kid.
	jwk, err := jwkset.NewJWKFromKey(rsaKey, jwkset.JWKOptions{
		Metadata: jwkset.JWKMetadataOptions{
			KID: kid,
		},
	})
	require.NoError(t, err)
	jwks := jwkset.JWKSMarshal{Keys: []jwkset.JWKMarshal{jwk.Marshal()}}

	// Prepare the token.
	claims := jwt.MapClaims{
		"iss": authServerURL.String(),
		"cid": clientID,
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Add(-time.Hour).Unix(),
		"nbf": time.Now().Add(-time.Hour).Unix(),
		"m2mProfile": map[string]any{
			"environment":  authServerURL.Host, // We're using the same server to serve responses for all endpoints.
			"organization": organization,
			"user":         "test@nobl9.com",
		},
	}
	jwtToken := jwt.NewWithClaims(jwtSigningAlgorithm, claims)
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

	// Start test server.
	srv.Start()

	// Prepare client.
	config, err := ReadConfig(
		ConfigOptionWithCredentials(clientID, clientSecret),
		ConfigOptionNoConfigFile())
	require.NoError(t, err)
	config.OktaOrgURL = oktaOrgURL
	config.OktaAuthServer = oktaAuthServer
	client, err = NewClient(config)
	require.NoError(t, err)
	client.SetUserAgent(userAgent)

	return client, srv
}
