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
	v1alphaAgent "github.com/nobl9/nobl9-go/manifest/v1alpha/agent"
	"github.com/nobl9/nobl9-go/sdk"
	objectsV1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
)

func Test_Objects_V1_V1alpha_Agent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	project := generateV1alphaProject(t)
	agentTypes := v1alpha.DataSourceTypeValues()
	allObjects := make([]manifest.Object, 0, len(agentTypes)+1)
	allObjects = append(allObjects, project)

	for i, typ := range agentTypes {
		agent := newV1alphaAgent(t,
			typ,
			v1alphaAgent.Metadata{
				Name:        generateName(),
				DisplayName: fmt.Sprintf("Agent %d", i),
				Project:     project.GetName(),
			},
		)
		if i == 0 {
			agent.Metadata.Project = defaultProject
		}
		allObjects = append(allObjects, agent)
	}

	v1Apply(t, ctx, allObjects[:2])
	// Since we can only apply a single Agent per request, we need to split the applies.
	for _, obj := range allObjects[2:] {
		v1Apply(t, ctx, []manifest.Object{obj})
	}
	t.Cleanup(func() {
		// Since we can only apply a single Agent per request, we need to split the applies.
		for _, obj := range allObjects[2:] {
			v1Delete(t, ctx, []manifest.Object{obj})
		}
		v1Delete(t, ctx, allObjects[:2])
	})
	inputs := manifest.FilterByKind[v1alphaAgent.Agent](allObjects)

	filterTests := map[string]struct {
		request    objectsV1.GetAgentsRequest
		expected   []v1alphaAgent.Agent
		returnsAll bool
	}{
		"all": {
			request:    objectsV1.GetAgentsRequest{Project: sdk.ProjectsWildcard},
			expected:   inputs,
			returnsAll: true,
		},
		"default project": {
			request:    objectsV1.GetAgentsRequest{},
			expected:   []v1alphaAgent.Agent{inputs[0]},
			returnsAll: true,
		},
		"filter by project": {
			request: objectsV1.GetAgentsRequest{
				Project: project.GetName(),
			},
			expected: inputs[1:],
		},
		"filter by name": {
			request: objectsV1.GetAgentsRequest{
				Project: project.GetName(),
				Names:   []string{inputs[3].Metadata.Name},
			},
			expected: []v1alphaAgent.Agent{inputs[3]},
		},
	}
	for name, test := range filterTests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			actual, err := client.Objects().V1().GetV1alphaAgents(ctx, test.request)
			require.NoError(t, err)
			if !test.returnsAll {
				require.Len(t, actual, len(test.expected))
			}
			assertSubset(t, actual, test.expected, assertV1alphaAgentsAreEqual)
		})
	}
}

type dataSourceTypeGetter interface {
	GetDataSourceType() v1alpha.DataSourceType
}

func newV1alphaAgent(
	t *testing.T,
	typ v1alpha.DataSourceType,
	metadata v1alphaAgent.Metadata,
) v1alphaAgent.Agent {
	t.Helper()
	variant := getExample[v1alphaAgent.Agent](t,
		manifest.KindAgent,
		func(example v1alphaExamples.Example) bool {
			return example.(dataSourceTypeGetter).GetDataSourceType() == typ
		},
	)
	variant.Spec.Description = objectDescription
	return v1alphaAgent.New(metadata, variant.Spec)
}

func assertV1alphaAgentsAreEqual(t *testing.T, expected, actual v1alphaAgent.Agent) {
	t.Helper()
	assert.NotNil(t, actual.Status)
	typ, _ := expected.Spec.GetType()
	assert.Equal(t, typ.String(), actual.Status.AgentType)
	actual.Status = nil
	actual.Spec.Interval = nil
	actual.Spec.Timeout = nil
	actual.Spec.Jitter = nil
	if expected.Spec.HistoricalDataRetrieval != nil {
		assert.NotEmpty(t, actual.Spec.HistoricalDataRetrieval.MinimumAgentVersion)
		actual.Spec.HistoricalDataRetrieval.MinimumAgentVersion = ""
	}
	assert.NotEmpty(t, actual.Spec.QueryDelay.MinimumAgentVersion)
	actual.Spec.QueryDelay.MinimumAgentVersion = ""
	assert.NotEmpty(t, actual.OktaClientID)
	actual.OktaClientID = ""
	assert.Equal(t, expected, actual)
}
