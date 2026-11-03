package sdk

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	logsv1 "github.com/nobl9/nobl9-go/sdk/endpoints/logs/v1"
)

func TestClient_Logs_V1_Query(t *testing.T) {
	query := logsv1.QueryRequest{
		Project: "project", Name: "source",
		From:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		To:    time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC),
		Limit: 2, PageToken: "previous", Levels: []string{"ERROR"}, SearchPhrase: "timeout",
		Slos: []logsv1.ResourceReference{{Name: "slo", Project: "project"}},
		Sort: logsv1.Sort{Column: "firstTimestamp", Direction: "desc"},
	}
	for name, want := range map[string]logsv1.QueryResponse{
		"empty continuation": {Data: []logsv1.Log{}, NextPageToken: "next"},
		"exhausted":          {Data: []logsv1.Log{}, Complete: true},
		"incomplete":         {Data: []logsv1.Log{}, Incomplete: true},
		"truncated group": {
			Data:  []logsv1.Log{{Message: "shortened", RepeatedCount: 42, Truncated: true}},
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
					var got logsv1.QueryRequest
					require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
					assert.Equal(t, query, got)
				},
			})
			t.Cleanup(srv.Close)
			got, err := client.Logs().V1().Query(t.Context(), query)
			require.NoError(t, err)
			assert.Equal(t, want, got)
			assert.Equal(t, 1, calls)
		})
	}
}

func TestClient_Logs_V1_Query_InvalidPageToken(t *testing.T) {
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
	_, err := client.Logs().V1().Query(t.Context(), logsv1.QueryRequest{})
	var apiErr *HTTPError
	require.ErrorAs(t, err, &apiErr)
	require.Len(t, apiErr.Errors, 1)
	assert.Equal(t, "InvalidPageToken", apiErr.Errors[0].Code)
}
