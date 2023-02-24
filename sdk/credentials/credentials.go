package credentials

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

/* #nosec G101 */
const (
	oktaTokenEndpointPattern = "%s/oauth2/%s/v1/token" //nolint: gosec
)

// Credentials stores Okta service-to-service app credentials
type Credentials struct {
	ClientID     string
	ClientSecret string
	AccessToken  string
}

func (credentials Credentials) getTokenReqAuthHeader() string {
	encoded := base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", credentials.ClientID, credentials.ClientSecret)))
	return fmt.Sprintf("Basic %s", encoded)
}

// GetBearerHeader returns an authorization header which should be included if not empty in requests to
// the resource server
func (credentials Credentials) GetBearerHeader() string {
	if credentials.AccessToken == "" {
		return ""
	}
	return fmt.Sprintf("Bearer %s", credentials.AccessToken)
}

type m2mTokenResponse struct {
	TokenType   string  `json:"token_type"`
	ExpiresIn   float64 `json:"expires_in"`
	AccessToken string  `json:"access_token"`
	Scope       string  `json:"scope"`
}

func (credentials *Credentials) RefreshOrRequestAccessToken(
	ctx context.Context,
	oktaOrgURL,
	oktaAuthServer string,
	disableOkta bool,
	client *http.Client,
) (tokenUpdated bool, err error) {
	if disableOkta {
		return false, nil
	}
	if credentials.checkAccessToken(oktaOrgURL, oktaAuthServer, client) != nil {
		if err := credentials.requestAccessToken(ctx, oktaOrgURL, oktaAuthServer, client); err != nil {
			return false, fmt.Errorf("error getting new access token from the customer identity provider: %w", err)
		}
		// Access token was updated so we need to update config file.
		return true, nil
	}

	return false, nil
}

func (credentials *Credentials) requestAccessToken(
	ctx context.Context,
	oktaOrgURL, oktaAuthServer string,
	client *http.Client,
) error {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("scope", "m2m")
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf(oktaTokenEndpointPattern, oktaOrgURL, oktaAuthServer),
		strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", credentials.getTokenReqAuthHeader())
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request to the customer identity provider: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode == http.StatusOK {
		target := m2mTokenResponse{}
		if err = json.NewDecoder(resp.Body).Decode(&target); err != nil {
			return fmt.Errorf(
				"cannot access the token, error decoding reply from the customer identity provider: %w",
				err)
		}
		credentials.AccessToken = target.AccessToken
		return nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf(
			"cannot access the token, customer identity provider replied with %d",
			resp.StatusCode)
	}
	return fmt.Errorf(
		"cannot access the token, customer identity provider replied with %d %s",
		resp.StatusCode,
		body)
}

func (credentials Credentials) checkAccessToken(oktaOrgURL, oktaAuthServer string, client *http.Client) error {
	tokenClaims, err := verifyAccessToken(context.TODO(), credentials.AccessToken, oktaOrgURL, oktaAuthServer, client)
	if err != nil {
		return err
	}
	const secondsToExpire = 60
	if tokenClaims["exp"].(float64) <= float64(time.Now().Unix()+secondsToExpire) {
		return errors.New("token will expire soon")
	}
	if credentials.ClientID != "" && (tokenClaims["cid"].(string) != credentials.ClientID) {
		return errors.New("mismatch between the client id and token's cid claim")
	}
	return nil
}
