package v1alpha

import (
	"github.com/nobl9/nobl9-go/manifest"
)

const (
	genericFieldKind           = "apiVersion"
	genericFieldVersion        = "kind"
	genericFieldMetadata       = "metadata"
	genericFieldName           = "name"
	genericFieldProject        = "project"
	genericFieldOrganization   = "organization"
	genericFieldManifestSource = "manifestSrc"
)

// GenericObject represents a generic map[string]interface{} representation of manifest.Object.
// It's useful for scenarios where an implementation does not want to be tied to specific
// v1alpha versions.
type GenericObject map[string]interface{}

func (g GenericObject) GetVersion() string {
	version, _ := g[genericFieldVersion].(string)
	return version
}

func (g GenericObject) GetKind() manifest.Kind {
	name, _ := g[genericFieldKind].(string)
	kind, _ := manifest.ParseKind(name)
	return kind
}

func (g GenericObject) GetName() string {
	meta, ok := g[genericFieldMetadata].(map[string]interface{})
	if !ok {
		return ""
	}
	name, _ := meta[genericFieldName].(string)
	return name
}

func (g GenericObject) Validate() error {
	return nil
}

func (g GenericObject) GetProject() string {
	meta, ok := g[genericFieldMetadata].(map[string]interface{})
	if !ok {
		return ""
	}
	name, _ := meta[genericFieldProject].(string)
	return name
}

func (g GenericObject) SetProject(project string) manifest.Object {
	switch g.GetKind() {
	case 0, manifest.KindProject, manifest.KindRoleBinding, manifest.KindUserGroup:
		return g
	default:
		meta, ok := g[genericFieldMetadata].(map[string]interface{})
		if !ok {
			return g
		}
		meta[genericFieldProject] = project
		g[genericFieldMetadata] = meta
		return g
	}
}

func (g GenericObject) GetOrganization() string {
	org, _ := g[genericFieldOrganization].(string)
	return org
}

func (g GenericObject) SetOrganization(org string) manifest.Object {
	g[genericFieldOrganization] = org
	return g
}

func (g GenericObject) GetManifestSource() string {
	src, _ := g[genericFieldManifestSource].(string)
	return src
}

func (g GenericObject) SetManifestSource(src string) manifest.Object {
	g[genericFieldManifestSource] = src
	return g
}
