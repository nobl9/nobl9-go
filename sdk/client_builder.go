package sdk

import (
	"net/http"

	"github.com/nobl9/nobl9-go/sdk/retryhttp"
)

// ClientBuilder allows constructing Client using builder pattern (https://refactoring.guru/design-patterns/builder).
type ClientBuilder struct {
	config    *Config
	http      *http.Client
	userAgent string
}

// NewClientBuilder creates a new ClientBuilder instance.
func NewClientBuilder(config *Config) *ClientBuilder {
	return &ClientBuilder{config: config}
}

// WithUserAgent allows setting a custom name for userAgent HTTP header in Client requests.
func (b *ClientBuilder) WithUserAgent(userAgent string) *ClientBuilder {
	b.userAgent = userAgent
	return b
}

// WithHTTPClient allows supplying a custom http.Client for the client to use.
// Note that the access token life cycle management is done by credentials,
// which become part of default http.Client request middleware chain, making sure
// the token is up-to-date before each request.
func (b *ClientBuilder) WithHTTPClient(client *http.Client) *ClientBuilder {
	b.http = client
	return b
}

// Build figures out which parts were supplied for ClientBuilder and sets the defaults for the Client it constructs.
func (b *ClientBuilder) Build() (*Client, error) {
	if b.userAgent == "" {
		b.userAgent = getDefaultUserAgent()
	}
	creds, err := newCredentials(b.config)
	if err != nil {
		return nil, err
	}
	if b.http == nil {
		b.http = retryhttp.NewClient(b.config.Timeout, creds)
	}
	client := &Client{
		HTTP:        b.http,
		Config:      b.config,
		credentials: creds,
		userAgent:   b.userAgent,
	}
	if err = client.loadConfig(); err != nil {
		return nil, err
	}
	return client, nil
}
