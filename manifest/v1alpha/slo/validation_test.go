package slo

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

//go:embed test_data/expected_metadata_error.txt
var expectedMetadataError string

func TestValidate_Metadata(t *testing.T) {
	err := validate(SLO{
		Kind: manifest.KindSLO,
		Metadata: Metadata{
			Name:        strings.Repeat("MY SLO", 20),
			DisplayName: strings.Repeat("my-slo", 10),
			Project:     strings.Repeat("MY PROJECT", 20),
			Labels: v1alpha.Labels{
				"L O L": []string{"dip", "dip"},
			},
		},
		Spec: Spec{
			Description: strings.Repeat("l", 2000),
		},
		ManifestSource: "/home/me/slo.yaml",
	})
	assert.ErrorContains(t, err, expectedMetadataError)
}

func TestValidate_Spec_BudgetingMethod(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validSLO()
		for _, method := range []string{
			BudgetingMethodOccurrences.String(),
			BudgetingMethodTimeslices.String(),
		} {
			slo.Spec.BudgetingMethod = method
			err := validate(slo)
			assert.NoError(t, err)
		}
	})
	t.Run("empty method", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.BudgetingMethod = ""
		err := validate(slo)
		assertContainsErrors(t, err, validation.ErrorCodeRequired)
	})
	t.Run("invalid method", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.BudgetingMethod = "invalid"
		err := validate(slo)
		assertContainsErrors(t, err, "'invalid' is not a valid budgeting method")
	})
}

func assertContainsErrors(t *testing.T, err error, expectedErrors ...string) {
	t.Helper()
	require.Error(t, err)
	var objErr *v1alpha.ObjectError
	require.ErrorAs(t, err, &objErr)
	require.Len(t,
		objErr.Errors,
		len(expectedErrors),
		"v1alpha.ObjectError contains a different number of errors than expected")
	for _, expectedErr := range expectedErrors {
		found := false
		for _, ruleErr := range objErr.Errors {
			if strings.Contains(ruleErr.Error(), expectedErr) || validation.HasErrorCode(ruleErr, expectedErr) {
				found = true
			}
		}
		require.Truef(t, found, "expected '%s' error was not found", expectedErr)
	}
}

func validSLO() SLO {
	return New(
		Metadata{
			Name:        "my-slo",
			DisplayName: "My SLO",
			Project:     "default",
			Labels: v1alpha.Labels{
				"team":   []string{"green", "orange"},
				"region": []string{"eu-central-1"},
			},
		},
		Spec{
			Description:   "Example slo",
			AlertPolicies: []string{"my-policy-name"},
			Attachments: []Attachment{
				{
					DisplayName: ptr("Grafana Dashboard"),
					URL:         "https://loki.my-org.dev/grafana/d/dnd48",
				},
			},
			BudgetingMethod: BudgetingMethodOccurrences.String(),
			Service:         "prometheus",
			Indicator: Indicator{
				MetricSource: MetricSourceSpec{
					Project: "default",
					Name:    "prometheus",
					Kind:    manifest.KindAgent,
				},
			},
			Objectives: []Objective{
				{
					ObjectiveBase: ObjectiveBase{
						DisplayName: "",
						Value:       0,
						Name:        "",
						NameChanged: false,
					},
					BudgetTarget: ptr(0.9),
					CountMetrics: &CountMetricsSpec{
						Incremental: ptr(false),
						GoodMetric: &MetricSpec{
							Prometheus: &PrometheusMetric{
								PromQL: ptr(`sum(rate(prometheus_http_requests_total{code=~"^2.*"}[1h]))`),
							},
						},
						TotalMetric: &MetricSpec{
							Prometheus: &PrometheusMetric{
								PromQL: ptr(`sum(rate(prometheus_http_requests_total[1h]))`),
							},
						},
					},
				},
			},
			TimeWindows: []TimeWindow{
				{
					Unit:      "Day",
					Count:     1,
					IsRolling: true,
				},
			},
		},
	)
}

func ptr[T any](v T) *T { return &v }
