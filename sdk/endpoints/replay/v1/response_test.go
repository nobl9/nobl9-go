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

func TestReplayAvailabilityReasonUnmarshal(t *testing.T) {
	t.Parallel()

	var availability ReplayAvailability
	err := json.Unmarshal([]byte(`{"available":false,"reason":"single_query_not_supported"}`), &availability)

	require.NoError(t, err)
	assert.Equal(t, ReplaySingleQueryNotSupported, availability.Reason)
}
