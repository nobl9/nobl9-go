package sdk

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	promapi "github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"

	prometheusEndpoints "github.com/nobl9/nobl9-go/sdk/endpoints/prometheus"
	prometheusEndpointsV1 "github.com/nobl9/nobl9-go/sdk/endpoints/prometheus/v1"
)

var errMissingPrometheusCredentials = errors.New(
	"client id and client secret must be provided to use Prometheus API",
)

// Prometheus is used to access specific Prometheus-compatible API version.
func (c *Client) Prometheus() prometheusEndpoints.Versions {
	return prometheusEndpoints.NewVersions(c.newPrometheusAPI)
}

func (c *Client) newPrometheusAPI(ctx context.Context) (prometheusEndpointsV1.API, error) {
	if c.Config.ClientID == "" || c.Config.ClientSecret == "" {
		return nil, errMissingPrometheusCredentials
	}
	apiURL, err := c.getAPIURL(ctx)
	if err != nil {
		return nil, err
	}
	client, err := promapi.NewClient(promapi.Config{
		Address: apiURL.JoinPath("prometheus", "v1").String(),
		Client: newRetryableHTTPClient(
			c.Config.Timeout,
			prometheusBasicAuthRoundTripper{
				clientID:     c.Config.ClientID,
				clientSecret: c.Config.ClientSecret,
				userAgent:    c.userAgent,
			},
		),
	})
	if err != nil {
		return nil, err
	}
	return promv1.NewAPI(client), nil
}

type prometheusBasicAuthRoundTripper struct {
	clientID     string
	clientSecret string
	userAgent    string
	base         http.RoundTripper
}

func (r prometheusBasicAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if r.clientID == "" || r.clientSecret == "" {
		return nil, httpNonRetryableError{Err: errMissingPrometheusCredentials}
	}
	cloned := req.Clone(req.Context())
	cloned.SetBasicAuth(r.clientID, r.clientSecret)
	cloned.Header.Set(HeaderUserAgent, r.userAgent)
	base := r.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(cloned)
}
