package sdk

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
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

func TestCredentials_RefreshAccessToken(t *testing.T) {
	t.Run("don't run in offline mode", func(t *testing.T) {
		creds := &Credentials{offlineMode: true}
		err := creds.RefreshAccessToken(context.Background())
		require.NoError(t, err)
	})

	for name, test := range map[string]struct {
		// Make sure 'exp' claim is float64, otherwise claims.VerifyExpiresAt will always return false.
		Claims       jwt.MapClaims
		TokenFetched bool
	}{
		"request new token for the first time": {
			// Claims are not set, we need a new token.
			Claims:       nil,
			TokenFetched: true,
		},
		"refresh the token if it's expired": {
			// If the expiry is set to now, the offset should still catch it.
			Claims:       jwt.MapClaims{"exp": float64(time.Now().Add(tokenExpiryOffset).Unix())},
			TokenFetched: true,
		},
		"don't refresh the token if it's not expired": {
			Claims:       jwt.MapClaims{"exp": float64(time.Now().Add(time.Hour).Unix())},
			TokenFetched: false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			tokenProvider := &mockTokenProvider{}
			tokenParser := &mockTokenParser{}
			creds := &Credentials{
				claims:        test.Claims,
				TokenProvider: tokenProvider,
				TokenParser:   tokenParser,
			}
			err := creds.RefreshAccessToken(context.Background())
			require.NoError(t, err)
			expectedCalledTimes := 0
			if test.TokenFetched {
				expectedCalledTimes = 1
			}
			assert.Equal(t, expectedCalledTimes, tokenProvider.calledTimes)
			assert.Equal(t, expectedCalledTimes, tokenParser.calledTimes)
		})
	}

	t.Run("golden path", func(t *testing.T) {
		tokenProvider := &mockTokenProvider{
			token: "access-token",
		}
		tokenParser := &mockTokenParser{
			claims: jwt.MapClaims{
				"m2mProfile": map[string]interface{}{
					"user":         "test@user.com",
					"environment":  "app.nobl9.com",
					"organization": "my-org",
				},
			},
		}
		hookCalled := false
		creds := &Credentials{
			ClientID:        "client-id",
			ClientSecret:    "my-secret",
			TokenProvider:   tokenProvider,
			TokenParser:     tokenParser,
			PostRequestHook: func(token string) error { hookCalled = true; return nil },
		}
		err := creds.RefreshAccessToken(context.Background())
		require.NoError(t, err)
		assert.True(t, hookCalled, "PostRequestHook must be called")
		assert.Equal(t, "my-secret", tokenProvider.calledWithClientSecret)
		assert.Equal(t, "client-id", tokenProvider.calledWithClientID)
		assert.Equal(t, "access-token", tokenParser.calledWithToken)
		assert.Equal(t, "client-id", tokenParser.calledWithClientID)
		assert.Equal(t, "access-token", creds.AccessToken)
		assert.Equal(t, AccessTokenM2MProfile{
			User:         "test@user.com",
			Environment:  "app.nobl9.com",
			Organization: "my-org",
		}, creds.M2MProfile)
		assert.Equal(t, tokenParser.claims, creds.claims)
	})
}

func TestCredentials_SetAccessToken(t *testing.T) {
	tokenParser := &mockTokenParser{
		claims: jwt.MapClaims{
			"m2mProfile": map[string]interface{}{
				"user":         "test@user.com",
				"environment":  "app.nobl9.com",
				"organization": "my-org",
			},
		},
	}
	creds := &Credentials{
		ClientID:        "client-id",
		TokenParser:     tokenParser,
		PostRequestHook: func(token string) error { return errors.New("hook should not be called!") },
	}
	err := creds.SetAccessToken("access-token")
	require.NoError(t, err)
	assert.Equal(t, "access-token", tokenParser.calledWithToken)
	assert.Equal(t, "client-id", tokenParser.calledWithClientID)
	assert.Equal(t, "access-token", creds.AccessToken)
	assert.Equal(t, AccessTokenM2MProfile{
		User:         "test@user.com",
		Environment:  "app.nobl9.com",
		Organization: "my-org",
	}, creds.M2MProfile)
	assert.Equal(t, tokenParser.claims, creds.claims)
}

func TestCredentials_setNewToken(t *testing.T) {
	t.Run("don't call hook if parser fails", func(t *testing.T) {
		parserErr := errors.New("parser failed!")
		hookCalled := false
		creds := &Credentials{
			TokenParser:     &mockTokenParser{err: parserErr},
			PostRequestHook: func(token string) error { hookCalled = true; return nil },
		}
		err := creds.setNewToken("", true)
		require.Error(t, err)
		assert.Equal(t, parserErr, err)
		assert.False(t, hookCalled, "PostRequestHook should not be called")
	})

	t.Run("don't call hook If we can't decode m2mProfile", func(t *testing.T) {
		hookCalled := false
		creds := &Credentials{
			TokenParser:     &mockTokenParser{claims: jwt.MapClaims{"m2mProfile": "should be a map..."}},
			PostRequestHook: func(token string) error { hookCalled = true; return nil },
		}
		err := creds.setNewToken("", true)
		assert.Error(t, err)
		assert.False(t, hookCalled, "PostRequestHook should not be called")
	})

	t.Run("don't update Credentials state if hook fails", func(t *testing.T) {
		tokenParser := &mockTokenParser{
			claims: jwt.MapClaims{},
		}
		hookErr := errors.New("hook failed!")
		creds := &Credentials{
			TokenParser:     tokenParser,
			PostRequestHook: func(token string) error { return hookErr },
		}
		err := creds.setNewToken("my-token", true)
		require.Error(t, err)
		assert.ErrorIs(t, err, hookErr)
		assert.Empty(t, creds.AccessToken)
		assert.Empty(t, creds.M2MProfile)
		assert.Empty(t, creds.claims)
	})
}

type mockTokenProvider struct {
	calledTimes            int
	calledWithClientID     string
	calledWithClientSecret string

	token string
}

func (m *mockTokenProvider) RequestAccessToken(_ context.Context, clientID, clientSecret string) (token string, err error) {
	m.calledTimes++
	m.calledWithClientID = clientID
	m.calledWithClientSecret = clientSecret
	return m.token, nil
}

type mockTokenParser struct {
	calledTimes        int
	calledWithClientID string
	calledWithToken    string

	claims jwt.MapClaims
	err    error
}

func (m *mockTokenParser) Parse(token, clientID string) (jwt.MapClaims, error) {
	m.calledTimes++
	m.calledWithToken = token
	m.calledWithClientID = clientID
	return m.claims, m.err
}
