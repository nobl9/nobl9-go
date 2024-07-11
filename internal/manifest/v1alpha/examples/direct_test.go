package v1alphaExamples

import (
	"slices"
	"testing"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaDirect "github.com/nobl9/nobl9-go/manifest/v1alpha/direct"
)

func TestDirect_SupportsAllDirectTypes(t *testing.T) {
	variants := Direct()
	for _, typ := range v1alpha.DataSourceTypeValues() {
		if !v1alphaDirect.IsValidDirectType(typ) {
			continue
		}
		if !slices.ContainsFunc(variants, func(e Example) bool {
			return e.(directExample).typ == typ
		}) {
			t.Errorf("%T '%s' is not listed in the examples", typ, typ)
		}
	}
}
