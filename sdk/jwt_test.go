package sdk

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testIssuer = "https://accounts.nobl9.com/oauth2/ausdh151kj9OOWv5x191"

func TestJWTParser_Parse(t *testing.T) {
	t.Run("return error if either token or clientID are empty", func(t *testing.T) {
		_, err := new(JWTParser).Parse(context.Background(), "", "")
		require.Error(t, err)
		assert.Equal(t, errTokenParseMissingArguments, err)
	})

	t.Run("invalid token, return error", func(t *testing.T) {
		parser, err := NewJWTParser(testIssuer, "https://jwk.io/keys")
		require.NoError(t, err)

		_, err = parser.Parse(context.Background(), "fake-token", "123")
		require.Error(t, err)
		assert.IsType(t, &jwt.ValidationError{}, err)
	})

	t.Run("invalid algorithm, return error", func(t *testing.T) {
		parser, err := NewJWTParser(testIssuer, "https://jwk.io/keys")
		require.NoError(t, err)

		token, _ := signToken(t, jwt.New(jwt.SigningMethodRS512))
		_, err = parser.Parse(context.Background(), token, "123")
		require.Error(t, err)
		assert.Equal(t, "expecting JWT header field 'RS256' to contain 'alg' algorithm, was: 'RS512'", err.Error())
	})

	t.Run("missing key id header, return error", func(t *testing.T) {
		parser, err := NewJWTParser(testIssuer, "https://jwk.io/keys")
		require.NoError(t, err)

		token, _ := signToken(t, jwt.New(jwt.GetSigningMethod(jwtSigningAlgorithm.String())))
		_, err = parser.Parse(context.Background(), token, "123")
		require.Error(t, err)
		assert.Equal(t, "expecting JWT header to contain 'kid' field as a string, was: ''", err.Error())
	})

	t.Run("fetch jwk fails, return error", func(t *testing.T) {
		parser, err := NewJWTParser(testIssuer, "https://jwk.io/keys")
		require.NoError(t, err)

		jwtToken := jwt.New(jwt.GetSigningMethod(jwtSigningAlgorithm.String()))
		jwtToken.Header[jwk.KeyIDKey] = "123"
		token, _ := signToken(t, jwtToken)
		expectedErr := errors.New("fetch failed!")
		jwksFetchFunction = func(ctx context.Context, urlstring string, options ...jwk.FetchOption) (jwk.Set, error) {
			return nil, expectedErr
		}
		_, err = parser.Parse(context.Background(), token, "123")
		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("fetch jwk from set cache if present in cache", func(t *testing.T) {
		parser, err := NewJWTParser(testIssuer, "https://jwk.io/keys")
		require.NoError(t, err)

		set := jwk.NewSet()
		parser.jwksCache.Set("my-kid", set, time.Hour)
		jwksFetchCalledTimes := 0
		jwksFetchFunction = func(ctx context.Context, urlstring string, options ...jwk.FetchOption) (jwk.Set, error) {
			jwksFetchCalledTimes++
			return nil, nil
		}
		result, err := parser.getJWKSet(context.Background(), "my-kid")
		require.NoError(t, err)
		assert.Equal(t, 0, jwksFetchCalledTimes)
		assert.Equal(t, set, result)
	})

	t.Run("fetch jwk from set cache only once per multiple goroutines", func(t *testing.T) {
		parser, err := NewJWTParser(testIssuer, "https://jwk.io/keys")
		require.NoError(t, err)

		const kid = "my-kid"
		JWK := jwk.NewRSAPublicKey()
		require.NoError(t, JWK.Set(jwk.KeyIDKey, kid))
		set := jwk.NewSet()
		set.Add(JWK)
		jwksFetchCalledTimes := 0
		jwksFetchFunction = func(ctx context.Context, urlstring string, options ...jwk.FetchOption) (jwk.Set, error) {
			jwksFetchCalledTimes++
			return set, nil
		}
		n := 100
		wg := sync.WaitGroup{}
		wg.Add(n)
		for i := 0; i < n; i++ {
			go func() {
				defer wg.Done()
				result, err := parser.getJWKSet(context.Background(), kid)
				require.NoError(t, err)
				assert.Equal(t, set, result)
			}()
		}
		wg.Wait()
		// Fetch only once.
		assert.Equal(t, 1, jwksFetchCalledTimes)
	})

	t.Run("'kid' not found in set, return error", func(t *testing.T) {
		parser, err := NewJWTParser(testIssuer, "https://jwk.io/keys")
		require.NoError(t, err)

		JWK := jwk.NewRSAPublicKey()
		require.NoError(t, JWK.Set(jwk.KeyIDKey, "my-kid"))
		set := jwk.NewSet()
		set.Add(JWK)
		jwksFetchFunction = func(ctx context.Context, urlstring string, options ...jwk.FetchOption) (jwk.Set, error) {
			return set, nil
		}

		jwtToken := jwt.New(jwt.GetSigningMethod(jwtSigningAlgorithm.String()))
		jwtToken.Header[jwk.KeyIDKey] = "other-kid"
		token, _ := signToken(t, jwtToken)
		_, err = parser.Parse(context.Background(), token, "123")
		require.Error(t, err)
		assert.EqualError(t, err, "jwk not found for kid: other-kid (key id)")
	})

	t.Run("golden path", func(t *testing.T) {
		parser, err := NewJWTParser(testIssuer, "https://jwk.io/keys")
		require.NoError(t, err)

		// Create a signed token and use the generated public key to create JWK.
		rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		const kid = "my-kid"
		// Create a JSON Web Key with a key id matching the tokens' kid.
		JWK := jwk.NewRSAPublicKey()
		require.NoError(t, JWK.Set(jwk.KeyIDKey, kid))
		require.NoError(t, JWK.Set(jwk.AlgorithmKey, jwtSigningAlgorithm))
		require.NoError(t, JWK.FromRaw(&rsaKey.PublicKey))
		// Create a JWK Set and add a single JWK.
		set := jwk.NewSet()
		set.Add(JWK)
		parser.jwksCache.Set(kid, set, time.Hour)

		// Prepare the token.
		claims := jwt.MapClaims{
			"iss": testIssuer,
			"cid": "123",
			"exp": time.Now().Add(time.Hour).Unix(),
			"iat": time.Now().Add(-time.Hour).Unix(),
			"nbf": time.Now().Add(-time.Hour).Unix(),
			"m2mProfile": map[string]interface{}{
				"environment":  "dev.nobl9.com",
				"organization": "my-org",
				"user":         "test@nobl9.com",
			},
		}
		jwtToken := jwt.NewWithClaims(jwt.GetSigningMethod(jwtSigningAlgorithm.String()), claims)
		jwtToken.Header["kid"] = kid
		token, err := jwtToken.SignedString(rsaKey)
		require.NoError(t, err)

		result, err := parser.Parse(context.Background(), token, "123")
		require.NoError(t, err)

		assert.Contains(t, result, "m2mProfile")
		assert.Equal(t, claims["m2mProfile"], result["m2mProfile"])
	})
}

func TestJWTParser_Parse_VerifyClaims(t *testing.T) {
	parser, err := NewJWTParser(testIssuer, "https://jwk.io/keys")
	require.NoError(t, err)

	// Create a signed token and use the generated public key to create JWK.
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	const kid = "my-kid"
	// Create a JSON Web Key with a key id matching the tokens' kid.
	JWK := jwk.NewRSAPublicKey()
	require.NoError(t, JWK.Set(jwk.KeyIDKey, kid))
	require.NoError(t, JWK.Set(jwk.AlgorithmKey, jwtSigningAlgorithm))
	require.NoError(t, JWK.FromRaw(&rsaKey.PublicKey))
	// Create a JWK Set and add a single JWK.
	set := jwk.NewSet()
	set.Add(JWK)
	parser.jwksCache.Set(kid, set, time.Hour)

	for name, test := range map[string]struct {
		ExpectedError string
		Claims        jwt.MapClaims
	}{
		"wrong issuer": {
			ExpectedError: "issuer claim validation failed",
			Claims: map[string]interface{}{
				"iss": "not the one we expect!",
			},
		},
		"client id does not match claims 'cid'": {
			ExpectedError: "client id does not match token's 'cid' claim",
			Claims: map[string]interface{}{
				"iss": testIssuer,
				"cid": "333",
			},
		},
		"expiry": {
			ExpectedError: "exp (expiry) claim validation failed",
			Claims: map[string]interface{}{
				"iss": testIssuer,
				"exp": time.Now().Unix(),
				"cid": "123",
			},
		},
		"issued at": {
			ExpectedError: "iat (issued at) claim validation failed",
			Claims: map[string]interface{}{
				"iss": testIssuer,
				"cid": "123",
				"exp": time.Now().Add(time.Hour).Unix(),
				"iat": time.Now().Add(time.Hour).Unix(),
			},
		},
		"not before": {
			ExpectedError: "nbf (not before) claim validation failed",
			Claims: map[string]interface{}{
				"iss": testIssuer,
				"cid": "123",
				"exp": time.Now().Add(time.Hour).Unix(),
				"iat": time.Now().Add(-time.Hour).Unix(),
				"nbf": time.Now().Add(time.Hour).Unix(),
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			jwtToken := jwt.NewWithClaims(jwt.GetSigningMethod(jwtSigningAlgorithm.String()), test.Claims)
			jwtToken.Header["kid"] = kid
			token, err := jwtToken.SignedString(rsaKey)
			require.NoError(t, err)
			_, err = parser.Parse(context.Background(), token, "123")
			require.Error(t, err)
			assert.ErrorContains(t, err, test.ExpectedError)
		})
	}
}

func TestM2MProfileFromClaims(t *testing.T) {
	claims := jwt.MapClaims{
		"m2mProfile": map[string]interface{}{
			"user":         "test@nobl9.com",
			"organization": "my-org",
			"environment":  "dev.nobl9.com",
		},
	}
	m2mProfile, err := M2MProfileFromClaims(claims)
	require.NoError(t, err)
	expected := AccessTokenM2MProfile{
		User:         "test@nobl9.com",
		Organization: "my-org",
		Environment:  "dev.nobl9.com",
	}
	assert.Equal(t, expected, m2mProfile)
}

func signToken(t *testing.T, jwtToken *jwt.Token) (token string, key *rsa.PrivateKey) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	strToken, err := jwtToken.SignedString(key)
	require.NoError(t, err)
	return strToken, key
}
