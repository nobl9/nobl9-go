package v2

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAnnotationsModelToV1alpha_ReplayFacts(t *testing.T) {
	replayStart := time.Date(2023, 5, 1, 17, 10, 5, 0, time.UTC)
	replayEnd := time.Date(2023, 5, 2, 17, 10, 5, 0, time.UTC)

	t.Run("maps the root replay object into spec.replay when set", func(t *testing.T) {
		resp := getAnnotationModel{
			Name:    "annotation",
			Project: "project",
			SloName: "my-slo",
			Status: getAnnotationModelStatus{
				UpdatedAt: time.Date(2023, 5, 2, 17, 10, 5, 0, time.UTC),
				IsSystem:  true,
			},
			Replay: &getAnnotationModelReplay{
				PeriodStart:        replayStart,
				PeriodEnd:          replayEnd,
				ElapsedTimeSeconds: ptr(int64(3600)),
			},
		}

		result := getAnnotationsModelToV1alpha(resp)

		require.NotNil(t, result.Spec.Replay)
		assert.Equal(t, replayStart, result.Spec.Replay.PeriodStart)
		assert.Equal(t, replayEnd, result.Spec.Replay.PeriodEnd)
		require.NotNil(t, result.Spec.Replay.ElapsedTimeSeconds)
		assert.Equal(t, int64(3600), *result.Spec.Replay.ElapsedTimeSeconds)
	})

	t.Run("omits elapsed time when the transport leaves it absent", func(t *testing.T) {
		resp := getAnnotationModel{
			Name:    "annotation",
			Project: "project",
			SloName: "my-slo",
			Replay: &getAnnotationModelReplay{
				PeriodStart: replayStart,
				PeriodEnd:   replayEnd,
			},
		}

		result := getAnnotationsModelToV1alpha(resp)

		require.NotNil(t, result.Spec.Replay)
		assert.Nil(t, result.Spec.Replay.ElapsedTimeSeconds)
	})

	t.Run("leaves spec.replay nil when the transport omits the replay object", func(t *testing.T) {
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

		assert.Nil(t, result.Spec.Replay)
	})
}

func ptr[T any](v T) *T { return &v }
