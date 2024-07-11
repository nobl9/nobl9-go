//go:build e2e_test

package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

const objectDescription = "Object generated by e2e SDK tests"

var (
	testStartTime             = time.Now()
	objectsCounter            = atomic.Int64{}
	uniqueTestIdentifierLabel = struct {
		Key   string
		Value string
	}{
		Key:   "sdk-e2e-test-id",
		Value: strconv.Itoa(int(testStartTime.UnixNano())),
	}
	commonAnnotations = v1alpha.MetadataAnnotations{"origin": "sdk-e2e-test"}
)

var (
	timeRFC3339Regexp = regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`)
	userIDRegexp      = regexp.MustCompile(`[a-zA-Z0-9]{20}`)
)

type objectsEqualityAssertFunc[T manifest.Object] func(t *testing.T, expected, actual T)

func assertSubset[T manifest.Object](t *testing.T, actual, expected []T, f objectsEqualityAssertFunc[T]) {
	t.Helper()
	for i := range expected {
		found := false
		for j := range actual {
			if actual[j].GetName() == expected[i].GetName() {
				f(t, expected[i], actual[j])
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %T %s not found in the actual list", expected[i], expected[i].GetName())
		}
	}
}

func v1Apply[T manifest.Object](t *testing.T, ctx context.Context, inputs []T) {
	t.Helper()
	objects := make([]manifest.Object, 0, len(inputs))
	for _, input := range inputs {
		objects = append(objects, input)
	}
	err := client.Objects().V1().Apply(ctx, objects)
	require.NoError(t, err)
}

func v1Delete[T manifest.Object](t *testing.T, ctx context.Context, inputs []T) {
	t.Helper()
	objects := make([]manifest.Object, 0, len(inputs))
	for _, input := range inputs {
		objects = append(objects, input)
	}
	err := client.Objects().V1().Delete(ctx, objects)
	require.NoError(t, err)
}

// generateName generates a unique name for the test object.
func generateName() string {
	return fmt.Sprintf("sdk-e2e-%d-%d", objectsCounter.Add(1), testStartTime.UnixNano())
}

// annotateLabels adds origin label to the provided labels,
// so it's easier to locate the leftovers from these tests.
// It also adds unique test identifier label to the provided labels so that we can reliably retrieve objects created withing a given test without .
func annotateLabels(t *testing.T, labels v1alpha.Labels) v1alpha.Labels {
	t.Helper()
	if labels == nil {
		labels = make(v1alpha.Labels, 3)
	}
	labels["origin"] = []string{"sdk-e2e-test"}
	labels[uniqueTestIdentifierLabel.Key] = []string{uniqueTestIdentifierLabel.Value}
	labels["sdk-test-name"] = []string{t.Name()}
	return labels
}

// deepCopyObject creates a deep copy of the provided object using JSON encoding and decoding.
func deepCopyObject[T any](t *testing.T, object T) T {
	t.Helper()
	data, err := json.Marshal(object)
	require.NoError(t, err)
	var copied T
	require.NoError(t, json.Unmarshal(data, &copied))
	return copied
}

func filterSlice[T any](s []T, filter func(T) bool) []T {
	result := make([]T, 0, len(s))
	for i := range s {
		if filter(s[i]) {
			result = append(result, s[i])
		}
	}
	return result
}

func ptr[T any](v T) *T { return &v }
