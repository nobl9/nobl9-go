package v1alpha

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestThousandEyesConfigs_GetFor(t *testing.T) {
	trueCases := []struct{ testType, releaseChannel string }{
		{testType: ThousandEyesServerAvailability, releaseChannel: ReleaseChannelStable.String()},
		{testType: ThousandEyesServerAvailability, releaseChannel: ReleaseChannelBeta.String()},
		{testType: ThousandEyesServerAvailability, releaseChannel: ReleaseChannelAlpha.String()},
		{testType: ThousandEyesServerTotalTime, releaseChannel: ReleaseChannelStable.String()},
		{testType: ThousandEyesServerTotalTime, releaseChannel: ReleaseChannelBeta.String()},
	}

	for _, c := range trueCases {
		_, ok := ThousandEyesTestAgentConfig.GetFor(c.testType, c.releaseChannel)
		assert.True(t, ok)
	}

	falseCases := []struct{ testType, releaseChannel string }{
		{testType: ThousandEyesServerTotalTime, releaseChannel: ReleaseChannelAlpha.String()},
		{testType: "non-existent", releaseChannel: ReleaseChannelAlpha.String()},
		{testType: ThousandEyesServerTotalTime, releaseChannel: "non-existent"},
	}

	for _, c := range falseCases {
		_, ok := ThousandEyesTestAgentConfig.GetFor(c.testType, c.releaseChannel)
		assert.False(t, ok)
	}
}
