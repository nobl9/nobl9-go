//go:build e2e_test

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaAnnotation "github.com/nobl9/nobl9-go/manifest/v1alpha/annotation"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	"github.com/nobl9/nobl9-go/sdk"
	objectsV1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
	"github.com/nobl9/nobl9-go/tests/e2etestutils"
)

func Test_Objects_V1_V1alpha_Annotation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	project := generateV1alphaProject(t)

	service := newV1alphaService(t, v1alphaService.Metadata{
		Name:    e2etestutils.GenerateName(),
		Project: project.GetName(),
	})
	defaultProjectService := newV1alphaService(t, v1alphaService.Metadata{
		Name:    e2etestutils.GenerateName(),
		Project: defaultProject,
	})

	dataSourceType := v1alpha.Datadog
	direct := e2etestutils.ProvisionStaticDirect(t, dataSourceType)

	slo := e2etestutils.GetExampleObject[v1alphaSLO.SLO](t,
		manifest.KindSLO,
		e2etestutils.FilterExamplesByDataSourceType(dataSourceType),
	)
	slo.Spec.AnomalyConfig = nil
	slo.Metadata.Name = e2etestutils.GenerateName()
	slo.Metadata.Project = project.GetName()
	slo.Spec.Indicator.MetricSource = v1alphaSLO.MetricSourceSpec{
		Name:    direct.Metadata.Name,
		Project: direct.Metadata.Project,
		Kind:    manifest.KindDirect,
	}
	slo.Spec.AlertPolicies = nil
	slo.Spec.Service = service.Metadata.Name
	slo.Spec.Objectives[0].Name = "good"

	defaultProjectSLO := deepCopyObject(t, slo)
	defaultProjectSLO.Metadata.Name = e2etestutils.GenerateName()
	defaultProjectSLO.Metadata.Project = defaultProject
	defaultProjectSLO.Spec.Service = defaultProjectService.Metadata.Name

	allObjects := make([]manifest.Object, 0)
	allObjects = append(
		allObjects,
		project,
		service,
		defaultProjectService,
		slo,
		defaultProjectSLO,
	)

	annotations := []v1alphaAnnotation.Annotation{
		v1alphaAnnotation.New(
			v1alphaAnnotation.Metadata{
				Name:    e2etestutils.GenerateName(),
				Project: defaultProject,
			},
			v1alphaAnnotation.Spec{
				Slo:           defaultProjectSLO.Metadata.Name,
				ObjectiveName: "good",
				Description:   e2etestutils.GetObjectDescription(),
				StartTime:     mustParseTime("2024-05-01T12:00:00Z").UTC(),
				EndTime:       mustParseTime("2024-05-04T10:00:00Z").UTC(),
			},
		),
		v1alphaAnnotation.New(
			v1alphaAnnotation.Metadata{
				Name:    e2etestutils.GenerateName(),
				Project: project.GetName(),
			},
			v1alphaAnnotation.Spec{
				Slo:         slo.Metadata.Name,
				Description: e2etestutils.GetObjectDescription(),
				StartTime:   mustParseTime("2024-05-16T14:00:00Z").UTC(),
				EndTime:     mustParseTime("2024-05-16T15:00:00Z").UTC(),
			},
		),
		v1alphaAnnotation.New(
			v1alphaAnnotation.Metadata{
				Name:    e2etestutils.GenerateName(),
				Project: project.GetName(),
			},
			v1alphaAnnotation.Spec{
				Slo:         slo.Metadata.Name,
				Description: e2etestutils.GetObjectDescription(),
				StartTime:   mustParseTime("2024-05-17T14:00:00Z").UTC(),
				EndTime:     mustParseTime("2024-05-17T15:00:00Z").UTC(),
			},
		),
		v1alphaAnnotation.New(
			v1alphaAnnotation.Metadata{
				Name:    generateName(),
				Project: project.GetName(),
				Labels:  annotateLabels(t, v1alpha.Labels{"team": []string{"blue"}}),
			},
			v1alphaAnnotation.Spec{
				Slo:         slo.Metadata.Name,
				Description: objectDescription,
				StartTime:   mustParseTime("2024-05-18T14:00:00Z").UTC(),
				EndTime:     mustParseTime("2024-05-18T15:00:00Z").UTC(),
			},
		),
	}
	for _, annotation := range annotations {
		allObjects = append(allObjects, annotation)
	}

	e2etestutils.V1Apply(t, allObjects)
	t.Cleanup(func() { e2etestutils.V1Delete(t, allObjects) })
	inputs := manifest.FilterByKind[v1alphaAnnotation.Annotation](allObjects)

	filterTests := map[string]struct {
		request    objectsV1.GetAnnotationsRequest
		expected   []v1alphaAnnotation.Annotation
		returnsAll bool
	}{
		"all": {
			request:    objectsV1.GetAnnotationsRequest{Project: sdk.ProjectsWildcard},
			expected:   manifest.FilterByKind[v1alphaAnnotation.Annotation](allObjects),
			returnsAll: true,
		},
		"default project": {
			request:    objectsV1.GetAnnotationsRequest{},
			expected:   []v1alphaAnnotation.Annotation{inputs[0]},
			returnsAll: true,
		},
		"filter by project": {
			request: objectsV1.GetAnnotationsRequest{
				Project: project.GetName(),
			},
			expected: inputs[1:],
		},
		"filter by name": {
			request: objectsV1.GetAnnotationsRequest{
				Project: project.GetName(),
				Names:   []string{inputs[1].Metadata.Name},
			},
			expected: []v1alphaAnnotation.Annotation{inputs[1]},
		},
		"filter by slo name": {
			request: objectsV1.GetAnnotationsRequest{
				Project: project.GetName(),
				SLOName: slo.Metadata.Name,
			},
			expected: inputs[1:],
		},
		"filter by from": {
			request: objectsV1.GetAnnotationsRequest{
				Project: project.GetName(),
				From:    mustParseTime("2024-05-17T10:00:00Z"),
			},
			expected: []v1alphaAnnotation.Annotation{inputs[2]},
		},
		"filter by to": {
			request: objectsV1.GetAnnotationsRequest{
				Project: project.GetName(),
				To:      mustParseTime("2024-05-16T20:00:00Z"),
			},
			expected: []v1alphaAnnotation.Annotation{inputs[1]},
		},
	}
	for name, test := range filterTests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			test.request.UserAnnotations = ptr(true)
			test.request.SystemAnnotations = ptr(false)
			actual, err := client.Objects().V1().GetV1alphaAnnotations(ctx, test.request)
			require.NoError(t, err)
			if !test.returnsAll {
				require.Len(t, actual, len(test.expected))
			}
			assertSubset(t, actual, test.expected, assertV1alphaAnnotationsAreEqual)
		})
	}
}

func assertV1alphaAnnotationsAreEqual(t *testing.T, expected, actual v1alphaAnnotation.Annotation) {
	t.Helper()
	if assert.NotNil(t, actual.Status) {
		assert.False(t, actual.Status.IsSystem)
	}
	actual.Status = nil
	assert.Regexp(t, userIDRegexp, actual.Spec.CreatedBy)
	actual.Spec.CreatedBy = ""
	assert.Equal(t, expected, actual)
}
