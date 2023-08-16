package sdk

import (
	"net/http"
)

// ClientBuilder allows constructing Client using builder pattern (https://refactoring.guru/design-patterns/builder).
type ClientBuilder struct {
	config    *Config
	http      *http.Client
	userAgent string
}

// NewClientBuilder creates a new ClientBuilder instance.
// To fully configure the Client you must also supply ClientBuilder with Credentials instance,
// either by running ClientBuilder.WithDefaultCredentials or ClientBuilder.WithCredentials.
// Example::
//
//	config, err := sdk.ReadConfig()
//	if err != nil {
//	  panic(err)
//	}
//	client, err := sdk.NewClientBuilder(config).Build()
func NewClientBuilder(config *Config) *ClientBuilder {
	return &ClientBuilder{config: config}
}

// WithUserAgent allows setting a custom name for UserAgent HTTP header in Client requests.
func (b *ClientBuilder) WithUserAgent(userAgent string) *ClientBuilder {
	b.userAgent = userAgent
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

// Build figures out which parts were supplied for ClientBuilder and sets the defaults for the Client it constructs.
func (b *ClientBuilder) Build() (*Client, error) {
	if b.userAgent == "" {
		b.userAgent = getDefaultUserAgent()
	}
	client := &Client{
		HTTP:      b.http,
		UserAgent: b.userAgent,
	}
	return client, nil
}
