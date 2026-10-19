package agent

import (
	"testing"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestValidateSpec_DatadogSite(t *testing.T) {
	for _, site := range []string{
		"com", "eu",
		"datadoghq.com", "us3.datadoghq.com", "us5.datadoghq.com", "datadoghq.eu",
		"ddog-gov.com", "us2.ddog-gov.com", "ap1.datadoghq.com", "ap2.datadoghq.com", "uk1.datadoghq.com",
	} {
		t.Run("accepts "+site, func(t *testing.T) {
			agent := validAgent(v1alpha.Datadog)
			agent.Spec.Datadog.Site = site
			err := validate(agent)
			testutils.AssertNoError(t, agent, err)
		})
	}
	for _, site := range []string{
		" ",
		"https://api.datadoghq.com",
		"datadoghq.com/",
		"datadoghq.com#x",
		"datadoghq.com:443",
		"api.datadoghq.com",
	} {
		t.Run("rejects "+site, func(t *testing.T) {
			agent := validAgent(v1alpha.Datadog)
			agent.Spec.Datadog.Site = site
			err := validate(agent)
			testutils.AssertContainsErrors(t, agent, err, 1, testutils.ExpectedError{
				Prop: "spec.datadog.site",
			})
		})
	}
}
