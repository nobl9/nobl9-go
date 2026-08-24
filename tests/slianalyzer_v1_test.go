//go:build e2e_test

package tests

import (
	"context"
	"errors"
	"fmt"
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
	test := newSLIAnalyzerTest(t)

	t.Run("create analysis", test.createAnalysis)
	t.Run("get analysis", test.getAnalysis)
	t.Run("list analyses", test.listAnalyses)
	t.Run("update analysis", test.updateAnalysis)
	t.Run("wait for imported data", test.waitForImportedData)
	t.Run("get timeseries", test.getTimeseries)
	t.Run("get statistics", test.getStats)
	t.Run("get histogram", test.getHistogram)
	t.Run("create calculation", test.createCalculation)
	t.Run("get calculation", test.getCalculation)
	t.Run("get summary", test.getSummary)
	t.Run("generate SLO", test.generateSLO)
	t.Run("delete analysis", test.deleteAnalysis)
}

type sliAnalyzerTest struct {
	project      string
	metricSource string
	displayName  string
	analysis     slianalyzerV1.Analysis
	deleted      bool
}

func newSLIAnalyzerTest(t *testing.T) *sliAnalyzerTest {
	t.Helper()

	direct := e2etestutils.ProvisionStaticDirect(t, v1alpha.Prometheus)
	test := &sliAnalyzerTest{
		project:      direct.Metadata.Project,
		metricSource: direct.Metadata.Name,
		displayName:  e2etestutils.GenerateName(),
	}
	cleanupContext := context.WithoutCancel(t.Context())
	t.Cleanup(func() { test.cleanup(t, cleanupContext) })
	return test
}

func (s *sliAnalyzerTest) createAnalysis(t *testing.T) {
	promQL := "(vector(1) and (vector(time() % 300) < 120)) or vector(0)"
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
			MetricSpec: slianalyzerV1.AnalysisMetricSpec{
				Kind:         manifest.KindDirect,
				MetricSource: s.metricSource,
				RawMetric: &v1alphaSLO.MetricSpec{
					Prometheus: &v1alphaSLO.PrometheusMetric{PromQL: &promQL},
				},
			},
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

func (s *sliAnalyzerTest) waitForImportedData(t *testing.T) {
	s.analysis = waitForSLIAnalysisStatus(t, s.analysis, map[slianalyzerV1.Status]bool{
		slianalyzerV1.StatusImportCompleted: true,
		slianalyzerV1.StatusImportFailed:    true,
	})
	require.Equal(t, slianalyzerV1.StatusImportCompleted, s.analysis.Status)
}

func (s *sliAnalyzerTest) getTimeseries(t *testing.T) {
	timeseries, err := client.SLIAnalyzer().V1().GetTimeseries(
		t.Context(),
		s.analysis.Metadata.Project,
		s.analysis.Metadata.Name,
	)
	require.NoError(t, err)
	assert.NotEmpty(t, timeseries.RawMetric)
}

func (s *sliAnalyzerTest) getStats(t *testing.T) {
	stats, err := client.SLIAnalyzer().V1().GetStats(
		t.Context(),
		s.analysis.Metadata.Project,
		s.analysis.Metadata.Name,
	)
	require.NoError(t, err)
	assert.NotNil(t, stats.RawMetric)
}

func (s *sliAnalyzerTest) getHistogram(t *testing.T) {
	histogram, err := client.SLIAnalyzer().V1().GetHistogram(
		t.Context(),
		s.analysis.Metadata.Project,
		s.analysis.Metadata.Name,
	)
	require.NoError(t, err)
	assert.NotEmpty(t, histogram.Bins)
}

func (s *sliAnalyzerTest) createCalculation(t *testing.T) {
	require.NoError(t, client.SLIAnalyzer().V1().CreateCalculation(
		t.Context(),
		s.analysis.Metadata.Project,
		s.analysis.Metadata.Name,
		slianalyzerV1.CreateCalculationRequest{
			Value:           130,
			BudgetTarget:    0.99,
			TimeSliceTarget: 0.99,
			BudgetingMethod: v1alphaSLO.BudgetingMethodTimeslices.String(),
			Operator:        v1alpha.GreaterThanEqual.String(),
		},
	))
	s.analysis = waitForSLIAnalysisStatus(t, s.analysis, map[slianalyzerV1.Status]bool{
		slianalyzerV1.StatusCalculationCompleted: true,
		slianalyzerV1.StatusCalculationFailed:    true,
	})
	require.Equal(t, slianalyzerV1.StatusCalculationCompleted, s.analysis.Status)
}

func (s *sliAnalyzerTest) getCalculation(t *testing.T) {
	calculation, err := client.SLIAnalyzer().V1().GetCalculation(
		t.Context(),
		s.analysis.Metadata.Project,
		s.analysis.Metadata.Name,
	)
	require.NoError(t, err)
	assert.NotEmpty(t, calculation)
}

func (s *sliAnalyzerTest) getSummary(t *testing.T) {
	_, err := client.SLIAnalyzer().V1().GetSummary(
		t.Context(),
		s.analysis.Metadata.Project,
		s.analysis.Metadata.Name,
	)
	require.NoError(t, err)
}

func (s *sliAnalyzerTest) generateSLO(t *testing.T) {
	slo, err := client.SLIAnalyzer().V1().GenerateSLO(
		t.Context(),
		s.analysis.Metadata.Project,
		s.analysis.Metadata.Name,
	)
	require.NoError(t, err)
	assert.Equal(t, manifest.KindSLO, slo.Kind)
	assert.Equal(t, s.analysis.Metadata.Project, slo.Metadata.Project)
	assert.NotEmpty(t, slo.Spec.Objectives)
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

func waitForSLIAnalysisStatus(
	t *testing.T,
	analysis slianalyzerV1.Analysis,
	terminalStatuses map[slianalyzerV1.Status]bool,
) slianalyzerV1.Analysis {
	t.Helper()

	response, err := tryExecuteRequest(t, func() (slianalyzerV1.Analysis, error) {
		response, err := client.SLIAnalyzer().V1().GetAnalysis(
			t.Context(),
			analysis.Metadata.Project,
			analysis.Metadata.Name,
		)
		if err != nil {
			return response, err
		}
		if !terminalStatuses[response.Status] {
			return response, fmt.Errorf(
				"analysis %q has status %q",
				analysis.Metadata.Name,
				response.Status,
			)
		}
		return response, nil
	})
	require.NoError(t, err)
	return response
}
