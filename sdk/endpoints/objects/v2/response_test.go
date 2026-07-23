package v2

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAnnotationsModelToV1alpha_ReplayStatusFields(t *testing.T) {
	replayStart := time.Date(2023, 5, 1, 17, 10, 5, 0, time.UTC)
	replayEnd := time.Date(2023, 5, 2, 17, 10, 5, 0, time.UTC)

	t.Run("copies the Replay block when set", func(t *testing.T) {
		resp := getAnnotationModel{
			Name:    "annotation",
			Project: "project",
			SloName: "my-slo",
			Status: getAnnotationModelStatus{
				UpdatedAt: time.Date(2023, 5, 2, 17, 10, 5, 0, time.UTC),
				IsSystem:  true,
				Replay: &getAnnotationModelReplay{
					PeriodStart:        &replayStart,
					PeriodEnd:          &replayEnd,
					ElapsedTimeSeconds: ptr(int64(3600)),
				},
			},
		}

		result := getAnnotationsModelToV1alpha(resp)

		require.NotNil(t, result.Status)
		require.NotNil(t, result.Status.Replay)
		require.NotNil(t, result.Status.Replay.PeriodStart)
		require.NotNil(t, result.Status.Replay.PeriodEnd)
		require.NotNil(t, result.Status.Replay.ElapsedTimeSeconds)
		assert.Equal(t, replayStart, *result.Status.Replay.PeriodStart)
		assert.Equal(t, replayEnd, *result.Status.Replay.PeriodEnd)
		assert.Equal(t, int64(3600), *result.Status.Replay.ElapsedTimeSeconds)
	})

	t.Run("leaves the Replay block nil when absent", func(t *testing.T) {
		resp := getAnnotationModel{
			Name:    "annotation",
			Project: "project",
			SloName: "my-slo",
			Status: getAnnotationModelStatus{
				UpdatedAt: time.Date(2023, 5, 2, 17, 10, 5, 0, time.UTC),
				IsSystem:  false,
			},
		}

		result := getAnnotationsModelToV1alpha(resp)

		require.NotNil(t, result.Status)
		assert.Nil(t, result.Status.Replay)
	})
}

func ptr[T any](v T) *T { return &v }
