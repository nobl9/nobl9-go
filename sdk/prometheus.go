package sdk

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"

	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"

	prometheusEndpoints "github.com/nobl9/nobl9-go/sdk/endpoints/prometheus"
	prometheusEndpointsV1 "github.com/nobl9/nobl9-go/sdk/endpoints/prometheus/v1"
)

// Prometheus is used to access specific Prometheus-compatible API version.
func (c *Client) Prometheus() prometheusEndpoints.Versions {
	return prometheusEndpoints.NewVersions(c.newPrometheusAPI)
}

func (c *Client) newPrometheusAPI(context.Context) (prometheusEndpointsV1.API, error) {
	c.prometheusAPI.mu.Lock()
	defer c.prometheusAPI.mu.Unlock()

	if c.prometheusAPI.api != nil {
		return c.prometheusAPI.api, nil
	}
	api := promv1.NewAPI(prometheusClient{client: c})
	c.prometheusAPI.api = api
	return api, nil
}

type prometheusAPIStore struct {
	mu  sync.Mutex
	api prometheusEndpointsV1.API
}

type prometheusClient struct {
	client *Client
}

func (c prometheusClient) URL(ep string, args map[string]string) *url.URL {
	p := path.Join("prometheus", "v1", ep)
	for arg, val := range args {
		p = strings.ReplaceAll(p, ":"+arg, val)
	}
	return &url.URL{Path: p}
}

func (c prometheusClient) Do(ctx context.Context, req *http.Request) (*http.Response, []byte, error) {
	requestCtx := ctx
	if requestCtx == nil {
		requestCtx = req.Context()
	}
	if requestCtx == nil {
		requestCtx = context.Background()
	}
	sdkReq, err := c.client.CreateRequest(
		requestCtx,
		req.Method,
		req.URL.Path,
		req.Header.Clone(),
		req.URL.Query(),
		req.Body,
	)
	if err != nil {
		return nil, nil, err
	}
	sdkReq.GetBody = req.GetBody
	sdkReq.ContentLength = req.ContentLength
	resp, err := c.client.HTTP.Do(sdkReq)
	if err != nil {
		return nil, nil, err
	}

	var body []byte
	done := make(chan error, 1)
	go func() {
		var buf bytes.Buffer
		_, readErr := buf.ReadFrom(resp.Body)
		body = buf.Bytes()
		done <- readErr
	}()

	select {
	case <-requestCtx.Done():
		_ = resp.Body.Close()
		<-done
		return resp, nil, requestCtx.Err()
	case err := <-done:
		_ = resp.Body.Close()
		return resp, body, err
	}
}
