package credentials

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
)

// nolint: gochecknoglobals
var (
	ErrJWKSetNotFound = errors.New("jwk not found for kid")
	jwksCache         = cache.New(time.Hour, time.Hour)
	jwkSetMu          = &sync.Mutex{}
)

func getJWKSet(ctx context.Context, authServerURL, kid string, client *http.Client) (jwk.Set, error) {
	// There are three scenarios under which a token might not be found in the cache:
	// 1. cache is empty right after the startup
	// 2. cache expired
	// 3. signing keys rotation happened
	if keySet, found := jwksCache.Get(kid); found {
		return keySet.(jwk.Set), nil
	}
	// Let's perform only one concurrent request to Okta's jwks endpoint.
	jwkSetMu.Lock()
	defer jwkSetMu.Unlock()
	// Check again because if a goroutine waited in queue to get the lock there is a chance that the goroutine before
	// it performed FetchHTTP call and cached its results already.
	if keySet, found := jwksCache.Get(kid); found {
		return keySet.(jwk.Set), nil
	}

	keysURL := fmt.Sprintf("%s/v1/keys", authServerURL)

	jwkSet, err := jwk.Fetch(ctx, keysURL, jwk.WithHTTPClient(client))
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
		jwksCache.SetDefault(kid, keySet)
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

type clockSkewTolerantMapClaims jwt.MapClaims

func (claims clockSkewTolerantMapClaims) Valid(authServerURL string) error {
	if !jwt.MapClaims(claims).VerifyIssuer(authServerURL, true) {
		return errors.New("issuer claim validation failed")
	}
	const allowedClockSkewSeconds = 120
	now := time.Now().Unix()
	if !jwt.MapClaims(claims).VerifyExpiresAt(now-allowedClockSkewSeconds, true) {
		return errors.New("expiry claim validation failed")
	}
	if !jwt.MapClaims(claims).VerifyIssuedAt(now+allowedClockSkewSeconds, true) {
		return errors.New("iat claim validation failed")
	}
	if !jwt.MapClaims(claims).VerifyNotBefore(now+allowedClockSkewSeconds, false) {
		return errors.New("nbf claim validation failed")
	}

	return nil
}

// verifyAccessToken performs basic token signature and expiration claim validation.
func verifyAccessToken(
	ctx context.Context,
	token, oktaOrgURL, oktaAuthServer string,
	client *http.Client,
) (map[string]interface{}, error) {
	const signingAlgorithm = "RS256"
	jwtParser := &jwt.Parser{
		ValidMethods:         []string{signingAlgorithm},
		SkipClaimsValidation: true, // We'll perform claims validation ourself to account for clock skew.
	}
	unverifiedJwtToken, _, err := jwtParser.ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		return nil, err
	}
	alg, ok := unverifiedJwtToken.Header["alg"].(string)
	if !ok || alg != signingAlgorithm {
		return nil, errors.Errorf("expecting JWT header to have %s alg", signingAlgorithm)
	}
	kid, ok := unverifiedJwtToken.Header["kid"].(string)
	if !ok {
		return nil, errors.New("expecting JWT header to have string kid")
	}
	authServerURL := fmt.Sprintf("%s/%s/%s", oktaOrgURL, "oauth2", oktaAuthServer)
	jwkSet, err := getJWKSet(ctx, authServerURL, kid, client)
	if err != nil {
		return nil, err
	}

	rawClaims, err := jws.VerifySet([]byte(token), jwkSet)
	if err != nil {
		return nil, err
	}

	var claims clockSkewTolerantMapClaims
	if err := json.Unmarshal(rawClaims, &claims); err != nil {
		return nil, err
	}
	if err := claims.Valid(authServerURL); err != nil {
		return nil, err
	}
	return claims, nil
}
