package sdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	prometheusV1 "github.com/nobl9/nobl9-go/sdk/endpoints/prometheus/v1"
)

func TestClient_Prometheus_Query(t *testing.T) {
	query := `reliability{project="default"}`
	ts := time.Unix(1_700_000_000, 0).UTC()
	client, srv := preparePrometheusTestClient(t, func(t *testing.T, w http.ResponseWriter, r *http.Request) {
		assertPrometheusRequest(t, r, http.MethodPost, "/api/prometheus/v1/api/v1/query")
		require.NoError(t, r.ParseForm())
		assert.Equal(t, query, r.Form.Get("query"))
		assert.Equal(t, "1700000000", r.Form.Get("time"))
		assert.Equal(t, "30s", r.Form.Get("timeout"))
		assert.Equal(t, "1m0s", r.Form.Get("lookback_delta"))
		assert.Equal(t, "10", r.Form.Get("limit"))
		writePrometheusResponse(t, w, map[string]any{
			"resultType": "vector",
			"result": []any{
				map[string]any{
					"metric": map[string]string{
						"__name__": "reliability",
						"project":  "default",
					},
					"value": []any{float64(1_700_000_000), "1"},
				},
			},
		}, []string{"slow query"})
	})
	defer srv.Close()

	value, warnings, err := client.Prometheus().V1().Query(
		context.Background(),
		prometheusV1.QueryRequest{
			Query: query,
			Time:  ts,
			Options: []promv1.Option{
				promv1.WithTimeout(30 * time.Second),
				promv1.WithLookbackDelta(time.Minute),
				promv1.WithLimit(10),
			},
		},
	)
	require.NoError(t, err)

	assert.Equal(t, promv1.Warnings{"slow query"}, warnings)
	require.Equal(t, model.ValVector, value.Type())
	vector := value.(model.Vector)
	require.Len(t, vector, 1)
	assert.Equal(t, model.Metric{"__name__": "reliability", "project": "default"}, vector[0].Metric)
	assert.Equal(t, model.SampleValue(1), vector[0].Value)
}

func TestClient_Prometheus_QueryRange(t *testing.T) {
	query := `reliability{project="default"}`
	start := time.Unix(1_700_000_000, 0).UTC()
	end := start.Add(5 * time.Minute)
	client, srv := preparePrometheusTestClient(t, func(t *testing.T, w http.ResponseWriter, r *http.Request) {
		assertPrometheusRequest(t, r, http.MethodPost, "/api/prometheus/v1/api/v1/query_range")
		require.NoError(t, r.ParseForm())
		assert.Equal(t, query, r.Form.Get("query"))
		assert.Equal(t, "1700000000", r.Form.Get("start"))
		assert.Equal(t, "1700000300", r.Form.Get("end"))
		assert.Equal(t, "60", r.Form.Get("step"))
		assert.Equal(t, "15s", r.Form.Get("timeout"))
		assert.Equal(t, "5m0s", r.Form.Get("lookback_delta"))
		assert.Equal(t, "25", r.Form.Get("limit"))
		writePrometheusResponse(t, w, map[string]any{
			"resultType": "matrix",
			"result":     []any{},
		}, nil)
	})
	defer srv.Close()

	value, warnings, err := client.Prometheus().V1().QueryRange(
		context.Background(),
		prometheusV1.QueryRangeRequest{
			Query: query,
			Range: promv1.Range{Start: start, End: end, Step: time.Minute},
			Options: []promv1.Option{
				promv1.WithTimeout(15 * time.Second),
				promv1.WithLookbackDelta(5 * time.Minute),
				promv1.WithLimit(25),
			},
		},
	)
	require.NoError(t, err)

	assert.Empty(t, warnings)
	assert.Equal(t, model.ValMatrix, value.Type())
}

func TestClient_Prometheus_MetadataEndpoints(t *testing.T) {
	start := time.Unix(1_700_000_000, 0).UTC()
	end := start.Add(time.Minute)
	client, srv := preparePrometheusTestClient(t, func(t *testing.T, w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/prometheus/v1/api/v1/series":
			assertPrometheusRequest(t, r, http.MethodPost, r.URL.Path)
			require.NoError(t, r.ParseForm())
			assert.Equal(t, []string{"up"}, r.Form["match[]"])
			assert.Equal(t, "1700000000", r.Form.Get("start"))
			assert.Equal(t, "1700000060", r.Form.Get("end"))
			assert.Equal(t, "1", r.Form.Get("limit"))
			writePrometheusResponse(t, w, []map[string]string{{"__name__": "up", "job": "api"}}, nil)
		case "/api/prometheus/v1/api/v1/labels":
			assertPrometheusRequest(t, r, http.MethodPost, r.URL.Path)
			require.NoError(t, r.ParseForm())
			assert.Equal(t, []string{"up"}, r.Form["match[]"])
			writePrometheusResponse(t, w, []string{"__name__", "job"}, nil)
		case "/api/prometheus/v1/api/v1/label/job/values":
			assertPrometheusRequest(t, r, http.MethodGet, r.URL.Path)
			require.NoError(t, r.ParseForm())
			assert.Equal(t, []string{"up"}, r.Form["match[]"])
			writePrometheusResponse(t, w, []string{"api", "worker"}, nil)
		case "/api/prometheus/v1/api/v1/metadata":
			assertPrometheusRequest(t, r, http.MethodGet, r.URL.Path)
			assert.Equal(t, "up", r.URL.Query().Get("metric"))
			assert.Equal(t, "10", r.URL.Query().Get("limit"))
			writePrometheusResponse(t, w, map[string]any{
				"up": []map[string]string{{"type": "gauge", "help": "Up", "unit": ""}},
			}, nil)
		case "/api/prometheus/v1/api/v1/status/buildinfo":
			assertPrometheusRequest(t, r, http.MethodGet, r.URL.Path)
			writePrometheusResponse(t, w, map[string]string{
				"version":   "1.0.0",
				"revision":  "abc123",
				"branch":    "main",
				"buildUser": "nobl9",
				"buildDate": "2026-04-28T00:00:00Z",
				"goVersion": "go1.25.9",
			}, nil)
		default:
			t.Fatalf("unsupported path: %s", r.URL.Path)
		}
	})
	defer srv.Close()

	endpoints := client.Prometheus().V1()

	series, seriesWarnings, err := endpoints.Series(
		context.Background(),
		prometheusV1.SeriesRequest{
			Matches:   []string{"up"},
			StartTime: start,
			EndTime:   end,
			Options:   []promv1.Option{promv1.WithLimit(1)},
		},
	)
	require.NoError(t, err)
	assert.Empty(t, seriesWarnings)
	assert.Equal(t, []model.LabelSet{{"__name__": "up", "job": "api"}}, series)

	labelNames, labelWarnings, err := endpoints.LabelNames(
		context.Background(),
		prometheusV1.LabelNamesRequest{
			Matches:   []string{"up"},
			StartTime: start,
			EndTime:   end,
		},
	)
	require.NoError(t, err)
	assert.Empty(t, labelWarnings)
	assert.Equal(t, []string{"__name__", "job"}, labelNames)

	labelValues, labelValueWarnings, err := endpoints.LabelValues(
		context.Background(),
		prometheusV1.LabelValuesRequest{
			Label:     "job",
			Matches:   []string{"up"},
			StartTime: start,
			EndTime:   end,
		},
	)
	require.NoError(t, err)
	assert.Empty(t, labelValueWarnings)
	assert.Equal(t, model.LabelValues{"api", "worker"}, labelValues)

	metadata, err := endpoints.Metadata(
		context.Background(),
		prometheusV1.MetadataRequest{Metric: "up", Limit: "10"},
	)
	require.NoError(t, err)
	assert.Equal(t, promv1.MetricType("gauge"), metadata["up"][0].Type)
	assert.Equal(t, "Up", metadata["up"][0].Help)

	buildInfo, err := endpoints.Buildinfo(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "1.0.0", buildInfo.Version)
	assert.Equal(t, "abc123", buildInfo.Revision)
	assert.Equal(t, "go1.25.9", buildInfo.GoVersion)
}

func TestClient_Prometheus_ReturnsPrometheusError(t *testing.T) {
	client, srv := preparePrometheusTestClient(t, func(t *testing.T, w http.ResponseWriter, r *http.Request) {
		assertPrometheusRequest(t, r, http.MethodPost, "/api/prometheus/v1/api/v1/query")
		w.WriteHeader(http.StatusUnprocessableEntity)
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"status":    "error",
			"errorType": "execution",
			"error":     "invalid query",
		}))
	})
	defer srv.Close()

	_, _, err := client.Prometheus().V1().Query(
		context.Background(),
		prometheusV1.QueryRequest{Query: "invalid"},
	)
	require.Error(t, err)

	var promErr *promv1.Error
	require.ErrorAs(t, err, &promErr)
	assert.Equal(t, promv1.ErrExec, promErr.Type)
	assert.Equal(t, "invalid query", promErr.Msg)
}

func TestClient_Prometheus_UsesSDKHTTPClientAuthorization(t *testing.T) {
	client, srv := preparePrometheusTestClient(t, func(t *testing.T, w http.ResponseWriter, r *http.Request) {
		assertPrometheusRequest(t, r, http.MethodPost, "/api/prometheus/v1/api/v1/query")
		assert.Equal(t, "Bearer prom-token", r.Header.Get(HeaderAuthorization))
		writePrometheusResponse(t, w, map[string]any{
			"resultType": "vector",
			"result":     []any{},
		}, nil)
	})
	defer srv.Close()

	client.credentials.accessToken = "prom-token"
	client.credentials.claims = &jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	client.credentials.clientID = client.Config.ClientID
	client.credentials.clientSecret = client.Config.ClientSecret
	client.Config.DisableOkta = false

	_, _, err := client.Prometheus().V1().Query(context.Background(), prometheusV1.QueryRequest{Query: "up"})
	require.NoError(t, err)
}

func preparePrometheusTestClient(
	t *testing.T,
	handler func(t *testing.T, w http.ResponseWriter, r *http.Request),
) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(t, w, r)
	}))
	config, err := ReadConfig(
		ConfigOptionWithCredentials("client-id", "client-secret"),
		ConfigOptionNoConfigFile(),
	)
	require.NoError(t, err)
	config.DisableOkta = true
	config.URL = parseTestURL(t, srv.URL+"/api")
	client, err := NewClient(config)
	require.NoError(t, err)
	return client, srv
}

func assertPrometheusRequest(t *testing.T, r *http.Request, method, path string) {
	t.Helper()
	assert.Equal(t, method, r.Method)
	assert.Equal(t, path, r.URL.Path)
	assert.Empty(t, r.Header.Get(HeaderOrganization))
	assert.Empty(t, r.Header.Get(HeaderProject))
	_, _, ok := r.BasicAuth()
	assert.False(t, ok)
}

func writePrometheusResponse(t *testing.T, w http.ResponseWriter, data any, warnings []string) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
		"status":   "success",
		"data":     data,
		"warnings": warnings,
	}))
}

func parseTestURL(t *testing.T, rawURL string) *url.URL {
	t.Helper()
	u, err := url.Parse(rawURL)
	require.NoError(t, err)
	return u
}
