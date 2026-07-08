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

func TestCategory_Replay(t *testing.T) {
	category, err := ParseCategory("Replay")
	require.NoError(t, err)
	assert.Equal(t, CategoryReplay, category)
	assert.True(t, CategoryReplay.IsValid())
	assert.Contains(t, GetSystemCategories(), CategoryReplay)
	assert.NotContains(t, GetUserCategories(), CategoryReplay)
}
