package v1alphaExamples

import (
	"slices"
	"testing"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestAgent_SupportsAllAgentTypes(t *testing.T) {
	variants := Agent()
	for _, typ := range v1alpha.DataSourceTypeValues() {
		if !slices.ContainsFunc(variants, func(e Example) bool {
			return e.(agentExample).typ == typ
		}) {
			t.Errorf("%T '%s' is not listed in the examples", typ, typ)
		}
	}
}
