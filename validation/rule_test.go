package validation

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestSinglRule(t *testing.T) {
	r := NewSingleRule[int](func(v int) error {
		if v < 0 {
			return errors.Errorf("must be positive")
		}
		return nil
	})

	err := r.Validate(0)
	assert.Empty(t, err)
	err = r.Validate(-1)
	assert.EqualError(t, err, "must be positive")
}

func TestSinglRule_WithErrorCode(t *testing.T) {
	r := NewSingleRule[int](func(v int) error {
		if v < 0 {
			return errors.Errorf("must be positive")
		}
		return nil
	}).WithErrorCode(ErrorCode("test"))

	err := r.Validate(0)
	assert.Empty(t, err)
	err = r.Validate(-1)
	assert.EqualError(t, err, "must be positive")
	assert.Equal(t, "test", err.(*RuleError).Code)
}

func TestSinglRule_WithDetals(t *testing.T) {
	for _, test := range []struct {
		Error         string
		Details       string
		ExpectedError string
	}{
		{
			Error:         "this is error",
			Details:       "details",
			ExpectedError: "this is error; details",
		},
		{
			Error:         "this is error",
			Details:       "",
			ExpectedError: "this is error",
		},
		{
			Error:         "",
			Details:       "details",
			ExpectedError: "details",
		},
	} {
		r := NewSingleRule[int](func(v int) error {
			if v < 0 {
				return errors.Errorf(test.Error)
			}
			return nil
		}).
			WithErrorCode(ErrorCode("test")).
			WithDetails(test.Details)

		err := r.Validate(0)
		assert.Empty(t, err)
		err = r.Validate(-1)
		assert.EqualError(t, err, test.ExpectedError)
		assert.Equal(t, "test", err.(*RuleError).Code)
	}
}
