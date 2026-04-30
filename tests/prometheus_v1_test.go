//go:build e2e_test

package tests

import (
	"testing"

	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	prometheusV1 "github.com/nobl9/nobl9-go/sdk/endpoints/prometheus/v1"
)

func Test_Prometheus_V1_Buildinfo(t *testing.T) {
	t.Parallel()

	buildInfo, err := client.Prometheus().V1().Buildinfo(t.Context())
	require.NoError(t, err)
	assert.NotEmpty(t, buildInfo.Version)
	assert.NotEmpty(t, buildInfo.Revision)
	assert.NotEmpty(t, buildInfo.Branch)
	assert.NotEmpty(t, buildInfo.GoVersion)
}

func Test_Prometheus_V1_LabelNames(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		request  prometheusV1.LabelNamesRequest
		expected []string
	}{
		"all labels": {
			expected: []string{
				"__name__",
				"component_objective",
				"component_project",
				"component_slo",
				"objective",
				"project",
				"service",
				"slo",
			},
		},
		"SLO metric labels": {
			request: prometheusV1.LabelNamesRequest{
				Matches: []string{"reliability"},
			},
			expected: []string{
				"__name__",
				"objective",
				"project",
				"service",
				"slo",
			},
		},
		"component metric labels": {
			request: prometheusV1.LabelNamesRequest{
				Matches: []string{"component_weight"},
			},
			expected: []string{
				"__name__",
				"component_objective",
				"component_project",
				"component_slo",
				"objective",
				"project",
				"service",
				"slo",
			},
		},
		"unknown metric": {
			request: prometheusV1.LabelNamesRequest{
				Matches: []string{"definitely_unknown_metric"},
			},
			expected: []string{},
		},
		"unsupported label matcher": {
			request: prometheusV1.LabelNamesRequest{
				Matches: []string{`reliability{component_slo="component"}`},
			},
			expected: []string{},
		},
		"limit": {
			request: prometheusV1.LabelNamesRequest{
				Options: []promv1.Option{promv1.WithLimit(3)},
			},
			expected: []string{
				"__name__",
				"component_objective",
				"component_project",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			labelNames, warnings, err := client.Prometheus().V1().LabelNames(t.Context(), tt.request)

			require.NoError(t, err)
			assert.Empty(t, warnings)
			assert.Equal(t, tt.expected, labelNames)
		})
	}
}
