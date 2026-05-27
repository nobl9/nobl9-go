package sdk

import (
	"context"
	"net/http"
	"sync"

	promapi "github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"

	prometheusEndpoints "github.com/nobl9/nobl9-go/sdk/endpoints/prometheus"
	prometheusEndpointsV1 "github.com/nobl9/nobl9-go/sdk/endpoints/prometheus/v1"
)

// Prometheus is used to access specific Prometheus-compatible API version.
func (c *Client) Prometheus() prometheusEndpoints.Versions {
	return prometheusEndpoints.NewVersions(c.newPrometheusAPI)
}

func (c *Client) newPrometheusAPI(ctx context.Context) (prometheusEndpointsV1.API, error) {
	c.prometheusAPI.mu.Lock()
	defer c.prometheusAPI.mu.Unlock()

	if c.prometheusAPI.api != nil {
		return c.prometheusAPI.api, nil
	}
	api, err := c.createPrometheusAPI(ctx)
	if err != nil {
		return nil, err
	}
	c.prometheusAPI.api = api
	return api, nil
}

type prometheusAPIStore struct {
	mu  sync.Mutex
	api prometheusEndpointsV1.API
}

func (c *Client) createPrometheusAPI(ctx context.Context) (prometheusEndpointsV1.API, error) {
	apiURL, err := c.getAPIURL(ctx)
	if err != nil {
		return nil, err
	}
	client, err := promapi.NewClient(promapi.Config{
		Address: apiURL.JoinPath("prometheus", "v1").String(),
		Client:  c.newPrometheusHTTPClient(),
	})
	if err != nil {
		return nil, err
	}
	return promv1.NewAPI(client), nil
}

func (c *Client) newPrometheusHTTPClient() *http.Client {
	client := *c.HTTP
	client.Transport = prometheusRoundTripper{
		client: c,
		next:   c.HTTP.Transport,
	}
	return &client
}

type prometheusRoundTripper struct {
	client *Client
	next   http.RoundTripper
}

func (r prometheusRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header = req.Header.Clone()
	if req.Header == nil {
		req.Header = make(http.Header)
	}

	org, err := r.client.credentials.GetOrganization(req.Context())
	if err != nil {
		return nil, httpNonRetryableError{Err: err}
	}
	req.Header.Set(HeaderOrganization, org)
	req.Header.Set(HeaderUserAgent, r.client.userAgent)

	if r.next == nil {
		return http.DefaultTransport.RoundTrip(req)
	}
	return r.next.RoundTrip(req)
}
