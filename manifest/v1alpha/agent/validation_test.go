package agent

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

//go:embed test_data/expected_error.txt
var expectedError string

func TestValidate_AllErrors(t *testing.T) {
	err := validate(Agent{
		Kind: manifest.KindProject,
		Metadata: Metadata{
			Name:        strings.Repeat("MY AGENT", 20),
			DisplayName: strings.Repeat("my-agent", 10),
			Project:     strings.Repeat("MY PROJECT", 20),
		},
		Spec: Spec{
			Description: strings.Repeat("l", 2000),
			Prometheus: &PrometheusConfig{
				URL: ptr("https://prometheus-service.monitoring:8080"),
			},
		},
		ManifestSource: "/home/me/agent.yaml",
	})
	assert.Equal(t, strings.TrimSuffix(expectedError, "\n"), err.Error())
}

func TestValidateSpec(t *testing.T) {

}

func validAgent() Agent {
	return New(
		Metadata{
			Name:        "prometheus",
			DisplayName: "Prometheus Agent",
			Project:     "default",
		}, Spec{
			Description:    "Example Prometheus Agent",
			ReleaseChannel: v1alpha.ReleaseChannelStable,
			Prometheus: &PrometheusConfig{
				URL: ptr("https://prometheus-service.monitoring:8080"),
			},
		})
}

func ptr[T any](v T) *T { return &v }
