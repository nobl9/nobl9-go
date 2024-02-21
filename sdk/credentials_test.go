package sdk

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	refreshTests := map[string]struct {
		// Make sure 'exp' claim is float64, otherwise claims.VerifyExpiresAt will always return false.
		Claims       *jwtClaims
		TokenFetched bool
	}{
		"request new token for the first time": {
			// Claims are not set, we need a new token.
			Claims:       nil,
			TokenFetched: true,
		},
		"refresh the token if it's expired": {
			// If the expiry is set to now, the offset should still catch it.
			Claims: &jwtClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpiryOffset)),
				},
			},
			TokenFetched: true,
		},
		"don't refresh the token if it's not expired": {
			Claims: &jwtClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				},
			},
			TokenFetched: false,
		},
	}
	for name, test := range refreshTests {
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

	credentialsTests := map[string]struct {
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
	}
	for name, test := range credentialsTests {
		t.Run(name, func(t *testing.T) {
			tokenProvider := &mockTokenProvider{}
			tokenParser := &mockTokenParser{}
			creds := &credentials{
				config: &Config{
					ClientID:     test.ClientID,
					ClientSecret: test.ClientSecret,
					DisableOkta:  false,
				},
				claims: &jwtClaims{
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
					},
				},
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
			claims: jwtClaims{
				M2MProfile: stringOrObject[jwtClaimM2MProfile]{Value: &jwtClaimM2MProfile{
					User:         "test@user.com",
					Organization: "my-org",
					Environment:  "app.nobl9.com",
				}},
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
		assert.Equal(t, jwtClaimM2MProfile{
			User:         "test@user.com",
			Environment:  "app.nobl9.com",
			Organization: "my-org",
		}, *creds.claims.M2MProfile.Value)
		assert.Equal(t, tokenParser.claims, *creds.claims)
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
			claims: jwtClaims{
				M2MProfile: stringOrObject[jwtClaimM2MProfile]{Value: &jwtClaimM2MProfile{
					User:         "test@user.com",
					Organization: "my-org",
					Environment:  "app.nobl9.com",
				},
				}},
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
		assert.Equal(t, jwtClaimM2MProfile{
			User:         "test@user.com",
			Environment:  "app.nobl9.com",
			Organization: "my-org",
		}, *creds.claims.M2MProfile.Value)
		assert.Equal(t, tokenParser.claims, *creds.claims)
	})

	t.Run("golden path, agent token", func(t *testing.T) {
		tokenProvider := &mockTokenProvider{
			token: "access-token",
		}
		tokenParser := &mockTokenParser{
			claims: jwtClaims{
				AgentProfile: stringOrObject[jwtClaimAgentProfile]{Value: &jwtClaimAgentProfile{
					User:         "test@user.com",
					Organization: "my-org",
					Environment:  "app.nobl9.com",
					Name:         "my-agent",
					Project:      "default",
				}},
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
		assert.Equal(t, jwtClaimAgentProfile{
			User:         "test@user.com",
			Environment:  "app.nobl9.com",
			Organization: "my-org",
			Name:         "my-agent",
			Project:      "default",
		}, *creds.claims.AgentProfile.Value)
		assert.Equal(t, tokenParser.claims, *creds.claims)
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

	t.Run("don't update credentials state if hook fails", func(t *testing.T) {
		tokenParser := &mockTokenParser{
			claims: jwtClaims{},
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
		_, isNonRetryableError := err.(httpNonRetryableError)
		assert.True(t, isNonRetryableError, "err is of type httpNonRetryableError")
	})

	t.Run("set auth header if not present", func(t *testing.T) {
		creds := &credentials{
			config:      &Config{},
			accessToken: "my-token",
			claims: &jwtClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				},
			},
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
			config:      &Config{},
			accessToken: "my-old-token",
			claims: &jwtClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now()), // expired
				},
			},
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
			config:      &Config{},
			accessToken: "my-old-token",
			claims: &jwtClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)), // not expired
				},
			},
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
			claims: jwtClaims{
				M2MProfile: stringOrObject[jwtClaimM2MProfile]{Value: &jwtClaimM2MProfile{
					Environment: "my-env",
				}},
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				},
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
			claims: jwtClaims{
				M2MProfile: stringOrObject[jwtClaimM2MProfile]{Value: &jwtClaimM2MProfile{
					Organization: "my-org",
				}},
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				},
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

	claims jwtClaims
	err    error
}

func (m *mockTokenParser) Parse(token, clientID string) (*jwtClaims, error) {
	m.calledTimes++
	m.calledWithToken = token
	m.calledWithClientID = clientID
	return &m.claims, m.err
}
