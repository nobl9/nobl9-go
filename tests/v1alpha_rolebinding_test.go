//go:build e2e_test

package tests

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	v1alphaRoleBinding "github.com/nobl9/nobl9-go/manifest/v1alpha/rolebinding"
	"github.com/nobl9/nobl9-go/sdk"
	objectsV1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
	"github.com/nobl9/nobl9-go/tests/e2etestutils"
)

func Test_Objects_V1_V1alpha_RoleBinding(t *testing.T) {
	t.Parallel()

	project := generateV1alphaProject(t)
	e2etestutils.V1Apply(t, []manifest.Object{project})
	implicitBindings, err := client.Objects().V1().GetV1alphaRoleBindings(t.Context(),
		objectsV1.GetRoleBindingsRequest{Project: project.GetName()})
	require.NoError(t, err)
	require.Len(t, implicitBindings, 1)
	implicitProjectBinding := implicitBindings[0]

	inputs := []v1alphaRoleBinding.RoleBinding{
		v1alphaRoleBinding.New(
			v1alphaRoleBinding.Metadata{Name: e2etestutils.GenerateName()},
			v1alphaRoleBinding.Spec{
				AccountID: ptr(e2etestutils.GenerateName()),
				RoleRef:   "organization-blank",
			},
		),
		v1alphaRoleBinding.New(
			v1alphaRoleBinding.Metadata{Name: e2etestutils.GenerateName()},
			v1alphaRoleBinding.Spec{
				GroupRef: ptr(e2etestutils.GenerateName()),
				RoleRef:  "organization-blank",
			},
		),
		v1alphaRoleBinding.New(
			v1alphaRoleBinding.Metadata{Name: e2etestutils.GenerateName()},
			v1alphaRoleBinding.Spec{
				AccountID:  ptr(e2etestutils.GenerateName()),
				RoleRef:    "project-viewer",
				ProjectRef: project.GetName(),
			},
		),
		v1alphaRoleBinding.New(
			v1alphaRoleBinding.Metadata{Name: e2etestutils.GenerateName()},
			v1alphaRoleBinding.Spec{
				GroupRef:   ptr(e2etestutils.GenerateName()),
				RoleRef:    "project-viewer",
				ProjectRef: project.GetName(),
			},
		),
		v1alphaRoleBinding.New(
			v1alphaRoleBinding.Metadata{Name: e2etestutils.GenerateName()},
			v1alphaRoleBinding.Spec{
				AccountID:  ptr(e2etestutils.GenerateName()),
				RoleRef:    "project-viewer",
				ProjectRef: defaultProject,
			},
		),
		v1alphaRoleBinding.New(
			v1alphaRoleBinding.Metadata{Name: e2etestutils.GenerateName()},
			v1alphaRoleBinding.Spec{
				GroupRef:   ptr(e2etestutils.GenerateName()),
				RoleRef:    "project-viewer",
				ProjectRef: defaultProject,
			},
		),
	}
	e2etestutils.V1Apply(t, inputs)
	t.Cleanup(func() {
		// Organization role bindings cannot be deleted.
		filterOrganizationBindings := func(r v1alphaRoleBinding.RoleBinding) bool {
			return !strings.HasPrefix(r.Spec.RoleRef, "organization-")
		}
		e2etestutils.V1Delete(t, filterSlice(inputs, filterOrganizationBindings))
	})

	filterTests := map[string]struct {
		request    objectsV1.GetRoleBindingsRequest
		expected   []v1alphaRoleBinding.RoleBinding
		returnsAll bool
	}{
		"all": {
			request:    objectsV1.GetRoleBindingsRequest{Project: sdk.ProjectsWildcard},
			expected:   userFieldsBackwardCompatible(inputs),
			returnsAll: true,
		},
		"default project": {
			request:    objectsV1.GetRoleBindingsRequest{},
			expected:   userFieldsBackwardCompatible([]v1alphaRoleBinding.RoleBinding{inputs[4], inputs[5]}),
			returnsAll: true,
		},
		"filter by project": {
			request: objectsV1.GetRoleBindingsRequest{
				Project: project.GetName(),
			},
			expected: userFieldsBackwardCompatible(
				[]v1alphaRoleBinding.RoleBinding{implicitProjectBinding, inputs[2], inputs[3]},
			),
		},
		"filter by name": {
			request: objectsV1.GetRoleBindingsRequest{
				Project: project.GetName(),
				Names:   []string{inputs[2].Metadata.Name},
			},
			expected: userFieldsBackwardCompatible(
				[]v1alphaRoleBinding.RoleBinding{inputs[2]},
			),
		},
	}
	for name, test := range filterTests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			actual, err := client.Objects().V1().GetV1alphaRoleBindings(t.Context(), test.request)
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

func userFieldsBackwardCompatible(bindings []v1alphaRoleBinding.RoleBinding) []v1alphaRoleBinding.RoleBinding {
	userFieldsBackwardCompatibleBindings := make([]v1alphaRoleBinding.RoleBinding, 0, len(bindings))
	for _, binding := range bindings {
		// nolint: staticcheck
		binding.Spec.User = binding.Spec.AccountID
		userFieldsBackwardCompatibleBindings = append(userFieldsBackwardCompatibleBindings, binding)
	}

	return userFieldsBackwardCompatibleBindings
}
