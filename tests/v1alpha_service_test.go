//go:build e2e_test

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	"github.com/nobl9/nobl9-go/sdk"
	objectsV1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
	"github.com/nobl9/nobl9-go/tests/e2etestutils"
)

func Test_Objects_V1_V1alpha_Service(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	project := generateV1alphaProject(t)
	project.Metadata.DisplayName = "Project 1"
	allObjects := []manifest.Object{
		project,
		newV1alphaService(t,
			v1alphaService.Metadata{
				Name:        e2etestutils.GenerateName(),
				DisplayName: "Service 1",
				Project:     defaultProject,
				Labels:      v1alpha.Labels{"team": []string{"orange"}},
				Annotations: commonAnnotations,
			},
		),
		newV1alphaService(t,
			v1alphaService.Metadata{
				Name:        e2etestutils.GenerateName(),
				DisplayName: "Service 2",
				Project:     project.GetName(),
				Labels:      v1alpha.Labels{"team": []string{"orange"}},
			},
		),
		newV1alphaService(t,
			v1alphaService.Metadata{
				Name:        e2etestutils.GenerateName(),
				DisplayName: "Service 3",
				Project:     project.GetName(),
				Labels:      v1alpha.Labels{"team": []string{"green"}},
				Annotations: commonAnnotations,
			},
		),
		newV1alphaService(t,
			v1alphaService.Metadata{
				Name:        e2etestutils.GenerateName(),
				DisplayName: "Service 4",
				Project:     project.GetName(),
				Labels:      v1alpha.Labels{"team": []string{"orange"}},
			},
		),
	}

	e2etestutils.V2Apply(t, allObjects)
	t.Cleanup(func() { e2etestutils.V2Delete(t, allObjects) })
	inputs := manifest.FilterByKind[v1alphaService.Service](allObjects)

	filterTests := map[string]struct {
		request    objectsV1.GetServicesRequest
		expected   []v1alphaService.Service
		returnsAll bool
	}{
		"all": {
			request:    objectsV1.GetServicesRequest{Project: sdk.ProjectsWildcard},
			expected:   inputs,
			returnsAll: true,
		},
		"default project": {
			request:    objectsV1.GetServicesRequest{},
			expected:   []v1alphaService.Service{inputs[0]},
			returnsAll: true,
		},
		"filter by project": {
			request: objectsV1.GetServicesRequest{
				Project: project.GetName(),
			},
			expected: []v1alphaService.Service{inputs[1], inputs[2], inputs[3]},
		},
		"filter by name": {
			request: objectsV1.GetServicesRequest{
				Project: project.GetName(),
				Names:   []string{inputs[1].Metadata.Name},
			},
			expected: []v1alphaService.Service{inputs[1]},
		},
		"filter by label": {
			request: objectsV1.GetServicesRequest{
				Project: project.GetName(),
				Labels:  e2etestutils.AnnotateLabels(t, v1alpha.Labels{"team": []string{"green"}}),
			},
			expected: []v1alphaService.Service{inputs[2]},
		},
		"filter by label and name": {
			request: objectsV1.GetServicesRequest{
				Project: project.GetName(),
				Names:   []string{inputs[3].Metadata.Name},
				Labels:  e2etestutils.AnnotateLabels(t, v1alpha.Labels{"team": []string{"orange"}}),
			},
			expected: []v1alphaService.Service{inputs[3]},
		},
	}
	for name, test := range filterTests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			actual, err := client.Objects().V1().GetV1alphaServices(ctx, test.request)
			require.NoError(t, err)
			if !test.returnsAll {
				require.Len(t, actual, len(test.expected))
			}
			assertSubset(t, actual, test.expected, assertV1alphaServicesAreEqual)
		})
	}
}

func newV1alphaService(
	t *testing.T,
	metadata v1alphaService.Metadata,
) v1alphaService.Service {
	t.Helper()
	metadata.Labels = e2etestutils.AnnotateLabels(t, metadata.Labels)
	metadata.Annotations = commonAnnotations
	return v1alphaService.New(metadata, v1alphaService.Spec{Description: e2etestutils.GetObjectDescription()})
}

func assertV1alphaServicesAreEqual(t *testing.T, expected, actual v1alphaService.Service) {
	t.Helper()
	assert.NotNil(t, actual.Status)
	actual.Status = nil
	assert.Equal(t, expected, actual)
}
