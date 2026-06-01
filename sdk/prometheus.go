package sdk

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
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
	return promv1.NewAPI(prometheusAPIClient{Client: client}), nil
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

type prometheusAPIClient struct {
	promapi.Client
}

func (c prometheusAPIClient) Do(ctx context.Context, req *http.Request) (*http.Response, []byte, error) {
	resp, body, err := c.Client.Do(ctx, req)
	if err != nil || resp == nil || resp.StatusCode < 300 {
		return resp, body, err
	}
	return resp, body, newPrometheusHTTPError(resp, body)
}

type prometheusAPIResponse struct {
	Status    string           `json:"status"`
	ErrorType promv1.ErrorType `json:"errorType"`
	Error     string           `json:"error"`
	Data      json.RawMessage  `json:"data,omitempty"`
	Warnings  []string         `json:"warnings,omitempty"`
}

func newPrometheusHTTPError(resp *http.Response, body []byte) error {
	apiErrors := prometheusAPIErrors(resp, body)
	httpErr := HTTPError{
		StatusCode: resp.StatusCode,
		TraceID:    resp.Header.Get(HeaderTraceID),
		APIErrors:  apiErrors,
	}
	if resp.Request != nil {
		if resp.Request.URL != nil {
			httpErr.URL = resp.Request.URL.String()
		}
		httpErr.Method = resp.Request.Method
	}
	return &httpErr
}

func prometheusAPIErrors(resp *http.Response, body []byte) APIErrors {
	if len(body) == 0 {
		return APIErrors{Errors: []APIError{{Title: "unknown error"}}}
	}
	if strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") {
		var response prometheusAPIResponse
		if err := json.Unmarshal(body, &response); err == nil && response.Status == "error" && response.Error != "" {
			return APIErrors{Errors: []APIError{{Title: prometheusErrorTitle(response)}}}
		}
	}
	apiErrors, err := newGenericAPIErrors(body)
	if err != nil {
		return APIErrors{Errors: []APIError{{Title: "unknown error"}}}
	}
	return apiErrors
}

func prometheusErrorTitle(response prometheusAPIResponse) string {
	if response.ErrorType == "" {
		return response.Error
	}
	return string(response.ErrorType) + ": " + response.Error
}
