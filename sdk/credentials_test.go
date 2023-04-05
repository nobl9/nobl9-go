package sdk

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentials_SetAuthorizationHeader(t *testing.T) {
	t.Run("access token is not set, do not set the header", func(t *testing.T) {
		creds := &Credentials{AccessToken: ""}
		req := &http.Request{}
		creds.SetAuthorizationHeader(req)
		assert.Empty(t, req.Header)
	})

	t.Run("set the header", func(t *testing.T) {
		creds := &Credentials{AccessToken: "123"}
		req := &http.Request{}
		creds.SetAuthorizationHeader(req)
		require.Contains(t, req.Header, HeaderAuthorization)
		assert.Equal(t, "Bearer 123", req.Header.Get(HeaderAuthorization))
	})
}
