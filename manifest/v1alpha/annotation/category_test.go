package annotation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCategory_GoodOverTotalAnomaly(t *testing.T) {
	category, err := ParseCategory("GoodOverTotalAnomaly")
	require.NoError(t, err)
	assert.Equal(t, CategoryGoodOverTotalAnomaly, category)
	assert.True(t, CategoryGoodOverTotalAnomaly.IsValid())
	assert.Contains(t, GetSystemCategories(), CategoryGoodOverTotalAnomaly)
	assert.NotContains(t, GetUserCategories(), CategoryGoodOverTotalAnomaly)
}

func TestCategory_SLIDensityAnomalies(t *testing.T) {
	for _, expected := range []Category{
		CategoryMissingPairAnomaly,
		CategoryDensityMismatchAnomaly,
		CategoryDensityChangeAnomaly,
	} {
		actual, err := ParseCategory(expected.String())
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
		assert.True(t, expected.IsValid())
		assert.Contains(t, GetSystemCategories(), expected)
		assert.NotContains(t, GetUserCategories(), expected)
	}
}
