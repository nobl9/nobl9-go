package v1

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
)

func TestReplayListItemUnmarshal(t *testing.T) {
	t.Parallel()

	var item ReplayListItem
	err := json.Unmarshal([]byte(`{
		"project": "default",
		"slo": "latency",
		"createdAt": "2026-01-01T00:00:00Z",
		"status": "in progress"
	}`), &item)

	require.NoError(t, err)
	assert.Equal(t, "2026-01-01T00:00:00Z", item.CreatedAt)
	assert.Equal(t, ReplayListStatusInProgress, item.Status)
}

func TestReplayWithStatusUnmarshal(t *testing.T) {
	t.Parallel()

	var replay ReplayWithStatus
	err := json.Unmarshal([]byte(`{
		"project": "default",
		"slo": "latency",
		"status": {
			"source": "user",
			"status": "in progress",
			"cancellation": "possible",
			"triggeredBy": "user@example.com",
			"unit": "Hour",
			"startTime": "2026-01-01T00:00:00Z",
			"value": 1
		}
	}`), &replay)

	require.NoError(t, err)
	assert.Equal(t, ReplayListStatusInProgress, replay.Status.Status)
	assert.Equal(t, ReplayCancellationStatusPossible, replay.Status.Cancellation)
	assert.Equal(t, DurationUnitHour, replay.Status.Unit)
}

func TestReplayStatusToProcessStatus(t *testing.T) {
	t.Parallel()

	status := ReplayStatus{
		Source:       ReplaySourceUser,
		Status:       ReplayListStatusInProgress,
		Cancellation: ReplayCancellationStatusRequested,
		CanceledBy:   "canceler@example.com",
		TriggeredBy:  "user@example.com",
		Unit:         DurationUnitHour,
		StartTime:    "2026-01-01T00:00:00Z",
		EndTime:      "2026-01-01T01:00:00Z",
		Value:        1,
	}

	assert.Equal(t, v1alphaSLO.ProcessStatus{
		Status:       ReplayListStatusInProgress.String(),
		Cancellation: ReplayCancellationStatusRequested.String(),
		CanceledBy:   "canceler@example.com",
		TriggeredBy:  "user@example.com",
		Unit:         DurationUnitHour.String(),
		Value:        1,
		StartTime:    "2026-01-01T00:00:00Z",
		EndTime:      "2026-01-01T01:00:00Z",
	}, status.ToProcessStatus())
}

func TestReplayAvailabilityReasonUnmarshal(t *testing.T) {
	t.Parallel()

	var availability ReplayAvailability
	err := json.Unmarshal([]byte(`{"available":false,"reason":"single_query_not_supported"}`), &availability)

	require.NoError(t, err)
	assert.Equal(t, ReplaySingleQueryNotSupported, availability.Reason)
}
