package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOneOf(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		err := OneOf("this", "that").Validate("that")
		assert.NoError(t, err)
	})
	t.Run("fails", func(t *testing.T) {
		err := OneOf("this", "that").Validate("those")
		require.Error(t, err)
		assert.EqualError(t, err, "must be one of [this, that]")
		assert.True(t, HasErrorCode(err, ErrorCodeOneOf))
	})
}
