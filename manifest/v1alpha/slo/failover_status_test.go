package slo

import (
	"encoding/json"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

const sloFailoverStatusYAML = `
failover:
  state: failover
  objectives:
    - name: good
      countMetrics:
        good:
          active:
            region: us-east-2
        total:
          active:
            region: us-east-1
    - name: raw
      rawMetric:
        active:
          region: us-east-2
`

func TestStatus_Failover_JSON(t *testing.T) {
	t.Run("round-trips the failover status of a ratio and a raw objective", func(t *testing.T) {
		// arrange
		var status Status
		require.NoError(t, yaml.UnmarshalWithOptions([]byte(sloFailoverStatusYAML), &status, yaml.Strict()))

		// act
		actual, err := json.Marshal(status)

		// assert
		require.NoError(t, err)
		expected, err := yaml.YAMLToJSON([]byte(sloFailoverStatusYAML))
		require.NoError(t, err)
		assert.JSONEq(t, string(expected), string(actual))
		assert.Equal(t, &SLOFailoverStatus{
			State: v1alpha.FailoverStateFailover,
			Objectives: []SLOFailoverObjective{
				{
					Name: "good",
					CountMetrics: &SLOFailoverCountMetrics{
						Good:  &SLOFailoverMetric{Active: v1alpha.FailoverTarget{Region: "us-east-2"}},
						Total: &SLOFailoverMetric{Active: v1alpha.FailoverTarget{Region: "us-east-1"}},
					},
				},
				{
					Name:      "raw",
					RawMetric: &SLOFailoverMetric{Active: v1alpha.FailoverTarget{Region: "us-east-2"}},
				},
			},
		}, status.Failover)
	})
	t.Run("omits failover when the SLO has no failover status", func(t *testing.T) {
		// act
		data, err := json.Marshal(Status{UpdatedAt: "2026-11-13T10:00:00Z"})

		// assert
		require.NoError(t, err)
		assert.JSONEq(t, `{"updatedAt":"2026-11-13T10:00:00Z"}`, string(data))
	})
}
