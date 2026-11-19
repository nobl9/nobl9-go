package v1

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
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
	assert.Nil(t, item.IsComposite)
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
	err := json.Unmarshal([]byte(`{"available":false,"reason":"formatted reason from server"}`), &availability)

	require.NoError(t, err)
	assert.Equal(t, ReplayAvailabilityReason("formatted reason from server"), availability.Reason)
}

func TestReplayAvailabilityFailoverUnmarshal(t *testing.T) {
	t.Parallel()

	var availability ReplayAvailability
	err := json.Unmarshal([]byte(`{
		"available": false,
		"reason": "failover_period_overlap",
		"message": "Replay is unavailable: Direct p/cw failed over.",
		"failover": {
			"metricSource": {"name": "cw-failover-test", "project": "cloudwatch-failover-test", "kind": "Direct"},
			"from": {"region": "us-west-2"},
			"to": {"region": "us-east-2"},
			"periodStart": "2026-11-09T10:00:00Z",
			"periodEnd": "2026-11-09T11:30:00Z"
		}
	}`), &availability)

	require.NoError(t, err)
	periodEnd := time.Date(2026, 11, 9, 11, 30, 0, 0, time.UTC)
	assert.Equal(t, ReplayAvailability{
		Available: false,
		Reason:    ReplayFailoverPeriodOverlap,
		Message:   "Replay is unavailable: Direct p/cw failed over.",
		Failover: &ReplayFailover{
			MetricSource: v1alphaSLO.MetricSourceSpec{
				Name:    "cw-failover-test",
				Project: "cloudwatch-failover-test",
				Kind:    manifest.KindDirect,
			},
			From:        v1alpha.FailoverTarget{Region: "us-west-2"},
			To:          v1alpha.FailoverTarget{Region: "us-east-2"},
			PeriodStart: time.Date(2026, 11, 9, 10, 0, 0, 0, time.UTC),
			PeriodEnd:   &periodEnd,
		},
	}, availability)
}
