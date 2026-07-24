package annotation

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpec_ReplayFactsSerialization(t *testing.T) {
	replayStart := time.Date(2023, 5, 1, 17, 10, 5, 0, time.UTC)
	replayEnd := time.Date(2023, 5, 2, 17, 10, 5, 0, time.UTC)

	t.Run("all Replay facts present are marshaled under spec.replay", func(t *testing.T) {
		a := validAnnotation()
		a.Spec.Category = CategoryReplay
		a.Spec.Replay = &ReplayFacts{
			PeriodStart:        replayStart,
			PeriodEnd:          replayEnd,
			ElapsedTimeSeconds: ptr(int64(3600)),
		}

		data, err := json.Marshal(a)
		require.NoError(t, err)

		replay := decodeReplay(t, data)
		assert.Equal(t, "2023-05-01T17:10:05Z", replay["periodStart"])
		assert.Equal(t, "2023-05-02T17:10:05Z", replay["periodEnd"])
		assert.Equal(t, float64(3600), replay["elapsedTimeSeconds"])
	})

	t.Run("spec.replay is the only manifest home for Replay facts", func(t *testing.T) {
		a := validAnnotation()
		a.Spec.Category = CategoryReplay
		a.Spec.Replay = &ReplayFacts{
			PeriodStart:        replayStart,
			PeriodEnd:          replayEnd,
			ElapsedTimeSeconds: ptr(int64(3600)),
		}
		a.Status = &Status{UpdatedAt: "2023-05-02T17:10:05Z", IsSystem: true}

		data, err := json.Marshal(a)
		require.NoError(t, err)

		var decoded map[string]any
		require.NoError(t, json.Unmarshal(data, &decoded))

		_, rootReplay := decoded["replay"]
		assert.False(t, rootReplay, "expected no root replay object")
		if status, ok := decoded["status"].(map[string]any); ok {
			_, statusReplay := status["replay"]
			assert.False(t, statusReplay, "expected no status.replay object")
		}
		spec, ok := decoded["spec"].(map[string]any)
		require.True(t, ok, "spec object should be present")
		for _, flat := range []string{
			"periodStart", "periodEnd", "elapsedTimeSeconds",
			"replayPeriodStart", "replayPeriodEnd", "replayElapsedTimeSeconds",
		} {
			_, found := spec[flat]
			assert.Falsef(t, found, "expected no flat Replay field %q on spec", flat)
		}
		_, specReplay := spec["replay"]
		assert.True(t, specReplay, "expected spec.replay to be present")
	})

	t.Run("replay block absent is omitted from spec", func(t *testing.T) {
		a := validAnnotation()
		a.Spec.Category = CategoryComment
		a.Spec.Replay = nil

		data, err := json.Marshal(a)
		require.NoError(t, err)

		spec := decodeSpec(t, data)
		_, found := spec["replay"]
		assert.False(t, found, "expected key %q to be absent from spec", "replay")
	})

	t.Run("replay block with only the period bounds omits elapsedTimeSeconds", func(t *testing.T) {
		a := validAnnotation()
		a.Spec.Category = CategoryReplay
		a.Spec.Replay = &ReplayFacts{
			PeriodStart: replayStart,
			PeriodEnd:   replayEnd,
		}

		data, err := json.Marshal(a)
		require.NoError(t, err)

		replay := decodeReplay(t, data)
		assert.Equal(t, "2023-05-01T17:10:05Z", replay["periodStart"])
		assert.Equal(t, "2023-05-02T17:10:05Z", replay["periodEnd"])
		_, found := replay["elapsedTimeSeconds"]
		assert.False(t, found, "expected elapsedTimeSeconds to be absent from spec.replay")
	})

	t.Run("marshal-unmarshal-marshal round-trip is stable", func(t *testing.T) {
		a := validAnnotation()
		a.Spec.Category = CategoryReplay
		a.Spec.Replay = &ReplayFacts{
			PeriodStart:        replayStart,
			PeriodEnd:          replayEnd,
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

// decodeSpec reads a JSON-encoded Annotation and returns its spec object.
func decodeSpec(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(data, &decoded))
	spec, ok := decoded["spec"].(map[string]any)
	require.True(t, ok, "spec object should be present")
	return spec
}

// decodeReplay returns the spec.replay object from a marshaled Annotation.
func decodeReplay(t *testing.T, data []byte) map[string]any {
	t.Helper()
	replay, ok := decodeSpec(t, data)["replay"].(map[string]any)
	require.True(t, ok, "spec.replay object should be present")
	return replay
}

func ptr[T any](v T) *T { return &v }
