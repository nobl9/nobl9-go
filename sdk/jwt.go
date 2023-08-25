package sdk

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/mitchellh/mapstructure"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
)

const (
	jwtSigningAlgorithm        = jwa.RS256
	jwtAllowedClockSkewSeconds = 120
	jwtKeysRequestTimeout      = 5 * time.Second

	jwtTokenClaimM2MProfile   = "m2mProfile"
	jwtTokenClaimAgentProfile = "agentProfile"
	jwtTokenClaimCID          = "cid"
)

var errTokenParseMissingArguments = errors.New("token and/or client id missing in jwtParser.Parse call")

// tokenType describes what kind of token and specific claims do we expect.
type tokenType int

const (
	tokenTypeM2M tokenType = iota + 1
	tokenTypeAgent
)

func tokenTypeFromClaims(claims jwt.MapClaims) (typ tokenType) {
	isZero := func(v interface{}) bool {
		vo := reflect.ValueOf(v)
		if !vo.IsValid() {
			return true
		}
		return vo.IsZero()
	}
	switch {
	case !isZero(claims[jwtTokenClaimM2MProfile]):
		typ = tokenTypeM2M
	case !isZero(claims[jwtTokenClaimAgentProfile]):
		typ = tokenTypeAgent
	}
	return
}

// m2mProfileFromClaims returns accessTokenM2MProfile object parsed from m2mProfile claim of provided token.
func m2mProfileFromClaims(claims jwt.MapClaims) (accessTokenM2MProfile, error) {
	var profile accessTokenM2MProfile
	err := mapstructure.Decode(claims[jwtTokenClaimM2MProfile], &profile)
	return profile, err
}

// agentProfileFromClaims returns accessTokenAgentProfile object parsed from agentProfile claim of provided token.
func agentProfileFromClaims(claims jwt.MapClaims) (accessTokenAgentProfile, error) {
	var profile accessTokenAgentProfile
	err := mapstructure.Decode(claims[jwtTokenClaimAgentProfile], &profile)
	return profile, err
}

type (
	getJWTIssuerFunc   = func() string
	getJWKFetchURLFunc = func() string
)

type jwtParser struct {
	HTTP           *http.Client
	getJWKFetchURL getJWKFetchURLFunc
	getIssuer      getJWTIssuerFunc
	jwksCache      *cache.Cache
	jwkSetMu       *sync.Mutex
}

func newJWTParser(issuer getJWTIssuerFunc, jwkFetchURL getJWKFetchURLFunc) *jwtParser {
	return &jwtParser{
		HTTP:           newRetryableHTTPClient(jwtKeysRequestTimeout, nil),
		jwksCache:      cache.New(time.Hour, time.Hour),
		jwkSetMu:       new(sync.Mutex),
		getJWKFetchURL: jwkFetchURL,
		getIssuer:      issuer,
	}
}

// Parse parses provided JWT and performs basic token signature and expiration claim validation.
func (j *jwtParser) Parse(token, clientID string) (jwt.MapClaims, error) {
	if token == "" || clientID == "" {
		return nil, errTokenParseMissingArguments
	}
	jwtParser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwtSigningAlgorithm.String()}),
		jwt.WithoutClaimsValidation()) // We'll perform claims validation ourselves to account for clock skew.
	unverifiedJwtToken, _, err := jwtParser.ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		return nil, err
	}
	// Parser will also check if 'alg' header is set. We should still be extra cautious here.
	alg, ok := unverifiedJwtToken.Header[jwk.AlgorithmKey].(string)
	if !ok || alg != jwtSigningAlgorithm.String() {
		return nil, errors.Errorf("expecting JWT header field '%s' to contain '%s' algorithm, was: '%s'",
			jwtSigningAlgorithm, jwk.AlgorithmKey, alg)
	}
	kid, ok := unverifiedJwtToken.Header[jwk.KeyIDKey].(string)
	if !ok || kid == "" {
		return nil, errors.Errorf("expecting JWT header to contain '%s' field as a string, was: '%s'",
			jwk.KeyIDKey, kid)
	}
	jwkSet, err := j.getJWKSet(kid)
	if err != nil {
		return nil, err
	}

	// This check is only run for clarity, jws.VerifySet will detect 'kid' mismatch too,
	// but the error it returns is ambiguous.
	if _, found := jwkSet.LookupKeyID(kid); !found {
		return nil, errors.Errorf("jwk not found for kid: %s (key id)", kid)
	}
	// One might think that it would be useful to cache the error for a given kid. Unfortunately we can't
	// easily do that because /v1/keys endpoint might return stale information for a bit after signing keys
	// rotation. /v1/keys docs: "Note: The information returned from this endpoint could lag slightly, but will
	// eventually be up-to-date."
	rawClaims, err := jws.VerifySet([]byte(token), jwkSet)
	if err != nil {
		return nil, err
	}

	var claims jwt.MapClaims
	if err = json.Unmarshal(rawClaims, &claims); err != nil {
		return nil, err
	}
	if err = j.verifyClaims(claims, clientID); err != nil {
		return nil, err
	}
	return claims, err
}

var jwksFetchFunction = jwk.Fetch

func (j *jwtParser) getJWKSet(kid string) (jwk.Set, error) {
	// There are three scenarios under which a token might not be found in the cache:
	// 1. Cache is empty right after the startup.
	// 2. Cache expired.
	// 3. Signing keys have been rotated.
	if keySet, found := j.jwksCache.Get(kid); found {
		return keySet.(jwk.Set), nil
	}
	// Perform only one concurrent request to Okta's jwks endpoint.
	j.jwkSetMu.Lock()
	defer j.jwkSetMu.Unlock()
	// If a goroutine waited in queue to get the lock there is a chance that another goroutine already did the job.
	if keySet, found := j.jwksCache.Get(kid); found {
		return keySet.(jwk.Set), nil
	}

	// Fetch doesn't perform retries. Use background context because we don't want client disconnects to interrupt
	// JWKS cache population process while other clients might be waiting for it.
	jwkSet, err := jwksFetchFunction(context.Background(), j.getJWKFetchURL(), jwk.WithHTTPClient(j.HTTP))
	if err != nil {
		return nil, err
	}
	j.jwksCache.SetDefault(kid, jwkSet)
	return jwkSet, nil
}

func (j *jwtParser) verifyClaims(claims jwt.MapClaims, clientID string) error {
	claimsJSON := func() string {
		data, _ := json.Marshal(claims)
		return string(data)
	}

	if !claims.VerifyIssuer(j.getIssuer(), true) {
		return errors.Errorf("issuer claim validation failed, issuer: %s, claims: %v", j.getIssuer(), claimsJSON())
	}
	// We're using 'cid' instead of audience ('aud') for some reason ¯\_(ツ)_/¯.
	if cid, ok := claims[jwtTokenClaimCID].(string); !ok || cid != clientID {
		return errors.Errorf("client id does not match token's 'cid' claim, clientID: %s, claims: %v", clientID, claimsJSON())
	}
	// By adding the skew we're saying that we might be behind the clock.
	nowWithOffset := time.Now().Unix() + jwtAllowedClockSkewSeconds
	if !claims.VerifyExpiresAt(nowWithOffset, true) {
		return errors.Errorf("exp (expiry) claim validation failed, ts: %d, claims: %v", nowWithOffset, claimsJSON())
	}
	if !claims.VerifyIssuedAt(nowWithOffset, true) {
		return errors.Errorf("iat (issued at) claim validation failed, ts: %d, claims: %v", nowWithOffset, claimsJSON())
	}
	if !claims.VerifyNotBefore(nowWithOffset, false) {
		return errors.Errorf("nbf (not before) claim validation failed, ts: %d, claims: %v", nowWithOffset, claimsJSON())
	}
	return nil
}
