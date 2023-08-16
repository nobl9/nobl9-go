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

// AccessTokenParser parses and verifies fetched access token.
type AccessTokenParser interface {
	Parse(token, clientID string) (jwt.MapClaims, error)
}

// AccessTokenProvider fetches the access token based on client it and client secret.
type AccessTokenProvider interface {
	RequestAccessToken(ctx context.Context, clientID, clientSecret string) (token string, err error)
}

// AccessTokenPostRequestHook is run whenever a new token request finishes successfully.
// It can be used, for example, to update persistent access token storage.
type AccessTokenPostRequestHook = func(token string) error

// accessTokenM2MProfile stores information specific to an Okta M2M application.
type accessTokenM2MProfile struct {
	User         string `json:"user"`
	Organization string `json:"organization"`
	Environment  string `json:"environment"`
}

// accessTokenAgentProfile stores information specific to an Okta Agent application.
type accessTokenAgentProfile struct {
	User         string `json:"user"`
	Organization string `json:"organization"`
	Environment  string `json:"environment"`
	Name         string `json:"name"`
	Project      string `json:"project"`
}

func newCredentials(config *Config) (*Credentials, error) {
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
	return &Credentials{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenParser:  parser,
		TokenProvider: newOktaClient(func() string {
			return oktaAuthServerURL(config.OktaOrgURL, config.OktaAuthServer).String()
		}),
	}, nil
}

// Credentials stores and manages IDP app credentials and claims.
// It governs access token life cycle, providing means of refreshing it
// and exposing claims delivered with the token.
// Currently, the only supported IDP is Okta.
type Credentials struct {
	// Required to fetch the token.
	ClientID     string
	ClientSecret string

	// Set after the token is fetched.
	AccessToken string
	// Extracted from claims.
	// Organization and Environment, if accessed before the first request
	// is executed, will be empty as the token was not yet fetched.
	// To force them to be set earlier you could provide the access token
	// to Credentials or call Credentials.RefreshAccessToken manually.
	Organization string
	Environment  string
	// Claims.
	m2mProfile   accessTokenM2MProfile
	agentProfile accessTokenAgentProfile
	tokenType    tokenType
	claims       jwt.MapClaims

	HTTP *http.Client
	// TokenParser is used to verify the token and its claims.
	TokenParser AccessTokenParser
	// TokenProvider is used to provide an access token.
	TokenProvider AccessTokenProvider
	// PostRequestHook is not run in offline mode.
	PostRequestHook AccessTokenPostRequestHook

	// offlineMode turns Credentials.RefreshAccessToken into a noop.
	offlineMode bool
	mu          sync.Mutex
}

// It's important for this to be clean client, request middleware in Go is kinda clunky
// and requires chaining multiple http clients, timeouts and retries should be handled
// by the predecessors of this one.
var credentialsCleanHTTPClient = &http.Client{}

// RoundTrip is responsible for making sure the access token is set and also update it
// if the expiry is imminent. It also sets the HeaderOrganization.
// It will wrap any errors returned from RefreshAccessToken
// in retryhttp.NonRetryableError to ensure the request is not retried by the wrapping client.
func (creds *Credentials) RoundTrip(req *http.Request) (*http.Response, error) {
	tokenUpdated, err := creds.RefreshAccessToken(req.Context())
	if err != nil {
		return nil, retryhttp.NonRetryableError{Err: err}
	}
	if _, authHeaderSet := req.Header[HeaderAuthorization]; tokenUpdated || !authHeaderSet {
		creds.SetAuthorizationHeader(req)
	}
	return credentialsCleanHTTPClient.Do(req)
}

// SetOfflineMode turns RefreshAccessToken into a noop.
func (creds *Credentials) SetOfflineMode() {
	creds.offlineMode = true
}

// SetAuthorizationHeader sets an authorization header which should be included
// if access token was set in request to the resource server.
func (creds *Credentials) SetAuthorizationHeader(r *http.Request) {
	if creds.AccessToken == "" {
		return
	}
	if r.Header == nil {
		r.Header = http.Header{}
	}
	r.Header.Set(HeaderAuthorization, fmt.Sprintf("Bearer %s", creds.AccessToken))
}

// SetAccessToken allows setting new access token without using TokenProvider.
// The provided token will be still parsed using setNewToken function.
// In offline mode this is a noop.
func (creds *Credentials) SetAccessToken(token string) error {
	if creds.offlineMode {
		return nil
	}
	creds.mu.Lock()
	defer creds.mu.Unlock()
	return creds.setNewToken(token, false)
}

// RefreshAccessToken checks the AccessToken expiry with an offset to detect if the token
// is soon to be expired. If so, it wll request a new token and update the Credentials state.
// If the token was not yet set, it will request a new one all the same.
func (creds *Credentials) RefreshAccessToken(ctx context.Context) (updated bool, err error) {
	if creds.offlineMode {
		return
	}
	if !creds.shouldRefresh() {
		return
	}
	creds.mu.Lock()
	defer creds.mu.Unlock()
	if !creds.shouldRefresh() {
		return
	}
	if err = creds.requestNewToken(ctx); err == nil {
		updated = true
	}
	return
}

// tokenExpiryOffset is added to the current time reading to make sure the token won't expiry before
// it reaches the API server.
const tokenExpiryOffset = 2 * time.Minute

// shouldRefresh defines token expiry policy for the JWT managed by Credentials.
func (creds *Credentials) shouldRefresh() bool {
	return len(creds.claims) == 0 || !creds.claims.VerifyExpiresAt(time.Now().Add(tokenExpiryOffset).Unix(), true)
}

// requestNewToken uses TokenProvider to fetch the new token and parse it via setNewToken function.
func (creds *Credentials) requestNewToken(ctx context.Context) (err error) {
	token, err := creds.TokenProvider.RequestAccessToken(ctx, creds.ClientID, creds.ClientSecret)
	if err != nil {
		return errors.Wrap(err, "error getting new access token from IDP")
	}
	return creds.setNewToken(token, true)
}

// setNewToken parses and verifies the provided JWT using TokenParser.
// It will then decode 'm2mProfile' from the extracted claims and set
// the new values for m2mProfile, AccessToken and claims Credentials fields.
func (creds *Credentials) setNewToken(token string, withHook bool) error {
	claims, err := creds.TokenParser.Parse(token, creds.ClientID)
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
	if withHook && creds.PostRequestHook != nil {
		if err = creds.PostRequestHook(token); err != nil {
			return errors.Wrap(err, "failed to execute access token post hook")
		}
	}
	// We can now update the token and it's claims.
	creds.AccessToken = token
	switch tokenTyp {
	case tokenTypeM2M:
		creds.Organization = m2mProfile.Organization
		creds.Environment = m2mProfile.Environment
	case tokenTypeAgent:
		creds.Organization = agentProfile.Organization
		creds.Environment = agentProfile.Environment
	}
	creds.tokenType = tokenTyp
	creds.m2mProfile = m2mProfile
	creds.agentProfile = agentProfile
	creds.claims = claims
	return nil
}
