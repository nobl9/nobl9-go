package v1

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReplayListItemStatusUnmarshal(t *testing.T) {
	t.Parallel()

	var item ReplayListItem
	err := json.Unmarshal([]byte(`{"project":"default","slo":"latency","status":"in progress"}`), &item)

	require.NoError(t, err)
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
	assert.Equal(t, ReplayStatusInProgress, replay.Status.Status)
	assert.Equal(t, ReplayCancellationStatusPossible, replay.Status.Cancellation)
	assert.Equal(t, DurationUnitHour, replay.Status.Unit)
}

func TestReplayAvailabilityReasonUnmarshal(t *testing.T) {
	t.Parallel()

	var availability ReplayAvailability
	err := json.Unmarshal([]byte(`{"available":false,"reason":"single_query_not_supported"}`), &availability)

	require.NoError(t, err)
	assert.Equal(t, ReplaySingleQueryNotSupported, availability.Reason)
}
