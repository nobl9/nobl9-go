package v1alphaExamples

import (
	"slices"
	"testing"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestAlertMethod_SupportsAllAlertMethodTypes(t *testing.T) {
	variants := AlertMethod()
	for _, methodType := range v1alpha.AlertMethodTypeValues() {
		if !slices.ContainsFunc(variants, func(e Example) bool {
			return e.(alertMethodExample).methodType == methodType
		}) {
			t.Errorf("%T '%s' is not listed in the examples", methodType, methodType)
		}
	}
}
