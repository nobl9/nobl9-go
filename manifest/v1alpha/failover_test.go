package v1alpha

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFailoverStatus_JSON(t *testing.T) {
	t.Run("writes the target object under active", func(t *testing.T) {
		// arrange
		status := FailoverStatus{
			State:              FailoverStateFailover,
			LastEventID:        "evt-0002",
			LastEventTimestamp: "2026-11-13T10:00:00Z",
			Active:             FailoverTarget{Region: "us-east-2"},
		}

		// act
		data, err := json.Marshal(status)

		// assert
		require.NoError(t, err)
		assert.JSONEq(t, `{"state":"failover","lastEventId":"evt-0002",`+
			`"lastEventTimestamp":"2026-11-13T10:00:00Z","active":{"region":"us-east-2"}}`, string(data))
	})
	t.Run("omits the event fields of a primary state with no event", func(t *testing.T) {
		// arrange
		status := FailoverStatus{State: FailoverStatePrimary, Active: FailoverTarget{Region: "us-west-2"}}

		// act
		data, err := json.Marshal(status)

		// assert
		require.NoError(t, err)
		assert.JSONEq(t, `{"state":"primary","active":{"region":"us-west-2"}}`, string(data))
	})
}
