package sdk

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/sdk/retryhttp"
)

// accessTokenParser parses and verifies fetched access token.
type accessTokenParser interface {
	Parse(token, clientID string) (jwt.MapClaims, error)
}

// accessTokenProvider fetches the access token based on client it and client secret.
type accessTokenProvider interface {
	RequestAccessToken(ctx context.Context, clientID, clientSecret string) (token string, err error)
}

// AccessTokenPostRequestHook is run whenever a new token request finishes successfully.
// It can be used, for example, to update persistent access token storage.
type AccessTokenPostRequestHook = func(token string) error

// accessTokenM2MProfile stores information specific to an Okta M2M application.
type accessTokenM2MProfile struct {
	User         string `json:"user"`
	Organization string `json:"Organization"`
	Environment  string `json:"environment"`
}

// accessTokenAgentProfile stores information specific to an Okta Agent application.
type accessTokenAgentProfile struct {
	User         string `json:"user"`
	Organization string `json:"Organization"`
	Environment  string `json:"environment"`
	Name         string `json:"name"`
	Project      string `json:"project"`
}

func newCredentials(config *Config) (*credentials, error) {
	parser, err := newJWTParser(
		func() string {
			return oktaAuthServerURL(config.OktaOrgURL, config.OktaAuthServer).String()
		},
		func() string {
			return oktaKeysEndpoint(oktaAuthServerURL(config.OktaOrgURL, config.OktaAuthServer)).String()
		})
	if err != nil {
		return nil, err
	}
	return &credentials{
		config:      config,
		tokenParser: parser,
		tokenProvider: newOktaClient(func() string {
			return oktaTokenEndpoint(oktaAuthServerURL(config.OktaOrgURL, config.OktaAuthServer)).String()
		}),
	}, nil
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
	// Organization and Environment, if accessed before the first request
	// is executed, will be empty as the token was not yet fetched.
	// To force them to be set earlier you could provide the access token
	// to credentials or call RefreshAccessToken manually.
	Organization string
	Environment  string
	// Claims.
	m2mProfile   accessTokenM2MProfile
	agentProfile accessTokenAgentProfile
	tokenType    tokenType
	claims       jwt.MapClaims

	HTTP *http.Client
	// tokenParser is used to verify the token and its claims.
	tokenParser accessTokenParser
	// tokenProvider is used to provide an access token.
	tokenProvider accessTokenProvider
	// PostRequestHook is not run in offline mode.
	PostRequestHook AccessTokenPostRequestHook

	mu sync.Mutex
}

// It's important for this to be clean client, request middleware in Go is kinda clunky
// and requires chaining multiple http clients, timeouts and retries should be handled
// by the predecessors of this one.
var cleanCredentialsHTTPClient = &http.Client{}

// RoundTrip is responsible for making sure the access token is set and also update it
// if the expiry is imminent. It also sets the HeaderOrganization.
// It will wrap any errors returned from RefreshAccessToken
// in retryhttp.NonRetryableError to ensure the request is not retried by the wrapping client.
func (c *credentials) RoundTrip(req *http.Request) (*http.Response, error) {
	tokenUpdated, err := c.RefreshAccessToken(req.Context())
	if err != nil {
		return nil, retryhttp.NonRetryableError{Err: err}
	}
	if _, authHeaderSet := req.Header[HeaderAuthorization]; tokenUpdated || !authHeaderSet {
		c.SetAuthorizationHeader(req)
	}
	return cleanCredentialsHTTPClient.Do(req)
}

// SetAuthorizationHeader sets an authorization header which should be included
// if access token was set in request to the resource server.
func (c *credentials) SetAuthorizationHeader(r *http.Request) {
	if c.accessToken == "" {
		return
	}
	if r.Header == nil {
		r.Header = http.Header{}
	}
	r.Header.Set(HeaderAuthorization, fmt.Sprintf("Bearer %s", c.accessToken))
}

// SetAccessToken allows setting new access token without using tokenProvider.
// The provided token will be still parsed using setNewToken function.
// In offline mode this is a noop.
func (c *credentials) SetAccessToken(token string) error {
	if c.config.DisableOkta {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.setNewToken(token, false)
}

// RefreshAccessToken checks the accessToken expiry with an offset to detect if the token
// is soon to be expired. If so, it wll request a new token and update the credentials state.
// If the token was not yet set, it will request a new one all the same.
func (c *credentials) RefreshAccessToken(ctx context.Context) (updated bool, err error) {
	if c.config.DisableOkta {
		return
	}
	if !c.shouldRefresh() {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.shouldRefresh() {
		return
	}
	if err = c.requestNewToken(ctx); err == nil {
		updated = true
	}
	return
}

// tokenExpiryOffset is added to the current time reading to make sure the token won't expiry before
// it reaches the API server.
const tokenExpiryOffset = 2 * time.Minute

// shouldRefresh defines token expiry policy for the JWT managed by credentials.
func (c *credentials) shouldRefresh() bool {
	return len(c.claims) == 0 || !c.claims.VerifyExpiresAt(time.Now().Add(tokenExpiryOffset).Unix(), true)
}

// requestNewToken uses tokenProvider to fetch the new token and parse it via setNewToken function.
func (c *credentials) requestNewToken(ctx context.Context) (err error) {
	token, err := c.tokenProvider.RequestAccessToken(ctx, c.config.ClientID, c.config.ClientSecret)
	if err != nil {
		return errors.Wrap(err, "error getting new access token from IDP")
	}
	return c.setNewToken(token, true)
}

// setNewToken parses and verifies the provided JWT using tokenParser.
// It will then decode 'm2mProfile' from the extracted claims and set
// the new values for m2mProfile, accessToken and claims credentials fields.
func (c *credentials) setNewToken(token string, withHook bool) error {
	claims, err := c.tokenParser.Parse(token, c.config.ClientID)
	if err != nil {
		return err
	}
	var (
		m2mProfile   accessTokenM2MProfile
		agentProfile accessTokenAgentProfile
	)
	tokenTyp := tokenTypeFromClaims(claims)
	switch tokenTyp {
	case tokenTypeM2M:
		m2mProfile, err = m2mProfileFromClaims(claims)
		if err != nil {
			return errors.Wrap(err, "failed to decode JWT claims to m2m profile object")
		}
	case tokenTypeAgent:
		agentProfile, err = agentProfileFromClaims(claims)
		if err != nil {
			return errors.Wrap(err, "failed to decode JWT claims to agent profile object")
		}
	}
	if withHook && c.PostRequestHook != nil {
		if err = c.PostRequestHook(token); err != nil {
			return errors.Wrap(err, "failed to execute access token post hook")
		}
	}
	// We can now update the token and it's claims.
	c.accessToken = token
	switch tokenTyp {
	case tokenTypeM2M:
		c.Organization = m2mProfile.Organization
		c.Environment = m2mProfile.Environment
	case tokenTypeAgent:
		c.Organization = agentProfile.Organization
		c.Environment = agentProfile.Environment
	}
	c.tokenType = tokenTyp
	c.m2mProfile = m2mProfile
	c.agentProfile = agentProfile
	c.claims = claims
	return nil
}
