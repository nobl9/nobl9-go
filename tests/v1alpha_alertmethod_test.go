//go:build e2e_test

package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaAlertMethod "github.com/nobl9/nobl9-go/manifest/v1alpha/alertmethod"
	"github.com/nobl9/nobl9-go/sdk"
	objectsV1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
	"github.com/nobl9/nobl9-go/tests/e2etestutils"
)

func Test_Objects_V1_V1alpha_AlertMethod(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	project := generateV1alphaProject(t)
	alertMethodTypes := v1alpha.AlertMethodTypeValues()
	allObjects := make([]manifest.Object, 0, len(alertMethodTypes)+1)
	allObjects = append(allObjects, project)

	for i, typ := range alertMethodTypes {
		method := newV1alphaAlertMethod(t,
			typ,
			v1alphaAlertMethod.Metadata{
				Name:        e2etestutils.GenerateName(),
				DisplayName: fmt.Sprintf("Alert Method %d", i),
				Project:     project.GetName(),
			},
		)
		if i == 0 {
			method.Metadata.Project = defaultProject
		}
		allObjects = append(allObjects, method)
	}

	e2etestutils.V1Apply(t, allObjects)
	t.Cleanup(func() { e2etestutils.V1Delete(t, allObjects) })
	inputs := manifest.FilterByKind[v1alphaAlertMethod.AlertMethod](allObjects)

	filterTests := map[string]struct {
		request    objectsV1.GetAlertMethodsRequest
		expected   []v1alphaAlertMethod.AlertMethod
		returnsAll bool
	}{
		"all": {
			request:    objectsV1.GetAlertMethodsRequest{Project: sdk.ProjectsWildcard},
			expected:   inputs,
			returnsAll: true,
		},
		"default project": {
			request:    objectsV1.GetAlertMethodsRequest{},
			expected:   []v1alphaAlertMethod.AlertMethod{inputs[0]},
			returnsAll: true,
		},
		"filter by project": {
			request: objectsV1.GetAlertMethodsRequest{
				Project: project.GetName(),
			},
			expected: inputs[1:],
		},
		"filter by name": {
			request: objectsV1.GetAlertMethodsRequest{
				Project: project.GetName(),
				Names:   []string{inputs[3].Metadata.Name},
			},
			expected: []v1alphaAlertMethod.AlertMethod{inputs[3]},
		},
	}
	for name, test := range filterTests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			actual, err := client.Objects().V1().GetV1alphaAlertMethods(ctx, test.request)
			require.NoError(t, err)
			if !test.returnsAll {
				require.Len(t, actual, len(test.expected))
			}
			assertSubset(t, actual, test.expected, assertV1alphaAlertMethodsAreEqual)
		})
	}
}

func newV1alphaAlertMethod(
	t *testing.T,
	typ v1alpha.AlertMethodType,
	metadata v1alphaAlertMethod.Metadata,
) v1alphaAlertMethod.AlertMethod {
	t.Helper()
	variant := e2etestutils.GetExampleObject[v1alphaAlertMethod.AlertMethod](t,
		manifest.KindAlertMethod,
		e2etestutils.FilterExamplesByAlertMethodType(typ),
	)
	variant.Spec.Description = e2etestutils.GetObjectDescription()
	return v1alphaAlertMethod.New(metadata, variant.Spec)
}

func assertV1alphaAlertMethodsAreEqual(t *testing.T, expected, actual v1alphaAlertMethod.AlertMethod) {
	t.Helper()
	expected = deepCopyObject(t, expected)
	typ, err := expected.Spec.GetType()
	require.NoError(t, err)
	switch typ {
	case v1alpha.AlertMethodTypeDiscord:
		expected.Spec.Discord.URL = "[hidden]"
	case v1alpha.AlertMethodTypeJira:
		expected.Spec.Jira.APIToken = "[hidden]"
	case v1alpha.AlertMethodTypeOpsgenie:
		expected.Spec.Opsgenie.Auth = "[hidden]"
	case v1alpha.AlertMethodTypePagerDuty:
		expected.Spec.PagerDuty.IntegrationKey = "[hidden]"
	case v1alpha.AlertMethodTypeServiceNow:
		expected.Spec.ServiceNow.Password = "[hidden]"
	case v1alpha.AlertMethodTypeSlack:
		expected.Spec.Slack.URL = "[hidden]"
	case v1alpha.AlertMethodTypeTeams:
		expected.Spec.Teams.URL = "[hidden]"
	case v1alpha.AlertMethodTypeEmail:
	case v1alpha.AlertMethodTypeWebhook:
		expected.Spec.Webhook.URL = "[hidden]"
		for i, header := range expected.Spec.Webhook.Headers {
			if header.IsSecret {
				expected.Spec.Webhook.Headers[i].Value = "[hidden]"
			}
		}
	}
	assert.Equal(t, expected, actual)
}
