package v1alphaExamples

import (
	"slices"
	"testing"

	internal "github.com/nobl9/nobl9-go/internal/manifest/v1alpha/slo"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSLOVariants(t *testing.T) {
	dataSourceTypes := v1alpha.DataSourceTypeValues()
	examples := make(map[v1alpha.DataSourceType]struct {
		Threshold             bool
		CountMetricsGood      bool
		CountMetricsBad       bool
		CountMetricsGoodTotal bool
	})
	for _, variant := range SLO() {
		v, ok := variant.(interface {
			GetDataSourceType() v1alpha.DataSourceType
		})
    require.True(t, ok)
		example := examples[v.GetDataSourceType()]
		switch v.MetricVariant {
		case metricVariantThreshold:
			example.Threshold = true
		case metricVariantGoodRatio:
			example.CountMetricsGood = true
		case metricVariantBadRatio:
			example.CountMetricsBad = true
		case metricVariantSingleQueryGoodRatio:
			example.CountMetricsGoodTotal = true
		}
		examples[v.DataSourceType] = example
	}

	for _, dataSourceType := range dataSourceTypes {
		example, ok := examples[dataSourceType]
		require.True(t, ok, "missing examples for %s", dataSourceType)
		if slices.Contains(internal.BadOverTotalEnabledSources, dataSourceType) {
			assert.True(t, example.CountMetricsBad, "bad over total is enabled for %s, missing examples", dataSourceType)
		} else {
			assert.False(t, example.CountMetricsBad, "bad over total is disabled for %s, correct the examples", dataSourceType)
		}
		if slices.Contains(internal.SingleQueryGoodOverTotalEnabledSources, dataSourceType) {
			assert.True(t, example.CountMetricsGoodTotal, "single query goodTotal is enabled for %s, missing examples", dataSourceType)
		} else {
			assert.False(t, example.CountMetricsGoodTotal, "single query goodTotal is disabled for %s, correct the examples", dataSourceType)
		}
	}
}
