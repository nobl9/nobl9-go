//go:build e2e_test

package tests

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/stretchr/testify/require"

	v1alphaExamples "github.com/nobl9/nobl9-go/internal/manifest/v1alpha/examples"
	"github.com/nobl9/nobl9-go/manifest"
)

type exampleWrapper struct {
	v1alphaExamples.Example
	rawObject []byte
}

var examplesRegistry = func() map[manifest.Kind][]exampleWrapper {
	kindToExamples := map[manifest.Kind][]v1alphaExamples.Example{
		manifest.KindProject:          v1alphaExamples.Project(),
		manifest.KindService:          v1alphaExamples.Service(),
		manifest.KindAlertMethod:      v1alphaExamples.AlertMethod(),
		manifest.KindSLO:              v1alphaExamples.SLO(),
		manifest.KindAgent:            v1alphaExamples.Agent(),
		manifest.KindDirect:           v1alphaExamples.Direct(),
		manifest.KindAlertPolicy:      v1alphaExamples.AlertPolicy(),
		manifest.KindAlertSilence:     v1alphaExamples.AlertSilence(),
		manifest.KindAnnotation:       v1alphaExamples.Annotation(),
		manifest.KindBudgetAdjustment: v1alphaExamples.BudgetAdjustment(),
		manifest.KindDataExport:       v1alphaExamples.DataExport(),
		manifest.KindRoleBinding:      v1alphaExamples.RoleBinding(),
	}
	wrapped := make(map[manifest.Kind][]exampleWrapper, len(kindToExamples))
	for kind, examples := range kindToExamples {
		wrapped[kind] = make([]exampleWrapper, 0, len(examples))
		for _, example := range examples {
			object := example.GetObject()
			rawObject, err := json.Marshal(object)
			if err != nil {
				log.Panicf("failed to marshal example %T object: %v", object, err)
			}
			wrapped[kind] = append(wrapped[kind], exampleWrapper{
				Example:   example,
				rawObject: rawObject,
			})
		}
	}
	return wrapped
}()

type examplesFilter = func(example v1alphaExamples.Example) bool

func getExample[T any](t *testing.T, kind manifest.Kind, filter examplesFilter) *T {
	t.Helper()
	examples, ok := examplesRegistry[kind]
	if !ok {
		require.True(t, ok, "%s kind not found in registry", kind)
	}
	decode := func(rawObject []byte) *T {
		var object T
		if err := json.Unmarshal(rawObject, &object); err != nil {
			log.Panicf("failed to unmarshal example %T object: %v", object, err)
		}
		return &object
	}
	if filter == nil {
		return decode(examples[0].rawObject)
	}
	for _, example := range examples {
		if filter(example.Example) {
			return decode(example.rawObject)
		}
	}
	t.Fatalf("example not found for kind %s", kind)
	return nil
}
