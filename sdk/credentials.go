package sdk

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
)

// AccessTokenParserInterface parses and verifies fetched access token.
type AccessTokenParserInterface interface {
	Parse(ctx context.Context, token, clientID string) (jwt.MapClaims, error)
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

func DefaultCredentials(oktaOrgURL, oktaAuthServer string) (*Credentials, error) {
	authServerURL, err := OktaAuthServer(oktaOrgURL, oktaAuthServer)
	if err != nil {
		return nil, err
	}
	parser, err := NewAccessTokenParser(authServerURL, OktaKeysEndpoint(authServerURL))
	if err != nil {
		return nil, err
	}
	provider, err := NewOktaClient(oktaOrgURL, oktaAuthServer)
	if err != nil {
		return nil, err
	}
	return &Credentials{
		authServerURL: authServerURL,
		TokenParser:   parser,
		TokenProvider: provider,
	}, nil
}

// Credentials stores and manages IDP service-to-service app credentials.
// Currently, the only supported IDP is Okta.
type Credentials struct {
	ClientID     string
	ClientSecret string
	AccessToken  string
	M2MProfile   AccessTokenM2MProfile

	HTTP *http.Client
	// TokenParser is used to verify the token and its claims.
	TokenParser AccessTokenParserInterface
	// TokenProvider is used to provide an access token.
	TokenProvider AccessTokenProvider
	// PostRequestHook is not run in offline mode.
	PostRequestHook AccessTokenPostRequestHook

	authServerURL string
	offlineMode   bool
}

// OfflineMode turns RefreshOrRequestAccessToken into a no-op.
func (creds *Credentials) OfflineMode() {
	creds.offlineMode = true
}

// GetBearerHeader returns an authorization header which should be included if not empty in requests to
// the resource server
func (creds *Credentials) GetBearerHeader() string {
	if creds.AccessToken == "" {
		return ""
	}
	return fmt.Sprintf("Bearer %s", creds.AccessToken)
}

func (creds *Credentials) RefreshOrRequestAccessToken(ctx context.Context) error {
	if creds.offlineMode {
		return nil
	}
	claims, err := creds.TokenParser.Parse(ctx, creds.AccessToken, creds.ClientID)
	switch err {
	case ErrTokenVerifyExpirationDateFailed:
		token, err := creds.TokenProvider.RequestAccessToken(ctx, creds.ClientID, creds.ClientSecret)
		if err != nil {
			return errors.Wrap(err, "error getting new access token from IDP")
		}
		if err = creds.PostRequestHook(token); err != nil {
			return errors.Wrap(err, "failed to execute access token post hook")
		}
		m2mProfile, err := M2MProfileFromClaims(claims)
		if err != nil {
			return errors.Wrap(err, "failed to decode JWT claims to m2m profile object")
		}
		creds.M2MProfile = m2mProfile
		creds.AccessToken = token
		return nil
	case nil:
		return nil
	default:
		return err
	}
}
