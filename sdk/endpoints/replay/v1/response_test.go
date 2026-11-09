package v1

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
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

func TestReplayAvailabilityReason_CloudWatchFailoverPeriodOverlap(t *testing.T) {
	t.Parallel()

	t.Run("parses the wire value into the alias", func(t *testing.T) {
		t.Parallel()

		// act
		reason, err := ParseReplayAvailabilityReason("cloudwatch_failover_period_overlap")

		// assert
		require.NoError(t, err)
		assert.Equal(t, ReplayCloudWatchFailoverPeriodOverlap, reason)
		assert.True(t, reason.IsValid())
	})
}

func TestReplayAvailability_RegionFailover(t *testing.T) {
	t.Parallel()

	t.Run("decodes the details of a closed failover period", func(t *testing.T) {
		t.Parallel()

		// arrange
		body := `{
			"available": false,
			"reason": "cloudwatch_failover_period_overlap",
			"regionFailover": {
				"dataSourceKind": "Direct",
				"dataSourceName": "cw-failover-test",
				"dataSourceProject": "cloudwatch-failover-test",
				"primaryRegion": "us-west-2",
				"failoverRegion": "us-east-2",
				"periodStart": "2026-11-09T10:00:00Z",
				"periodEnd": "2026-11-09T11:30:00Z"
			}
		}`

		// act
		var availability ReplayAvailability
		err := json.Unmarshal([]byte(body), &availability)

		// assert
		require.NoError(t, err)
		periodEnd := time.Date(2026, 11, 9, 11, 30, 0, 0, time.UTC)
		assert.Equal(t, ReplayAvailability{
			Available: false,
			Reason:    ReplayCloudWatchFailoverPeriodOverlap,
			RegionFailover: &ReplayRegionFailover{
				DataSourceKind:    manifest.KindDirect,
				DataSourceName:    "cw-failover-test",
				DataSourceProject: "cloudwatch-failover-test",
				PrimaryRegion:     "us-west-2",
				FailoverRegion:    "us-east-2",
				PeriodStart:       time.Date(2026, 11, 9, 10, 0, 0, 0, time.UTC),
				PeriodEnd:         &periodEnd,
			},
		}, availability)
	})

	t.Run("encodes an open failover period without periodEnd", func(t *testing.T) {
		t.Parallel()

		// arrange
		availability := ReplayAvailability{
			Reason: ReplayCloudWatchFailoverPeriodOverlap,
			RegionFailover: &ReplayRegionFailover{
				DataSourceKind:    manifest.KindAgent,
				DataSourceName:    "cw-agent",
				DataSourceProject: "cloudwatch-failover-test",
				PrimaryRegion:     "us-west-2",
				FailoverRegion:    "us-east-2",
				PeriodStart:       time.Date(2026, 11, 9, 10, 0, 0, 0, time.UTC),
			},
		}

		// act
		encoded, err := json.Marshal(availability)

		// assert
		require.NoError(t, err)
		assert.JSONEq(t, `{
			"available": false,
			"reason": "cloudwatch_failover_period_overlap",
			"regionFailover": {
				"dataSourceKind": "Agent",
				"dataSourceName": "cw-agent",
				"dataSourceProject": "cloudwatch-failover-test",
				"primaryRegion": "us-west-2",
				"failoverRegion": "us-east-2",
				"periodStart": "2026-11-09T10:00:00Z"
			}
		}`, string(encoded))
	})

	t.Run("encodes another reason without regionFailover", func(t *testing.T) {
		t.Parallel()

		// arrange
		availability := ReplayAvailability{Reason: ReplayAgentVersionDoesNotSupportReplay}

		// act
		encoded, err := json.Marshal(availability)

		// assert
		require.NoError(t, err)
		assert.JSONEq(t, `{"available":false,"reason":"agent_version_does_not_support_replay"}`, string(encoded))
	})
}
