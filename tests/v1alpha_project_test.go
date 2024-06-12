//go:build e2e_test

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	objectsV1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
)

func Test_Objects_V1_V1alpha_Project(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	inputs := []v1alphaProject.Project{
		v1alphaProject.New(
			v1alphaProject.Metadata{
				Name:        generateName(),
				Labels:      annotateLabels(t, v1alpha.Labels{"team": []string{"green"}}),
				Annotations: commonAnnotations,
			},
			v1alphaProject.Spec{
				Description: objectDescription,
			},
		),
		v1alphaProject.New(
			v1alphaProject.Metadata{
				Name:        generateName(),
				Labels:      annotateLabels(t, v1alpha.Labels{"team": []string{"orange"}}),
				Annotations: commonAnnotations,
			},
			v1alphaProject.Spec{
				Description: objectDescription,
			},
		),
		v1alphaProject.New(
			v1alphaProject.Metadata{
				Name:        generateName(),
				Labels:      annotateLabels(t, v1alpha.Labels{"team": []string{"orange"}}),
				Annotations: commonAnnotations,
			},
			v1alphaProject.Spec{
				Description: objectDescription,
			},
		),
	}

	v1Apply(t, ctx, inputs)
	t.Cleanup(func() { v1Delete(t, ctx, inputs) })

	filterTests := map[string]struct {
		request    objectsV1.GetProjectsRequest
		expected   []v1alphaProject.Project
		returnsAll bool
	}{
		"get all": {
			request:    objectsV1.GetProjectsRequest{},
			expected:   inputs,
			returnsAll: true,
		},
		"filter by name": {
			request: objectsV1.GetProjectsRequest{
				Names: []string{inputs[0].Metadata.Name},
			},
			expected: []v1alphaProject.Project{inputs[0]},
		},
		"filter by label": {
			request: objectsV1.GetProjectsRequest{
				Labels: annotateLabels(t, v1alpha.Labels{"team": []string{"orange"}}),
			},
			expected: []v1alphaProject.Project{inputs[1], inputs[2]},
		},
		"filter by label and name": {
			request: objectsV1.GetProjectsRequest{
				Names:  []string{inputs[2].Metadata.Name},
				Labels: annotateLabels(t, v1alpha.Labels{"team": []string{"orange"}}),
			},
			expected: []v1alphaProject.Project{inputs[2]},
		},
	}
	for name, test := range filterTests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			actual, err := client.Objects().V1().GetV1alphaProjects(ctx, test.request)
			require.NoError(t, err)
			if !test.returnsAll {
				require.Len(t, actual, len(test.expected))
			}
			assertSubset(t, actual, test.expected, assertProjectsAreEqual)
		})
	}
}

func assertProjectsAreEqual(t *testing.T, expected, actual v1alphaProject.Project) {
	t.Helper()
	assert.Regexp(t, timeRFC3339Regexp, actual.Spec.CreatedAt)
	assert.Regexp(t, userIDRegexp, actual.Spec.CreatedBy)
	actual.Spec.CreatedAt = ""
	actual.Spec.CreatedBy = ""
	assert.Equal(t, expected, actual)
}
