package e2etestutils

import (
	"encoding/json"
	"log"
	"slices"
	"sort"
	"testing"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaExamples "github.com/nobl9/nobl9-go/manifest/v1alpha/examples"
)

// ExamplesFilter defines a function shape used to filter [v1alphaExamples.Example].
type ExamplesFilter func(example v1alphaExamples.Example) bool

// GetAllExamples returns all examples for the given [manifest.Kind].
func GetAllExamples(t *testing.T, kind manifest.Kind) []v1alphaExamples.Example {
	t.Helper()
	examples := getExamples(t, kind)
	if len(examples) == 0 {
		t.Fatalf("%s kind not found in registry", kind)
	}
	return slices.Clone(examples)
}

// GetExample returns the first [ExampleObject] matching given [ExamplesFilter].
// If no example was found, the test will fail immediately .
func GetExample(t *testing.T, kind manifest.Kind, filter ExamplesFilter) v1alphaExamples.Example {
	t.Helper()
	examples := getExamples(t, kind)
	if len(examples) == 0 {
		t.Fatalf("%s kind not found in registry", kind)
	}
	if filter == nil {
		return examples[0]
	}
	for _, example := range examples {
		if filter(example) {
			return example
		}
	}
	t.Fatalf("example not found for kind %s", kind)
	return nil
}

// GetExampleObject returns a concrete [manifest.Object] implementation as specified by the T type constraint.
// Under the hood [GetExample] is called, refer to its documentation for more details on how the filter is applied.
func GetExampleObject[T manifest.Object](t *testing.T, kind manifest.Kind, filter ExamplesFilter) T {
	t.Helper()
	genericObject := GetExample(t, kind, filter).GetObject()
	data, err := json.Marshal(genericObject)
	if err != nil {
		t.Fatalf("failed to marshal example %T object: %v", genericObject, err)
	}
	var object T
	if err := json.Unmarshal(data, &object); err != nil {
		log.Panicf("failed to unmarshal example %T object: %v", object, err)
	}
	return object
}

// FilterExamplesByDataSourceType is an [ExamplesFilter] which filters examples
// by the provided [v1alpha.DataSourceType].
func FilterExamplesByDataSourceType(dataSourceType v1alpha.DataSourceType) ExamplesFilter {
	return func(example v1alphaExamples.Example) bool {
		dataSourceGetter, ok := example.(v1alphaExamples.DataSourceTypeGetter)
		if !ok {
			return false
		}
		return dataSourceGetter.GetDataSourceType() == dataSourceType
	}
}

// FilterExamplesByAlertMethodType is an [ExamplesFilter] which filters examples
// by the provided [v1alpha.AlertMethodType].
func FilterExamplesByAlertMethodType(typ v1alpha.AlertMethodType) ExamplesFilter {
	return func(example v1alphaExamples.Example) bool {
		alertMethodGetter, ok := example.(v1alphaExamples.AlertMethodTypeGetter)
		if !ok {
			return false
		}
		return alertMethodGetter.GetAlertMethodType() == typ
	}
}

// examplesRegistry MUST NOT be accessed directly, use [getExamples] instead.
var examplesRegistry = newMapCache[manifest.Kind, []v1alphaExamples.Example]()

func getExamples(t *testing.T, kind manifest.Kind) []v1alphaExamples.Example {
	t.Helper()

	if v, ok := examplesRegistry.Load(kind); ok {
		return v
	}

	examplesRegistry.Lock()
	defer examplesRegistry.Unlock()

	// In case multiple goroutines were waiting on the locker,
	// so we don't do the work multiple times.
	if v, ok := examplesRegistry.LoadUnsafe(kind); ok {
		return v
	}

	var examples []v1alphaExamples.Example
	switch kind {
	case manifest.KindProject:
		examples = v1alphaExamples.Project()
	case manifest.KindService:
		examples = v1alphaExamples.Service()
	case manifest.KindAlertMethod:
		examples = v1alphaExamples.AlertMethod()
	case manifest.KindSLO:
		examples = v1alphaExamples.SLO()
	case manifest.KindAgent:
		examples = v1alphaExamples.Agent()
	case manifest.KindDirect:
		examples = v1alphaExamples.Direct()
	case manifest.KindAlertPolicy:
		examples = v1alphaExamples.AlertPolicy()
	case manifest.KindAlertSilence:
		examples = v1alphaExamples.AlertSilence()
	case manifest.KindAnnotation:
		examples = v1alphaExamples.Annotation()
	case manifest.KindBudgetAdjustment:
		examples = v1alphaExamples.BudgetAdjustment()
	case manifest.KindDataExport:
		examples = v1alphaExamples.DataExport()
	case manifest.KindRoleBinding:
		examples = v1alphaExamples.RoleBinding()
	default:
		return nil
	}

	sort.Slice(examples, func(i, j int) bool {
		if examples[i].GetVariant() != examples[j].GetVariant() {
			return examples[i].GetVariant() < examples[j].GetVariant()
		}
		return examples[i].GetSubVariant() < examples[j].GetSubVariant()
	})
	examplesRegistry.StoreUnsafe(kind, examples)
	return examples
}
