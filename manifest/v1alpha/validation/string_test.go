package validation

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringRequired(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		err := StringRequired().Validate("test")
		assert.NoError(t, err)
	})
	t.Run("fails", func(t *testing.T) {
		err := StringRequired().Validate("")
		assert.Error(t, err)
	})
}

func TestStringLength(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		err := StringLength(0, 20).Validate("test")
		assert.NoError(t, err)
	})
	t.Run("fails, upper bound", func(t *testing.T) {
		err := StringLength(0, 2).Validate("test")
		assert.Error(t, err)
	})
	t.Run("fails, lower bound", func(t *testing.T) {
		err := StringLength(10, 20).Validate("test")
		assert.Error(t, err)
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
	})
}
