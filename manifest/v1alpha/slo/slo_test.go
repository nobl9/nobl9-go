package slo

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpec_HasCompositeObjectives(t *testing.T) {
	t.Run("passes - composite SLO", func(t *testing.T) {
		slo := validCompositeSLO()

		assert.Equal(t, true, slo.Spec.HasCompositeObjectives())
	})

	t.Run("passes - normal SLO", func(t *testing.T) {
		slo := validSLO()

		assert.Equal(t, false, slo.Spec.HasCompositeObjectives())
	})
}

func TestAnomalyConfigNoData_TreatZeroAsNoDataSerialization(t *testing.T) {
	tests := map[string]struct {
		value           *bool
		expectedJSON    string
		expectedPresent bool
	}{
		"unset is omitted": {
			expectedJSON:    `{"alertMethods":[{"name":"my-name"}]}`,
			expectedPresent: false,
		},
		"true is preserved": {
			value:           ptr(true),
			expectedJSON:    `{"alertMethods":[{"name":"my-name"}],"treatZeroAsNoData":true}`,
			expectedPresent: true,
		},
		"false is preserved": {
			value:           ptr(false),
			expectedJSON:    `{"alertMethods":[{"name":"my-name"}],"treatZeroAsNoData":false}`,
			expectedPresent: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			noData := AnomalyConfigNoData{
				AlertMethods: []AnomalyConfigAlertMethod{
					{Name: "my-name"},
				},
				TreatZeroAsNoData: test.value,
			}

			data, err := json.Marshal(noData)
			require.NoError(t, err)
			assert.JSONEq(t, test.expectedJSON, string(data))

			var decoded map[string]any
			require.NoError(t, json.Unmarshal(data, &decoded))
			_, found := decoded["treatZeroAsNoData"]
			assert.Equal(t, test.expectedPresent, found)
		})
	}
}
