package alertpolicy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMeasurement_Validation(t *testing.T) {
	for _, value := range getMeasurements() {
		t.Run("passes", func(t *testing.T) {
			rule := MeasurementValidation()
			err := rule.Validate(value.String())
			assert.NoError(t, err)
		})
	}

	t.Run("not valid", func(t *testing.T) {
		rule := MeasurementValidation()
		err := rule.Validate("not valid enum")
		assert.Error(t, err)
	})
}
