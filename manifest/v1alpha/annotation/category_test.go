package annotation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCategory_BurnAnomaly(t *testing.T) {
	category, err := ParseCategory("BurnAnomaly")
	require.NoError(t, err)
	assert.Equal(t, CategoryBurnAnomaly, category)
	assert.True(t, CategoryBurnAnomaly.IsValid())
	assert.Contains(t, GetSystemCategories(), CategoryBurnAnomaly)
	assert.NotContains(t, GetUserCategories(), CategoryBurnAnomaly)
}
