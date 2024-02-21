package sdk

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/url"
	"sync"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"golang.org/x/time/rate"
)

const (
	jwtLeeway          = 120 * time.Second
	jwksRequestTimeout = 10 * time.Second
)

var jwtSigningAlgorithm = jwt.SigningMethodRS256

var errTokenParseMissingArguments = errors.New("token and/or client id missing in jwtParser.Parse call")

// Ensure we implement [jwt.ClaimsValidator] at compile time so we know our custom [jwtClaims.Validate] method is used.
var _ jwt.ClaimsValidator = (*jwtClaims)(nil)

type jwtClaims struct {
	jwt.RegisteredClaims
	ClaimID      string                `json:"cid"`
	M2MProfile   *jwtClaimM2MProfile   `json:"m2mProfile,omitempty"`
	AgentProfile *jwtClaimAgentProfile `json:"agentProfile,omitempty"`

	expectedClientID string
	expectedIssuer   string
}

// jwtClaimM2MProfile stores information specific to an Okta M2M application.
type jwtClaimM2MProfile struct {
	User         string `json:"user"`
	Organization string `json:"organization"`
	Environment  string `json:"environment"`
}

// jwtClaimAgentProfile stores information specific to an Okta Agent application.
type jwtClaimAgentProfile struct {
	User         string `json:"user"`
	Organization string `json:"organization"`
	Environment  string `json:"environment"`
	Name         string `json:"name"`
	Project      string `json:"project"`
}

func (j jwtClaims) Validate() error {
	claimsJSON := func() string {
		data, _ := json.Marshal(j)
		return string(data)
	}
	if j.Issuer != j.expectedIssuer {
		return errors.Errorf("issuer claim '%s' is not equal to '%s', JWT claims: %v",
			j.Issuer, j.expectedIssuer, claimsJSON())
	}
	// We're using 'cid' instead of audience ('aud') for some reason ¯\_(ツ)_/¯.
	if j.ClaimID != j.expectedClientID {
		return errors.Errorf("claim id '%s' does not match '%s' client id, JWT claims: %v",
			j.ClaimID, j.expectedClientID, claimsJSON())
	}
	if j.M2MProfile == nil && j.AgentProfile == nil {
		return errors.New("expected either 'm2mProfile' or 'agentProfile' to be set in JWT claims, but none were found")
	}
	if j.M2MProfile != nil && j.AgentProfile != nil {
		return errors.New("expected either 'm2mProfile' or 'agentProfile' to be set in JWT claims, but both found")
	}
	return nil
}

// tokenType describes what kind of token and specific claims do we expect.
type tokenType int

const (
	tokenTypeM2M tokenType = iota + 1
	tokenTypeAgent
)

func (j jwtClaims) getTokenType() tokenType {
	if j.M2MProfile != nil {
		return tokenTypeM2M
	}
	if j.AgentProfile != nil {
		return tokenTypeAgent
	}
	return 0
}

type jwtParser struct {
	parser      *jwt.Parser
	keyfunc     jwt.Keyfunc
	once        sync.Once
	issuer      string
	jwkFetchURL string
}

func newJWTParser(issuer, jwkFetchURL string) *jwtParser {
	return &jwtParser{
		parser: jwt.NewParser(
			jwt.WithValidMethods([]string{jwtSigningAlgorithm.Alg()}),
			// Applies to "exp", "nbf" and "iat" claims.
			// We're adding negative leeway for a stricter approach which will report expiration
			// some time before it actually expires.
			jwt.WithLeeway(-jwtLeeway),
			jwt.WithExpirationRequired(),
			// "exp" amd "nbf" claims are always verified, "iat" is optional as per JWT RFC.
			jwt.WithIssuedAt(),
		),
		issuer:      issuer,
		jwkFetchURL: jwkFetchURL,
	}
}

// Parse parses provided JWT and performs basic token signature and expiration claim validation.
func (j *jwtParser) Parse(tokenString, clientID string) (*jwtClaims, error) {
	if tokenString == "" || clientID == "" {
		return nil, errTokenParseMissingArguments
	}
	var err error
	j.once.Do(func() { err = j.initKeyfunc() })
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize JWT parser keyfunc.Keyfunc")
	}
	claims := jwtClaims{
		expectedClientID: clientID,
		expectedIssuer:   j.issuer,
	}
	if _, err := j.parser.ParseWithClaims(tokenString, &claims, j.keyfunc); err != nil {
		return nil, err
	}
	return &claims, nil
}

// initKeyfunc should be called as late as possible, that's why it's placed in [jwtParser.Parse] method.
// The reason is keyfunc library immediately attempts to fetch keys from the server, otherwise,
// it might be counter-intuitive that such a resource-intensive operation is executed within constructor.
func (j *jwtParser) initKeyfunc() error {
	jwkStorage, err := newJWKStorage(j.jwkFetchURL)
	if err != nil {
		return errors.Wrapf(err, "failed to create a jwkset.Storage with the server's URL: %s", j.jwkFetchURL)
	}
	keyFunc, err := keyfunc.New(keyfunc.Options{Storage: jwkStorage})
	if err != nil {
		return errors.Wrap(err, "failed to create a keyfunc.Keyfunc")
	}
	j.keyfunc = keyFunc.Keyfunc
	return nil
}

// newJWKStorage is almost a direct copy of the [jwkset.NewDefaultHTTPClientCtx].
// One notable change is that we're setting NoErrorReturnFirstHTTPReq to false,
// this ensures that if an error occurs when fetching keys inside the constructor,
// it is returned immediately.
// We also modify the timeout value.
func newJWKStorage(jwkFetchURL string) (jwkset.Storage, error) {
	parsed, err := url.ParseRequestURI(jwkFetchURL)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse given URL %q", jwkFetchURL)
	}
	jwkFetchURI := parsed.String()
	refreshErrorHandler := func(ctx context.Context, err error) {
		slog.Default().ErrorContext(ctx, "Failed to refresh HTTP JWK Set from remote HTTP resource.",
			"error", err,
			"url", jwkFetchURI,
		)
	}
	options := jwkset.HTTPClientStorageOptions{
		NoErrorReturnFirstHTTPReq: false,
		RefreshErrorHandler:       refreshErrorHandler,
		RefreshInterval:           time.Hour,
		HTTPTimeout:               jwksRequestTimeout,
	}
	storage, err := jwkset.NewStorageFromHTTP(parsed, options)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create HTTP client storage for %q", jwkFetchURI)
	}
	return jwkset.NewHTTPClient(jwkset.HTTPClientOptions{
		HTTPURLs:          map[string]jwkset.Storage{jwkFetchURI: storage},
		RefreshUnknownKID: rate.NewLimiter(rate.Every(5*time.Minute), 1),
	})
}
