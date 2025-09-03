//go:build e2e_test

package tests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	"github.com/nobl9/nobl9-go/sdk"
	v2 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v2"
	"github.com/nobl9/nobl9-go/tests/e2etestutils"
)

func Test_Objects_V2_Apply_And_Delete(t *testing.T) {
	if client, err := sdk.DefaultClient(); err != nil {
		t.Errorf("failed to create %T: %v", client, err)
		t.FailNow()
	}
	// We're making sure that per-request settings for dry-run are overriding the client settings.
	client.WithDryRun()

	project := generateV1alphaProject(t)
	service := newV1alphaService(t, v1alphaService.Metadata{
		Name:    e2etestutils.GenerateName(),
		Project: project.GetName(),
	})
	objects := []manifest.Object{project, service}
	t.Cleanup(func() { e2etestutils.V2Delete(t, objects) })

	t.Run("dry-run apply objects", func(t *testing.T) {
		err := client.Objects().V2().Apply(t.Context(), v2.ApplyRequest{Objects: objects}.WithDryRun(true))
		require.NoError(t, err)
		requireObjectsNotExists(t, objects...)
	})

	t.Run("apply objects", func(t *testing.T) {
		err := client.Objects().V2().Apply(t.Context(), v2.ApplyRequest{Objects: objects}.WithDryRun(false))
		require.NoError(t, err)
		requireObjectsExists(t, objects...)
	})

	t.Run("dry-run delete objects", func(t *testing.T) {
		err := client.Objects().V2().Delete(t.Context(), v2.DeleteRequest{Objects: objects}.WithDryRun(true))
		require.NoError(t, err)
		requireObjectsExists(t, objects...)
	})

	t.Run("delete objects", func(t *testing.T) {
		err := client.Objects().V2().Delete(t.Context(), v2.DeleteRequest{Objects: objects}.WithDryRun(false))
		require.NoError(t, err)
		requireObjectsNotExists(t, objects...)
	})

	t.Run("re-apply objects", func(t *testing.T) {
		err := client.Objects().V2().Apply(t.Context(), v2.ApplyRequest{Objects: objects}.WithDryRun(false))
		require.NoError(t, err)
		requireObjectsExists(t, objects...)
	})

	t.Run("delete service by name", func(t *testing.T) {
		err := client.Objects().V2().DeleteByName(t.Context(), v2.DeleteByNameRequest{
			Kind:    manifest.KindService,
			Names:   []string{service.GetName()},
			Project: project.GetName(),
		}.WithDryRun(false))
		require.NoError(t, err)
		requireObjectsNotExists(t, service)
	})

	t.Run("dry-run delete project by name", func(t *testing.T) {
		err := client.Objects().V2().DeleteByName(t.Context(), v2.DeleteByNameRequest{
			Kind:  manifest.KindProject,
			Names: []string{project.GetName()},
		}.WithDryRun(true))
		require.NoError(t, err)
		requireObjectsExists(t, project)
	})

	t.Run("delete project by name", func(t *testing.T) {
		err := client.Objects().V2().DeleteByName(t.Context(), v2.DeleteByNameRequest{
			Kind:  manifest.KindProject,
			Names: []string{project.GetName()},
		}.WithDryRun(false))
		require.NoError(t, err)
		requireObjectsNotExists(t, project)
	})
}
