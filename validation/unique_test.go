package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSliceUnique(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		err := SliceUnique(SelfHashFunc[string]()).Validate([]string{"a", "b", "c"})
		assert.NoError(t, err)
	})
	t.Run("fails", func(t *testing.T) {
		err := SliceUnique(SelfHashFunc[string]()).Validate([]string{"a", "b", "c", "b"})
		require.Error(t, err)
		assert.EqualError(t, err, "elements are not unique, index 1 collides with index 3")
		assert.True(t, HasErrorCode(err, ErrorCodeSliceUnique))
	})
}
