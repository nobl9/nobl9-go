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

	"github.com/nobl9/nobl9-go/internal/sdk"
)

func TestCredentials_SetAuthorizationHeader(t *testing.T) {
	t.Run("access token is not set, do not set the header", func(t *testing.T) {
		creds := &credentials{accessToken: ""}
		req := &http.Request{}
		creds.setAuthorizationHeader(req)
		assert.Empty(t, req.Header)
	})

	t.Run("set the header", func(t *testing.T) {
		creds := &credentials{accessToken: "123"}
		req := &http.Request{}
		creds.setAuthorizationHeader(req)
		require.Contains(t, req.Header, HeaderAuthorization)
		assert.Equal(t, "Bearer 123", req.Header.Get(HeaderAuthorization))
	})
}

func TestCredentials_RefreshAccessToken(t *testing.T) {
	t.Run("don't run in offline mode", func(t *testing.T) {
		creds := &credentials{config: &Config{DisableOkta: true, Organization: "my-org"}}
		tokenUpdated, err := creds.refreshAccessToken(context.Background())
		require.NoError(t, err)
		assert.False(t, tokenUpdated)
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
			creds := &credentials{
				config:        &Config{DisableOkta: false},
				claims:        test.Claims,
				tokenProvider: tokenProvider,
				tokenParser:   tokenParser,
			}
			_, err := creds.refreshAccessToken(context.Background())
			require.NoError(t, err)
			expectedCalledTimes := 0
			if test.TokenFetched {
				expectedCalledTimes = 1
			}
			assert.Equal(t, expectedCalledTimes, tokenProvider.calledTimes)
			assert.Equal(t, expectedCalledTimes, tokenParser.calledTimes)
		})
	}

	for name, test := range map[string]struct {
		ClientID     string
		ClientSecret string
		CalledTimes  int
	}{
		"clientID changed": {
			ClientID:     "new-id",
			ClientSecret: "old-secret",
			CalledTimes:  1,
		},
		"clientSecret changed": {
			ClientID:     "old-secret",
			ClientSecret: "new-secret",
			CalledTimes:  1,
		},
		"credentials did not change": {
			ClientID:     "old-id",
			ClientSecret: "old-secret",
			CalledTimes:  0,
		},
	} {
		t.Run(name, func(t *testing.T) {
			tokenProvider := &mockTokenProvider{}
			tokenParser := &mockTokenParser{}
			creds := &credentials{
				config: &Config{
					ClientID:     test.ClientID,
					ClientSecret: test.ClientSecret,
					DisableOkta:  false,
				},
				claims:        jwt.MapClaims{"exp": float64(time.Now().Add(time.Hour).Unix())},
				tokenProvider: tokenProvider,
				tokenParser:   tokenParser,
				clientID:      "old-id",
				clientSecret:  "old-secret",
			}
			_, err := creds.refreshAccessToken(context.Background())
			require.NoError(t, err)
			assert.Equal(t, test.CalledTimes, tokenProvider.calledTimes)
			assert.Equal(t, test.CalledTimes, tokenParser.calledTimes)
		})
	}

	t.Run("parse token when Config.AccessToken is set", func(t *testing.T) {
		tokenProvider := &mockTokenProvider{}
		tokenParser := &mockTokenParser{}
		creds := &credentials{
			config:        &Config{AccessToken: "token"},
			tokenProvider: tokenProvider,
			tokenParser:   tokenParser,
		}
		_, err := creds.refreshAccessToken(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 0, tokenProvider.calledTimes)
		assert.Equal(t, 1, tokenParser.calledTimes)
	})

	t.Run("do not parse token when Config.AccessToken is set, but token was already fetched", func(t *testing.T) {
		tokenProvider := &mockTokenProvider{}
		tokenParser := &mockTokenParser{}
		creds := &credentials{
			config:        &Config{AccessToken: "token"},
			tokenProvider: tokenProvider,
			tokenParser:   tokenParser,
			accessToken:   "already fetched",
		}
		_, err := creds.refreshAccessToken(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 1, tokenProvider.calledTimes)
		assert.Equal(t, 1, tokenParser.calledTimes)
	})

	t.Run("set new access token", func(t *testing.T) {
		tokenParser := &mockTokenParser{
			claims: jwt.MapClaims{
				"m2mProfile": map[string]interface{}{
					"user":         "test@user.com",
					"environment":  "app.nobl9.com",
					"organization": "my-org",
				},
			},
		}
		tokenProvider := &mockTokenProvider{}
		creds := &credentials{
			config:        &Config{AccessToken: "token"},
			tokenParser:   tokenParser,
			tokenProvider: tokenProvider,
		}
		updated, err := creds.refreshAccessToken(context.Background())
		require.NoError(t, err)
		assert.False(t, updated)
		assert.Equal(t, 0, tokenProvider.calledTimes)
		assert.Equal(t, 1, tokenParser.calledTimes)
		assert.Equal(t, "token", tokenParser.calledWithToken)
		assert.Equal(t, "token", creds.accessToken)
		assert.Equal(t, "app.nobl9.com", creds.environment)
		assert.Equal(t, "my-org", creds.organization)
		assert.Equal(t, tokenTypeM2M, creds.tokenType)
		assert.Equal(t, accessTokenM2MProfile{
			User:         "test@user.com",
			Environment:  "app.nobl9.com",
			Organization: "my-org",
		}, creds.m2mProfile)
		assert.Equal(t, tokenParser.claims, creds.claims)
	})

	t.Run("try setting new access token", func(t *testing.T) {
		tokenParser := &mockTokenParser{err: errors.New("token error")}
		tokenProvider := &mockTokenProvider{}
		creds := &credentials{
			config:        &Config{AccessToken: "token"},
			tokenParser:   tokenParser,
			tokenProvider: tokenProvider,
		}
		// Provider will always drop the error, but at least we make sure,
		// the provider is called once parser fails for the first time.
		_, err := creds.refreshAccessToken(context.Background())
		require.Error(t, err)
		assert.Equal(t, 1, tokenProvider.calledTimes)
		assert.Equal(t, 2, tokenParser.calledTimes)
	})

	t.Run("golden path, m2m token", func(t *testing.T) {
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
		creds := &credentials{
			config: &Config{
				ClientID:     "client-id",
				ClientSecret: "my-secret",
			},
			tokenProvider:   tokenProvider,
			tokenParser:     tokenParser,
			postRequestHook: func(token string) error { hookCalled = true; return nil },
		}
		tokenUpdated, err := creds.refreshAccessToken(context.Background())
		require.NoError(t, err)
		assert.True(t, tokenUpdated, "accessToken must be updated")
		assert.True(t, hookCalled, "postRequestHook must be called")
		assert.Equal(t, "my-secret", tokenProvider.calledWithClientSecret)
		assert.Equal(t, "client-id", tokenProvider.calledWithClientID)
		assert.Equal(t, "access-token", tokenParser.calledWithToken)
		assert.Equal(t, "client-id", tokenParser.calledWithClientID)
		assert.Equal(t, "access-token", creds.accessToken)
		assert.Equal(t, "app.nobl9.com", creds.environment)
		assert.Equal(t, "my-org", creds.organization)
		assert.Equal(t, tokenTypeM2M, creds.tokenType)
		assert.Equal(t, accessTokenM2MProfile{
			User:         "test@user.com",
			Environment:  "app.nobl9.com",
			Organization: "my-org",
		}, creds.m2mProfile)
		assert.Equal(t, tokenParser.claims, creds.claims)
	})

	t.Run("golden path, agent token", func(t *testing.T) {
		tokenProvider := &mockTokenProvider{
			token: "access-token",
		}
		tokenParser := &mockTokenParser{
			claims: jwt.MapClaims{
				"agentProfile": map[string]interface{}{
					"user":         "test@user.com",
					"environment":  "app.nobl9.com",
					"organization": "my-org",
					"name":         "my-agent",
					"project":      "default",
				},
			},
		}
		hookCalled := false
		creds := &credentials{
			config: &Config{
				ClientID:     "client-id",
				ClientSecret: "my-secret",
			},
			tokenProvider:   tokenProvider,
			tokenParser:     tokenParser,
			postRequestHook: func(token string) error { hookCalled = true; return nil },
		}
		tokenUpdated, err := creds.refreshAccessToken(context.Background())
		require.NoError(t, err)
		assert.True(t, tokenUpdated, "accessToken must be updated")
		assert.True(t, hookCalled, "postRequestHook must be called")
		assert.Equal(t, "my-secret", tokenProvider.calledWithClientSecret)
		assert.Equal(t, "client-id", tokenProvider.calledWithClientID)
		assert.Equal(t, "access-token", tokenParser.calledWithToken)
		assert.Equal(t, "client-id", tokenParser.calledWithClientID)
		assert.Equal(t, "access-token", creds.accessToken)
		assert.Equal(t, "app.nobl9.com", creds.environment)
		assert.Equal(t, "my-org", creds.organization)
		assert.Equal(t, tokenTypeAgent, creds.tokenType)
		assert.Equal(t, accessTokenAgentProfile{
			User:         "test@user.com",
			Environment:  "app.nobl9.com",
			Organization: "my-org",
			Name:         "my-agent",
			Project:      "default",
		}, creds.agentProfile)
		assert.Equal(t, tokenParser.claims, creds.claims)
	})
}

func TestCredentials_setNewToken(t *testing.T) {
	t.Run("don't call hook if parser fails", func(t *testing.T) {
		parserErr := errors.New("parser failed!")
		hookCalled := false
		creds := &credentials{
			config:          &Config{},
			tokenParser:     &mockTokenParser{err: parserErr},
			postRequestHook: func(token string) error { hookCalled = true; return nil },
		}
		err := creds.setNewToken("")
		require.Error(t, err)
		assert.Equal(t, parserErr, err)
		assert.False(t, hookCalled, "postRequestHook should not be called")
	})

	t.Run("don't call hook If we can't decode m2mProfile", func(t *testing.T) {
		hookCalled := false
		creds := &credentials{
			config:          &Config{},
			tokenParser:     &mockTokenParser{claims: jwt.MapClaims{"m2mProfile": "should be a map..."}},
			postRequestHook: func(token string) error { hookCalled = true; return nil },
		}
		err := creds.setNewToken("")
		assert.Error(t, err)
		assert.False(t, hookCalled, "postRequestHook should not be called")
	})

	t.Run("don't update credentials state if hook fails", func(t *testing.T) {
		tokenParser := &mockTokenParser{
			claims: jwt.MapClaims{},
		}
		hookErr := errors.New("hook failed!")
		creds := &credentials{
			config:          &Config{},
			tokenParser:     tokenParser,
			postRequestHook: func(token string) error { return hookErr },
		}
		err := creds.setNewToken("my-token")
		require.Error(t, err)
		assert.ErrorIs(t, err, hookErr)
		assert.Empty(t, creds.accessToken)
		assert.Empty(t, creds.m2mProfile)
		assert.Empty(t, creds.claims)
	})
}

// nolint: bodyclose
func TestClient_RoundTrip(t *testing.T) {
	t.Run("wrap errors with httpNonRetryableError", func(t *testing.T) {
		tokenProvider := &mockTokenProvider{err: errors.New("token fetching failed!")}
		creds := &credentials{
			config:        &Config{},
			tokenProvider: tokenProvider,
			tokenParser:   &mockTokenParser{},
		}

		req := &http.Request{}
		_, err := creds.RoundTrip(req)
		require.Error(t, err)
		_, isNonRetryableError := err.(sdk.HttpNonRetryableError)
		assert.True(t, isNonRetryableError, "err is of type httpNonRetryableError")
	})

	t.Run("set auth header if not present", func(t *testing.T) {
		creds := &credentials{
			config:        &Config{},
			accessToken:   "my-token",
			claims:        jwt.MapClaims{"exp": float64(time.Now().Add(time.Hour).Unix())},
			tokenProvider: &mockTokenProvider{},
			tokenParser:   &mockTokenParser{},
		}

		req := &http.Request{}
		_, _ = creds.RoundTrip(req)
		require.Contains(t, req.Header, HeaderAuthorization)
		assert.Equal(t, "Bearer my-token", req.Header.Get(HeaderAuthorization))
	})

	t.Run("update auth header if token was updated", func(t *testing.T) {
		creds := &credentials{
			config:        &Config{},
			accessToken:   "my-old-token",
			claims:        jwt.MapClaims{"exp": float64(time.Now().Unix())}, // expired
			tokenProvider: &mockTokenProvider{token: "my-new-token"},
			tokenParser:   &mockTokenParser{},
		}

		req := &http.Request{Header: http.Header{HeaderAuthorization: []string{"Bearer my-old-token"}}}
		_, _ = creds.RoundTrip(req)
		require.Contains(t, req.Header, HeaderAuthorization)
		assert.Equal(t, "Bearer my-new-token", req.Header.Get(HeaderAuthorization))
	})

	t.Run("don't update auth header if token was not updated", func(t *testing.T) {
		creds := &credentials{
			config:        &Config{},
			accessToken:   "my-old-token",
			claims:        jwt.MapClaims{"exp": float64(time.Now().Add(time.Hour).Unix())}, // not expired
			tokenProvider: &mockTokenProvider{token: "my-new-token"},
			tokenParser:   &mockTokenParser{},
		}

		req := &http.Request{Header: http.Header{HeaderAuthorization: []string{"Bearer my-old-token"}}}
		_, _ = creds.RoundTrip(req)
		require.Contains(t, req.Header, HeaderAuthorization)
		assert.Equal(t, "Bearer my-old-token", req.Header.Get(HeaderAuthorization))
	})
}

func TestCredentials_GetEnvironment(t *testing.T) {
	tokenProvider := &mockTokenProvider{}
	creds := &credentials{
		config:        &Config{},
		tokenProvider: tokenProvider,
		tokenParser: &mockTokenParser{
			claims: jwt.MapClaims{
				jwtTokenClaimM2MProfile: accessTokenM2MProfile{
					Environment: "my-env",
				},
				"exp": float64(time.Now().Add(time.Hour).Unix()),
			},
		},
	}

	env, err := creds.GetEnvironment(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, tokenProvider.calledTimes)
	assert.Equal(t, "my-env", env)

	// Make sure token is not requested again.
	env, err = creds.GetEnvironment(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, tokenProvider.calledTimes)
	assert.Equal(t, "my-env", env)
}

func TestCredentials_GetOrganization(t *testing.T) {
	tokenProvider := &mockTokenProvider{}
	creds := &credentials{
		config:        &Config{},
		tokenProvider: tokenProvider,
		tokenParser: &mockTokenParser{
			claims: jwt.MapClaims{
				jwtTokenClaimM2MProfile: accessTokenM2MProfile{
					Organization: "my-org",
				},
				"exp": float64(time.Now().Add(time.Hour).Unix()),
			},
		},
	}

	org, err := creds.GetOrganization(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, tokenProvider.calledTimes)
	assert.Equal(t, "my-org", org)

	// Make sure token is not requested again.
	org, err = creds.GetOrganization(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, tokenProvider.calledTimes)
	assert.Equal(t, "my-org", org)
}

type mockTokenProvider struct {
	calledTimes            int
	calledWithClientID     string
	calledWithClientSecret string

	token string
	err   error
}

func (m *mockTokenProvider) RequestAccessToken(
	_ context.Context,
	clientID, clientSecret string,
) (token string, err error) {
	m.calledTimes++
	m.calledWithClientID = clientID
	m.calledWithClientSecret = clientSecret
	return m.token, m.err
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
