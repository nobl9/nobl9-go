package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	internal "github.com/nobl9/nobl9-go/internal/sdk"
)

// QueryLogs retrieves exactly one page of collection logs using the caller's credentials.
// It does not follow continuation tokens automatically. An invalid or expired token
// requires restarting without PageToken; a forbidden response requires project access.
func (e endpoints) QueryLogs(ctx context.Context, request QueryLogsRequest) (response QueryLogsResponse, err error) {
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
