package sdk

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testIssuer = "https://accounts.nobl9.com/oauth2/ausdh151kj9OOWv5x191"

func TestJWTParser_Parse(t *testing.T) {
	t.Run("return error if either token or clientID are empty", func(t *testing.T) {
		_, err := new(jwtParser).Parse("", "")
		require.Error(t, err)
		assert.Equal(t, errTokenParseMissingArguments, err)
	})

	t.Run("invalid token, return error", func(t *testing.T) {
		parser := newJWTParser(testIssuer)

		_, err := parser.Parse("fake-token", "123")
		require.Error(t, err)
		assert.ErrorIs(t, err, jwt.ErrTokenMalformed)
	})

	t.Run("golden path", func(t *testing.T) {
		// Create a signed token and use the generated public key to create JWK.
		rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		const kid = "my-kid"

		for profile, claims := range map[string]jwtClaims{
			"m2mProfile": {
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:    testIssuer,
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
					NotBefore: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
					IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Hour)),
				},
				ClaimID: "123",
				M2MProfile: stringOrObject[jwtClaimM2MProfile]{Value: &jwtClaimM2MProfile{
					User:         "dev.nobl9.com",
					Organization: "my-org",
					Environment:  "test@nobl9.com",
				}},
				expectedIssuer:   "https://accounts.nobl9.com/oauth2/ausdh151kj9OOWv5x191",
				expectedClientID: "123",
			},
			"agentProfile": {
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:    testIssuer,
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
					NotBefore: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
					IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Hour)),
				},
				ClaimID: "123",
				AgentProfile: stringOrObject[jwtClaimAgentProfile]{Value: &jwtClaimAgentProfile{
					User:         "dev.nobl9.com",
					Organization: "my-org",
					Environment:  "test@nobl9.com",
					Project:      "default",
					Name:         "prometheus",
				}},
				expectedIssuer:   "https://accounts.nobl9.com/oauth2/ausdh151kj9OOWv5x191",
				expectedClientID: "123",
			},
		} {
			t.Run(profile, func(t *testing.T) {
				// Prepare the token.
				jwtToken := jwt.NewWithClaims(jwtSigningAlgorithm, claims)
				jwtToken.Header["kid"] = kid
				token, err := jwtToken.SignedString(rsaKey)
				require.NoError(t, err)

				parser := newJWTParser(testIssuer)

				result, err := parser.Parse(token, "123")
				require.NoError(t, err)
				assert.Equal(t, claims, *result)
			})
		}
	})
}

func TestJWTParser_Parse_VerifyClaims(t *testing.T) {
	// Create a signed token and use the generated public key to create JWK.
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	const kid = "my-kid"

	validAgentProfile := jwtClaimAgentProfile{
		User:         "John Wick",
		Organization: "nobl9-dev",
		Environment:  "dev.nobl9.com",
		Name:         "test",
		Project:      "default",
	}
	validM2MProfile := jwtClaimM2MProfile{
		User:         "John Wick",
		Organization: "nobl9-dev",
		Environment:  "dev.nobl9.com",
	}

	tests := map[string]struct {
		ErrorMessage string
		ErrorIs      error
		Claims       jwt.MapClaims
	}{
		"wrong issuer": {
			ErrorMessage: "issuer claim 'not the one we expect!' is not equal to " +
				"'https://accounts.nobl9.com/oauth2/ausdh151kj9OOWv5x191'",
			Claims: map[string]any{
				"iss":        "not the one we expect!",
				"m2mprofile": validM2MProfile,
			},
		},
		"client id does not match claims 'cid'": {
			ErrorMessage: "token has invalid claims: claim id '333' does not match '123' client id",
			Claims: map[string]any{
				"iss":        testIssuer,
				"exp":        time.Now().Add(time.Hour).Unix(),
				"cid":        "333",
				"m2mprofile": validM2MProfile,
			},
		},
		"expiry required": {
			ErrorIs: errTokenMissingExpiryClaim,
			Claims: map[string]any{
				"iss":        testIssuer,
				"cid":        "123",
				"m2mprofile": validM2MProfile,
			},
		},
		"barely expired, but not quite": {
			Claims: map[string]any{
				"iss":        testIssuer,
				"cid":        "123",
				"exp":        time.Now().Add((2 * time.Minute) + (1 * time.Second)).Unix(),
				"m2mprofile": validM2MProfile,
			},
		},
		"expiry": {
			ErrorIs: errTokenExpired,
			Claims: map[string]any{
				"iss":        testIssuer,
				"cid":        "123",
				"exp":        time.Now().Add((2 * time.Minute) - (1 * time.Second)).Unix(),
				"m2mprofile": validM2MProfile,
			},
		},
		"no profile": {
			ErrorMessage: "expected either 'm2mProfile' or 'agentProfile' to be set in JWT claims, but none were found",
			Claims: map[string]any{
				"iss": testIssuer,
				"cid": "123",
				"exp": time.Now().Add(time.Hour).Unix(),
			},
		},
		"both profiles set": {
			ErrorMessage: "expected either 'm2mProfile' or 'agentProfile' to be set in JWT claims, but both were found",
			Claims: map[string]any{
				"iss":          testIssuer,
				"cid":          "123",
				"exp":          time.Now().Add(time.Hour).Unix(),
				"m2mprofile":   validM2MProfile,
				"agentProfile": validAgentProfile,
			},
		},
		"agent profile empty string": {
			Claims: map[string]any{
				"iss":          testIssuer,
				"cid":          "123",
				"exp":          time.Now().Add(time.Hour).Unix(),
				"m2mprofile":   validM2MProfile,
				"agentProfile": "",
			},
		},
		"m2m profile empty string": {
			Claims: map[string]any{
				"iss":          testIssuer,
				"cid":          "123",
				"exp":          time.Now().Add(time.Hour).Unix(),
				"m2mprofile":   "",
				"agentProfile": validAgentProfile,
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			jwtToken := jwt.NewWithClaims(jwtSigningAlgorithm, test.Claims)
			jwtToken.Header["kid"] = kid
			token, err := jwtToken.SignedString(rsaKey)
			require.NoError(t, err)
			parser := newJWTParser(testIssuer)

			_, err = parser.Parse(token, "123")
			switch {
			case test.ErrorIs != nil:
				require.Error(t, err)
				assert.ErrorIs(t, err, test.ErrorIs)
			case test.ErrorMessage != "":
				require.Error(t, err)
				assert.ErrorContains(t, err, test.ErrorMessage)
			default:
				require.NoError(t, err)
			}
		})
	}
}
