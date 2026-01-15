package slo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComponentAggregationMethodValues(t *testing.T) {
	methods := ComponentAggregationMethodValues()

	assert.Len(t, methods, 2)
	assert.Contains(t, methods, ComponentAggregationMethodReliability)
	assert.Contains(t, methods, ComponentAggregationMethodErrorBudgetState)
}

func TestComponentAggregationMethodNames(t *testing.T) {
	names := ComponentAggregationMethodNames()

	assert.Len(t, names, 2)
	assert.Contains(t, names, "Reliability")
	assert.Contains(t, names, "ErrorBudgetState")
}

func TestComponentAggregationMethodDefault(t *testing.T) {
	assert.Equal(t, ComponentAggregationMethodReliability, ComponentAggregationMethodDefault)
}

func TestComponentAggregationMethodConstants(t *testing.T) {
	assert.Equal(t, ComponentAggregationMethod("Reliability"), ComponentAggregationMethodReliability)
	assert.Equal(t, ComponentAggregationMethod("ErrorBudgetState"), ComponentAggregationMethodErrorBudgetState)
}

func TestComponentAggregationMethod_String(t *testing.T) {
	assert.Equal(t, "Reliability", ComponentAggregationMethodReliability.String())
	assert.Equal(t, "ErrorBudgetState", ComponentAggregationMethodErrorBudgetState.String())
}

func TestComponentAggregationMethod_IsValid(t *testing.T) {
	assert.True(t, ComponentAggregationMethodReliability.IsValid())
	assert.True(t, ComponentAggregationMethodErrorBudgetState.IsValid())
	assert.False(t, ComponentAggregationMethod("Invalid").IsValid())
	assert.False(t, ComponentAggregationMethod("").IsValid())
}

func TestParseComponentAggregationMethod(t *testing.T) {
	t.Run("valid values", func(t *testing.T) {
		method, err := ParseComponentAggregationMethod("Reliability")
		require.NoError(t, err)
		assert.Equal(t, ComponentAggregationMethodReliability, method)

		method, err = ParseComponentAggregationMethod("ErrorBudgetState")
		require.NoError(t, err)
		assert.Equal(t, ComponentAggregationMethodErrorBudgetState, method)
	})

	t.Run("case insensitive", func(t *testing.T) {
		method, err := ParseComponentAggregationMethod("reliability")
		require.NoError(t, err)
		assert.Equal(t, ComponentAggregationMethodReliability, method)

		method, err = ParseComponentAggregationMethod("errorbudgetstate")
		require.NoError(t, err)
		assert.Equal(t, ComponentAggregationMethodErrorBudgetState, method)
	})

	t.Run("invalid value", func(t *testing.T) {
		_, err := ParseComponentAggregationMethod("Invalid")
		assert.ErrorIs(t, err, ErrInvalidComponentAggregationMethod)
	})

	t.Run("empty value", func(t *testing.T) {
		_, err := ParseComponentAggregationMethod("")
		assert.ErrorIs(t, err, ErrInvalidComponentAggregationMethod)
	})
}
