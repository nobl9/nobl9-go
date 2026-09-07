package v1alpha

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
)

func TestZscalerDefaults(t *testing.T) {
	assert.Equal(t, 20*time.Minute, GetQueryDelayDefaults()[Zscaler].Duration())
	for _, kind := range []manifest.Kind{manifest.KindAgent, manifest.KindDirect} {
		t.Run(kind.String(), func(t *testing.T) {
			duration, err := GetDataRetrievalMaxDuration(kind, Zscaler)
			require.NoError(t, err)
			assert.Equal(t, 14*24*time.Hour, duration.Duration())
		})
	}
}
