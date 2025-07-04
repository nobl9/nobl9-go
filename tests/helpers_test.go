//go:build e2e_test

package tests

import (
	"encoding/json"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

var (
	timeRFC3339Regexp = regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`)
	userIDRegexp      = regexp.MustCompile(`[a-zA-Z0-9]{20}`)
	commonAnnotations = v1alpha.MetadataAnnotations{"origin": "sdk-e2e-test"}
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

func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func tryExecuteRequest[T any](t *testing.T, reqFunc func() (T, error)) (T, error) {
	t.Helper()
	ticker := time.NewTicker(5 * time.Second)
	timer := time.NewTimer(time.Minute)
	defer ticker.Stop()
	defer timer.Stop()
	var (
		response T
		err      error
	)
	for {
		select {
		case <-ticker.C:
			response, err = reqFunc()
			if err == nil {
				return response, nil
			}
		case <-timer.C:
			t.Error("timeout")
			return response, err
		}
	}
}

func ptr[T any](v T) *T { return &v }
