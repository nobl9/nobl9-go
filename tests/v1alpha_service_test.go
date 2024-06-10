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
)

func Test_Objects_V1_V1alpha_Service(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	project := generateV1alphaProject(t)
	allObjects := []manifest.Object{
		project,
		v1alphaService.New(
			v1alphaService.Metadata{
				Name:        generateName(),
				Project:     sdk.DefaultProject,
				Labels:      annotateLabels(t, v1alpha.Labels{"team": []string{"orange"}}),
				Annotations: commonAnnotations,
			},
			v1alphaService.Spec{
				Description: objectDescription,
			},
		),
		v1alphaService.New(
			v1alphaService.Metadata{
				Name:        generateName(),
				Project:     project.GetName(),
				Labels:      annotateLabels(t, v1alpha.Labels{"team": []string{"orange"}}),
				Annotations: commonAnnotations,
			},
			v1alphaService.Spec{
				Description: objectDescription,
			},
		),
		v1alphaService.New(
			v1alphaService.Metadata{
				Name:        generateName(),
				Project:     project.GetName(),
				Labels:      annotateLabels(t, v1alpha.Labels{"team": []string{"green"}}),
				Annotations: commonAnnotations,
			},
			v1alphaService.Spec{
				Description: objectDescription,
			},
		),
		v1alphaService.New(
			v1alphaService.Metadata{
				Name:        generateName(),
				Project:     project.GetName(),
				Labels:      annotateLabels(t, v1alpha.Labels{"team": []string{"orange"}}),
				Annotations: commonAnnotations,
			},
			v1alphaService.Spec{
				Description: objectDescription,
			},
		),
	}

	v1Apply(t, ctx, allObjects)
	t.Cleanup(func() { v1Delete(t, ctx, allObjects) })
	inputs := manifest.FilterByKind[v1alphaService.Service](allObjects)

	filterTests := map[string]struct {
		request    objectsV1.GetServicesRequest
		expected   []v1alphaService.Service
		returnsAll bool
	}{
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
				Labels:  annotateLabels(t, v1alpha.Labels{"team": []string{"green"}}),
			},
			expected: []v1alphaService.Service{inputs[2]},
		},
		"filter by label and name": {
			request: objectsV1.GetServicesRequest{
				Project: project.GetName(),
				Names:   []string{inputs[3].Metadata.Name},
				Labels:  annotateLabels(t, v1alpha.Labels{"team": []string{"orange"}}),
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
			assertSubset(t, actual, test.expected, assertServicesAreEqual)
		})
	}
}

func assertServicesAreEqual(t *testing.T, expected, actual v1alphaService.Service) {
	t.Helper()
	assert.NotNil(t, actual.Status)
	actual.Status = nil
	assert.Equal(t, expected, actual)
}
