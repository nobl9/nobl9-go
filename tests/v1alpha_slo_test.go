//go:build e2e_test

package tests

import (
	"cmp"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaAlertMethod "github.com/nobl9/nobl9-go/manifest/v1alpha/alertmethod"
	v1alphaAlertPolicy "github.com/nobl9/nobl9-go/manifest/v1alpha/alertpolicy"
	v1alphaExamples "github.com/nobl9/nobl9-go/manifest/v1alpha/examples"
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	"github.com/nobl9/nobl9-go/sdk"
	objectsV1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
	"github.com/nobl9/nobl9-go/tests/e2etestutils"
)

const (
	slosPerService          = 50
	sloListSortFixtureLabel = "slo-list-sort-fixture"
)

// nolint: gocognit
func Test_Objects_V1_V1alpha_SLO(t *testing.T) {
	t.Parallel()
	// Prepare dependencies.
	project := generateV1alphaProject(t)
	defaultProjectService := newV1alphaService(t, v1alphaService.Metadata{
		Name:    e2etestutils.GenerateName(),
		Project: defaultProject,
	})
	alertMethod := newV1alphaAlertMethod(t, v1alpha.AlertMethodTypeSlack, v1alphaAlertMethod.Metadata{
		Name:    e2etestutils.GenerateName(),
		Project: project.GetName(),
	})
	alertPolicyExample := e2etestutils.GetExample(t, manifest.KindAlertPolicy, nil)
	alertPolicy := newV1alphaAlertPolicy(t, v1alphaAlertPolicy.Metadata{
		Name:    e2etestutils.GenerateName(),
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

	sloExamples := e2etestutils.GetAllExamples(t, manifest.KindSLO)
	// Composite SLOs depend on other SLOs. Example SLOs are being sorted so that Composite SLOs are placed at the end,
	// allowing them to depend on the SLOs listed before them.
	slices.SortStableFunc(sloExamples, func(i, j v1alphaExamples.Example) int {
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
			Name:        e2etestutils.GenerateName(),
			DisplayName: fmt.Sprintf("SLO %d", i),
			Project:     project.GetName(),
			Labels:      e2etestutils.AnnotateLabels(t, v1alpha.Labels{}),
			Annotations: commonAnnotations,
		}
		// Generate new service for every `slosPerService` SLOs to meet the quota.
		if i%slosPerService == 0 {
			service = newV1alphaService(t, v1alphaService.Metadata{
				Name:    e2etestutils.GenerateName(),
				Project: project.GetName(),
			})
			dependencies = append(dependencies, service)
		}
		slo.Spec.Service = service.GetName()
		slo.Spec.AlertPolicies = []string{alertPolicy.GetName()}

		if slo.Spec.HasCompositeObjectives() {
			for componentIndex, component := range slo.Spec.Objectives[0].Composite.Objectives {
				componentSlo := slos[len(slos)-1-componentIndex].(v1alphaSLO.SLO)
				component.Project = componentSlo.Metadata.Project
				component.SLO = componentSlo.Metadata.Name
				component.Objective = componentSlo.Spec.Objectives[0].Name
				slo.Spec.Objectives[0].Composite.Objectives[componentIndex] = component
			}
		} else {
			slo.Spec.AnomalyConfig.NoData.AlertMethods = []v1alphaSLO.AnomalyConfigAlertMethod{
				{
					Name:    alertMethod.Metadata.Name,
					Project: alertMethod.Metadata.Project,
				},
			}
			slo.Spec.AnomalyConfig.NoData.AlertAfter = ptr("1h")

			metricSpecs := slo.Spec.AllMetricSpecs()
			require.Greater(t, len(metricSpecs), 0, "expected at least 1 metric spec")
			sourceType := metricSpecs[0].DataSourceType()

			e2etestutils.ProvisionDataSourceForSLO(t, &slo)

			switch i {
			case 0:
				slo.Metadata.Project = defaultProject
				slo.Spec.Service = defaultProjectService.GetName()
				// We don't need to have these field filled,
				// the first SLO is only here to test default project querying.
				slo.Spec.AlertPolicies = nil
				slo.Spec.AnomalyConfig = nil
			case 1:
				slo.Metadata.Labels["team"] = []string{"green"}
				slo.Spec.AnomalyConfig.NoData.TreatZeroAsNoData = ptr(true)
			case 2:
				slo.Metadata.Labels["team"] = []string{"orange"}
				slo.Spec.AnomalyConfig.NoData.TreatZeroAsNoData = ptr(false)
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

	serviceNameFilterSLOs, serviceNameFilterDependencies := prepareObjectsForServiceNameFilteringTests(t)
	for _, slo := range serviceNameFilterSLOs {
		slos = append(slos, slo)
	}
	dependencies = append(dependencies, serviceNameFilterDependencies...)

	t.Cleanup(func() {
		slices.Reverse(slos)
		e2etestutils.V1DeleteBatch(t, slos, 50)
		e2etestutils.V1Delete(t, dependencies)
	})
	e2etestutils.V1Apply(t, dependencies)
	batchedSLOs := slos[:len(slos)-len(serviceNameFilterSLOs)]
	e2etestutils.V1ApplyBatch(t, batchedSLOs, 50)
	for _, slo := range serviceNameFilterSLOs {
		e2etestutils.V1Apply(t, []v1alphaSLO.SLO{slo})
	}
	inputs := manifest.FilterByKind[v1alphaSLO.SLO](slos)
	require.Greater(t, len(inputs), 3)
	sortInputs := inputs[len(inputs)-len(serviceNameFilterSLOs):]
	require.Len(t, sortInputs, len(serviceNameFilterSLOs))
	updatedSortInput := len(sortInputs) - 1
	sortInputs[updatedSortInput].Metadata.DisplayName += " updated"
	serviceNameFilterSLOs[updatedSortInput] = sortInputs[updatedSortInput]
	e2etestutils.V1Apply(t, []v1alphaSLO.SLO{sortInputs[updatedSortInput]})
	sortFixtureLabels := e2etestutils.AnnotateLabels(t, v1alpha.Labels{
		sloListSortFixtureLabel: []string{""},
	})

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
				Labels:  e2etestutils.AnnotateLabels(t, v1alpha.Labels{"team": []string{"green"}}),
			},
			expected: []v1alphaSLO.SLO{inputs[1]},
		},
		"filter by label and name": {
			request: objectsV1.GetSLOsRequest{
				Project: project.GetName(),
				Names:   []string{inputs[3].Metadata.Name},
				Labels:  e2etestutils.AnnotateLabels(t, v1alpha.Labels{"team": []string{"orange"}}),
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
				Labels:   e2etestutils.AnnotateLabels(t, v1alpha.Labels{"service-name-filter": []string{"foo", "bar"}}),
			},
			expected: serviceNameFilterSLOs[1:3],
		},
		"filter by project, label, service and name": {
			request: objectsV1.GetSLOsRequest{
				Project:  serviceNameFilterSLOs[2].GetProject(),
				Names:    []string{serviceNameFilterSLOs[2].GetName()},
				Labels:   e2etestutils.AnnotateLabels(t, v1alpha.Labels{"service-name-filter": []string{"foo"}}),
				Services: []string{serviceNameFilterSLOs[2].Spec.Service},
			},
			expected: serviceNameFilterSLOs[2:3],
		},
	}
	for name, test := range filterTests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, err := client.Objects().V1().GetV1alphaSLOs(t.Context(), test.request)
			require.NoError(t, err)
			if !test.returnsAll {
				require.Equal(t, len(test.expected), len(actual),
					"actual number of SLOs does not match the expected")
			}
			assertSubset(t, actual, test.expected, assertV1alphaSLOsAreEqual)
		})
	}

	t.Run("pagination and sorting", func(t *testing.T) {
		projectSortValues := map[string]string{defaultProject: defaultProject}
		for _, project := range manifest.FilterByKind[v1alphaProject.Project](dependencies) {
			projectSortValues[project.Metadata.Name] = displayNameOrName(
				project.Metadata.DisplayName,
				project.Metadata.Name,
			)
		}
		serviceSortValues := make(map[string]string)
		for _, service := range manifest.FilterByKind[v1alphaService.Service](dependencies) {
			serviceSortValues[service.Metadata.Project+"/"+service.Metadata.Name] = displayNameOrName(
				service.Metadata.DisplayName,
				service.Metadata.Name,
			)
		}
		sortTests := []struct {
			column objectsV1.GetSLOsSortColumn
			value  func(v1alphaSLO.SLO) string
		}{
			{
				column: objectsV1.GetSLOsSortColumnProject,
				value:  func(slo v1alphaSLO.SLO) string { return projectSortValues[slo.Metadata.Project] },
			},
			{
				column: objectsV1.GetSLOsSortColumnService,
				value: func(slo v1alphaSLO.SLO) string {
					return serviceSortValues[slo.Metadata.Project+"/"+slo.Spec.Service]
				},
			},
			{
				column: objectsV1.GetSLOsSortColumnSLO,
				value: func(slo v1alphaSLO.SLO) string {
					return displayNameOrName(slo.Metadata.DisplayName, slo.Metadata.Name)
				},
			},
			{
				column: objectsV1.GetSLOsSortColumnLastModifiedAt,
				value: func(slo v1alphaSLO.SLO) string {
					if slo.Status == nil {
						return ""
					}
					return slo.Status.UpdatedAt
				},
			},
		}
		directions := []objectsV1.GetSLOsSortDirection{
			objectsV1.GetSLOsSortDirectionAsc,
			objectsV1.GetSLOsSortDirectionDesc,
		}

		for _, sortTest := range sortTests {
			for _, direction := range directions {
				t.Run(string(sortTest.column)+"_"+string(direction), func(t *testing.T) {
					request := objectsV1.GetSLOsRequest{
						Project: sdk.ProjectsWildcard,
						Labels:  sortFixtureLabels,
						Sort: &objectsV1.GetSLOsSort{
							Column:    sortTest.column,
							Direction: direction,
						},
					}
					all, err := client.Objects().V1().GetV1alphaSLOs(t.Context(), request)
					require.NoError(t, err)
					require.Len(t, all, len(sortInputs))
					var expected []v1alphaSLO.SLO
					if sortTest.column == objectsV1.GetSLOsSortColumnLastModifiedAt {
						for _, slo := range all {
							require.NotEmpty(t, sortTest.value(slo))
						}
						expected = slices.Clone(sortInputs)
						if direction == objectsV1.GetSLOsSortDirectionDesc {
							slices.Reverse(expected)
						}
					} else {
						expected = expectedSLOListOrder(t, all, sortTest.value, direction)
					}
					require.Equal(t, sloListIdentities(expected), sloListIdentities(all))

					request.Pagination = &objectsV1.GetSLOsPagination{Limit: 3, Offset: 1}
					page, err := client.Objects().V1().GetV1alphaSLOs(t.Context(), request)
					require.NoError(t, err)
					require.Equal(t, sloListIdentities(expected[1:4]), sloListIdentities(page))
				})
			}
		}
	})
	t.Run("invalid list query", func(t *testing.T) {
		tests := []struct {
			name    string
			query   url.Values
			wantErr string
		}{
			{
				name:    "zero limit",
				query:   url.Values{"pagination.limit": []string{"0"}},
				wantErr: "'pagination.limit' with value '0'",
			},
			{
				name:    "limit above maximum",
				query:   url.Values{"pagination.limit": []string{"1001"}},
				wantErr: "must be less than or equal to '1000'",
			},
			{
				name:    "offset without limit",
				query:   url.Values{"pagination.offset": []string{"1"}},
				wantErr: "'pagination.limit' with value '0'",
			},
			{
				name: "negative offset",
				query: url.Values{
					"pagination.limit":  []string{"1"},
					"pagination.offset": []string{"-1"},
				},
				wantErr: "must be greater than or equal to '0'",
			},
			{
				name: "offset above maximum",
				query: url.Values{
					"pagination.limit":  []string{"1"},
					"pagination.offset": []string{"2147483648"},
				},
				wantErr: "must be less than or equal to '2147483647'",
			},
			{
				name:    "invalid sort column",
				query:   url.Values{"sort.column": []string{"createdAt"}},
				wantErr: "must be one of: project, service, slo, lastModifiedAt",
			},
			{
				name:    "invalid sort direction",
				query:   url.Values{"sort.direction": []string{"up"}},
				wantErr: "must be one of: asc, desc",
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				_, err := client.Objects().V1().Get(
					t.Context(),
					manifest.KindSLO,
					http.Header{sdk.HeaderProject: []string{project.GetName()}},
					test.query,
				)
				require.Error(t, err)
				var httpErr *sdk.HTTPError
				require.ErrorAs(t, err, &httpErr)
				require.Equal(t, http.StatusBadRequest, httpErr.StatusCode)
				require.ErrorContains(t, err, test.wantErr)
			})
		}
	})
}

func displayNameOrName(displayName, name string) string {
	if displayName != "" {
		return displayName
	}
	return name
}

func expectedSLOListOrder(
	t *testing.T,
	slos []v1alphaSLO.SLO,
	value func(v1alphaSLO.SLO) string,
	direction objectsV1.GetSLOsSortDirection,
) []v1alphaSLO.SLO {
	t.Helper()

	distinct := make(map[string]struct{}, len(slos))
	for _, slo := range slos {
		sortValue := value(slo)
		require.NotEmpty(t, sortValue)
		distinct[sortValue] = struct{}{}
	}
	require.Greater(t, len(distinct), 1)

	expected := slices.Clone(slos)
	slices.SortFunc(expected, func(left, right v1alphaSLO.SLO) int {
		comparison := cmp.Compare(value(left), value(right))
		if direction == objectsV1.GetSLOsSortDirectionDesc {
			comparison = -comparison
		}
		if comparison != 0 {
			return comparison
		}
		if comparison = cmp.Compare(left.Metadata.Project, right.Metadata.Project); comparison != 0 {
			return comparison
		}
		if comparison = cmp.Compare(left.Spec.Service, right.Spec.Service); comparison != 0 {
			return comparison
		}
		return cmp.Compare(left.Metadata.Name, right.Metadata.Name)
	})
	return expected
}

func sloListIdentities(slos []v1alphaSLO.SLO) []string {
	result := make([]string, len(slos))
	for i := range slos {
		result[i] = slos[i].Metadata.Project + "/" + slos[i].Metadata.Name
	}
	return result
}

func prepareObjectsForServiceNameFilteringTests(t *testing.T) (slos []v1alphaSLO.SLO, dependencies []manifest.Object) {
	t.Helper()

	agentType := v1alpha.Prometheus
	agent := e2etestutils.ProvisionStaticAgent(t, v1alpha.Prometheus)

	// Projects.
	project1 := newV1alphaProject(t, v1alphaProject.Metadata{
		Name:        "z-" + e2etestutils.GenerateName(),
		DisplayName: "a-project",
	})
	project2 := newV1alphaProject(t, v1alphaProject.Metadata{
		Name: "m-" + e2etestutils.GenerateName(),
	})
	// Services.
	service1Proj1 := newV1alphaService(t, v1alphaService.Metadata{
		Name:        "z-" + e2etestutils.GenerateName(),
		DisplayName: "a-service",
		Project:     project1.GetName(),
	})
	service2Proj1 := newV1alphaService(t, v1alphaService.Metadata{
		Name:    "m-" + e2etestutils.GenerateName(),
		Project: project1.GetName(),
	})
	service1Proj2 := newV1alphaService(t, v1alphaService.Metadata{
		Name:        service1Proj1.GetName(),
		DisplayName: "z-service",
		Project:     project2.GetName(),
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
	sloTemplate := e2etestutils.GetExampleObject[v1alphaSLO.SLO](t,
		manifest.KindSLO,
		e2etestutils.FilterExamplesByDataSourceType(agentType),
	)

	for _, params := range []struct {
		namePrefix  string
		displayName string
		project     string
		service     string
		labels      v1alpha.Labels
	}{
		{
			namePrefix:  "z",
			displayName: "a-slo",
			project:     project1.GetName(),
			service:     service1Proj1.GetName(),
			labels:      v1alpha.Labels{},
		},
		{
			namePrefix: "b",
			project:    project1.GetName(),
			service:    service1Proj1.GetName(),
			labels:     v1alpha.Labels{"service-name-filter": []string{"bar"}},
		},
		{
			namePrefix:  "a",
			displayName: "shared-slo",
			project:     project1.GetName(),
			service:     service1Proj1.GetName(),
			labels:      v1alpha.Labels{"service-name-filter": []string{"foo"}},
		},
		{
			namePrefix:  "y",
			displayName: "shared-slo",
			project:     project1.GetName(),
			service:     service2Proj1.GetName(),
			labels:      v1alpha.Labels{},
		},
		{
			namePrefix:  "c",
			displayName: "z-slo",
			project:     project2.GetName(),
			service:     service1Proj2.GetName(),
			labels:      v1alpha.Labels{},
		},
	} {
		params.labels[sloListSortFixtureLabel] = []string{""}
		slo := clone(t, sloTemplate)
		slo.Metadata = v1alphaSLO.Metadata{
			Name:        params.namePrefix + "-" + e2etestutils.GenerateName(),
			DisplayName: params.displayName,
			Project:     params.project,
			Labels:      e2etestutils.AnnotateLabels(t, params.labels),
			Annotations: commonAnnotations,
		}
		slo.Spec.Service = params.service
		slo.Spec.AlertPolicies = nil
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
