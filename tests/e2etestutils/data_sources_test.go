// cspell:ignore splunkobservability appdynamics unsuffixed
package e2etestutils

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestStaticFixtureName(t *testing.T) {
	tests := map[string]struct {
		kind    manifest.Kind
		typ     v1alpha.DataSourceType
		channel v1alpha.ReleaseChannel
		want    string
	}{
		"agent stable keeps legacy unsuffixed name": {
			kind:    manifest.KindAgent,
			typ:     v1alpha.SplunkObservability,
			channel: v1alpha.ReleaseChannelStable,
			want:    "e2e-agent-splunkobservability",
		},
		"agent unset channel keeps legacy unsuffixed name": {
			kind:    manifest.KindAgent,
			typ:     v1alpha.SplunkObservability,
			channel: 0,
			want:    "e2e-agent-splunkobservability",
		},
		"agent beta gets channel suffix": {
			kind:    manifest.KindAgent,
			typ:     v1alpha.SplunkObservability,
			channel: v1alpha.ReleaseChannelBeta,
			want:    "e2e-agent-splunkobservability-beta",
		},
		"agent alpha gets channel suffix": {
			kind:    manifest.KindAgent,
			typ:     v1alpha.SplunkObservability,
			channel: v1alpha.ReleaseChannelAlpha,
			want:    "e2e-agent-splunkobservability-alpha",
		},
		"direct stable keeps legacy unsuffixed name": {
			kind:    manifest.KindDirect,
			typ:     v1alpha.AppDynamics,
			channel: v1alpha.ReleaseChannelStable,
			want:    "e2e-direct-appdynamics",
		},
		"direct beta gets channel suffix": {
			kind:    manifest.KindDirect,
			typ:     v1alpha.AppDynamics,
			channel: v1alpha.ReleaseChannelBeta,
			want:    "e2e-direct-appdynamics-beta",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, staticFixtureName(tc.kind, tc.typ, tc.channel))
		})
	}
}
