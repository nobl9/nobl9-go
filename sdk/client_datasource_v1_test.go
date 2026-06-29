package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	datasourceV1 "github.com/nobl9/nobl9-go/sdk/endpoints/datasource/v1"
)

func TestClient_DataSource_V1_Query(t *testing.T) {
	from := time.Date(2024, 5, 15, 12, 0, 0, 0, time.UTC)
	to := time.Date(2024, 5, 15, 12, 5, 0, 0, time.UTC)
	promQL := "up"
	expectedData := datasourceV1.QueryResponse{
		TimeSeries: []datasourceV1.TimeSeries{
			{
				Measurement: "raw",
				Timestamps:  []int64{1715760000, 1715760060},
				Values:      []float64{0.99, 0.98},
			},
		},
	}

	client, srv := prepareTestClient(t, endpointConfig{
		Path: "api/agentcommander/v2/commands/timeseries/execute",
		ResponseFunc: func(t *testing.T, w http.ResponseWriter) {
			require.NoError(t, json.NewEncoder(w).Encode(expectedData))
		},
		TestRequestFunc: func(t *testing.T, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)

			var request struct {
				DataSource v1alphaSLO.MetricSourceSpec `json:"datasource"`
				Command    struct {
					Payload struct {
						RawMetric    *v1alphaSLO.RawMetricSpec    `json:"rawMetric,omitempty"`
						CountMetrics *v1alphaSLO.CountMetricsSpec `json:"countMetrics,omitempty"`
						TimeRange    datasourceV1.TimeRange       `json:"timeRange"`
					} `json:"payload"`
				} `json:"command"`
			}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
			assert.Equal(t, v1alphaSLO.MetricSourceSpec{
				Name:    "prometheus-direct",
				Project: "default",
				Kind:    manifest.KindDirect,
			}, request.DataSource)
			require.NotNil(t, request.Command.Payload.RawMetric)
			require.NotNil(t, request.Command.Payload.RawMetric.MetricQuery)
			require.NotNil(t, request.Command.Payload.RawMetric.MetricQuery.Prometheus)
			require.NotNil(t, request.Command.Payload.RawMetric.MetricQuery.Prometheus.PromQL)
			assert.Equal(t, promQL, *request.Command.Payload.RawMetric.MetricQuery.Prometheus.PromQL)
			assert.Nil(t, request.Command.Payload.CountMetrics)
			assert.Equal(t, from, request.Command.Payload.TimeRange.From)
			assert.Equal(t, to, request.Command.Payload.TimeRange.To)
		},
	})
	defer srv.Close()

	response, err := client.DataSource().V1().Query(context.Background(), datasourceV1.QueryRequest{
		DataSource: v1alphaSLO.MetricSourceSpec{
			Name:    "prometheus-direct",
			Project: "default",
			Kind:    manifest.KindDirect,
		},
		Query: datasourceV1.Query{
			RawMetric: &v1alphaSLO.RawMetricSpec{
				MetricQuery: &v1alphaSLO.MetricSpec{
					Prometheus: &v1alphaSLO.PrometheusMetric{
						PromQL: &promQL,
					},
				},
			},
		},
		TimeRange: datasourceV1.TimeRange{
			From: from,
			To:   to,
		},
	})
	require.NoError(t, err)
	assert.Equal(t, expectedData, response)
}

func TestClient_DataSource_V1_Query_ReturnsAPIErrors(t *testing.T) {
	client, srv := prepareTestClient(t, endpointConfig{
		Path: "api/agentcommander/v2/commands/timeseries/execute",
		ResponseFunc: func(t *testing.T, w http.ResponseWriter) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)
			require.NoError(t, json.NewEncoder(w).Encode(APIErrors{
				Errors: []APIError{
					{
						Title:  "unsupported datasource",
						Code:   "unsupported_datasource",
						Detail: "datasource type version does not support V2 commands",
					},
				},
			}))
		},
		TestRequestFunc: func(t *testing.T, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
		},
	})
	defer srv.Close()

	_, err := client.DataSource().V1().Query(context.Background(), datasourceV1.QueryRequest{})
	require.Error(t, err)

	var httpErr *HTTPError
	require.True(t, errors.As(err, &httpErr))
	assert.Equal(t, http.StatusUnprocessableEntity, httpErr.StatusCode)
	require.Len(t, httpErr.Errors, 1)
	assert.Equal(t, APIError{
		Title:  "unsupported datasource",
		Code:   "unsupported_datasource",
		Detail: "datasource type version does not support V2 commands",
	}, httpErr.Errors[0])
}
