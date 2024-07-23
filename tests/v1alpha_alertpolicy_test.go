//go:build e2e_test

package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1alphaExamples "github.com/nobl9/nobl9-go/internal/manifest/v1alpha/examples"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaAlertMethod "github.com/nobl9/nobl9-go/manifest/v1alpha/alertmethod"
	v1alphaAlertPolicy "github.com/nobl9/nobl9-go/manifest/v1alpha/alertpolicy"
	"github.com/nobl9/nobl9-go/sdk"
	objectsV1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
)

func Test_Objects_V1_V1alpha_AlertPolicy(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	project := generateV1alphaProject(t)
	alertMethod := newV1alphaAlertMethod(t, v1alpha.AlertMethodTypeSlack, v1alphaAlertMethod.Metadata{
		Name:        generateName(),
		DisplayName: "Alert Method",
		Project:     project.GetName(),
	})
	examples := examplesRegistry[manifest.KindAlertPolicy]
	allObjects := make([]manifest.Object, 0, len(examples)+2)
	allObjects = append(allObjects, project)
	allObjects = append(allObjects, alertMethod)

	for i, example := range examples {
		policy := newV1alphaAlertPolicy(t,
			v1alphaAlertPolicy.Metadata{
				Name:        generateName(),
				DisplayName: fmt.Sprintf("Alert Policy %d", i),
				Project:     project.GetName(),
			},
			example.GetVariant(),
			example.GetSubVariant(),
		)
		policy.Spec.AlertMethods = []v1alphaAlertPolicy.AlertMethodRef{
			{
				Metadata: v1alphaAlertPolicy.AlertMethodRefMetadata{
					Name:    alertMethod.Metadata.Name,
					Project: alertMethod.Metadata.Project,
				},
			},
		}
		for i := range policy.Spec.Conditions {
			if policy.Spec.Conditions[i].AlertingWindow == "" && policy.Spec.Conditions[i].LastsForDuration == "" {
				policy.Spec.Conditions[i].LastsForDuration = "0m"
			}
		}
		switch i {
		case 0:
			policy.Metadata.Project = defaultProject
		case 1:
			policy.Metadata.Labels["team"] = []string{"green"}
		case 2:
			policy.Metadata.Labels["team"] = []string{"orange"}
		case 3:
			policy.Metadata.Labels["team"] = []string{"orange"}
		}
		allObjects = append(allObjects, policy)
	}

	v1Apply(t, allObjects)
	t.Cleanup(func() { v1Delete(t, allObjects) })
	inputs := manifest.FilterByKind[v1alphaAlertPolicy.AlertPolicy](allObjects)

	filterTests := map[string]struct {
		request    objectsV1.GetAlertPolicyRequest
		expected   []v1alphaAlertPolicy.AlertPolicy
		returnsAll bool
	}{
		"all": {
			request:    objectsV1.GetAlertPolicyRequest{Project: sdk.ProjectsWildcard},
			expected:   manifest.FilterByKind[v1alphaAlertPolicy.AlertPolicy](allObjects),
			returnsAll: true,
		},
		"default project": {
			request:    objectsV1.GetAlertPolicyRequest{},
			expected:   []v1alphaAlertPolicy.AlertPolicy{inputs[0]},
			returnsAll: true,
		},
		"filter by project": {
			request: objectsV1.GetAlertPolicyRequest{
				Project: project.GetName(),
			},
			expected: inputs[1:],
		},
		"filter by name": {
			request: objectsV1.GetAlertPolicyRequest{
				Project: project.GetName(),
				Names:   []string{inputs[4].Metadata.Name},
			},
			expected: []v1alphaAlertPolicy.AlertPolicy{inputs[4]},
		},
		"filter by label": {
			request: objectsV1.GetAlertPolicyRequest{
				Project: project.GetName(),
				Labels:  annotateLabels(t, v1alpha.Labels{"team": []string{"green"}}),
			},
			expected: []v1alphaAlertPolicy.AlertPolicy{inputs[1]},
		},
		"filter by label and name": {
			request: objectsV1.GetAlertPolicyRequest{
				Project: project.GetName(),
				Names:   []string{inputs[3].Metadata.Name},
				Labels:  annotateLabels(t, v1alpha.Labels{"team": []string{"orange"}}),
			},
			expected: []v1alphaAlertPolicy.AlertPolicy{inputs[3]},
		},
	}
	for name, test := range filterTests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			actual, err := client.Objects().V1().GetV1alphaAlertPolicies(ctx, test.request)
			require.NoError(t, err)
			if !test.returnsAll {
				require.Len(t, actual, len(test.expected))
			}
			assertSubset(t, actual, test.expected, assertV1alphaAlertPoliciesAreEqual)
		})
	}
}

func newV1alphaAlertPolicy(
	t *testing.T,
	metadata v1alphaAlertPolicy.Metadata,
	variant,
	subVariant string,
) v1alphaAlertPolicy.AlertPolicy {
	t.Helper()
	metadata.Labels = annotateLabels(t, metadata.Labels)
	metadata.Annotations = commonAnnotations
	ap := getExample[v1alphaAlertPolicy.AlertPolicy](t,
		manifest.KindAlertPolicy,
		func(example v1alphaExamples.Example) bool {
			return example.GetVariant() == variant && example.GetSubVariant() == subVariant
		},
	)
	ap.Spec.Description = objectDescription
	return v1alphaAlertPolicy.New(metadata, ap.Spec)
}

func assertV1alphaAlertPoliciesAreEqual(t *testing.T, expected, actual v1alphaAlertPolicy.AlertPolicy) {
	t.Helper()
	assert.Equal(t, expected, actual)
}
