//go:build e2e_test

package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/sdk"
)

const (
	objectDescription          = "Object generated by e2e SDK tests"
	objectPersistedDescription = objectDescription + ". This object is persisted across all tests, do not delete it."
)

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
	commonAnnotations  = v1alpha.MetadataAnnotations{"origin": "sdk-e2e-test"}
	applyAndDeleteLock = newApplyAndDeleteLocker()
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

func v1Apply[T manifest.Object](t *testing.T, objects []T) {
	t.Helper()
	v1ApplyOrDeleteBatch(t, generifyObjects(objects), false, len(objects)+1)
}

func v1Delete[T manifest.Object](t *testing.T, objects []T) {
	t.Helper()
	v1ApplyOrDeleteBatch(t, generifyObjects(objects), true, len(objects)+1)
}

func v1ApplyBatch[T manifest.Object](t *testing.T, objects []T, batchSize int) {
	t.Helper()
	v1ApplyOrDeleteBatch(t, generifyObjects(objects), false, batchSize)
}

func v1DeleteBatch[T manifest.Object](t *testing.T, objects []T, batchSize int) {
	t.Helper()
	v1ApplyOrDeleteBatch(t, generifyObjects(objects), true, batchSize)
}

// v1ApplyOrDeleteBatch applies or deletes objects in batches.
// The batch size is determined by the batchSize parameter.
// The operations on each batch are executed concurrently.
func v1ApplyOrDeleteBatch(
	t *testing.T,
	objects []manifest.Object,
	delete bool,
	batchSize int,
) {
	t.Helper()
	ctx := context.Background()
	group, ctx := errgroup.WithContext(ctx)
	group.SetLimit(runtime.NumCPU())
	for i, j := 0, 0; i < len(objects); i += batchSize {
		j += batchSize
		if j > len(objects) {
			j = len(objects)
		}
		batch := objects[i:j]
		group.Go(func() error {
			applyAndDeleteLock.Lock()
			defer applyAndDeleteLock.Unlock()
			if delete {
				return client.Objects().V1().Delete(ctx, batch)
			}
			return client.Objects().V1().Apply(ctx, batch)
		})
	}
	err := group.Wait()
	var urlErr *url.Error
	if errors.As(err, &urlErr) && urlErr.Timeout() {
		// Unlock the lock to allow other tests to proceed,
		// including the retry, which otherwise would cause a deadlock.
		applyAndDeleteLock.Unlock()

		waitFor := 30 * time.Second
		t.Logf("timeout encountered, the apply/delete operation will be retried in %s; test: %s; error: %v",
			waitFor, t.Name(), err)
		time.Sleep(waitFor)
		v1ApplyOrDeleteBatch(t, objects, delete, batchSize)
	} else {
		require.NoError(t, err)
	}
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

type noopLocker struct{}

func (n noopLocker) Lock()   {}
func (n noopLocker) Unlock() {}

func newApplyAndDeleteLocker() sync.Locker {
	sequential, _ := strconv.ParseBool(os.Getenv(sdk.EnvPrefix + "TEST_RUN_SEQUENTIAL_APPLY_AND_DELETE"))
	if sequential {
		fmt.Println("Running apply and delete operations sequentially")
		return new(sync.Mutex)
	}
	return noopLocker{}
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

func generifyObjects[T manifest.Object](objects []T) []manifest.Object {
	result := make([]manifest.Object, len(objects))
	for i := range objects {
		result[i] = objects[i]
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

func ptr[T any](v T) *T { return &v }
