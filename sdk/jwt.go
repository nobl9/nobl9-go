package sdk

import (
	"encoding/json"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

const jwtLeeway = 2 * time.Minute

var jwtSigningAlgorithm = jwt.SigningMethodRS256

var (
	errTokenParseMissingArguments = errors.New("token and/or client id missing in jwtParser.Parse call")
	errTokenMissingExpiryClaim    = errors.New("token is missing 'exp' claim")
	errTokenExpired               = errors.New("token is expired")
)

// Ensure we implement [jwt.ClaimsValidator] at compile time so we know our custom [jwtClaims.Validate] method is used.
var _ jwt.ClaimsValidator = (*jwtClaims)(nil)

type jwtClaims struct {
	jwt.RegisteredClaims
	ClaimID      string                               `json:"cid"`
	M2MProfile   stringOrObject[jwtClaimM2MProfile]   `json:"m2mProfile,omitzero"`
	AgentProfile stringOrObject[jwtClaimAgentProfile] `json:"agentProfile,omitzero"`

	expectedClientID string
	expectedIssuer   string
}

type jwtClaimsProfile interface {
	jwtClaimAgentProfile | jwtClaimM2MProfile
}

// stringOrObject has to be used to wrap our profiles as currently
// they can either contain the profile object or an empty string.
//
// TODO: Once PC-12146 is done, it can be removed.
type stringOrObject[T jwtClaimsProfile] struct {
	Value *T
}

func (s *stringOrObject[T]) UnmarshalJSON(data []byte) error {
	if len(data) == 2 && string(data) == `""` {
		return nil
	}
	return json.Unmarshal(data, &s.Value)
}

func (s stringOrObject[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Value)
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
	if j.M2MProfile.Value == nil && j.AgentProfile.Value == nil {
		return errors.New("expected either 'm2mProfile' or 'agentProfile' to be set in JWT claims, but none were found")
	}
	if j.M2MProfile.Value != nil && j.AgentProfile.Value != nil {
		return errors.New("expected either 'm2mProfile' or 'agentProfile' to be set in JWT claims, but both were found")
	}
	if j.ExpiresAt == nil || j.ExpiresAt.IsZero() {
		return errTokenMissingExpiryClaim
	}
	// if 15:00 is after 15:00+00:02=15:02 then it is expired
	if time.Now().After((j.ExpiresAt).Add(-jwtLeeway)) {
		return errTokenExpired
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
	if j.M2MProfile.Value != nil {
		return tokenTypeM2M
	}
	if j.AgentProfile.Value != nil {
		return tokenTypeAgent
	}
	return 0
}

type jwtParser struct {
	parser *jwt.Parser
	issuer string
}

func newJWTParser(issuer string) *jwtParser {
	return &jwtParser{
		parser: jwt.NewParser(
			jwt.WithValidMethods([]string{jwtSigningAlgorithm.Alg()}),
			// Applies to "exp", "nbf" and "iat" claims.
			jwt.WithLeeway(jwtLeeway),
			jwt.WithExpirationRequired(),
			// "exp" and "nbf" claims are always verified, "iat" is optional as per JWT RFC.
			jwt.WithIssuedAt(),
		),
		issuer: issuer,
	}
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
	if _, _, err := j.parser.ParseUnverified(tokenString, &claims); err != nil {
		return nil, err
	}
	if err := claims.Validate(); err != nil {
		return nil, errors.Wrap(err, "token has invalid claims")
	}
	return &claims, nil
}
