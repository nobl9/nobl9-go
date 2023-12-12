package alertpolicy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSeverity_Validation(t *testing.T) {
	for _, value := range getSeverityLevels() {
		t.Run("passes", func(t *testing.T) {
			rule := severityValidation()
			err := rule.Validate(value.String())
			assert.NoError(t, err)
		})
	}

	t.Run("not valid", func(t *testing.T) {
		rule := severityValidation()
		err := rule.Validate("not valid enum")
		assert.Error(t, err)
	})
}
