package v2

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaAnnotation "github.com/nobl9/nobl9-go/manifest/v1alpha/annotation"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
)

func TestGetAnnotationsModelToV1alpha_Replay(t *testing.T) {
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

func TestGetAnnotationsModelToV1alpha_Failover(t *testing.T) {
	t.Run("maps the root failover object into spec.failover when set", func(t *testing.T) {
		// arrange
		var resp getAnnotationModel
		require.NoError(t, json.Unmarshal([]byte(`{
  "name": "failover-123",
  "project": "default",
  "slo": "my-slo",
  "category": "Failover",
  "startTime": "2026-10-20T10:00:00Z",
  "status": {"updatedAt": "2026-10-20T10:00:05Z", "isSystem": true},
  "failover": {
    "metricSource": {"name": "cloudwatch-agent", "project": "infra", "kind": "Agent"},
    "from": {"region": "us-east-1"},
    "to": {"region": "us-west-2"},
    "eventId": "event-1"
  }
}`), &resp))

		// act
		result := getAnnotationsModelToV1alpha(resp)

		// assert
		assert.Equal(t, &v1alphaAnnotation.Failover{
			MetricSource: v1alphaSLO.MetricSourceSpec{Name: "cloudwatch-agent", Project: "infra", Kind: manifest.KindAgent},
			From:         v1alpha.FailoverTarget{Region: "us-east-1"},
			To:           v1alpha.FailoverTarget{Region: "us-west-2"},
			EventID:      "event-1",
		}, result.Spec.Failover)
	})

	t.Run("leaves spec.failover nil when the transport omits the failover object", func(t *testing.T) {
		// arrange
		resp := getAnnotationModel{
			Name:     "annotation",
			Project:  "project",
			SloName:  "my-slo",
			Category: v1alphaAnnotation.CategoryComment,
		}

		// act
		result := getAnnotationsModelToV1alpha(resp)

		// assert
		assert.Nil(t, result.Spec.Failover)
	})
}

func ptr[T any](v T) *T { return &v }
