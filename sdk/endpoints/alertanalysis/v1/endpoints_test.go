package v1

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest/v1alpha/alertpolicy"
)

func TestEndpoints_StartAnalysis(t *testing.T) {
	client := &fakeClient{responseBody: `{"analysisId":"analysis-id"}`}
	endpoints := NewEndpoints(client)
	startTime := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)

	response, err := endpoints.StartAnalysis(t.Context(), StartAnalysisRequest{
		SLO:       "slo-name",
		Project:   "project-name",
		Objective: "objective-name",
		StartTime: startTime,
		EndTime:   endTime,
		AlertPolicy: alertpolicy.New(
			alertpolicy.Metadata{Name: "alert-policy", Project: "project-name"},
			alertpolicy.Spec{
				Severity: alertpolicy.SeverityHigh.String(),
				Conditions: []alertpolicy.AlertCondition{{
					Measurement:    alertpolicy.MeasurementAverageBurnRate.String(),
					Value:          2,
					AlertingWindow: "10m",
				}},
			},
		),
	})

	require.NoError(t, err)
	assert.Equal(t, "analysis-id", response.AnalysisID)
	require.Len(t, client.requests, 1)
	req := client.requests[0]
	assert.Equal(t, http.MethodPost, req.Method)
	assert.Equal(t, "/api/alerting/v1/analysis", req.URL.Path)

	var payload StartAnalysisRequest
	require.NoError(t, json.NewDecoder(req.Body).Decode(&payload))
	assert.Equal(t, "slo-name", payload.SLO)
	assert.Equal(t, "project-name", payload.Project)
	assert.Equal(t, "objective-name", payload.Objective)
	assert.Equal(t, startTime, payload.StartTime)
	assert.Equal(t, endTime, payload.EndTime)
}

func TestEndpoints_GetAnalysis(t *testing.T) {
	client := &fakeClient{
		responseBody: `{
			"alerts": [],
			"alertPolicy": {
				"apiVersion": "n9/v1alpha",
				"kind": "AlertPolicy",
				"metadata": {
					"name": "alert-policy",
					"project": "project-name"
				},
				"spec": {
					"severity": "High",
					"conditions": [{
						"measurement": "averageBurnRate",
						"value": 2,
						"alertingWindow": "10m"
					}],
					"alertMethods": null
				}
			},
			"startTime": "2026-05-01T12:00:00Z",
			"endTime": "2026-05-02T12:00:00Z",
			"status": "done",
			"detectionStatus": "ready",
			"timeseriesStatus": "ready",
			"timeseries": [{
				"measurement": "remaining_budget",
				"timestamps": [1, 2],
				"values": [99.9, 99.8],
				"attributes": {"objective": "objective-name"}
			}]
		}`,
	}
	endpoints := NewEndpoints(client)
	from := time.Date(2026, 5, 1, 13, 0, 0, 123, time.UTC)
	to := time.Date(2026, 5, 1, 14, 0, 0, 456, time.UTC)
	includeTimeseries := false

	response, err := endpoints.GetAnalysis(t.Context(), GetAnalysisRequest{
		AnalysisID:        "analysis-id",
		From:              &from,
		To:                &to,
		IncludeTimeseries: &includeTimeseries,
	})

	require.NoError(t, err)
	assert.Equal(t, AlertAnalysisStatusDone, response.Status)
	assert.Equal(t, AnalysisReadinessStatusReady, response.DetectionStatus)
	assert.Equal(t, AnalysisReadinessStatusReady, response.TimeseriesStatus)
	require.Len(t, response.Timeseries, 1)
	assert.Equal(t, "remaining_budget", response.Timeseries[0].Measurement)

	require.Len(t, client.requests, 1)
	req := client.requests[0]
	assert.Equal(t, http.MethodGet, req.Method)
	assert.Equal(t, "/api/alerting/v1/analysis/analysis-id", req.URL.Path)
	assert.Equal(t, from.Format(time.RFC3339Nano), req.URL.Query().Get("from"))
	assert.Equal(t, to.Format(time.RFC3339Nano), req.URL.Query().Get("to"))
	assert.Equal(t, "false", req.URL.Query().Get("includeTimeseries"))
}

func TestEndpoints_GetAnalysisDoesNotSetOptionalQueryParams(t *testing.T) {
	client := &fakeClient{responseBody: `{
		"alerts": [],
		"alertPolicy": {
			"apiVersion": "n9/v1alpha",
			"kind": "AlertPolicy",
			"metadata": {
				"name": "alert-policy",
				"project": "project-name"
			},
			"spec": {
				"severity": "High",
				"conditions": [{
					"measurement": "averageBurnRate",
					"value": 2,
					"alertingWindow": "10m"
				}],
				"alertMethods": null
			}
		},
		"startTime": "2026-05-01T12:00:00Z",
		"endTime": "2026-05-02T12:00:00Z",
		"status": "calculating_alerts",
		"detectionStatus": "running",
		"timeseriesStatus": "pending"
	}`}
	endpoints := NewEndpoints(client)

	_, err := endpoints.GetAnalysis(t.Context(), GetAnalysisRequest{AnalysisID: "analysis-id"})

	require.NoError(t, err)
	require.Len(t, client.requests, 1)
	assert.Empty(t, client.requests[0].URL.Query())
}

func TestEndpoints_RetryAnalysis(t *testing.T) {
	client := &fakeClient{responseBody: `{"analysisId":"retried-analysis-id"}`}
	endpoints := NewEndpoints(client)

	response, err := endpoints.RetryAnalysis(t.Context(), "analysis-id")

	require.NoError(t, err)
	assert.Equal(t, "retried-analysis-id", response.AnalysisID)
	require.Len(t, client.requests, 1)
	req := client.requests[0]
	assert.Equal(t, http.MethodPost, req.Method)
	assert.Equal(t, "/api/alerting/v1/analysis/analysis-id/retry", req.URL.Path)
}

type fakeClient struct {
	requests     []*http.Request
	responseBody string
}

func (f *fakeClient) CreateRequest(
	ctx context.Context,
	method, endpoint string,
	headers http.Header,
	q url.Values,
	body io.Reader,
) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, "https://example.com/api/"+endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header = headers
	if q != nil {
		req.URL.RawQuery = q.Encode()
	}
	return req, nil
}

func (f *fakeClient) Do(req *http.Request) (*http.Response, error) {
	f.requests = append(f.requests, req)
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(f.responseBody)),
	}, nil
}
