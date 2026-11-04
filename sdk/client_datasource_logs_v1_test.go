package sdk

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	datasourceV1 "github.com/nobl9/nobl9-go/sdk/endpoints/datasource/v1"
)

func TestClient_DataSource_V1_QueryLogs(t *testing.T) {
	query := datasourceV1.QueryLogsRequest{
		Project: "project", Name: "source",
		From:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		To:    time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC),
		Limit: 2, PageToken: "previous", Levels: []string{"ERROR"}, SearchPhrase: "timeout",
		Slos: []datasourceV1.ResourceReference{{Name: "slo", Project: "project"}},
		Sort: datasourceV1.LogSort{Column: "firstTimestamp", Direction: "desc"},
	}
	for name, want := range map[string]datasourceV1.QueryLogsResponse{
		"empty continuation": {Data: []datasourceV1.Log{}, NextPageToken: "next"},
		"exhausted":          {Data: []datasourceV1.Log{}, Complete: true},
		"incomplete":         {Data: []datasourceV1.Log{}, Incomplete: true},
		"truncated group": {
			Data:  []datasourceV1.Log{{Message: "shortened", RepeatedCount: 42, Truncated: true}},
			Total: 1, Truncated: true, Complete: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			calls := 0
			client, srv := prepareTestClient(t, endpointConfig{
				Path: "api/logs/v1/query",
				ResponseFunc: func(t *testing.T, w http.ResponseWriter) {
					require.NoError(t, json.NewEncoder(w).Encode(want))
				},
				TestRequestFunc: func(t *testing.T, r *http.Request) {
					calls++
					assert.Equal(t, http.MethodPost, r.Method)
					assert.Equal(t, query.Project, r.Header.Get("Project"))
					var got datasourceV1.QueryLogsRequest
					require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
					assert.Equal(t, query, got)
				},
			})
			t.Cleanup(srv.Close)
			got, err := client.DataSource().V1().QueryLogs(t.Context(), query)
			require.NoError(t, err)
			assert.Equal(t, want, got)
			assert.Equal(t, 1, calls)
		})
	}
}

func TestClient_DataSource_V1_QueryLogs_InvalidPageToken(t *testing.T) {
	client, srv := prepareTestClient(t, endpointConfig{
		Path: "api/logs/v1/query",
		ResponseFunc: func(t *testing.T, w http.ResponseWriter) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			response := APIErrors{Errors: []APIError{{
				Code: "InvalidPageToken", Title: "Refresh the logs to start again.",
			}}}
			require.NoError(t, json.NewEncoder(w).Encode(response))
		},
	})
	t.Cleanup(srv.Close)
	_, err := client.DataSource().V1().QueryLogs(t.Context(), datasourceV1.QueryLogsRequest{})
	var apiErr *HTTPError
	require.ErrorAs(t, err, &apiErr)
	require.Len(t, apiErr.Errors, 1)
	assert.Equal(t, "InvalidPageToken", apiErr.Errors[0].Code)
}
