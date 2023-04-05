package sdk

import (
	"net/http"
	"time"

	"github.com/nobl9/nobl9-go/sdk/retryhttp"
)

// ClientBuilder allows constructing Client using builder pattern (https://refactoring.guru/design-patterns/builder).
type ClientBuilder struct {
	http           *http.Client
	timeout        time.Duration
	credentials    *Credentials
	userAgent      string
	clientID       string
	clientSecret   string
	oktaOrgURL     string
	oktaAuthServer string
	offlineMode    bool
	apiURL         string
}

// NewClientBuilder accepts the required minimum of arguments.
// To fully configure the Client
func NewClientBuilder(userAgent string) ClientBuilder {
	return ClientBuilder{userAgent: userAgent}
}

func (b ClientBuilder) WithCredentials(credentials *Credentials) ClientBuilder {
	b.credentials = credentials
	return b
}

func (b ClientBuilder) WithOfflineMode() ClientBuilder {
	b.offlineMode = true
	return b
}

func (b ClientBuilder) WithDefaultCredentials(oktaOrgURL, oktaAuthServer, clientID, clientSecret string) ClientBuilder {
	b.oktaOrgURL = oktaOrgURL
	b.oktaAuthServer = oktaAuthServer
	b.clientID = clientID
	b.clientSecret = clientSecret
	return b
}

func (b ClientBuilder) WithHTTPClient(client *http.Client) ClientBuilder {
	b.http = client
	return b
}

func (b ClientBuilder) WithTimeout(timeout time.Duration) ClientBuilder {
	b.timeout = timeout
	return b
}

func (b ClientBuilder) WithApiURL(apiURL string) ClientBuilder {
	b.apiURL = apiURL
	return b
}

func (b ClientBuilder) Build() (*Client, error) {
	authServerURL, err := OktaAuthServer(b.oktaOrgURL, b.oktaAuthServer)
	if err != nil {
		return nil, err
	}
	if b.credentials == nil {
		b.credentials, err = DefaultCredentials(b.clientID, b.clientSecret, authServerURL)
		if err != nil {
			return nil, err
		}
	}
	if b.offlineMode {
		b.credentials.offlineMode = true
	}
	if b.http == nil {
		if b.timeout == 0 {
			b.timeout = Timeout
		}
		b.http = retryhttp.NewClient(b.timeout, b.credentials)
	}
	return &Client{
		HTTP:        b.http,
		Credentials: b.credentials,
		UserAgent:   b.userAgent,
		apiURL:      b.apiURL,
	}, nil
}
