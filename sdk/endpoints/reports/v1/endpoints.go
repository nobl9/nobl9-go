package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	endpointsHelpers "github.com/nobl9/nobl9-go/internal/endpoints"
	v1alphaReport "github.com/nobl9/nobl9-go/manifest/v1alpha/report"
)

const reportsAPIPath = "reports/v1"

//go:generate ../../../../bin/ifacemaker -y " " -f ./*.go -s endpoints -i Endpoints -o endpoints_interface.go -p "$GOPACKAGE"

// NewEndpoints returns the Reports v1 endpoints.
func NewEndpoints(client endpointsHelpers.Client) Endpoints {
	return endpoints{client: client}
}

type endpoints struct {
	client endpointsHelpers.Client
}

// GetUsageSummary returns resource usage and quota status for the current organization.
func (e endpoints) GetUsageSummary(ctx context.Context) (response UsageSummary, err error) {
	err = e.requestJSON(ctx, http.MethodGet, path.Join(reportsAPIPath, "usage-summary"), nil, nil, &response)
	return response, err
}

// GetReliabilityRollup returns data for a saved Reliability Roll-up report identified by its metadata name.
func (e endpoints) GetReliabilityRollup(
	ctx context.Context,
	name string,
	params ReliabilityRollupRequest,
) (response ReliabilityRollup, err error) {
	err = e.requestJSON(ctx, http.MethodGet,
		path.Join(reportsAPIPath, "report", name, "reliability-rollup"), params.queryValues(), nil, &response)
	return response, err
}

// GenerateReliabilityRollup returns data for a Report manifest without saving or updating a report.
func (e endpoints) GenerateReliabilityRollup(
	ctx context.Context,
	report v1alphaReport.Report,
	params ReliabilityRollupRequest,
) (response ReliabilityRollup, err error) {
	buf := new(bytes.Buffer)
	if err = json.NewEncoder(buf).Encode(report); err != nil {
		return response, fmt.Errorf("failed to encode request body: %w", err)
	}
	err = e.requestJSON(ctx, http.MethodPost,
		path.Join(reportsAPIPath, "report", "reliability-rollup"), params.queryValues(), buf, &response)
	return response, err
}

func (e endpoints) requestJSON(
	ctx context.Context,
	method, endpoint string,
	query url.Values,
	body io.Reader,
	response any,
) error {
	req, err := e.client.CreateRequest(ctx, method, endpoint, nil, query, body)
	if err != nil {
		return err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if err = json.NewDecoder(resp.Body).Decode(response); err != nil {
		return fmt.Errorf("failed to decode response body: %w", err)
	}
	return nil
}
