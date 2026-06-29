package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	endpointsHelpers "github.com/nobl9/nobl9-go/internal/endpoints"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
)

const (
	queryAPIPath = "agentcommander/v2/commands/timeseries/execute"
)

//go:generate ../../../../bin/ifacemaker -y " " -f ./*.go -s endpoints -i Endpoints -o endpoints_interface.go -p "$GOPACKAGE"

func NewEndpoints(client endpointsHelpers.Client) Endpoints {
	return endpoints{client: client}
}

type endpoints struct {
	client endpointsHelpers.Client
}

func (e endpoints) Query(ctx context.Context, request QueryRequest) (response QueryResponse, err error) {
	buf := new(bytes.Buffer)
	if err = json.NewEncoder(buf).Encode(newExecuteRequest(request)); err != nil {
		return response, errors.Wrap(err, "failed to encode request body")
	}
	req, err := e.client.CreateRequest(ctx, http.MethodPost, queryAPIPath, nil, nil, buf)
	if err != nil {
		return response, err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return response, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, errors.Wrap(err, "failed to decode response body")
	}
	return response, nil
}

func newExecuteRequest(request QueryRequest) executeRequest {
	return executeRequest{
		DataSource: request.DataSource,
		Command: commandRequest{
			Payload: commandPayload{
				RawMetric:    request.Query.RawMetric,
				CountMetrics: request.Query.CountMetrics,
				TimeRange:    request.TimeRange,
			},
		},
	}
}

type executeRequest struct {
	DataSource v1alphaSLO.MetricSourceSpec `json:"datasource"`
	Command    commandRequest              `json:"command"`
}

type commandRequest struct {
	Payload commandPayload `json:"payload"`
}

type commandPayload struct {
	RawMetric    *v1alphaSLO.RawMetricSpec    `json:"rawMetric,omitempty"`
	CountMetrics *v1alphaSLO.CountMetricsSpec `json:"countMetrics,omitempty"`
	TimeRange    TimeRange                    `json:"timeRange"`
}
