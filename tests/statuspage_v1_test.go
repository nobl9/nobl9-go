//go:build e2e_test

package tests

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	statusPageV1 "github.com/nobl9/nobl9-go/sdk/endpoints/statuspage/v1"
)

func Test_StatusPage_V1_GetStatus(t *testing.T) {
	t.Parallel()

	response, err := client.StatusPage().V1().GetStatus(t.Context())
	require.NoError(t, err)
	for _, component := range response.Components {
		assertStatusComponent(t, component)
	}
}

func Test_StatusPage_V1_ListDisruptions(t *testing.T) {
	t.Parallel()

	const limit = 10
	for _, state := range []statusPageV1.DisruptionState{
		statusPageV1.DisruptionStateImpacting,
		statusPageV1.DisruptionStateCleared,
	} {
		response, err := client.StatusPage().V1().ListDisruptions(
			t.Context(),
			statusPageV1.ListDisruptionsRequest{State: state, Limit: limit},
		)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(response.Disruptions), limit)
		assert.LessOrEqual(t, int64(len(response.Disruptions)), response.Total)
		for _, disruption := range response.Disruptions {
			assert.NotEmpty(t, disruption.ID)
			assert.NotEmpty(t, disruption.Severity)
			assert.Equal(t, state == statusPageV1.DisruptionStateCleared, disruption.IsCleared)
		}
	}
}

func assertStatusComponent(t *testing.T, component statusPageV1.StatusComponent) {
	t.Helper()
	assert.NotEmpty(t, component.ID)
	assert.NotEmpty(t, component.Name)
	assert.True(t, slices.Contains([]string{
		statusPageV1.ComponentStatusOperational,
		statusPageV1.ComponentStatusDegradedPerformance,
		statusPageV1.ComponentStatusMajorOutage,
	}, component.Status), "unexpected component status %q", component.Status)
	for _, child := range component.Children {
		assertStatusComponent(t, child)
	}
}
