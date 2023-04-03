package sdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
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

	jwtHeaderAlgorithm = "alg"
	jwtHeaderKeyID     = "kid"

	jwtTokenClaimProfile = "m2mProfile"
	jwtTokenClaimCID     = "cid"
)

var (
	ErrJWKSetNotFound                  = errors.New("jwk not found for kid (key id)")
	ErrTokenVerifyExpirationDateFailed = errors.New("expiry claim validation failed")
	ErrTokenVerifyIssuedAtFailed       = errors.New("iat (issued at) claim validation failed")
	ErrTokenVerifyNotBeforeFailed      = errors.New("nbf (not before) claim validation failed")
	ErrTokenInvalidCID                 = errors.New("client id does not match token's 'cid' claim")
)

type AccessTokenParser struct {
	HTTP        *http.Client
	jwkFetchURL string
	issuer      string
	jwksCache   *cache.Cache
	jwkSetMu    *sync.Mutex
}

func NewAccessTokenParser(issuer, jwkFetchURL string) (*AccessTokenParser, error) {
	if _, err := url.Parse(jwkFetchURL); err != nil {
		return nil, errors.Wrapf(err, "invalid JWK fetching URL: %s", jwkFetchURL)
	}
	return &AccessTokenParser{
		HTTP:        newRetryableHTTPClient(jwtKeysRequestTimeout, nil),
		jwksCache:   cache.New(time.Hour, time.Hour),
		jwkSetMu:    new(sync.Mutex),
		jwkFetchURL: jwkFetchURL,
		issuer:      issuer,
	}, nil
}

// M2MProfileFromClaims returns AccessTokenM2MProfile object parsed from m2mProfile claim of provided token.
func M2MProfileFromClaims(claims jwt.MapClaims) (AccessTokenM2MProfile, error) {
	var accessKeyProfile AccessTokenM2MProfile
	err := mapstructure.Decode(claims[jwtTokenClaimProfile], &accessKeyProfile)
	return accessKeyProfile, err
}

// Parse parses provided JWT and performs basic token signature and expiration claim validation.
func (a *AccessTokenParser) Parse(ctx context.Context, token, clientID string) (jwt.MapClaims, error) {
	jwtParser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwtSigningAlgorithm.String()}),
		jwt.WithoutClaimsValidation()) // We'll perform claims validation ourselves to account for clock skew.
	unverifiedJwtToken, _, err := jwtParser.ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		return nil, err
	}
	alg, ok := unverifiedJwtToken.Header[jwtHeaderAlgorithm].(string)
	if !ok {
		return nil, errors.Errorf("expecting JWT header to contain '%s' field as a string, was: '%s'",
			jwtHeaderAlgorithm, alg)
	}
	if alg != jwtSigningAlgorithm.String() {
		return nil, errors.Errorf("expecting JWT header field '%s' to contain '%s' algorithm, was: '%s'",
			jwtSigningAlgorithm, jwtHeaderAlgorithm, alg)
	}
	kid, ok := unverifiedJwtToken.Header[jwtHeaderKeyID].(string)
	if !ok {
		return nil, errors.Errorf("expecting JWT header to contain '%s' filed as a string, was: '%s'",
			jwtHeaderKeyID, kid)
	}
	jwkSet, err := a.getJWKSet(ctx, kid)
	if err != nil {
		return nil, err
	}

	rawClaims, err := jws.VerifySet([]byte(token), jwkSet)
	if err != nil {
		return nil, err
	}

	var claims jwt.MapClaims
	if err = json.Unmarshal(rawClaims, &claims); err != nil {
		return nil, err
	}
	if err = a.verifyClaims(claims, clientID); err != nil {
		return nil, err
	}
	return claims, err
}

func (a *AccessTokenParser) getJWKSet(ctx context.Context, kid string) (jwk.Set, error) {
	// There are three scenarios under which a token might not be found in the cache:
	// 1. cache is empty right after the startup
	// 2. cache expired
	// 3. signing keys rotation happened
	if keySet, found := a.jwksCache.Get(kid); found {
		return keySet.(jwk.Set), nil
	}
	// Perform only one concurrent request to Okta's jwks endpoint.
	a.jwkSetMu.Lock()
	defer a.jwkSetMu.Unlock()
	// If a goroutine waited in queue to get the lock there is a chance that another goroutine already did the job.
	if keySet, found := a.jwksCache.Get(kid); found {
		return keySet.(jwk.Set), nil
	}

	// Fetch doesn't perform retries. Use background context because we don't want client disconnects to interrupt
	// JWKS cache population process while other clients might be waiting for it.
	ctx = context.Background()
	jwkSet, err := jwk.Fetch(ctx, a.jwkFetchURL, jwk.WithHTTPClient(a.HTTP))
	if err != nil {
		return nil, err
	}

	kidMatchingKeySets := make(map[string]jwk.Set)
	for it := jwkSet.Iterate(ctx); it.Next(ctx); {
		key := it.Pair().Value.(jwk.Key)
		if _, ok := kidMatchingKeySets[key.KeyID()]; !ok {
			kidMatchingKeySets[key.KeyID()] = jwk.NewSet()
		}
		kidMatchingKeySets[key.KeyID()].Add(key)
	}
	for kid, keySet := range kidMatchingKeySets {
		a.jwksCache.SetDefault(kid, keySet)
	}

	if _, found := jwkSet.LookupKeyID(kid); !found {
		// One might think that it would be useful to cache the error for a given kid. Unfortunately we can't
		// easily do that because /v1/keys endpoint might return stale information for a bit after signing keys
		// rotation. /v1/keys docs: "Note: The information returned from this endpoint could lag slightly, but will
		// eventually be up-to-date."
		return nil, ErrJWKSetNotFound
	}

	return kidMatchingKeySets[kid], nil
}

func (a *AccessTokenParser) verifyClaims(claims jwt.MapClaims, clientID string) error {
	if !claims.VerifyIssuer(a.issuer, true) {
		return errors.New("issuer claim validation failed")
	}
	if cid, ok := claims[jwtTokenClaimCID].(string); !ok || cid != clientID {
		return ErrTokenInvalidCID
	}
	now := time.Now().Unix()
	if !claims.VerifyExpiresAt(now-jwtAllowedClockSkewSeconds, true) {
		return ErrTokenVerifyExpirationDateFailed
	}
	if !claims.VerifyIssuedAt(now+jwtAllowedClockSkewSeconds, true) {
		return ErrTokenVerifyIssuedAtFailed
	}
	if !claims.VerifyNotBefore(now+jwtAllowedClockSkewSeconds, false) {
		return ErrTokenVerifyNotBeforeFailed
	}
	return nil
}
