//go:build e2e_test

package tests

import (
	"fmt"
	"slices"
	"testing"
	"time"

	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaExamples "github.com/nobl9/nobl9-go/manifest/v1alpha/examples"
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	prometheusV1 "github.com/nobl9/nobl9-go/sdk/endpoints/prometheus/v1"
	"github.com/nobl9/nobl9-go/tests/e2etestutils"
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
				Limit: 3,
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

func Test_Prometheus_V1_LabelValues(t *testing.T) {
	t.Parallel()

	objects := preparePrometheusLabelValuesObjects(t)
	primaryProject := objects.projects[0].GetName()
	secondaryProject := objects.projects[1].GetName()
	availabilitySLO := objects.slos[0]
	latencySLO := objects.slos[1]
	externalSLO := objects.slos[2]
	compositeSLO := objects.slos[3]

	tests := map[string]struct {
		request        prometheusV1.LabelValuesRequest
		expected       []string
		containsSubset bool
	}{
		"empty matcher returns all values for __name__": {
			request: prometheusV1.LabelValuesRequest{
				Label:   model.MetricNameLabel,
				Matches: []string{},
			},
			expected: []string{
				"budget",
				"burn_rate",
				"component_delay",
				"component_impact",
				"component_was_delayed",
				"component_weight",
				"composite_max_delay",
				"count_good",
				"count_total",
				"reliability",
				"sli_ratio_received_bad",
				"sli_ratio_received_good",
				"sli_ratio_received_total",
				"sli_ratio_used_good",
				"sli_ratio_used_total",
				"sli_threshold",
				"sli_threshold_status",
				"target",
				"threshold",
				"time_slice_allowance",
				"time_slice_reliability",
				"time_slice_status",
			},
		},
		"empty matcher returns all values for projects": {
			request: prometheusV1.LabelValuesRequest{
				Label:   "project",
				Matches: []string{},
			},
			expected: []string{
				alertTestProject1,
				alertTestProject2,
				primaryProject,
				secondaryProject,
			},
			containsSubset: true,
		},
		"metric name values filtered by SLO metric matcher": {
			request: prometheusV1.LabelValuesRequest{
				Label: model.MetricNameLabel,
				Matches: []string{
					fmt.Sprintf(`reliability{project="%s",slo="%s"}`, primaryProject, availabilitySLO.GetName()),
				},
			},
			expected: []string{"reliability"},
		},
		"metric name values filtered by component metric matcher": {
			request: prometheusV1.LabelValuesRequest{
				Label: model.MetricNameLabel,
				Matches: []string{
					fmt.Sprintf(`component_weight{project="%s",component_slo="%s"}`, primaryProject, availabilitySLO.GetName()),
				},
			},
			expected: []string{"component_weight"},
		},
		"SLO values filtered by project": {
			request: prometheusV1.LabelValuesRequest{
				Label:   "slo",
				Matches: []string{fmt.Sprintf(`reliability{project="%s"}`, primaryProject)},
			},
			expected: []string{
				availabilitySLO.GetName(),
				compositeSLO.GetName(),
				latencySLO.GetName(),
			},
		},
		"SLO values filtered by secondary project": {
			request: prometheusV1.LabelValuesRequest{
				Label:   "slo",
				Matches: []string{fmt.Sprintf(`reliability{project="%s"}`, secondaryProject)},
			},
			expected: []string{externalSLO.GetName()},
		},
		"project values from multiple projects": {
			request: prometheusV1.LabelValuesRequest{
				Label: "project",
				Matches: []string{
					fmt.Sprintf(`reliability{slo="%s"}`, availabilitySLO.GetName()),
					fmt.Sprintf(`reliability{slo="%s"}`, externalSLO.GetName()),
				},
			},
			expected: []string{
				primaryProject,
				secondaryProject,
			},
		},
		"SLO values filtered by project and service": {
			request: prometheusV1.LabelValuesRequest{
				Label: "slo",
				Matches: []string{
					fmt.Sprintf(`reliability{project="%s",service="%s"}`, primaryProject, objects.services[0].GetName()),
				},
			},
			expected: []string{
				availabilitySLO.GetName(),
				latencySLO.GetName(),
			},
		},
		"SLO values filtered by project service and SLO": {
			request: prometheusV1.LabelValuesRequest{
				Label: "slo",
				Matches: []string{
					fmt.Sprintf(
						`reliability{project="%s",service="%s",slo="%s"}`,
						primaryProject,
						objects.services[0].GetName(),
						latencySLO.GetName(),
					),
				},
			},
			expected: []string{latencySLO.GetName()},
		},
		"SLO values filtered by mismatched service and SLO": {
			request: prometheusV1.LabelValuesRequest{
				Label: "slo",
				Matches: []string{
					fmt.Sprintf(
						`reliability{project="%s",service="%s",slo="%s"}`,
						primaryProject,
						objects.services[1].GetName(),
						availabilitySLO.GetName(),
					),
				},
			},
			expected: []string{},
		},
		"SLO values filtered by service regexp": {
			request: prometheusV1.LabelValuesRequest{
				Label: "slo",
				Matches: []string{
					fmt.Sprintf(
						`reliability{project="%s",service=~"^%s$"}`,
						primaryProject,
						objects.services[0].GetName(),
					),
				},
			},
			expected: []string{
				availabilitySLO.GetName(),
				latencySLO.GetName(),
			},
		},
		"SLO values filtered by negative regexp": {
			request: prometheusV1.LabelValuesRequest{
				Label: "slo",
				Matches: []string{
					fmt.Sprintf(`reliability{project="%s",slo!~".*-latency"}`, primaryProject),
				},
			},
			expected: []string{
				availabilitySLO.GetName(),
				compositeSLO.GetName(),
			},
		},
		"SLO values from multiple selectors": {
			request: prometheusV1.LabelValuesRequest{
				Label: "slo",
				Matches: []string{
					fmt.Sprintf(
						`reliability{project="%s",service="%s",slo="%s"}`,
						primaryProject,
						objects.services[0].GetName(),
						availabilitySLO.GetName(),
					),
					fmt.Sprintf(`component_weight{project="%s",component_slo="%s"}`, primaryProject, latencySLO.GetName()),
				},
			},
			expected: []string{
				availabilitySLO.GetName(),
				compositeSLO.GetName(),
			},
		},
		"service values filtered by SLO matcher": {
			request: prometheusV1.LabelValuesRequest{
				Label: "service",
				Matches: []string{
					fmt.Sprintf(`reliability{project="%s",slo="%s"}`, primaryProject, availabilitySLO.GetName()),
				},
			},
			expected: []string{objects.services[0].GetName()},
		},
		"service values filtered by project service and SLO": {
			request: prometheusV1.LabelValuesRequest{
				Label: "service",
				Matches: []string{
					fmt.Sprintf(
						`reliability{project="%s",service="%s",slo="%s"}`,
						primaryProject,
						objects.services[0].GetName(),
						availabilitySLO.GetName(),
					),
				},
			},
			expected: []string{objects.services[0].GetName()},
		},
		"service values from multiple selectors": {
			request: prometheusV1.LabelValuesRequest{
				Label: "service",
				Matches: []string{
					fmt.Sprintf(`reliability{project="%s",service="%s"}`, primaryProject, objects.services[0].GetName()),
					fmt.Sprintf(`component_weight{project="%s",slo="%s"}`, primaryProject, compositeSLO.GetName()),
				},
			},
			expected: []string{
				objects.services[0].GetName(),
				objects.services[1].GetName(),
			},
		},
		"objective values filtered by negative SLO matcher": {
			request: prometheusV1.LabelValuesRequest{
				Label: "objective",
				Matches: []string{
					fmt.Sprintf(`reliability{project="%s",slo!="%s"}`, primaryProject, compositeSLO.GetName()),
				},
			},
			expected: []string{
				availabilitySLO.Spec.Objectives[0].Name,
				latencySLO.Spec.Objectives[0].Name,
			},
		},
		"objective values filtered by project service and SLO": {
			request: prometheusV1.LabelValuesRequest{
				Label: "objective",
				Matches: []string{
					fmt.Sprintf(
						`reliability{project="%s",service="%s",slo="%s"}`,
						primaryProject,
						objects.services[0].GetName(),
						availabilitySLO.GetName(),
					),
				},
			},
			expected: []string{availabilitySLO.Spec.Objectives[0].Name},
		},
		"component SLO values filtered by composite metric": {
			request: prometheusV1.LabelValuesRequest{
				Label: "component_slo",
				Matches: []string{
					fmt.Sprintf(`component_weight{project="%s",slo="%s"}`, primaryProject, compositeSLO.GetName()),
				},
			},
			expected: []string{
				availabilitySLO.GetName(),
				externalSLO.GetName(),
				latencySLO.GetName(),
			},
		},
		"component SLO values filtered by secondary component project": {
			request: prometheusV1.LabelValuesRequest{
				Label: "component_slo",
				Matches: []string{
					fmt.Sprintf(
						`component_weight{project="%s",component_project="%s"}`,
						primaryProject,
						secondaryProject,
					),
				},
			},
			expected: []string{externalSLO.GetName()},
		},
		"composite SLO values filtered by component matcher": {
			request: prometheusV1.LabelValuesRequest{
				Label: "slo",
				Matches: []string{
					fmt.Sprintf(
						`component_weight{project="%s",component_slo="%s"}`,
						primaryProject,
						availabilitySLO.GetName(),
					),
				},
			},
			expected: []string{compositeSLO.GetName()},
		},
		"composite SLO values filtered by component project SLO and objective": {
			request: prometheusV1.LabelValuesRequest{
				Label: "slo",
				Matches: []string{
					fmt.Sprintf(
						`component_weight{project="%s",component_project="%s",component_slo="%s",component_objective="%s"}`,
						primaryProject,
						primaryProject,
						latencySLO.GetName(),
						latencySLO.Spec.Objectives[0].Name,
					),
				},
			},
			expected: []string{compositeSLO.GetName()},
		},
		"component objective values filtered by component project": {
			request: prometheusV1.LabelValuesRequest{
				Label: "component_objective",
				Matches: []string{
					fmt.Sprintf(`component_weight{project="%s",component_project="%s"}`, primaryProject, primaryProject),
				},
			},
			expected: []string{
				availabilitySLO.Spec.Objectives[0].Name,
				latencySLO.Spec.Objectives[0].Name,
			},
		},
		"component objective values filtered by secondary component project": {
			request: prometheusV1.LabelValuesRequest{
				Label: "component_objective",
				Matches: []string{
					fmt.Sprintf(
						`component_weight{project="%s",component_project="%s"}`,
						primaryProject,
						secondaryProject,
					),
				},
			},
			expected: []string{externalSLO.Spec.Objectives[0].Name},
		},
		"component project values filtered by composite metric": {
			request: prometheusV1.LabelValuesRequest{
				Label: "component_project",
				Matches: []string{
					fmt.Sprintf(`component_weight{project="%s",slo="%s"}`, primaryProject, compositeSLO.GetName()),
				},
			},
			expected: []string{
				primaryProject,
				secondaryProject,
			},
		},
		"component project values filtered by component SLO": {
			request: prometheusV1.LabelValuesRequest{
				Label: "component_project",
				Matches: []string{
					fmt.Sprintf(
						`component_weight{project="%s",component_slo="%s"}`,
						primaryProject,
						externalSLO.GetName(),
					),
				},
			},
			expected: []string{secondaryProject},
		},
		"component label values filtered by non-component metric": {
			request: prometheusV1.LabelValuesRequest{
				Label: "component_slo",
				Matches: []string{
					fmt.Sprintf(`reliability{project="%s",slo="%s"}`, primaryProject, availabilitySLO.GetName()),
				},
			},
			expected: []string{},
		},
		"unknown metric": {
			request: prometheusV1.LabelValuesRequest{
				Label:   "slo",
				Matches: []string{fmt.Sprintf(`definitely_unknown_metric{project="%s"}`, primaryProject)},
			},
			expected: []string{},
		},
		"unsupported label": {
			request: prometheusV1.LabelValuesRequest{
				Label: "job",
			},
			expected: []string{},
		},
		"limit for __name__": {
			request: prometheusV1.LabelValuesRequest{
				Label:   model.MetricNameLabel,
				Matches: []string{},
				Limit:   3,
			},
			expected: []string{
				"budget",
				"burn_rate",
				"component_delay",
			},
		},
		"limit for slo": {
			request: prometheusV1.LabelValuesRequest{
				Label: "slo",
				Matches: []string{
					fmt.Sprintf(`reliability{project="%s"}`, primaryProject),
				},
				Limit: 2,
			},
			expected: []string{
				availabilitySLO.GetName(),
				compositeSLO.GetName(),
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			requirePrometheusLabelValues(t, tt.request, tt.expected, tt.containsSubset)
		})
	}
}

type prometheusLabelValuesObjects struct {
	projects []manifest.Object
	services []v1alphaService.Service
	slos     []v1alphaSLO.SLO
}

func preparePrometheusLabelValuesObjects(t *testing.T) prometheusLabelValuesObjects {
	t.Helper()

	projectNamePrefix := e2etestutils.GenerateName()
	primaryProject := newV1alphaProject(t, v1alphaProject.Metadata{
		Name: fmt.Sprintf("%s-primary", projectNamePrefix),
	})
	secondaryProject := newV1alphaProject(t, v1alphaProject.Metadata{
		Name: fmt.Sprintf("%s-secondary", projectNamePrefix),
	})
	serviceNamePrefix := e2etestutils.GenerateName()
	services := []v1alphaService.Service{
		newV1alphaService(t, v1alphaService.Metadata{
			Name:    fmt.Sprintf("%s-primary", serviceNamePrefix),
			Project: primaryProject.GetName(),
		}),
		newV1alphaService(t, v1alphaService.Metadata{
			Name:    fmt.Sprintf("%s-primary-composite", serviceNamePrefix),
			Project: primaryProject.GetName(),
		}),
		newV1alphaService(t, v1alphaService.Metadata{
			Name:    fmt.Sprintf("%s-secondary", serviceNamePrefix),
			Project: secondaryProject.GetName(),
		}),
	}

	sloNamePrefix := e2etestutils.GenerateName()
	availabilitySLO := newPrometheusLabelValuesSLO(
		t,
		primaryProject.GetName(),
		services[0].GetName(),
		fmt.Sprintf("%s-availability", sloNamePrefix),
		"availability",
	)
	latencySLO := newPrometheusLabelValuesSLO(
		t,
		primaryProject.GetName(),
		services[0].GetName(),
		fmt.Sprintf("%s-latency", sloNamePrefix),
		"latency",
	)
	externalSLO := newPrometheusLabelValuesSLO(
		t,
		secondaryProject.GetName(),
		services[2].GetName(),
		fmt.Sprintf("%s-external", sloNamePrefix),
		"external",
	)
	compositeSLO := newPrometheusLabelValuesCompositeSLO(
		t,
		primaryProject.GetName(),
		services[1].GetName(),
		fmt.Sprintf("%s-composite", sloNamePrefix),
		[]v1alphaSLO.SLO{availabilitySLO, externalSLO, latencySLO},
	)
	slos := []v1alphaSLO.SLO{availabilitySLO, latencySLO, externalSLO, compositeSLO}

	dependencies := []manifest.Object{primaryProject, secondaryProject, services[0], services[1], services[2]}
	e2etestutils.V1Apply(t, dependencies)
	e2etestutils.V1Apply(t, slos)
	t.Cleanup(func() {
		e2etestutils.V1Delete(t, slos)
		e2etestutils.V1Delete(t, dependencies)
	})

	return prometheusLabelValuesObjects{
		projects: []manifest.Object{primaryProject, secondaryProject},
		services: services,
		slos:     slos,
	}
}

func newPrometheusLabelValuesSLO(
	t *testing.T,
	project string,
	service string,
	name string,
	objective string,
) v1alphaSLO.SLO {
	t.Helper()

	slo := e2etestutils.GetExampleObject[v1alphaSLO.SLO](
		t,
		manifest.KindSLO,
		e2etestutils.FilterExamplesByDataSourceType(v1alpha.Prometheus),
	)
	slo.Metadata = v1alphaSLO.Metadata{
		Name:        name,
		DisplayName: objective,
		Project:     project,
		Labels:      e2etestutils.AnnotateLabels(t, v1alpha.Labels{}),
		Annotations: commonAnnotations,
	}
	slo.Spec.Service = service
	slo.Spec.AlertPolicies = nil
	slo.Spec.AnomalyConfig = nil
	slo.Spec.Objectives[0].Name = objective
	slo.Spec.Objectives[0].DisplayName = objective
	e2etestutils.ProvisionDataSourceForSLO(t, &slo)
	return slo
}

func newPrometheusLabelValuesCompositeSLO(
	t *testing.T,
	project string,
	service string,
	name string,
	components []v1alphaSLO.SLO,
) v1alphaSLO.SLO {
	t.Helper()

	slo := e2etestutils.GetExampleObject[v1alphaSLO.SLO](t, manifest.KindSLO, func(example v1alphaExamples.Example) bool {
		exampleSLO := example.GetObject().(v1alphaSLO.SLO)
		return exampleSLO.Spec.HasCompositeObjectives()
	})
	slo.Metadata = v1alphaSLO.Metadata{
		Name:        name,
		DisplayName: "composite",
		Project:     project,
		Labels:      e2etestutils.AnnotateLabels(t, v1alpha.Labels{}),
		Annotations: commonAnnotations,
	}
	slo.Spec.Service = service
	slo.Spec.AlertPolicies = nil
	slo.Spec.AnomalyConfig = nil
	slo.Spec.Objectives[0].Name = "composite"
	slo.Spec.Objectives[0].DisplayName = "composite"
	slo.Spec.Objectives[0].Composite.Objectives = make([]v1alphaSLO.CompositeObjective, 0, len(components))
	for _, component := range components {
		slo.Spec.Objectives[0].Composite.Objectives = append(
			slo.Spec.Objectives[0].Composite.Objectives,
			v1alphaSLO.CompositeObjective{
				Project:     component.GetProject(),
				SLO:         component.GetName(),
				Objective:   component.Spec.Objectives[0].Name,
				Weight:      1,
				WhenDelayed: v1alphaSLO.WhenDelayedCountAsGood,
			},
		)
	}
	return slo
}

func requirePrometheusLabelValues(
	t *testing.T,
	request prometheusV1.LabelValuesRequest,
	expected []string,
	containsSubset bool,
) {
	t.Helper()

	ticker := time.NewTicker(5 * time.Second)
	timer := time.NewTimer(time.Minute)
	defer ticker.Stop()
	defer timer.Stop()

	expectedValues := make(model.LabelValues, 0, len(expected))
	for _, v := range expected {
		expectedValues = append(expectedValues, model.LabelValue(v))
	}
	var (
		actualValues model.LabelValues
		warnings     promv1.Warnings
		err          error
	)
	for {
		actualValues, warnings, err = client.Prometheus().V1().LabelValues(t.Context(), request)
		if err == nil && len(warnings) == 0 &&
			labelValuesMatch(actualValues, expectedValues, containsSubset) {
			return
		}

		select {
		case <-ticker.C:
		case <-timer.C:
			require.NoError(t, err)
			require.Empty(t, warnings)
			if containsSubset {
				require.Subset(t, actualValues, expectedValues)
			} else {
				require.Equal(t, expectedValues, actualValues)
			}
			return
		}
	}
}

func labelValuesMatch(
	actualValues model.LabelValues,
	expectedValues model.LabelValues,
	containsSubset bool,
) bool {
	if containsSubset {
		return labelValuesContainSubset(actualValues, expectedValues)
	}
	return slices.Equal(expectedValues, actualValues)
}

func labelValuesContainSubset(actualValues, expectedValues model.LabelValues) bool {
	for _, expectedValue := range expectedValues {
		if !slices.Contains(actualValues, expectedValue) {
			return false
		}
	}
	return true
}
