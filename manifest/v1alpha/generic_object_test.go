package v1alpha

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/manifest"
)

func TestGenericObject_GetVersion(t *testing.T) {
	t.Run("returns version", func(t *testing.T) {
		assert.Equal(t, APIVersion, GenericObject{genericFieldVersion: APIVersion}.GetVersion())
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
			GenericObject{genericFieldMetadata: map[string]interface{}{genericFieldName: "default"}}.GetName())
	})
	t.Run("no metadata, not panics", func(t *testing.T) {
		assert.NotPanics(t, func() { GenericObject{}.GetName() })
	})
	t.Run("invalid metadata type, not panics", func(t *testing.T) {
		assert.NotPanics(t, func() { GenericObject{genericFieldMetadata: 123}.GetName() })
	})
	t.Run("empty, not panics", func(t *testing.T) {
		assert.NotPanics(t, func() { GenericObject{genericFieldMetadata: map[string]interface{}{}}.GetName() })
	})
	t.Run("invalid type, not panics", func(t *testing.T) {
		assert.NotPanics(t, func() {
			GenericObject{genericFieldMetadata: map[string]interface{}{genericFieldName: 1}}.GetName()
		})
	})
}

func TestGenericObject_GetProject(t *testing.T) {
	t.Run("returns version", func(t *testing.T) {
		assert.Equal(t, "default",
			GenericObject{genericFieldMetadata: map[string]interface{}{genericFieldProject: "default"}}.GetProject())
	})
	t.Run("no metadata, not panics", func(t *testing.T) {
		assert.NotPanics(t, func() { GenericObject{}.GetProject() })
	})
	t.Run("invalid metadata type, not panics", func(t *testing.T) {
		assert.NotPanics(t, func() { GenericObject{genericFieldMetadata: 123}.GetProject() })
	})
	t.Run("empty, not panics", func(t *testing.T) {
		assert.NotPanics(t, func() { GenericObject{genericFieldMetadata: map[string]interface{}{}}.GetProject() })
	})
	t.Run("invalid type, not panics", func(t *testing.T) {
		assert.NotPanics(t, func() {
			GenericObject{genericFieldMetadata: map[string]interface{}{genericFieldProject: 1}}.GetProject()
		})
	})
}

func TestGenericObject_SetProject(t *testing.T) {
	t.Run("sets project", func(t *testing.T) {
		o := GenericObject{genericFieldMetadata: map[string]interface{}{}}
		for _, kind := range []manifest.Kind{
			manifest.KindSLO,
			manifest.KindProject,
			manifest.KindService,
		} {
			o[genericFieldKind] = kind.String()
			res := o.SetProject("default").(manifest.ProjectScopedObject)
			assert.Equal(t, "default", res.GetProject())
		}
	})
	t.Run("do not set for specific kinds", func(t *testing.T) {
		o := GenericObject{genericFieldMetadata: map[string]interface{}{}}
		for _, kind := range []manifest.Kind{
			0,
			manifest.KindProject,
			manifest.KindRoleBinding,
			manifest.KindUserGroup,
		} {
			o[genericFieldKind] = kind.String()
			res := o.SetProject("default")
			assert.NotNil(t, res)
			assert.Nil(t, o[genericFieldMetadata].(map[string]interface{})[genericFieldProject])
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
