package v1

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
