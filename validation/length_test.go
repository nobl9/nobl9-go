package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringLength(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		err := StringLength(0, 4).Validate("test")
		assert.NoError(t, err)
	})
	t.Run("fails", func(t *testing.T) {
		for min, max := range map[int]int{
			0:  2,
			10: 20,
		} {
			err := StringLength(min, max).Validate("test")
			assert.Error(t, err)
			assert.True(t, HasErrorCode(err, ErrorCodeStringLength))
		}
	})
}

func TestSliceLength(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		err := SliceLength[[]string](0, 1).Validate([]string{"test"})
		assert.NoError(t, err)
	})
	t.Run("fails", func(t *testing.T) {
		for min, max := range map[int]int{
			0: 1,
			3: 10,
		} {
			err := SliceLength[[]string](min, max).Validate([]string{"test", "test"})
			assert.Error(t, err)
			assert.True(t, HasErrorCode(err, ErrorCodeSliceLength))
		}
	})
}

func TestMapLength(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		err := MapLength[map[string]string](0, 1).Validate(map[string]string{"this": "that"})
		assert.NoError(t, err)
	})
	t.Run("fails", func(t *testing.T) {
		for min, max := range map[int]int{
			0: 1,
			3: 10,
		} {
			err := MapLength[map[string]string](min, max).Validate(map[string]string{"a": "b", "c": "d"})
			assert.Error(t, err)
			assert.True(t, HasErrorCode(err, ErrorCodeMapLength))
		}
	})
}
