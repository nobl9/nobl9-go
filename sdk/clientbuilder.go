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

// NewClientBuilder creates a new ClientBuilder instance.
// To fully configure the Client you must also supply ClientBuilder with Credentials instance,
// either by running ClientBuilder.WithDefaultCredentials or ClientBuilder.WithCredentials.
// Recommended usage:
//
//	NewClientBuilder().WithDefaultCredentials().Build()
func NewClientBuilder(userAgent string) *ClientBuilder {
	return &ClientBuilder{userAgent: userAgent}
}

// WithCredentials allows setting an initialized Credentials instance.
func (b *ClientBuilder) WithCredentials(credentials *Credentials) *ClientBuilder {
	b.credentials = credentials
	return b
}

// WithOfflineMode if used will turn the Client.Credentials into a noop.
// If used in conjunction with WithCredentials or WithDefaultCredentials will render them useless.
func (b *ClientBuilder) WithOfflineMode() *ClientBuilder {
	b.offlineMode = true
	return b
}

// WithDefaultCredentials instructs the ClientBuilder to supply a default Credentials instance.
// It is recommended for most use cases over WithCredentials.
func (b *ClientBuilder) WithDefaultCredentials(oktaOrgURL, oktaAuthServer, clientID, clientSecret string) *ClientBuilder {
	b.oktaOrgURL = oktaOrgURL
	b.oktaAuthServer = oktaAuthServer
	b.clientID = clientID
	b.clientSecret = clientSecret
	return b
}

// WithHTTPClient allows supplying a custom http.Client for the client to use.
// Note that the access token life cycle management is done by Credentials,
// which become part of default http.Client request middleware chain, making sure
// the token is up to date before each request.
func (b *ClientBuilder) WithHTTPClient(client *http.Client) *ClientBuilder {
	b.http = client
	return b
}

// WithTimeout will only work for default HTTP client,
// it won't affect the client supplied with WithHTTPClient.
func (b *ClientBuilder) WithTimeout(timeout time.Duration) *ClientBuilder {
	b.timeout = timeout
	return b
}

// WithApiURL should only be used for development workflows as the URL is constructed from JWT claims.
func (b *ClientBuilder) WithApiURL(apiURL string) *ClientBuilder {
	b.apiURL = apiURL
	return b
}

// Build figures out which parts were supplied for ClientBuilder and sets the defaults for the Client it constructs.
func (b *ClientBuilder) Build() (*Client, error) {
	if b.offlineMode {
		b.credentials = &Credentials{}
		b.credentials.offlineMode = true
	} else if b.credentials == nil {
		authServerURL, err := OktaAuthServer(b.oktaOrgURL, b.oktaAuthServer)
		if err != nil {
			return nil, err
		}
		b.credentials, err = DefaultCredentials(b.clientID, b.clientSecret, authServerURL)
		if err != nil {
			return nil, err
		}
	}
	if b.http == nil {
		if b.timeout == 0 {
			b.timeout = Timeout
		}
		b.http = retryhttp.NewClient(b.timeout, b.credentials)
	}
	client := &Client{
		HTTP:        b.http,
		Credentials: b.credentials,
		UserAgent:   b.userAgent,
	}
	if b.apiURL != "" {
		if err := client.SetApiURL(b.apiURL); err != nil {
			return nil, err
		}
	}
	return client, nil
}
