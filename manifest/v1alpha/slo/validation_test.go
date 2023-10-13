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
	slo := validSLO()
	slo.Metadata = Metadata{
		Name:        strings.Repeat("MY SLO", 20),
		DisplayName: strings.Repeat("my-slo", 10),
		Project:     strings.Repeat("MY PROJECT", 20),
		Labels: v1alpha.Labels{
			"L O L": []string{"dip", "dip"},
		},
	}
	slo.Spec.Description = strings.Repeat("l", 2000)
	slo.ManifestSource = "/home/me/slo.yaml"
	err := validate(slo)
	require.Error(t, err)
	assert.EqualError(t, err, expectedMetadataError)
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
		assertContainsErrors(t, err, expectedError{
			Prop: "spec.budgetingMethod",
			Code: validation.ErrorCodeRequired,
		})
	})
	t.Run("invalid method", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.BudgetingMethod = "invalid"
		err := validate(slo)
		assertContainsErrors(t, err, expectedError{
			Prop:    "spec.budgetingMethod",
			Message: "'invalid' is not a valid budgeting method",
		})
	})
}

func TestValidate_Spec_Service(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.Service = "my-service"
		err := validate(slo)
		assert.NoError(t, err)
	})
	t.Run("fails", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.Service = "MY SERVICE"
		err := validate(slo)
		assertContainsErrors(t, err, expectedError{
			Prop: "spec.service",
			Code: validation.ErrorCodeStringIsDNSSubdomain,
		})
	})
}

func TestValidate_Spec_AlertPolicies(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.AlertPolicies = []string{"my-policy"}
		err := validate(slo)
		assert.NoError(t, err)
	})
	t.Run("fails", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.AlertPolicies = []string{"my-policy", "MY POLICY", "ok-policy"}
		err := validate(slo)
		assertContainsErrors(t, err, expectedError{
			Prop: "spec.alertPolicies[1]",
			Code: validation.ErrorCodeStringIsDNSSubdomain,
		})
	})
}

func TestValidate_Spec_Attachments(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		for _, attachments := range [][]Attachment{
			{},
			{{URL: "https://my-url.com"}},
			{{URL: "https://my-url.com"}, {URL: "http://insecure-url.pl", DisplayName: ptr("Dashboard")}},
		} {
			slo := validSLO()
			slo.Spec.Attachments = attachments
			err := validate(slo)
			assert.NoError(t, err)
		}
	})
	t.Run("fails, too many attachments", func(t *testing.T) {
		slo := validSLO()
		var attachments []Attachment
		for i := 0; i < 21; i++ {
			attachments = append(attachments, Attachment{})
		}
		slo.Spec.Attachments = attachments
		err := validate(slo)
		assertContainsErrors(t, err, expectedError{
			Prop: "spec.attachments",
			Code: validation.ErrorCodeSliceLength,
		})
	})
	t.Run("fails, invalid attachment", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.Attachments = []Attachment{{URL: "https://valid.com"}, {URL: ""}}
		err := validate(slo)
		assertContainsErrors(t, err, expectedError{
			Prop: "spec.attachments[1].url",
			Code: validation.ErrorCodeRequired,
		})
	})
	t.Run("fails, invalid attachment", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.Attachments = []Attachment{
			{URL: "https://this.com"},
			{URL: ".com"},
			{URL: "", DisplayName: ptr(strings.Repeat("l", 64))},
		}
		err := validate(slo)
		assertContainsErrors(t, err,
			expectedError{
				Prop: "spec.attachments[1].url",
				Code: validation.ErrorCodeStringURL,
			},
			expectedError{
				Prop: "spec.attachments[2].displayName",
				Code: validation.ErrorCodeStringLength,
			},
			expectedError{
				Prop: "spec.attachments[2].url",
				Code: validation.ErrorCodeRequired,
			},
		)
	})
}

type expectedError struct {
	Prop    string
	Code    string
	Message string
}

func assertContainsErrors(t *testing.T, err error, expectedErrors ...expectedError) {
	t.Helper()
	require.Error(t, err)
	var objErr *v1alpha.ObjectError
	require.ErrorAs(t, err, &objErr)
	require.Len(t,
		objErr.Errors,
		len(expectedErrors),
		"v1alpha.ObjectError contains a different number of errors than expected")
	for _, expected := range expectedErrors {
		found := false
		for _, actual := range objErr.Errors {
			var propErr *validation.PropertyError
			require.ErrorAs(t, actual, &propErr)
			if propErr.PropertyName != expected.Prop {
				continue
			}
			if expected.Message != "" && strings.Contains(actual.Error(), expected.Message) ||
				validation.HasErrorCode(actual, expected.Code) {
				found = true
			}
		}
		require.Truef(t, found, "expected '%v' error was not found", expected)
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
