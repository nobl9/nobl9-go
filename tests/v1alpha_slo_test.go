//go:build e2e_test

package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaAgent "github.com/nobl9/nobl9-go/manifest/v1alpha/agent"
	v1alphaAlertMethod "github.com/nobl9/nobl9-go/manifest/v1alpha/alertmethod"
	v1alphaAlertPolicy "github.com/nobl9/nobl9-go/manifest/v1alpha/alertpolicy"
	v1alphaDirect "github.com/nobl9/nobl9-go/manifest/v1alpha/direct"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	"github.com/nobl9/nobl9-go/sdk"
	objectsV1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
)

const slosPerService = 50

// nolint: gocognit
func Test_Objects_V1_V1alpha_SLO(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	// Prepare dependencies.
	project := generateV1alphaProject(t)
	defaultProjectService := newV1alphaService(t, v1alphaService.Metadata{
		Name:    generateName(),
		Project: defaultProject,
	})
	alertMethod := newV1alphaAlertMethod(t, v1alpha.AlertMethodTypeSlack, v1alphaAlertMethod.Metadata{
		Name:    generateName(),
		Project: project.GetName(),
	})
	alertPolicyExample := examplesRegistry[manifest.KindAlertPolicy][0].Example
	alertPolicy := newV1alphaAlertPolicy(t, v1alphaAlertPolicy.Metadata{
		Name:    generateName(),
		Project: project.GetName(),
	}, alertPolicyExample.GetVariant(), alertPolicyExample.GetSubVariant())
	alertPolicy.Spec.AlertMethods = []v1alphaAlertPolicy.AlertMethodRef{
		{
			Metadata: v1alphaAlertPolicy.AlertMethodRefMetadata{
				Name:    alertMethod.Metadata.Name,
				Project: alertMethod.Metadata.Project,
			},
		},
	}
	agentsAndDirects := append(
		v1alphaSLODependencyAgents(t),
		v1alphaSLODependencyDirects(t)...)

	sloExamples := examplesRegistry[manifest.KindSLO]

	// Composite SLOs depend on other SLOs. Example SLOs are being sorted so that Composite SLOs are placed at the end,
	// allowing them to depend on the SLOs listed before them.
	slices.SortStableFunc(sloExamples, func(i, j exampleWrapper) int {
		var intI, intJ int
		iSlo := i.GetObject().(v1alphaSLO.SLO)
		if iSlo.Spec.HasCompositeObjectives() {
			intI = 1
		}
		jSlo := j.GetObject().(v1alphaSLO.SLO)
		if jSlo.Spec.HasCompositeObjectives() {
			intJ = 1
		}
		return intI - intJ
	})

	slos := make([]manifest.Object, 0, len(sloExamples))
	dependencies := []manifest.Object{
		project,
		defaultProjectService,
		alertMethod,
		alertPolicy,
	}

	var service v1alphaService.Service
	for i, example := range sloExamples {
		slo := example.GetObject().(v1alphaSLO.SLO)
		slo.Metadata = v1alphaSLO.Metadata{
			Name:        generateName(),
			DisplayName: fmt.Sprintf("SLO %d", i),
			Project:     project.GetName(),
			Labels:      annotateLabels(t, v1alpha.Labels{}),
			Annotations: commonAnnotations,
		}
		// Generate new service for every `slosPerService` SLOs to meet the quota.
		if i%slosPerService == 0 {
			service = newV1alphaService(t, v1alphaService.Metadata{
				Name:    generateName(),
				Project: project.GetName(),
			})
			dependencies = append(dependencies, service)
		}
		slo.Spec.Service = service.GetName()
		slo.Spec.AlertPolicies = []string{alertPolicy.GetName()}

		if slo.Spec.HasCompositeObjectives() {
			for componentIndex, component := range slo.Spec.Objectives[0].Composite.Components.Objectives {
				componentSlo := slos[len(slos)-1-componentIndex].(v1alphaSLO.SLO)
				component.Project = componentSlo.Metadata.Project
				component.SLO = componentSlo.Metadata.Name
				component.Objective = componentSlo.Spec.Objectives[0].Name
				slo.Spec.Objectives[0].Composite.Components.Objectives[componentIndex] = component
			}
		} else {
			slo.Spec.AnomalyConfig.NoData.AlertMethods = []v1alphaSLO.AnomalyConfigAlertMethod{
				{
					Name:    alertMethod.Metadata.Name,
					Project: alertMethod.Metadata.Project,
				},
			}

			metricSpecs := slo.Spec.AllMetricSpecs()
			require.Greater(t, len(metricSpecs), 0, "expected at least 1 metric spec")
			sourceType := metricSpecs[0].DataSourceType()
			sources := filterSlice(agentsAndDirects, func(object manifest.Object) bool {
				if object.GetKind() != slo.Spec.Indicator.MetricSource.Kind {
					return false
				}
				var getType func() (v1alpha.DataSourceType, error)
				if direct, ok := object.(v1alphaDirect.Direct); ok {
					getType = direct.Spec.GetType
				} else if agent, ok := object.(v1alphaAgent.Agent); ok {
					getType = agent.Spec.GetType
				}
				require.NotNil(t, getType)
				typ, err := getType()
				require.NoError(t, err)
				return typ == sourceType
			})
			require.Greater(t, len(sources), 0, "expected at least 1 source of type %s", sourceType)
			source := sources[0]
			slo.Spec.Indicator.MetricSource.Name = source.GetName()
			slo.Spec.Indicator.MetricSource.Project = source.(manifest.ProjectScopedObject).GetProject()

			switch i {
			case 0:
				slo.Metadata.Project = defaultProject
				slo.Spec.Service = defaultProjectService.GetName()
				// We don't need to have these field filled,
				// the first SLO is only here to test default project querying.
				slo.Spec.AlertPolicies = []string{}
				slo.Spec.AnomalyConfig = nil
			case 1:
				slo.Metadata.Labels["team"] = []string{"green"}
			case 2:
				slo.Metadata.Labels["team"] = []string{"orange"}
			case 3:
				slo.Metadata.Labels["team"] = []string{"orange"}
			}
			// TODO: Remove this after PC-13575 is resolved.
			if slo.Spec.Indicator.MetricSource.Kind == manifest.KindAgent && sourceType == v1alpha.CloudWatch {
				skip := false
				for _, spec := range slo.Spec.AllMetricSpecs() {
					if spec.CloudWatch.AccountID != nil {
						skip = true
						break
					}
				}
				if skip {
					continue
				}
			}
		}
		slos = append(slos, slo)
	}

	serviceNameFilterSLOs, serviceNameFilterDependencies := prepareObjectsForServiceNameFilteringTests(
		t,
		agentsAndDirects[0].(v1alphaAgent.Agent),
	)
	for _, slo := range serviceNameFilterSLOs {
		slos = append(slos, slo)
	}
	dependencies = append(dependencies, serviceNameFilterDependencies...)

	t.Cleanup(func() {
		slices.Reverse(slos)
		v1DeleteBatch(t, slos, 50)
		v1Delete(t, dependencies)
	})
	v1Apply(t, dependencies)
	v1ApplyBatch(t, slos, 50)
	inputs := manifest.FilterByKind[v1alphaSLO.SLO](slos)

	filterTests := map[string]struct {
		request    objectsV1.GetSLOsRequest
		expected   []v1alphaSLO.SLO
		returnsAll bool
	}{
		"all": {
			request:    objectsV1.GetSLOsRequest{Project: sdk.ProjectsWildcard},
			expected:   inputs,
			returnsAll: true,
		},
		"default project": {
			request:    objectsV1.GetSLOsRequest{},
			expected:   []v1alphaSLO.SLO{inputs[0]},
			returnsAll: true,
		},
		"filter by project": {
			request: objectsV1.GetSLOsRequest{
				Project: project.GetName(),
			},
			expected: inputs[1 : len(inputs)-len(serviceNameFilterSLOs)],
		},
		"filter by name": {
			request: objectsV1.GetSLOsRequest{
				Project: project.GetName(),
				Names:   []string{inputs[3].Metadata.Name},
			},
			expected: []v1alphaSLO.SLO{inputs[3]},
		},
		"filter by label": {
			request: objectsV1.GetSLOsRequest{
				Project: project.GetName(),
				Labels:  annotateLabels(t, v1alpha.Labels{"team": []string{"green"}}),
			},
			expected: []v1alphaSLO.SLO{inputs[1]},
		},
		"filter by label and name": {
			request: objectsV1.GetSLOsRequest{
				Project: project.GetName(),
				Names:   []string{inputs[3].Metadata.Name},
				Labels:  annotateLabels(t, v1alpha.Labels{"team": []string{"orange"}}),
			},
			expected: []v1alphaSLO.SLO{inputs[3]},
		},
		"filter by one service": {
			request: objectsV1.GetSLOsRequest{
				Project:  serviceNameFilterSLOs[0].GetProject(),
				Services: []string{serviceNameFilterSLOs[0].Spec.Service},
			},
			expected: serviceNameFilterSLOs[0:3],
		},
		"filter by one service with project wildcard": {
			request: objectsV1.GetSLOsRequest{
				Project:  sdk.ProjectsWildcard,
				Services: []string{serviceNameFilterSLOs[0].Spec.Service},
			},
			expected: append(slices.Clone(serviceNameFilterSLOs[0:3]), serviceNameFilterSLOs[4]),
		},
		"filter by two services": {
			request: objectsV1.GetSLOsRequest{
				Project: serviceNameFilterSLOs[0].GetProject(),
				Services: []string{
					serviceNameFilterSLOs[0].Spec.Service,
					serviceNameFilterSLOs[3].Spec.Service,
				},
			},
			expected: serviceNameFilterSLOs[0:4],
		},
		"filter by project, label and service": {
			request: objectsV1.GetSLOsRequest{
				Project:  serviceNameFilterSLOs[1].GetProject(),
				Services: []string{serviceNameFilterSLOs[0].Spec.Service},
				Labels:   annotateLabels(t, v1alpha.Labels{"service-name-filter": []string{"foo", "bar"}}),
			},
			expected: serviceNameFilterSLOs[1:3],
		},
		"filter by project, label, service and name": {
			request: objectsV1.GetSLOsRequest{
				Project:  serviceNameFilterSLOs[2].GetProject(),
				Names:    []string{serviceNameFilterSLOs[2].GetName()},
				Labels:   annotateLabels(t, v1alpha.Labels{"service-name-filter": []string{"foo"}}),
				Services: []string{serviceNameFilterSLOs[2].Spec.Service},
			},
			expected: serviceNameFilterSLOs[2:3],
		},
	}
	for name, test := range filterTests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, err := client.Objects().V1().GetV1alphaSLOs(ctx, test.request)
			require.NoError(t, err)
			if !test.returnsAll {
				require.Equal(t, len(test.expected), len(actual),
					"actual number of SLOs does not match the expected")
			}
			assertSubset(t, actual, test.expected, assertV1alphaSLOsAreEqual)
		})
	}
}

func prepareObjectsForServiceNameFilteringTests(
	t *testing.T,
	agent v1alphaAgent.Agent,
) (
	slos []v1alphaSLO.SLO,
	dependencies []manifest.Object,
) {
	t.Helper()
	agentType, err := agent.Spec.GetType()
	require.NoError(t, err)

	// Projects.
	project1 := generateV1alphaProject(t)
	project2 := generateV1alphaProject(t)
	// Services.
	service1Proj1 := newV1alphaService(t, v1alphaService.Metadata{
		Name:    generateName(),
		Project: project1.GetName(),
	})
	service2Proj1 := newV1alphaService(t, v1alphaService.Metadata{
		Name:    generateName(),
		Project: project1.GetName(),
	})
	service1Proj2 := newV1alphaService(t, v1alphaService.Metadata{
		Name:    service1Proj1.GetName(),
		Project: project2.GetName(),
	})

	dependencies = append(
		dependencies,
		project1,
		project2,
		service1Proj1,
		service2Proj1,
		service1Proj2,
	)

	// SLOs.
	var sloTemplate v1alphaSLO.SLO
	for _, example := range examplesRegistry[manifest.KindSLO] {
		slo := example.GetObject().(v1alphaSLO.SLO)
		metricSpecs := slo.Spec.AllMetricSpecs()
		require.Greater(t, len(metricSpecs), 0, "expected at least 1 metric spec")
		if !slo.Spec.HasCompositeObjectives() && metricSpecs[0].DataSourceType() == agentType {
			sloTemplate = slo
			break
		}
	}
	require.NotNil(t, sloTemplate, "expected at least 1 SLO with metric source of type %s", agentType)

	for i, params := range []struct {
		project string
		service string
		labels  v1alpha.Labels
	}{
		{project1.GetName(), service1Proj1.GetName(), v1alpha.Labels{}},
		{project1.GetName(), service1Proj1.GetName(), v1alpha.Labels{"service-name-filter": []string{"bar"}}},
		{project1.GetName(), service1Proj1.GetName(), v1alpha.Labels{"service-name-filter": []string{"foo"}}},
		{project1.GetName(), service2Proj1.GetName(), v1alpha.Labels{}},
		{project2.GetName(), service1Proj2.GetName(), v1alpha.Labels{}},
	} {
		slo := clone(t, sloTemplate)
		slo.Metadata = v1alphaSLO.Metadata{
			Name:        generateName(),
			DisplayName: fmt.Sprintf("SLO filtered by service %d", i),
			Project:     params.project,
			Labels:      annotateLabels(t, params.labels),
			Annotations: commonAnnotations,
		}
		slo.Spec.Service = params.service
		slo.Spec.AlertPolicies = []string{}
		slo.Spec.AnomalyConfig = nil
		slo.Spec.Indicator.MetricSource = v1alphaSLO.MetricSourceSpec{
			Name:    agent.GetName(),
			Project: agent.GetProject(),
			Kind:    agent.GetKind(),
		}
		slos = append(slos, slo)
	}
	return slos, dependencies
}

func v1alphaSLODependencyAgents(t *testing.T) []manifest.Object {
	t.Helper()
	agentTypes := v1alpha.DataSourceTypeValues()
	agents := make([]manifest.Object, 0, len(agentTypes)+1)
	for _, typ := range agentTypes {
		agent := newV1alphaAgent(t,
			typ,
			v1alphaAgent.Metadata{
				Name:    fmt.Sprintf("sdk-e2e-agent-%s", strings.ToLower(typ.String())),
				Project: defaultProject,
			},
		)
		agent.Spec.Description = objectPersistedDescription
		agents = append(agents, agent)
	}
	v1ApplyBatch(t, agents, 1)
	return agents
}

func v1alphaSLODependencyDirects(t *testing.T) []manifest.Object {
	t.Helper()
	directTypes := filterSlice(v1alpha.DataSourceTypeValues(), v1alphaDirect.IsValidDirectType)
	directs := make([]manifest.Object, 0, len(directTypes)+1)
	for _, typ := range directTypes {
		direct := newV1alphaDirect(t,
			typ,
			v1alphaDirect.Metadata{
				Name:    fmt.Sprintf("sdk-e2e-direct-%s", strings.ToLower(typ.String())),
				Project: defaultProject,
			},
		)
		direct.Spec.Description = objectPersistedDescription
		directs = append(directs, direct)
	}
	v1Apply(t, directs)
	return directs
}

func assertV1alphaSLOsAreEqual(t *testing.T, expected, actual v1alphaSLO.SLO) {
	t.Helper()
	assert.NotNil(t, actual.Status)
	actual.Status = nil
	assert.NotNil(t, actual.Spec.CreatedAt)
	actual.Spec.CreatedAt = ""
	assert.NotNil(t, actual.Spec.CreatedBy)
	actual.Spec.CreatedBy = ""
	actual.Status = nil
	actual.Spec.TimeWindows[0].Period = nil
	assert.Equal(t, expected, actual)
}

func clone[T any](t *testing.T, object T) T {
	t.Helper()
	data, err := json.Marshal(object)
	require.NoError(t, err)
	var cloned T
	require.NoError(t, json.Unmarshal(data, &cloned))
	return cloned
}
