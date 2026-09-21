package v1_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	"github.com/nobl9/nobl9-go/sdk"
	slianalyzerv1 "github.com/nobl9/nobl9-go/sdk/endpoints/slianalyzer/v1"
)

func TestSLIAnalyzer_CanceledRequests(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name    string
		request func(context.Context, *sdk.Client) error
	}{
		{
			name: "create with malformed count metrics",
			request: func(ctx context.Context, client *sdk.Client) error {
				_, err := client.SLIAnalyzer().V1().CreateAnalysis(ctx, slianalyzerv1.CreateAnalysisRequest{
					Metadata: slianalyzerv1.AnalysisMetadata{DisplayName: "Analysis", Project: "test"},
					MetricSpec: slianalyzerv1.AnalysisMetricSpec{
						Kind: manifest.KindAgent, MetricSource: "prometheus",
						CountMetrics: &v1alphaSLO.CountMetricsSpec{
							GoodMetric:  &v1alphaSLO.MetricSpec{BigQuery: &v1alphaSLO.BigQueryMetric{}},
							TotalMetric: &v1alphaSLO.MetricSpec{},
						},
					},
					Period: slianalyzerv1.AnalysisPeriod{
						StartTime: "2026-09-01 00:00:00", EndTime: "2026-09-01 00:05:00", TimeZone: "UTC",
					},
				})
				return err
			},
		},
		{
			name: "update with missing fields",
			request: func(ctx context.Context, client *sdk.Client) error {
				return client.SLIAnalyzer().V1().UpdateAnalysis(ctx, "analysis", slianalyzerv1.UpdateAnalysisRequest{})
			},
		},
		{
			name: "calculation with invalid target",
			request: func(ctx context.Context, client *sdk.Client) error {
				return client.SLIAnalyzer().V1().CreateCalculation(ctx, "test", "analysis",
					slianalyzerv1.CreateCalculationRequest{BudgetTarget: 1, BudgetingMethod: "Occurrences", Operator: "lte"},
				)
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client, err := sdk.NewClient(&sdk.Config{
				DisableOkta: true, Organization: "test", Project: "test", Timeout: time.Second,
				URL: &url.URL{Scheme: "https", Host: "sli-analyzer.invalid"},
			})
			require.NoError(t, err)
			ctx, cancel := context.WithCancel(t.Context())
			cancel()
			require.ErrorIs(t, tt.request(ctx, client), context.Canceled)
		})
	}
}
