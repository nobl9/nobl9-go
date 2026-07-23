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

	t.Run("all Replay fields present are marshaled under status.replay", func(t *testing.T) {
		a := validAnnotation()
		a.Spec.Category = CategoryReplay
		a.Status = &Status{
			UpdatedAt: "2023-05-02T17:10:05Z",
			IsSystem:  true,
			Replay: &ReplayStatus{
				PeriodStart:        &replayStart,
				PeriodEnd:          &replayEnd,
				ElapsedTimeSeconds: ptr(int64(3600)),
			},
		}

		data, err := json.Marshal(a)
		require.NoError(t, err)

		replay := decodeReplay(t, data)
		assert.Equal(t, "2023-05-01T17:10:05Z", replay["periodStart"])
		assert.Equal(t, "2023-05-02T17:10:05Z", replay["periodEnd"])
		assert.Equal(t, float64(3600), replay["elapsedTimeSeconds"])
	})

	t.Run("replay block absent is omitted from status", func(t *testing.T) {
		a := validAnnotation()
		a.Spec.Category = CategoryComment
		a.Status = &Status{
			UpdatedAt: "2023-05-02T17:10:05Z",
			IsSystem:  false,
		}

		data, err := json.Marshal(a)
		require.NoError(t, err)

		status := decodeStatus(t, data)
		_, found := status["replay"]
		assert.False(t, found, "expected key %q to be absent from status", "replay")
	})

	t.Run("replay block with only the period bounds omits elapsedTimeSeconds", func(t *testing.T) {
		a := validAnnotation()
		a.Spec.Category = CategoryReplay
		a.Status = &Status{
			UpdatedAt: "2023-05-02T17:10:05Z",
			IsSystem:  true,
			Replay: &ReplayStatus{
				PeriodStart: &replayStart,
				PeriodEnd:   &replayEnd,
			},
		}

		data, err := json.Marshal(a)
		require.NoError(t, err)

		replay := decodeReplay(t, data)
		assert.Equal(t, "2023-05-01T17:10:05Z", replay["periodStart"])
		assert.Equal(t, "2023-05-02T17:10:05Z", replay["periodEnd"])
		_, found := replay["elapsedTimeSeconds"]
		assert.False(t, found, "expected elapsedTimeSeconds to be absent from status.replay")
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
			UpdatedAt: "2023-05-02T17:10:05Z",
			IsSystem:  true,
			Replay: &ReplayStatus{
				PeriodStart:        &replayStart,
				PeriodEnd:          &replayEnd,
				ElapsedTimeSeconds: ptr(int64(3600)),
			},
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

// decodeReplay returns the status.replay object from a marshaled Annotation.
func decodeReplay(t *testing.T, data []byte) map[string]any {
	t.Helper()
	replay, ok := decodeStatus(t, data)["replay"].(map[string]any)
	require.True(t, ok, "status.replay object should be present")
	return replay
}

func ptr[T any](v T) *T { return &v }
