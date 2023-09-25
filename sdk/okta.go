package sdk

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	oktaTokenEndpointPath = "v1/token"
	oktaKeysEndpointPath  = "v1/keys"
	oktaHeaderContentType = "application/x-www-form-urlencoded"

	oktaRequestTimeout = 5 * time.Second
)

func oktaAuthServerURL(oktaOrgURL *url.URL, oktaAuthServer string) *url.URL {
	return oktaOrgURL.JoinPath("oauth2", oktaAuthServer)
}

func oktaTokenEndpoint(authServerURL *url.URL) *url.URL {
	return authServerURL.JoinPath(oktaTokenEndpointPath)
}

func oktaKeysEndpoint(authServerURL *url.URL) *url.URL {
	return authServerURL.JoinPath(oktaKeysEndpointPath)
}

type getTokenEndpointFunc = func() string

type oktaClient struct {
	HTTP             *http.Client
	getTokenEndpoint getTokenEndpointFunc
}

func newOktaClient(getTokenEndpoint getTokenEndpointFunc) *oktaClient {
	return &oktaClient{
		HTTP:             newRetryableHTTPClient(oktaRequestTimeout, nil),
		getTokenEndpoint: getTokenEndpoint,
	}
}

type oktaTokenResponse struct {
	AccessToken string `json:"access_token"`
}

var errMissingClientCredentials = errors.New("client id and client secret must not be empty")

func (okta *oktaClient) RequestAccessToken(
	ctx context.Context,
	clientID, clientSecret string,
) (token string, err error) {
	if clientID == "" || clientSecret == "" {
		return "", errMissingClientCredentials
	}
	data := url.Values{
		"grant_type": {"client_credentials"},
		"scope":      {"m2m"},
	}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		okta.getTokenEndpoint(),
		strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", okta.authHeader(clientID, clientSecret))
	req.Header.Add("Content-Type", oktaHeaderContentType)

	resp, err := okta.HTTP.Do(req)
	if err != nil {
		return "", errors.Wrapf(err,
			"failed to execute POST %s request to IDP", okta.getTokenEndpoint())
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", errors.Errorf(
			"cannot access the token from POST %s, IDP replied with (status: %d): %s",
			okta.getTokenEndpoint(), resp.StatusCode, string(body))
	}
	var tr oktaTokenResponse
	if err = json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", errors.Wrapf(err,
			"cannot decode the token provided by IDP from POST %s", okta.getTokenEndpoint())
	}
	return tr.AccessToken, nil
}

func (okta *oktaClient) authHeader(clientID, clientSecret string) string {
	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", clientID, clientSecret))))
}
