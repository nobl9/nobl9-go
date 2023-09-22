package v1alpha

import (
	"github.com/nobl9/nobl9-go/manifest"
)

// GenericObject represents a generic map[string]interface{} representation of manifest.Object.
// It's useful for scenarios where an implementation does not want to be tied to specific
// v1alpha versions.
type GenericObject map[string]interface{}

func (g GenericObject) GetVersion() string {
	version, _ := g["apiVersion"].(string)
	return version
}

func (g GenericObject) GetKind() manifest.Kind {
	name, _ := g["kind"].(string)
	kind, _ := manifest.ParseKind(name)
	return kind
}

func (g GenericObject) GetName() string {
	meta, _ := g["metadata"].(map[string]interface{})
	name, _ := meta["name"].(string)
	return name
}

func (g GenericObject) Validate() error {
	return nil
}

func (g GenericObject) GetProject() string {
	meta, _ := g["metadata"].(map[string]interface{})
	name, _ := meta["project"].(string)
	return name
}

func (g GenericObject) SetProject(project string) manifest.Object {
	meta, _ := g["metadata"].(map[string]interface{})
	meta["project"] = project
	g["metadata"] = meta
	return g
}

func (g GenericObject) GetOrganization() string {
	org, _ := g["organization"].(string)
	return org
}

func (g GenericObject) SetOrganization(org string) manifest.Object {
	g["organization"] = org
	return g
}

func (g GenericObject) GetManifestSource() string {
	src, _ := g["manifestSrc"].(string)
	return src
}

func (g GenericObject) SetManifestSource(src string) manifest.Object {
	g["manifestSrc"] = src
	return g
}
