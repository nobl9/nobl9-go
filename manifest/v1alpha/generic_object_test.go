package v1alpha

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/manifest"
)

func TestGenericObject_GetVersion(t *testing.T) {
	t.Run("returns version", func(t *testing.T) {
		assert.Equal(
			t, manifest.VersionV1alpha, GenericObject{genericFieldVersion: manifest.VersionV1alpha.String()}.GetVersion(),
		)
	})
	t.Run("empty, not panics", func(t *testing.T) {
		assert.NotPanics(t, func() { GenericObject{}.GetVersion() })
	})
	t.Run("invalid type, not panics", func(t *testing.T) {
		assert.NotPanics(t, func() { GenericObject{genericFieldVersion: 1}.GetVersion() })
	})
}

func TestGenericObject_GetKind(t *testing.T) {
	t.Run("returns version", func(t *testing.T) {
		assert.Equal(t, manifest.KindProject, GenericObject{genericFieldKind: manifest.KindProject.String()}.GetKind())
	})
	t.Run("empty, not panics", func(t *testing.T) {
		assert.NotPanics(t, func() { GenericObject{}.GetKind() })
	})
	t.Run("invalid kind, not panics", func(t *testing.T) {
		assert.NotPanics(t, func() { GenericObject{genericFieldKind: "fake"}.GetKind() })
	})
	t.Run("invalid type, not panics", func(t *testing.T) {
		assert.NotPanics(t, func() { GenericObject{genericFieldKind: 1}.GetKind() })
	})
}

func TestGenericObject_GetName(t *testing.T) {
	t.Run("returns version", func(t *testing.T) {
		assert.Equal(t, "default",
			GenericObject{genericFieldMetadata: map[string]any{genericFieldName: "default"}}.GetName())
	})
	t.Run("no metadata, not panics", func(t *testing.T) {
		assert.NotPanics(t, func() { GenericObject{}.GetName() })
	})
	t.Run("invalid metadata type, not panics", func(t *testing.T) {
		assert.NotPanics(t, func() { GenericObject{genericFieldMetadata: 123}.GetName() })
	})
	t.Run("empty, not panics", func(t *testing.T) {
		assert.NotPanics(t, func() { GenericObject{genericFieldMetadata: map[string]any{}}.GetName() })
	})
	t.Run("invalid type, not panics", func(t *testing.T) {
		assert.NotPanics(t, func() {
			GenericObject{genericFieldMetadata: map[string]any{genericFieldName: 1}}.GetName()
		})
	})
}

func TestGenericObject_GetProject(t *testing.T) {
	t.Run("returns version", func(t *testing.T) {
		assert.Equal(t, "default",
			GenericObject{genericFieldMetadata: map[string]any{genericFieldProject: "default"}}.GetProject())
	})
	t.Run("no metadata, not panics", func(t *testing.T) {
		assert.NotPanics(t, func() { GenericObject{}.GetProject() })
	})
	t.Run("invalid metadata type, not panics", func(t *testing.T) {
		assert.NotPanics(t, func() { GenericObject{genericFieldMetadata: 123}.GetProject() })
	})
	t.Run("empty, not panics", func(t *testing.T) {
		assert.NotPanics(t, func() { GenericObject{genericFieldMetadata: map[string]any{}}.GetProject() })
	})
	t.Run("invalid type, not panics", func(t *testing.T) {
		assert.NotPanics(t, func() {
			GenericObject{genericFieldMetadata: map[string]any{genericFieldProject: 1}}.GetProject()
		})
	})
}

func TestGenericObject_SetProject(t *testing.T) {
	t.Run("sets project for project scoped kinds", func(t *testing.T) {
		for _, kind := range manifest.ProjectScopedKinds() {
			t.Run(kind.String(), func(t *testing.T) {
				obj := GenericObject{
					genericFieldKind:     kind.String(),
					genericFieldMetadata: map[string]any{},
				}.SetProject("foo")
				assert.Equal(t, "foo", obj.(GenericObject).GetProject())
			})
		}
	})
	t.Run("does not set project for organization scoped kinds", func(t *testing.T) {
		for _, kind := range append([]manifest.Kind{0}, manifest.KindValues()...) {
			if kind.ProjectScoped() {
				continue
			}
			t.Run(kind.String(), func(t *testing.T) {
				obj := GenericObject{
					genericFieldKind:     kind.String(),
					genericFieldMetadata: map[string]any{},
				}.SetProject("foo")
				assert.Empty(t, obj.(GenericObject).GetProject())
			})
		}
	})
	t.Run("invalid metadata type, not panics", func(t *testing.T) {
		assert.NotPanics(t, func() {
			GenericObject{
				genericFieldKind:     manifest.KindSLO.String(),
				genericFieldMetadata: 123,
			}.SetProject("default")
		})
	})
}

func TestGenericObject_GetOrganization(t *testing.T) {
	t.Run("returns version", func(t *testing.T) {
		assert.Equal(t, "my-org", GenericObject{genericFieldOrganization: "my-org"}.GetOrganization())
	})
	t.Run("empty, not panics", func(t *testing.T) {
		assert.NotPanics(t, func() { GenericObject{}.GetOrganization() })
	})
	t.Run("invalid type, not panics", func(t *testing.T) {
		assert.NotPanics(t, func() { GenericObject{genericFieldOrganization: 1}.GetOrganization() })
	})
}

func TestGenericObject_GetManifestSrc(t *testing.T) {
	assert.Equal(t, "/home/me/slo.yaml",
		GenericObject{genericFieldManifestSource: "/home/me/slo.yaml"}.GetManifestSource())
	t.Run("empty, not panics", func(t *testing.T) {
		assert.NotPanics(t, func() { GenericObject{}.GetManifestSource() })
	})
	t.Run("invalid type, not panics", func(t *testing.T) {
		assert.NotPanics(t, func() { GenericObject{genericFieldManifestSource: 1}.GetManifestSource() })
	})
}
