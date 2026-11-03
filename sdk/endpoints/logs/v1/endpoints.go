package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	endpointsHelpers "github.com/nobl9/nobl9-go/internal/endpoints"
	internal "github.com/nobl9/nobl9-go/internal/sdk"
)

//go:generate ../../../../bin/ifacemaker -y " " -f ./*.go -s endpoints -i Endpoints -o endpoints_interface.go -p "$GOPACKAGE"

// NewEndpoints returns the Logs v1 endpoints.
func NewEndpoints(client endpointsHelpers.Client) Endpoints { return endpoints{client: client} }

type endpoints struct{ client endpointsHelpers.Client }

// Query retrieves exactly one page of collection logs using the caller's credentials.
// It does not follow continuation tokens automatically. An invalid or expired token
// requires restarting without PageToken; a forbidden response requires project access.
func (e endpoints) Query(ctx context.Context, request QueryRequest) (response QueryResponse, err error) {
	buf := new(bytes.Buffer)
	if err = json.NewEncoder(buf).Encode(request); err != nil {
		return response, fmt.Errorf("encode logs query: %w", err)
	}
	headers := http.Header{internal.HeaderProject: []string{request.Project}}
	req, err := e.client.CreateRequest(ctx, http.MethodPost, "logs/v1/query", headers, nil, buf)
	if err != nil {
		return response, err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return response, err
	}
	defer func() { _ = resp.Body.Close() }()
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, fmt.Errorf("decode logs response: %w", err)
	}
	return response, nil
}
