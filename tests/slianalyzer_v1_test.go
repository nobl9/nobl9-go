//go:build e2e_test

package tests

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/twindow"
	"github.com/nobl9/nobl9-go/sdk"
	slianalyzerV1 "github.com/nobl9/nobl9-go/sdk/endpoints/slianalyzer/v1"
	"github.com/nobl9/nobl9-go/tests/e2etestutils"
)

func Test_SLIAnalyzer_V1(t *testing.T) {
	tests := []struct {
		name       string
		metricSpec slianalyzerV1.AnalysisMetricSpec
	}{
		{
			name: "raw metric",
			metricSpec: slianalyzerV1.AnalysisMetricSpec{
				RawMetric: &v1alphaSLO.MetricSpec{
					Prometheus: &v1alphaSLO.PrometheusMetric{
						PromQL: ptr("vector(1)"),
					},
				},
			},
		},
		{
			name: "count metrics",
			metricSpec: slianalyzerV1.AnalysisMetricSpec{
				CountMetrics: &v1alphaSLO.CountMetricsSpec{
					Incremental: ptr(true),
					GoodMetric: &v1alphaSLO.MetricSpec{
						Prometheus: &v1alphaSLO.PrometheusMetric{PromQL: ptr("vector(9)")},
					},
					TotalMetric: &v1alphaSLO.MetricSpec{
						Prometheus: &v1alphaSLO.PrometheusMetric{PromQL: ptr("vector(10)")},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test := newSLIAnalyzerTest(t, tt.metricSpec)
			for _, step := range []struct {
				name string
				run  func(*testing.T)
			}{
				{"create analysis", test.createAnalysis},
				{"get analysis", test.getAnalysis},
				{"list analyses", test.listAnalyses},
				{"update analysis", test.updateAnalysis},
				{"delete analysis", test.deleteAnalysis},
			} {
				if !t.Run(step.name, step.run) {
					return
				}
			}
		})
	}
}

type sliAnalyzerTest struct {
	project     string
	metricSpec  slianalyzerV1.AnalysisMetricSpec
	displayName string
	analysis    slianalyzerV1.Analysis
	deleted     bool
}

func newSLIAnalyzerTest(t *testing.T, metricSpec slianalyzerV1.AnalysisMetricSpec) *sliAnalyzerTest {
	t.Helper()

	agent := e2etestutils.ProvisionStaticAgent(t, v1alpha.Prometheus)
	metricSpec.Kind = manifest.KindAgent
	metricSpec.MetricSource = agent.Metadata.Name
	test := &sliAnalyzerTest{
		project:     agent.Metadata.Project,
		metricSpec:  metricSpec,
		displayName: e2etestutils.GenerateName(),
	}
	cleanupContext := context.WithoutCancel(t.Context())
	t.Cleanup(func() { test.cleanup(t, cleanupContext) })
	return test
}

func (s *sliAnalyzerTest) createAnalysis(t *testing.T) {
	endTime := time.Now().UTC().Truncate(time.Second)
	startTime := endTime.Add(-time.Hour)

	var err error
	s.analysis, err = client.SLIAnalyzer().V1().CreateAnalysis(
		t.Context(),
		slianalyzerV1.CreateAnalysisRequest{
			Metadata: slianalyzerV1.AnalysisMetadata{
				DisplayName: s.displayName,
				Project:     s.project,
			},
			MetricSpec: s.metricSpec,
			Period: slianalyzerV1.AnalysisPeriod{
				StartTime: startTime.Format(twindow.IsoDateTimeOnlyLayout),
				EndTime:   endTime.Format(twindow.IsoDateTimeOnlyLayout),
				TimeZone:  time.UTC.String(),
			},
		},
	)
	require.NoError(t, err)
	require.NotEmpty(t, s.analysis.Metadata.Name)
	assert.Equal(t, s.displayName, s.analysis.Metadata.DisplayName)
	assert.Equal(t, s.project, s.analysis.Metadata.Project)
	assert.Equal(t, s.metricSpec, s.analysis.MetricSpec)
	assert.Equal(t, slianalyzerV1.StatusFetchingHistoricalData, s.analysis.Status)
}

func (s *sliAnalyzerTest) getAnalysis(t *testing.T) {
	actual, err := client.SLIAnalyzer().V1().GetAnalysis(
		t.Context(),
		s.analysis.Metadata.Project,
		s.analysis.Metadata.Name,
	)
	require.NoError(t, err)
	assert.Equal(t, s.analysis.Metadata, actual.Metadata)
	assert.Equal(t, s.analysis.MetricSpec, actual.MetricSpec)
	assert.Equal(t, s.analysis.Period, actual.Period)
}

func (s *sliAnalyzerTest) listAnalyses(t *testing.T) {
	analyses, err := client.SLIAnalyzer().V1().ListAnalyses(t.Context())
	require.NoError(t, err)
	assert.True(t, slices.ContainsFunc(analyses, func(item slianalyzerV1.Analysis) bool {
		return item.Metadata.Project == s.analysis.Metadata.Project &&
			item.Metadata.Name == s.analysis.Metadata.Name
	}))
}

func (s *sliAnalyzerTest) updateAnalysis(t *testing.T) {
	updatedDisplayName := s.displayName + " updated"
	require.NoError(t, client.SLIAnalyzer().V1().UpdateAnalysis(
		t.Context(),
		s.analysis.Metadata.Name,
		slianalyzerV1.UpdateAnalysisRequest{
			Project:     s.analysis.Metadata.Project,
			DisplayName: updatedDisplayName,
		},
	))
	actual, err := client.SLIAnalyzer().V1().GetAnalysis(
		t.Context(),
		s.analysis.Metadata.Project,
		s.analysis.Metadata.Name,
	)
	require.NoError(t, err)
	assert.Equal(t, updatedDisplayName, actual.Metadata.DisplayName)
	s.analysis = actual
}

func (s *sliAnalyzerTest) deleteAnalysis(t *testing.T) {
	require.NoError(t, client.SLIAnalyzer().V1().DeleteAnalysis(
		t.Context(),
		s.analysis.Metadata.Project,
		s.analysis.Metadata.Name,
	))
	s.deleted = true

	_, err := client.SLIAnalyzer().V1().GetAnalysis(
		t.Context(),
		s.analysis.Metadata.Project,
		s.analysis.Metadata.Name,
	)
	require.Error(t, err)
	var httpErr *sdk.HTTPError
	require.True(t, errors.As(err, &httpErr))
	assert.Equal(t, http.StatusNotFound, httpErr.StatusCode)
}

func (s *sliAnalyzerTest) cleanup(t *testing.T, ctx context.Context) {
	t.Helper()

	if s.deleted || s.analysis.Metadata.Name == "" {
		return
	}
	assert.NoError(t, client.SLIAnalyzer().V1().DeleteAnalysis(
		ctx,
		s.analysis.Metadata.Project,
		s.analysis.Metadata.Name,
	))
}
