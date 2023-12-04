package alertpolicy

import (
	"embed"
	_ "embed"
	"path/filepath"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed test_data
var readerTestData embed.FS

func TestAlertCondition_UnmarshalYaml(t *testing.T) {
	tests := map[string]valueCase{
		"fails, wrong format": {
			yaml:     "condition_value_float64_serialization.yaml",
			expected: 0.00000002,
		},
		"fails, wrong unit in format": {
			yaml:     "condition_value_duration_serialization.yaml",
			expected: "5m",
		},
	}
	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := readerTestData.ReadFile(
				filepath.Join("test_data", testCase.yaml),
			)
			require.NoError(t, err)

			var condition AlertCondition
			err = yaml.Unmarshal(data, &condition)
			require.NoError(t, err)

			assert.NoError(t, err)
			assert.Equal(t, testCase.expected, condition.Value)
		})
	}
}

type valueCase struct {
	yaml     string
	expected interface{}
}
