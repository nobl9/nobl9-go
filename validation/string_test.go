package validation

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringLength(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		err := StringLength(0, 20).Validate("test")
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

func TestStringIsDNSSubdomain(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		for _, input := range []string{
			"test",
			"s",
			"test-this",
			"test-1-this",
			"test1-this",
			"123",
			strings.Repeat("l", 63),
		} {
			err := StringIsDNSSubdomain().Validate(input)
			assert.NoError(t, err)
		}
	})
	t.Run("fails", func(t *testing.T) {
		for _, input := range []string{
			"tesT",
			"",
			strings.Repeat("l", 64),
			"test?",
			"test this",
			"1_2",
			"LOL",
		} {
			err := StringIsDNSSubdomain().Validate(input)
			assert.Error(t, err)
			for _, e := range err.(ruleSetError) {
				assert.True(t, HasErrorCode(e, ErrorCodeStringIsDNSSubdomain))
			}
		}
	})
}

func TestStringDescription(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		err := StringDescription().Validate(strings.Repeat("l", 1050))
		assert.NoError(t, err)
	})
	t.Run("fails", func(t *testing.T) {
		err := StringDescription().Validate(strings.Repeat("l", 1051))
		assert.Error(t, err)
		assert.True(t, HasErrorCode(err, ErrorCodeStringDescription))
	})
}
