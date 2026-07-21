//go:build e2e_test

package tests

import (
	"errors"
	"net/http"
	"sync"
	"testing"

	"github.com/hashicorp/go-retryablehttp"
	client "github.com/influxdata/influxdb1-client"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	"github.com/nobl9/nobl9-go/sdk"
	v2 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v2"
	"github.com/nobl9/nobl9-go/tests/e2etestutils"
)

func Test_Objects_V2_Apply_And_Delete(t *testing.T) {
	dryRunClient, err := sdk.DefaultClient()
	if err != nil {
		t.Errorf("failed to create %T: %v", dryRunClient, err)
		t.FailNow()
	}
	// We're making sure that the client settings have no effect over v2 API.
	dryRunClient.WithDryRun()

	project := generateV1alphaProject(t)
	service := newV1alphaService(t, v1alphaService.Metadata{
		Name:    e2etestutils.GenerateName(),
		Project: project.GetName(),
	})
	objects := []manifest.Object{project, service}
	t.Cleanup(func() { e2etestutils.V1Delete(t, objects) })

	t.Run("dry-run apply objects", func(t *testing.T) {
		err = dryRunClient.Objects().V2().Apply(t.Context(), v2.ApplyRequest{Objects: objects, DryRun: true})
		require.NoError(t, err)
		requireObjectsNotExists(t, objects...)
	})

	t.Run("apply objects", func(t *testing.T) {
		err = dryRunClient.Objects().V2().Apply(t.Context(), v2.ApplyRequest{Objects: objects})
		require.NoError(t, err)
		requireObjectsExists(t, objects...)
	})

	t.Run("dry-run delete objects", func(t *testing.T) {
		err = dryRunClient.Objects().V2().Delete(t.Context(), v2.DeleteRequest{Objects: objects, DryRun: true})
		require.NoError(t, err)
		requireObjectsExists(t, objects...)
	})

	t.Run("delete objects", func(t *testing.T) {
		err = dryRunClient.Objects().V2().Delete(t.Context(), v2.DeleteRequest{Objects: objects})
		require.NoError(t, err)
		requireObjectsNotExists(t, objects...)
	})

	t.Run("re-apply objects", func(t *testing.T) {
		err = dryRunClient.Objects().V2().Apply(t.Context(), v2.ApplyRequest{Objects: objects})
		require.NoError(t, err)
		requireObjectsExists(t, objects...)
	})

	t.Run("delete service by name", func(t *testing.T) {
		err = dryRunClient.Objects().V2().DeleteByName(t.Context(), v2.DeleteByNameRequest{
			Kind:    manifest.KindService,
			Names:   []string{service.GetName()},
			Project: project.GetName(),
		})
		require.NoError(t, err)
		requireObjectsNotExists(t, service)
	})

	t.Run("dry-run delete project by name", func(t *testing.T) {
		err = dryRunClient.Objects().V2().DeleteByName(t.Context(), v2.DeleteByNameRequest{
			Kind:   manifest.KindProject,
			Names:  []string{project.GetName()},
			DryRun: true,
		})
		require.NoError(t, err)
		requireObjectsExists(t, project)
	})

	t.Run("delete project by name", func(t *testing.T) {
		err = dryRunClient.Objects().V2().DeleteByName(t.Context(), v2.DeleteByNameRequest{
			Kind:  manifest.KindProject,
			Names: []string{project.GetName()},
		})
		require.NoError(t, err)
		requireObjectsNotExists(t, project)
	})
}

// These tests are a non-perfect safeguard which ensures the API returns 409
// in case concurrent apply operations conflict with each other.
// It is impossible to reliably test this behavior on this level,
// in case there's no conflict, the test will still succeed, without proving anything.
// This is acceptable, run enough times, it will eventually fail and catch a regression.
func Test_Objects_V2_Apply_ConcurrentServiceRequests_ReturnSuccessOrConflict(t *testing.T) {
	const (
		rounds  = 3
		workers = 16
	)

	project := generateV1alphaProject(t)
	e2etestutils.V1Apply(t, []manifest.Object{project})
	t.Cleanup(func() { e2etestutils.V1Delete(t, []manifest.Object{project}) })

	noRetryClient := newObjectsV2NoRetryClient(t)
	ctx := t.Context()
	totalSuccesses := 0
	totalConflicts := 0
	for range rounds {
		service := newV1alphaService(t, v1alphaService.Metadata{
			Name:    e2etestutils.GenerateName(),
			Project: project.GetName(),
		})
		objects := []manifest.Object{service}
		t.Cleanup(func() { e2etestutils.V1Delete(t, objects) })

		start := make(chan struct{})
		results := make(chan error, workers)
		var wg sync.WaitGroup
		for range workers {
			wg.Go(func() {
				<-start
				requestObjects := []manifest.Object{service}
				results <- noRetryClient.Objects().V2().Apply(ctx, v2.ApplyRequest{Objects: requestObjects})
			})
		}
		close(start)
		wg.Wait()
		close(results)

		for err := range results {
			if err == nil {
				totalSuccesses++
				continue
			}
			var httpErr *sdk.HTTPError
			if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusConflict {
				totalConflicts++
				continue
			}
			require.NoError(t, err)
		}
		requireObjectsExists(t, service)
	}

	t.Logf("concurrent apply results: %d succeeded, %d returned HTTP %d",
		totalSuccesses, totalConflicts, http.StatusConflict)
}

func newObjectsV2NoRetryClient(t *testing.T) *sdk.Client {
	t.Helper()
	config := *client.Config
	noRetryClient, err := sdk.NewClient(&config)
	require.NoError(t, err)
	retryTransport, ok := noRetryClient.HTTP.Transport.(*retryablehttp.RoundTripper)
	require.Truef(t, ok, "unexpected HTTP transport %T", noRetryClient.HTTP.Transport)
	retryTransport.Client.RetryMax = 0
	return noRetryClient
}
