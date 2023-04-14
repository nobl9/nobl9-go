package sdk

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
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

// AccessTokenM2MProfile stores information specific to an Okta application.
type AccessTokenM2MProfile struct {
	User         string `json:"user,omitempty"`
	Organization string `json:"organization,omitempty"`
	Environment  string `json:"environment,omitempty"`
}

func DefaultCredentials(clientID, clientSecret string, authServerURL *url.URL) (*Credentials, error) {
	if clientID == "" || clientSecret == "" || authServerURL == nil {
		return nil, errors.New("clientID, clientSecret and AuthServerURL must all be provided for DefaultCredentials call")
	}
	parser, err := NewJWTParser(authServerURL.String(), OktaKeysEndpoint(authServerURL))
	if err != nil {
		return nil, err
	}
	return &Credentials{
		ClientID:      clientID,
		ClientSecret:  clientSecret,
		TokenParser:   parser,
		TokenProvider: NewOktaClient(authServerURL),
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
	M2MProfile  AccessTokenM2MProfile
	claims      jwt.MapClaims

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

func (creds *Credentials) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := creds.RefreshAccessToken(req.Context()); err != nil {
		return nil, err
	}
	creds.SetAuthorizationHeader(req)
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
func (creds *Credentials) SetAccessToken(token string) error {
	creds.mu.Lock()
	defer creds.mu.Unlock()
	return creds.setNewToken(token, false)
}

// RefreshAccessToken checks the AccessToken expiry with an offset to detect if the token
// is soon to be expired. If so, it wll request a new token and update the Credentials state.
// If the token was not yet set, it will request a new one all the same.
func (creds *Credentials) RefreshAccessToken(ctx context.Context) error {
	if creds.offlineMode {
		return nil
	}
	if !creds.shouldRefresh() {
		return nil
	}
	creds.mu.Lock()
	defer creds.mu.Unlock()
	if !creds.shouldRefresh() {
		return nil
	}
	return creds.requestNewToken(ctx)
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
// the new values for M2MProfile, AccessToken and claims Credentials fields.
func (creds *Credentials) setNewToken(token string, withHook bool) error {
	claims, err := creds.TokenParser.Parse(token, creds.ClientID)
	if err != nil {
		return err
	}
	m2mProfile, err := M2MProfileFromClaims(claims)
	if err != nil {
		return errors.Wrap(err, "failed to decode JWT claims to m2m profile object")
	}
	if withHook && creds.PostRequestHook != nil {
		if err = creds.PostRequestHook(token); err != nil {
			return errors.Wrap(err, "failed to execute access token post hook")
		}
	}
	// We can now update the token.
	creds.M2MProfile = m2mProfile
	creds.AccessToken = token
	creds.claims = claims
	return nil
}
