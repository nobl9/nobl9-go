package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"

	"github.com/goccy/go-yaml"

	endpointsHelpers "github.com/nobl9/nobl9-go/internal/endpoints"
	internalSDK "github.com/nobl9/nobl9-go/internal/sdk"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
)

const analysesAPIPath = "sli-analyses"

//go:generate ../../../../bin/ifacemaker -y " " -f ./*.go -s endpoints -i Endpoints -o endpoints_interface.go -p "$GOPACKAGE"

// NewEndpoints returns the SLI Analyzer v1 endpoints.
func NewEndpoints(client endpointsHelpers.Client) Endpoints {
	return endpoints{client: client}
}

type endpoints struct {
	client endpointsHelpers.Client
}

// CreateAnalysis creates an SLI analysis and starts data processing.
func (e endpoints) CreateAnalysis(
	ctx context.Context,
	request CreateAnalysisRequest,
) (response Analysis, err error) {
	buf := new(bytes.Buffer)
	if err = json.NewEncoder(buf).Encode(request); err != nil {
		return response, fmt.Errorf("failed to encode request body: %w", err)
	}
	req, err := e.client.CreateRequest(
		ctx,
		http.MethodPost,
		analysesAPIPath,
		http.Header{internalSDK.HeaderProject: {request.Metadata.Project}},
		nil,
		buf,
	)
	if err != nil {
		return response, err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return response, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, fmt.Errorf("failed to decode response body: %w", err)
	}
	return response, nil
}

// GetAnalysis returns an SLI analysis by project and name.
func (e endpoints) GetAnalysis(
	ctx context.Context,
	project, name string,
) (response Analysis, err error) {
	err = e.getJSONResource(
		ctx,
		project,
		path.Join(analysesAPIPath, name),
		&response,
	)
	return response, err
}

// ListAnalyses returns every SLI analysis visible to the current user.
func (e endpoints) ListAnalyses(ctx context.Context) (response []Analysis, err error) {
	req, err := e.client.CreateRequest(ctx, http.MethodGet, analysesAPIPath, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}
	return response, nil
}

// UpdateAnalysis updates the display name of an SLI analysis.
func (e endpoints) UpdateAnalysis(
	ctx context.Context,
	name string,
	request UpdateAnalysisRequest,
) error {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(request); err != nil {
		return fmt.Errorf("failed to encode request body: %w", err)
	}
	req, err := e.client.CreateRequest(
		ctx,
		http.MethodPut,
		path.Join(analysesAPIPath, name),
		http.Header{internalSDK.HeaderProject: {request.Project}},
		nil,
		buf,
	)
	if err != nil {
		return err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	return nil
}

// DeleteAnalysis deletes an SLI analysis by project and name.
func (e endpoints) DeleteAnalysis(ctx context.Context, project, name string) error {
	req, err := e.client.CreateRequest(
		ctx,
		http.MethodDelete,
		path.Join(analysesAPIPath, name),
		http.Header{internalSDK.HeaderProject: {project}},
		nil,
		nil,
	)
	if err != nil {
		return err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	return nil
}

// CreateCalculation starts a calculation for an SLI analysis.
func (e endpoints) CreateCalculation(
	ctx context.Context,
	project, name string,
	request CreateCalculationRequest,
) error {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(request); err != nil {
		return fmt.Errorf("failed to encode request body: %w", err)
	}
	req, err := e.client.CreateRequest(
		ctx,
		http.MethodPost,
		path.Join(analysesAPIPath, name, "calculation"),
		http.Header{internalSDK.HeaderProject: {project}},
		nil,
		buf,
	)
	if err != nil {
		return err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	return nil
}

// GetCalculation returns the error budget burn-down series for an SLI analysis.
func (e endpoints) GetCalculation(
	ctx context.Context,
	project, name string,
) (response Timeseries, err error) {
	err = e.getJSONResource(
		ctx,
		project,
		path.Join(analysesAPIPath, name, "calculation"),
		&response,
	)
	return response, err
}

// GetSummary returns the calculation summary for an SLI analysis.
func (e endpoints) GetSummary(
	ctx context.Context,
	project, name string,
) (response AnalysisCalculationSummary, err error) {
	err = e.getJSONResource(
		ctx,
		project,
		path.Join(analysesAPIPath, name, "summary"),
		&response,
	)
	return response, err
}

// GetTimeseries returns the aggregated metric series for an SLI analysis.
func (e endpoints) GetTimeseries(
	ctx context.Context,
	project, name string,
) (response AggregatedTimeseries, err error) {
	err = e.getJSONResource(
		ctx,
		project,
		path.Join(analysesAPIPath, name, "timeseries"),
		&response,
	)
	return response, err
}

// GetStats returns descriptive statistics for an SLI analysis.
func (e endpoints) GetStats(
	ctx context.Context,
	project, name string,
) (response AnalysisStats, err error) {
	err = e.getJSONResource(
		ctx,
		project,
		path.Join(analysesAPIPath, name, "timeseries", "stats"),
		&response,
	)
	return response, err
}

// GetHistogram returns histogram data for an SLI analysis.
func (e endpoints) GetHistogram(
	ctx context.Context,
	project, name string,
) (response AnalysisHistogram, err error) {
	err = e.getJSONResource(
		ctx,
		project,
		path.Join(analysesAPIPath, name, "timeseries", "histogram"),
		&response,
	)
	return response, err
}

// GenerateSLO returns an SLO generated from an SLI analysis.
func (e endpoints) GenerateSLO(
	ctx context.Context,
	project, name string,
) (response v1alphaSLO.SLO, err error) {
	req, err := e.client.CreateRequest(
		ctx,
		http.MethodGet,
		path.Join(analysesAPIPath, name, "yaml"),
		http.Header{internalSDK.HeaderProject: {project}},
		nil,
		nil,
	)
	if err != nil {
		return response, err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return response, err
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return response, fmt.Errorf("failed to read response body: %w", err)
	}
	if err = yaml.Unmarshal(data, &response); err != nil {
		return response, fmt.Errorf("failed to decode response body: %w", err)
	}
	return response, nil
}

func (e endpoints) getJSONResource(
	ctx context.Context,
	project, endpoint string,
	response any,
) error {
	req, err := e.client.CreateRequest(
		ctx,
		http.MethodGet,
		endpoint,
		http.Header{internalSDK.HeaderProject: {project}},
		nil,
		nil,
	)
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
