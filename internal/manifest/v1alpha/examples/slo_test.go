package v1alphaExamples

import (
	"slices"
	"testing"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestSLO_SupportsAllDataSourceTypes(t *testing.T) {
	variants := SLO()
	for _, dataSourceType := range v1alpha.DataSourceTypeValues() {
		if !slices.ContainsFunc(variants, func(e Example) bool {
			return e.(sloVariant).DataSourceType == dataSourceType
		}) {
			t.Errorf("%T '%s' is not listed in the examples", dataSourceType, dataSourceType)
		}
	}
}
