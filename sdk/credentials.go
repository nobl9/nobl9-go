package sdk

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// accessTokenParser parses and verifies fetched access token.
type accessTokenParser interface {
	Parse(token, clientID string) (*jwtClaims, error)
}

// accessTokenProvider fetches the access token based on client it and client secret.
type accessTokenProvider interface {
	RequestAccessToken(ctx context.Context, clientID, clientSecret string) (token string, err error)
}

// accessTokenPostRequestHook is run whenever a new token request finishes successfully.
// It can be used, for example, to update persistent access token storage.
type accessTokenPostRequestHook = func(token string) error

func newCredentials(config *Config) *credentials {
	if config.DisableOkta {
		return &credentials{config: config}
	}
	authServerURL := oktaAuthServerURL(config.OktaOrgURL, config.OktaAuthServer)
	return &credentials{
		config: config,
		tokenParser: newJWTParser(
			authServerURL.String(),
			oktaKeysEndpoint(authServerURL).String()),
		tokenProvider: newOktaClient(func() string {
			return oktaTokenEndpoint(authServerURL).String()
		}),
		postRequestHook: config.saveAccessToken,
	}
}

// credentials stores and manages IDP app credentials and claims.
// It governs access token life cycle, providing means of refreshing it
// and exposing claims delivered with the token.
// Currently, the only supported IDP is Okta.
type credentials struct {
	config *Config
	// Set after the token is fetched.
	accessToken string
	// Extracted from claims.
	// organization the token belongs to.
	organization string
	// environment extracted from token claims which is the HTTP host of the Client requests.
	environment string
	// Claims.
	tokenType tokenType
	claims    *jwtClaims

	HTTP *http.Client
	// tokenParser is used to verify the token and its claims.
	tokenParser accessTokenParser
	// tokenProvider is used to provide an access token.
	tokenProvider accessTokenProvider
	// postRequestHook is not run in offline mode.
	postRequestHook accessTokenPostRequestHook

	// These are independent of Config.ClientID and Config.ClientSecret.
	// They are set just before the token is fetched.
	clientID     string
	clientSecret string

	mu sync.Mutex
}

// GetEnvironment first ensures a token has been parsed before returning the environment,
// as it is extracted from the token claims.
// credentials.environment should no tbe accessed directly, but rather through this method.
func (c *credentials) GetEnvironment(ctx context.Context) (string, error) {
	if _, err := c.refreshAccessToken(ctx); err != nil {
		return "", errors.Wrap(err, "failed to get environment")
	}
	return c.environment, nil
}

// GetOrganization first ensures a token has been parsed before returning the organization,
// as it is extracted from the token claims.
// credentials.organization should no tbe accessed directly, but rather through this method.
func (c *credentials) GetOrganization(ctx context.Context) (string, error) {
	if c.config.DisableOkta {
		return c.config.Organization, nil
	}

	if _, err := c.refreshAccessToken(ctx); err != nil {
		return "", errors.Wrap(err, "failed to get organization")
	}
	return c.organization, nil
}

// It's important for this to be clean client, request middleware in Go is kinda clunky
// and requires chaining multiple HTTP clients, timeouts and retries should be handled
// by the predecessors of this one.
var cleanCredentialsHTTPClient = &http.Client{}

// RoundTrip is responsible for making sure the access token is set and also update it
// if the expiry is imminent. It also sets the HeaderOrganization.
// It will wrap any errors returned from refreshAccessToken
// in retryhttp.httpNonRetryableError to ensure the request is not retried by the wrapping client.
func (c *credentials) RoundTrip(req *http.Request) (*http.Response, error) {
	tokenUpdated, err := c.refreshAccessToken(req.Context())
	if err != nil {
		return nil, httpNonRetryableError{Err: err}
	}
	if _, authHeaderSet := req.Header[HeaderAuthorization]; tokenUpdated || !authHeaderSet {
		c.setAuthorizationHeader(req)
	}
	return cleanCredentialsHTTPClient.Do(req)
}

// setAuthorizationHeader sets an authorization header which should be included
// if access token was set in request to the resource server.
func (c *credentials) setAuthorizationHeader(r *http.Request) {
	if c.accessToken == "" {
		return
	}
	if r.Header == nil {
		r.Header = http.Header{}
	}
	r.Header.Set(HeaderAuthorization, fmt.Sprintf("Bearer %s", c.accessToken))
}

// refreshAccessToken checks the accessToken expiry with an offset to detect if the token
// is soon to be expired. If so, it will request a new token and update the credentials state.
// If the token was not yet set, it will request a new one all the same.
func (c *credentials) refreshAccessToken(ctx context.Context) (updated bool, err error) {
	if c.config.DisableOkta {
		return false, nil
	}
	if !c.shouldRefresh() {
		return false, nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.shouldRefresh() {
		return false, nil
	}
	// Special case when we provide access token via Config.
	if c.config.AccessToken != "" && c.accessToken == "" {
		// If we didn't succeed, simply try refreshing the token.
		if err = c.setNewToken(c.config.AccessToken); err == nil {
			return false, nil
		}
	}
	if err = c.requestNewToken(ctx); err != nil {
		return false, err
	}
	return true, nil
}

// tokenExpiryOffset is added to the current time reading to make sure the token won't expiry before
// it reaches the API server.
const tokenExpiryOffset = 2 * time.Minute

// shouldRefresh defines token expiry policy for the JWT managed by credentials
// or if the config.ClientID or config.ClientSecret have been updated.
func (c *credentials) shouldRefresh() bool {
	return c.claims == nil ||
		c.claims.ExpiresAt.Before(time.Now().Add(tokenExpiryOffset)) ||
		c.clientID != c.config.ClientID ||
		c.clientSecret != c.config.ClientSecret
}

// requestNewToken uses tokenProvider to fetch the new token and parse it via setNewToken function.
func (c *credentials) requestNewToken(ctx context.Context) (err error) {
	c.clientID = c.config.ClientID
	c.clientSecret = c.config.ClientSecret
	token, err := c.tokenProvider.RequestAccessToken(ctx, c.config.ClientID, c.config.ClientSecret)
	if err != nil {
		return errors.Wrap(err, "error getting new access token from IDP")
	}
	return c.setNewToken(token)
}

// setNewToken parses and verifies the provided JWT using tokenParser.
// It will then decode 'm2mProfile' from the extracted claims and set
// the new values for m2mProfile, accessToken and claims credentials fields.
func (c *credentials) setNewToken(token string) error {
	claims, err := c.tokenParser.Parse(token, c.config.ClientID)
	if err != nil {
		return err
	}
	if c.postRequestHook != nil {
		if err = c.postRequestHook(token); err != nil {
			return errors.Wrap(err, "failed to execute access token post hook")
		}
	}
	// We can now update the token and it's claims.
	c.accessToken = token
	c.tokenType = claims.getTokenType()
	switch c.tokenType {
	case tokenTypeM2M:
		c.organization = claims.M2MProfile.Organization
		c.environment = claims.M2MProfile.Environment
	case tokenTypeAgent:
		c.organization = claims.AgentProfile.Organization
		c.environment = claims.AgentProfile.Environment
	}
	c.claims = claims
	return nil
}
