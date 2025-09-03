//go:build e2e_test

package tests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	"github.com/nobl9/nobl9-go/sdk"
	"github.com/nobl9/nobl9-go/tests/e2etestutils"
)

func Test_Objects_V1_Apply_And_Delete(t *testing.T) {
	config, err := sdk.ReadConfig()
	require.NoError(t, err)

	mustCreateClient := func() *sdk.Client {
		cl, err := sdk.NewClient(config)
		require.NoError(t, err)
		return cl
	}

	project := generateV1alphaProject(t)
	service := newV1alphaService(t, v1alphaService.Metadata{
		Name:    e2etestutils.GenerateName(),
		Project: project.GetName(),
	})
	objects := []manifest.Object{project, service}
	t.Cleanup(func() { e2etestutils.V2Delete(t, objects) })

	t.Run("dry-run apply objects", func(t *testing.T) {
		err = mustCreateClient().WithDryRun().
			Objects().V1().Apply(t.Context(), objects)
		require.NoError(t, err)
		requireObjectsNotExists(t, objects...)
	})

	t.Run("apply objects", func(t *testing.T) {
		err = client.Objects().V1().Apply(t.Context(), objects)
		require.NoError(t, err)
		requireObjectsExists(t, objects...)
	})

	t.Run("dry-run delete objects", func(t *testing.T) {
		err = mustCreateClient().WithDryRun().
			Objects().V1().Delete(t.Context(), objects)
		require.NoError(t, err)
		requireObjectsExists(t, objects...)
	})

	t.Run("delete objects", func(t *testing.T) {
		err = client.Objects().V1().Delete(t.Context(), objects)
		require.NoError(t, err)
		requireObjectsNotExists(t, objects...)
	})

	t.Run("re-apply objects", func(t *testing.T) {
		err = client.Objects().V1().Apply(t.Context(), objects)
		require.NoError(t, err)
		requireObjectsExists(t, objects...)
	})

	t.Run("delete service by name", func(t *testing.T) {
		err = client.Objects().V1().DeleteByName(
			t.Context(),
			manifest.KindService,
			project.GetName(),
			service.GetName(),
		)
		require.NoError(t, err)
		requireObjectsNotExists(t, service)
	})

	t.Run("dry-run delete project by name", func(t *testing.T) {
		err = mustCreateClient().WithDryRun().
			Objects().V1().DeleteByName(t.Context(), manifest.KindProject, "", project.GetName())
		require.NoError(t, err)
		requireObjectsExists(t, project)
	})

	t.Run("delete project by name", func(t *testing.T) {
		err = client.Objects().V1().DeleteByName(t.Context(), manifest.KindProject, "", project.GetName())
		require.NoError(t, err)
		requireObjectsNotExists(t, project)
	})
}
