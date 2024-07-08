//go:build e2e_test

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	v1alphaRoleBinding "github.com/nobl9/nobl9-go/manifest/v1alpha/rolebinding"
	"github.com/nobl9/nobl9-go/sdk"
	objectsV1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
)

func Test_Objects_V1_V1alpha_RoleBinding(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	project := generateV1alphaProject(t)
	project.Metadata.DisplayName = "Project 1"
	allObjects := []manifest.Object{
		project,
		v1alphaRoleBinding.New(
			v1alphaRoleBinding.Metadata{Name: generateName()},
			v1alphaRoleBinding.Spec{
				GroupRef: ptr(generateName()),
				RoleRef:  "organization-blank",
			},
		),
		v1alphaRoleBinding.New(
			v1alphaRoleBinding.Metadata{Name: generateName()},
			v1alphaRoleBinding.Spec{
				User:       new(string),
				GroupRef:   new(string),
				RoleRef:    "",
				ProjectRef: "",
			},
		),
		v1alphaRoleBinding.New(
			v1alphaRoleBinding.Metadata{Name: generateName()},
			v1alphaRoleBinding.Spec{
				User:       new(string),
				GroupRef:   new(string),
				RoleRef:    "",
				ProjectRef: "",
			},
		),
		v1alphaRoleBinding.New(
			v1alphaRoleBinding.Metadata{Name: generateName()},
			v1alphaRoleBinding.Spec{
				User:       new(string),
				GroupRef:   new(string),
				RoleRef:    "",
				ProjectRef: "",
			},
		),
	}

	v1Apply(t, ctx, allObjects)
	t.Cleanup(func() { v1Delete(t, ctx, allObjects) })
	inputs := manifest.FilterByKind[v1alphaRoleBinding.RoleBinding](allObjects)

	filterTests := map[string]struct {
		request    objectsV1.GetRoleBindingsRequest
		expected   []v1alphaRoleBinding.RoleBinding
		returnsAll bool
	}{
		"all": {
			request:    objectsV1.GetRoleBindingsRequest{Project: sdk.ProjectsWildcard},
			expected:   manifest.FilterByKind[v1alphaRoleBinding.RoleBinding](allObjects),
			returnsAll: true,
		},
		"default project": {
			request:    objectsV1.GetRoleBindingsRequest{},
			expected:   []v1alphaRoleBinding.RoleBinding{inputs[0]},
			returnsAll: true,
		},
		"filter by project": {
			request: objectsV1.GetRoleBindingsRequest{
				Project: project.GetName(),
			},
			expected: []v1alphaRoleBinding.RoleBinding{inputs[1], inputs[2], inputs[3]},
		},
		"filter by name": {
			request: objectsV1.GetRoleBindingsRequest{
				Project: project.GetName(),
				Names:   []string{inputs[1].Metadata.Name},
			},
			expected: []v1alphaRoleBinding.RoleBinding{inputs[1]},
		},
	}
	for name, test := range filterTests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			actual, err := client.Objects().V1().GetV1alphaRoleBindings(ctx, test.request)
			require.NoError(t, err)
			if !test.returnsAll {
				require.Len(t, actual, len(test.expected))
			}
			assertSubset(t, actual, test.expected, assertV1alphaRoleBindingsAreEqual)
		})
	}
}

func assertV1alphaRoleBindingsAreEqual(t *testing.T, expected, actual v1alphaRoleBinding.RoleBinding) {
	t.Helper()
	assert.Equal(t, expected, actual)
}
