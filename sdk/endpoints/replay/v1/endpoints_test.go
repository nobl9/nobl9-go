package v1

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEndpointsRunRejectsOverflowBeforeUsingClient(t *testing.T) {
	t.Parallel()

	err := NewEndpoints(nil).Run(t.Context(), RunRequest{
		Project: "project",
		SLO:     "slo",
		Duration: Duration{
			Unit:  DurationUnitDay,
			Value: overflowingDurationValue(t),
		},
	})

	require.ErrorIs(t, err, ErrDurationOverflow)
}

func TestEndpointsGetAvailabilityRejectsOverflowBeforeUsingClient(t *testing.T) {
	t.Parallel()

	_, err := NewEndpoints(nil).GetAvailability(t.Context(), GetAvailabilityRequest{
		SLOName:       "slo",
		DurationUnit:  DurationUnitDay,
		DurationValue: overflowingDurationValue(t),
	})

	require.ErrorIs(t, err, ErrDurationOverflow)
}

func TestEndpointsDeleteRejectsInvalidRequestBeforeUsingClient(t *testing.T) {
	t.Parallel()

	err := NewEndpoints(nil).Delete(t.Context(), DeleteRequest{})

	require.Error(t, err)
}

func TestEndpointsCancelRejectsInvalidRequestBeforeUsingClient(t *testing.T) {
	t.Parallel()

	err := NewEndpoints(nil).Cancel(t.Context(), CancelRequest{})

	require.Error(t, err)
}

func TestEndpointsGetStatusRejectsInvalidRequestBeforeUsingClient(t *testing.T) {
	t.Parallel()

	_, err := NewEndpoints(nil).GetStatus(t.Context(), GetStatusRequest{})

	require.Error(t, err)
}
