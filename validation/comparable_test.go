package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEqual(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		err := EqualTo(1.1).Validate(1.1)
		assert.NoError(t, err)
	})
	t.Run("fails", func(t *testing.T) {
		err := EqualTo(1.1).Validate(1.3)
		require.Error(t, err)
		assert.EqualError(t, err, "1.3 should be equal to 1.1")
	})
}

func TestNotEqual(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		err := NotEqualTo(1.1).Validate(1.3)
		assert.NoError(t, err)
	})
	t.Run("fails", func(t *testing.T) {
		err := NotEqualTo(1.1).Validate(1.1)
		require.Error(t, err)
		assert.EqualError(t, err, "1.1 should be not equal to 1.1")
	})
}

func TestGreaterThan(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		err := GreaterThan(1).Validate(2)
		assert.NoError(t, err)
	})
	t.Run("fails", func(t *testing.T) {
		for n, v := range map[int]int{1: 1, 4: 2} {
			err := GreaterThan(n).Validate(v)
			require.Error(t, err)
			assert.EqualError(t, err, fmt.Sprintf("%v should be greater than %v", v, n))
		}
	})
}

func TestGreaterThanOrEqual(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		for n, v := range map[int]int{1: 1, 2: 4} {
			err := GreaterThanOrEqualTo(n).Validate(v)
			assert.NoError(t, err)
		}
	})
	t.Run("fails", func(t *testing.T) {
		err := GreaterThanOrEqualTo(4).Validate(2)
		require.Error(t, err)
		assert.EqualError(t, err, "2 should be greater than or equal to 4")
	})
}

func TestLessThan(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		err := LessThan(4).Validate(2)
		assert.NoError(t, err)
	})
	t.Run("fails", func(t *testing.T) {
		for n, v := range map[int]int{1: 1, 2: 4} {
			err := LessThan(n).Validate(v)
			require.Error(t, err)
			assert.EqualError(t, err, fmt.Sprintf("%v should be less than %v", v, n))
		}
	})
}

func TestLessThanOrEqual(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		for n, v := range map[int]int{1: 1, 4: 2} {
			err := LessThanOrEqualTo(n).Validate(v)
			assert.NoError(t, err)
		}
	})
	t.Run("fails", func(t *testing.T) {
		err := LessThanOrEqualTo(2).Validate(4)
		require.Error(t, err)
		assert.EqualError(t, err, "4 should be less than or equal to 2")
	})
}
