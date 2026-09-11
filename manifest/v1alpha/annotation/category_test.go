package annotation

import (
	"slices"
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
	assert.Contains(t, GetUserCategories(), CategoryReplay)
	assert.NotContains(t, GetSystemCategories(), CategoryReplay)
}

// Consumers tell system and user categories apart by checking membership of one
// list only, so a category missing from both lists is silently misclassified.
func TestCategory_EveryCategoryIsEitherSystemOrUser(t *testing.T) {
	listedTimes := make(map[Category]int)
	for _, category := range slices.Concat(GetSystemCategories(), GetUserCategories()) {
		listedTimes[category]++
	}
	for _, category := range CategoryValues() {
		assert.Equalf(t, 1, listedTimes[category],
			"category %q must be listed exactly once across systemCategories and userCategories, got %d",
			category, listedTimes[category])
		delete(listedTimes, category)
	}
	for category := range listedTimes {
		assert.Failf(t, "unknown category listed",
			"category %q is listed in systemCategories or userCategories but is not a valid Category", category)
	}
}
