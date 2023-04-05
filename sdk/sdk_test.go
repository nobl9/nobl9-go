package sdk

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetObject(t *testing.T) {
	urlScheme = "http"
	const (
		oktaAuthServer = "ausdh151kj9OOWv5x191"
		kid            = "my-kid"
		clientID       = "client-id"
		organization   = "my-org"
		project        = "non-default"
		userAgent      = "sloctl"
	)
	// Declare the test server, we can provide the handler later on since it's not started yet.
	srv := httptest.NewUnstartedServer(nil)
	// Our server url will be our oktaOrgURL.
	oktaOrgURL := "http://" + srv.Listener.Addr().String()
	authServerURL, err := OktaAuthServer(oktaOrgURL, oktaAuthServer)
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

	serviceResponse := []AnyJSONObj{
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

	var recordedRequest *http.Request
	// Define the handler for test server.
	srv.Config = &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path[1:] /* Trim leading '/' */ {
		case OktaTokenEndpoint(authServerURL).Path:
			require.NoError(t, json.NewEncoder(w).Encode(oktaTokenResponse{AccessToken: token}))
		case OktaKeysEndpoint(authServerURL).Path:
			require.NoError(t, json.NewEncoder(w).Encode(jwks))
		case "api/get/service":
			recordedRequest = r
			require.NoError(t, json.NewEncoder(w).Encode(serviceResponse))
		default:
			panic(fmt.Sprintf("unsupported path: %s", r.URL.Path))
		}
	})}

	// Prepare our client.
	client, err := DefaultClient(clientID, "super-secret", oktaOrgURL, oktaAuthServer, userAgent)
	require.NoError(t, err)

	// Start the test server.
	srv.Start()
	defer srv.Close()

	objects, err := client.GetObject(
		context.Background(),
		project,
		ObjectService,
		"2023-01-01T15:30:27Z",
		map[string][]string{"team": {"green", "purple"}},
		"service1", "service2",
	)
	// Verify response handling.
	require.NoError(t, err)
	require.Len(t, objects, 2)
	assert.Equal(t, serviceResponse, objects)
	// Verify request parameters.
	assert.Equal(t, organization, recordedRequest.Header.Get(HeaderOrganization))
	assert.Equal(t, project, recordedRequest.Header.Get(HeaderProject))
	assert.Equal(t, userAgent, recordedRequest.Header.Get(HeaderUserAgent))
	assert.Equal(t, "Bearer "+token, recordedRequest.Header.Get(HeaderAuthorization))
	assert.Equal(t, url.Values{
		QueryKeyName:         []string{"service1", "service2"},
		QueryKeyTime:         []string{"2023-01-01T15:30:27Z"},
		QueryKeyLabelsFilter: []string{"team:green,team:purple"},
	}, recordedRequest.URL.Query())
}
