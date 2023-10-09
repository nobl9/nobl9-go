package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNumberEqual(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		err := NumberEqual(1.1).Validate(1.1)
		assert.NoError(t, err)
	})
	t.Run("fails", func(t *testing.T) {
		err := NumberEqual(1.1).Validate(1.3)
		require.Error(t, err)
		assert.EqualError(t, err, "1.3 should be equal to 1.1")
	})
}

func TestNumberGreaterThan(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		err := NumberGreaterThan(1).Validate(2)
		assert.NoError(t, err)
	})
	t.Run("fails", func(t *testing.T) {
		for n, v := range map[int]int{1: 1, 4: 2} {
			err := NumberGreaterThan(n).Validate(v)
			require.Error(t, err)
			assert.EqualError(t, err, fmt.Sprintf("%v should be greater than %v", v, n))
		}
	})
}

func TestNumberGreaterThanOrEqual(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		for n, v := range map[int]int{1: 1, 2: 4} {
			err := NumberGreaterThanOrEqual(n).Validate(v)
			assert.NoError(t, err)
		}
	})
	t.Run("fails", func(t *testing.T) {
		err := NumberGreaterThanOrEqual(4).Validate(2)
		require.Error(t, err)
		assert.EqualError(t, err, "2 should be greater than or equal to 4")
	})
}

func TestNumberLessThan(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		err := NumberLessThan(4).Validate(2)
		assert.NoError(t, err)
	})
	t.Run("fails", func(t *testing.T) {
		for n, v := range map[int]int{1: 1, 2: 4} {
			err := NumberLessThan(n).Validate(v)
			require.Error(t, err)
			assert.EqualError(t, err, fmt.Sprintf("%v should be less than %v", v, n))
		}
	})
}

func TestNumberLessThanOrEqual(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		for n, v := range map[int]int{1: 1, 4: 2} {
			err := NumberLessThanOrEqual(n).Validate(v)
			assert.NoError(t, err)
		}
	})
	t.Run("fails", func(t *testing.T) {
		err := NumberLessThanOrEqual(2).Validate(4)
		require.Error(t, err)
		assert.EqualError(t, err, "4 should be less than or equal to 2")
	})
}
