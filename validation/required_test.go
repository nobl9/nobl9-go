package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringRequired(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		err := Required[any]().Validate("non-empty")
		assert.NoError(t, err)
	})
	t.Run("fails", func(t *testing.T) {
		err := Required[any]().Validate("")
		require.Error(t, err)
		assert.EqualError(t, err, "field is required but was empty")
	})
}
