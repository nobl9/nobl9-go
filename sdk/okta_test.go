package sdk

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOktaClient_RequestAccessToken(t *testing.T) {
	t.Run("return error if client id or client secret are missing", func(t *testing.T) {
		okta := oktaClient{}
		_, err := okta.RequestAccessToken(context.Background(), "123", "")
		require.Error(t, err)
		assert.Equal(t, errMissingClientCredentials, err)
		_, err = okta.RequestAccessToken(context.Background(), "", "secret")
		require.Error(t, err)
		assert.Equal(t, errMissingClientCredentials, err)
	})

	t.Run("handle context cancellation", func(t *testing.T) {
		okta := oktaClient{HTTP: new(http.Client), getTokenEndpoint: func() string { return "https://test.com/api" }}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := okta.RequestAccessToken(ctx, "123", "secret")
		assert.ErrorIs(t, err, context.Canceled)
	})

	var (
		respondWithStatusCode int
		respondWithPayload    []byte
		requestBody           []byte
		requestHeaders        http.Header
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestBody, _ = io.ReadAll(r.Body)
		requestHeaders = r.Header
		if respondWithPayload != nil {
			_, _ = w.Write(respondWithPayload)
		}
		w.WriteHeader(respondWithStatusCode)
	}))
	defer srv.Close()
	u, err := url.Parse(srv.URL)
	require.NoError(t, err)
	okta := oktaClient{HTTP: new(http.Client), getTokenEndpoint: func() string { return u.String() }}

	t.Run("return error for invalid status codes", func(t *testing.T) {
		for _, respondWithStatusCode = range []int{401, 409, 500, 300} {
			_, err := okta.RequestAccessToken(context.Background(), "123", "secret")
			require.Error(t, err)
			assert.ErrorContains(t, err, fmt.Sprintf("status: %d", respondWithStatusCode))
		}
	})

	t.Run("golden path", func(t *testing.T) {
		respondWithStatusCode = http.StatusOK
		respondWithPayload, _ = json.Marshal(oktaTokenResponse{AccessToken: "access-token"})

		token, err := okta.RequestAccessToken(context.Background(), "123", "secret")
		require.NoError(t, err)
		assert.Equal(t, "access-token", token)
		auth, err := io.ReadAll(base64.NewDecoder(base64.StdEncoding,
			strings.NewReader(strings.Split(requestHeaders.Get("Authorization"), " ")[1])))
		require.NoError(t, err)
		assert.Equal(t, "123:secret", string(auth))
		assert.Equal(t, oktaHeaderContentType, requestHeaders.Get("Content-Type"))
		assert.Equal(t, "grant_type=client_credentials&scope=m2m", string(requestBody))
	})
}
