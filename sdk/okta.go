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

	"github.com/nobl9/nobl9-go/sdk/retryhttp"
)

const (
	defaultOktaOrgURL       = "https://accounts.nobl9.com"
	defaultOktaAuthServerID = "auseg9kiegWKEtJZC416"

	oktaTokenEndpoint     = "v1/token"
	oktaKeysEndpoint      = "v1/keys"
	oktaHeaderContentType = "application/x-www-form-urlencoded"

	oktaRequestTimeout = 5 * time.Second
)

func DefaultOktaAuthServer() (*url.URL, error) {
	return OktaAuthServer(defaultOktaOrgURL, defaultOktaAuthServerID)
}

func OktaAuthServer(oktaOrgURL, oktaAuthServer string) (*url.URL, error) {
	authServerURL, err := url.Parse(oktaOrgURL)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid oktaOrgURL: %s", oktaOrgURL)
	}
	return authServerURL.JoinPath("oauth2", oktaAuthServer), nil
}

func OktaTokenEndpoint(authServerURL *url.URL) *url.URL {
	return authServerURL.JoinPath(oktaTokenEndpoint)
}

func OktaKeysEndpoint(authServerURL *url.URL) *url.URL {
	return authServerURL.JoinPath(oktaKeysEndpoint)
}

type OktaClient struct {
	HTTP                 *http.Client
	requestTokenEndpoint string
}

func NewOktaClient(authServerURL *url.URL) *OktaClient {
	return &OktaClient{
		HTTP:                 retryhttp.NewClient(oktaRequestTimeout, nil),
		requestTokenEndpoint: OktaTokenEndpoint(authServerURL).String(),
	}
}

type oktaTokenResponse struct {
	AccessToken string `json:"access_token"`
}

var errMissingClientCredentials = errors.New("client id and client secret must not be empty")

func (okta *OktaClient) RequestAccessToken(
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
		okta.requestTokenEndpoint,
		strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", okta.authHeader(clientID, clientSecret))
	req.Header.Add("Content-Type", oktaHeaderContentType)

	resp, err := okta.HTTP.Do(req)
	if err != nil {
		return "", errors.Wrapf(err,
			"failed to execute POST %s request to IDP", okta.requestTokenEndpoint)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", errors.Errorf(
			"cannot access the token from POST %s, IDP replied with (status: %d): %s",
			okta.requestTokenEndpoint, resp.StatusCode, string(body))
	}
	var tr oktaTokenResponse
	if err = json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", errors.Wrapf(err,
			"cannot decode the token provided by IDP from POST %s", okta.requestTokenEndpoint)
	}
	return tr.AccessToken, nil
}

func (okta *OktaClient) authHeader(clientID, clientSecret string) string {
	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", clientID, clientSecret))))
}
