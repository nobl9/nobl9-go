package sdk

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

const (
	jwtLeeway             = 120 * time.Second
	jwtKeysRequestTimeout = 5 * time.Second
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
		return errors.Errorf("issuer claim '%s' is not equal to '%s', JWT claims: %v", j.Issuer, j.expectedIssuer, claimsJSON())
	}
	// We're using 'cid' instead of audience ('aud') for some reason ¯\_(ツ)_/¯.
	if j.ClaimID != j.expectedClientID {
		return errors.Errorf("claim id '%s' does not match '%s' client id, JWT claims: %v", j.ClaimID, j.expectedClientID, claimsJSON())
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

func (c jwtClaims) getTokenType() tokenType {
	if c.M2MProfile != nil {
		return tokenTypeM2M
	}
	if c.AgentProfile != nil {
		return tokenTypeAgent
	}
	return 0
}

type jwtParser struct {
	HTTP    *http.Client
	parser  *jwt.Parser
	keyfunc jwt.Keyfunc
	once    sync.Once
	issuer  string
}

func newJWTParser(issuer, jwkFetchURL string) (*jwtParser, error) {
	// Create the keyfunc.Keyfunc.
	keyfunc, err := keyfunc.NewDefault([]string{jwkFetchURL})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a keyfunc.Keyfunc with the server's URL")
	}
	return &jwtParser{
		HTTP: newRetryableHTTPClient(jwtKeysRequestTimeout, nil),
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
		keyfunc: keyfunc.Keyfunc,
		issuer:  issuer,
	}, nil
}

// Parse parses provided JWT and performs basic token signature and expiration claim validation.
func (j *jwtParser) Parse(tokenString, clientID string) (*jwtClaims, error) {
	if tokenString == "" || clientID == "" {
		return nil, errTokenParseMissingArguments
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
