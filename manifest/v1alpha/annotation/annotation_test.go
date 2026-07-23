package annotation

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatus_ReplayFieldsSerialization(t *testing.T) {
	replayStart := time.Date(2023, 5, 1, 17, 10, 5, 0, time.UTC)
	replayEnd := time.Date(2023, 5, 2, 17, 10, 5, 0, time.UTC)

	t.Run("all Replay fields present are marshaled under status", func(t *testing.T) {
		a := validAnnotation()
		a.Spec.Category = CategoryReplay
		a.Status = &Status{
			UpdatedAt:          "2023-05-02T17:10:05Z",
			IsSystem:           true,
			ReplayPeriodStart:  &replayStart,
			ReplayPeriodEnd:    &replayEnd,
			ElapsedTimeSeconds: ptr(int64(3600)),
		}

		data, err := json.Marshal(a)
		require.NoError(t, err)

		status := decodeStatus(t, data)
		assert.Equal(t, "2023-05-01T17:10:05Z", status["replayPeriodStart"])
		assert.Equal(t, "2023-05-02T17:10:05Z", status["replayPeriodEnd"])
		assert.Equal(t, float64(3600), status["elapsedTimeSeconds"])
	})

	t.Run("Replay fields absent are omitted from status", func(t *testing.T) {
		a := validAnnotation()
		a.Spec.Category = CategoryComment
		a.Status = &Status{
			UpdatedAt: "2023-05-02T17:10:05Z",
			IsSystem:  false,
		}

		data, err := json.Marshal(a)
		require.NoError(t, err)

		status := decodeStatus(t, data)
		for _, key := range []string{"replayPeriodStart", "replayPeriodEnd", "elapsedTimeSeconds"} {
			_, found := status[key]
			assert.Falsef(t, found, "expected key %q to be absent from status", key)
		}
	})

	t.Run("nil status omits the status key entirely", func(t *testing.T) {
		a := validAnnotation()
		a.Status = nil

		data, err := json.Marshal(a)
		require.NoError(t, err)

		var decoded map[string]any
		require.NoError(t, json.Unmarshal(data, &decoded))
		_, found := decoded["status"]
		assert.False(t, found)
	})

	t.Run("marshal-unmarshal-marshal round-trip is stable", func(t *testing.T) {
		a := validAnnotation()
		a.Spec.Category = CategoryReplay
		a.Status = &Status{
			UpdatedAt:          "2023-05-02T17:10:05Z",
			IsSystem:           true,
			ReplayPeriodStart:  &replayStart,
			ReplayPeriodEnd:    &replayEnd,
			ElapsedTimeSeconds: ptr(int64(3600)),
		}

		first, err := json.Marshal(a)
		require.NoError(t, err)

		var decoded Annotation
		require.NoError(t, json.Unmarshal(first, &decoded))

		second, err := json.Marshal(decoded)
		require.NoError(t, err)

		assert.JSONEq(t, string(first), string(second))
	})
}

// decodeStatus unmarshals a marshaled Annotation and returns its status object.
func decodeStatus(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(data, &decoded))
	status, ok := decoded["status"].(map[string]any)
	require.True(t, ok, "status object should be present")
	return status
}

func ptr[T any](v T) *T { return &v }
