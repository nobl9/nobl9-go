package v1

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/pkg/errors"

	endpointsHelpers "github.com/nobl9/nobl9-go/internal/endpoints"
)

const (
	apiMCPServer = "/api/mcp"
)

//go:generate ../../../../bin/ifacemaker -y " " -f ./*.go -s endpoints -i Endpoints -o endpoints_interface.go -p "$GOPACKAGE"

func NewEndpoints(client endpointsHelpers.Client) Endpoints {
	return endpoints{client: client}
}

type endpoints struct {
	client endpointsHelpers.Client
}

// ProxyRequest forwards a request to the MCP server API with authentication headers.
// The body parameter should contain the MCP protocol message (JSON-RPC).
// The response body is returned as an io.ReadCloser for streaming - caller must close it.
// HTTP errors (status >= 300) are returned as errors. MCP protocol errors (JSON-RPC
// errors with status 200) are passed through in the response body.
func (e endpoints) ProxyRequest(ctx context.Context, body []byte) (io.ReadCloser, error) {
	req, err := e.client.CreateRequest(
		ctx,
		http.MethodPost,
		apiMCPServer,
		nil,
		nil,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create MCP proxy request")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "MCP server request failed")
	}

	return resp.Body, nil
}
