// Code is generated by "{{ .ProgramInvocation }}"; DO NOT EDIT.

package {{ .Package }}

import (
    "github.com/nobl9/nobl9-go/manifest"
    "github.com/nobl9/nobl9-go/manifest/v1alpha"
)

{{- range .Structs }}
// Ensure interfaces are implemented.
var _ manifest.Object = {{ .Name }}{}
{{- if .GenerateProjectScopedObject }}
var _ manifest.ProjectScopedObject = {{ .Name }}{}
{{- end }}
{{- if .GenerateV1alphaObjectContext }}
var _ v1alpha.ObjectContext = {{ .Name }}{}
{{- end }}

{{- if .GenerateObject }}

func ({{ .Receiver }} {{ .Name }}) GetVersion() manifest.Version {
  return {{ .Receiver }}.APIVersion
}

func ({{ .Receiver }} {{ .Name }}) GetKind() manifest.Kind {
  return {{ .Receiver }}.Kind
}

func ({{ .Receiver }} {{ .Name }}) GetName() string {
  return {{ .Receiver }}.Metadata.Name
}

func ({{ .Receiver }} {{ .Name }}) Validate() error {
  	if err := validate({{ .Receiver }}); err != nil {
  		return err
  	}
  	return nil
}

func ({{ .Receiver }} {{ .Name }}) GetManifestSource() string {
  return {{ .Receiver }}.ManifestSource
}

func ({{ .Receiver }} {{ .Name }}) SetManifestSource(src string) manifest.Object {
{{ .Receiver }}.ManifestSource = src
  return {{ .Receiver }}
}
{{- end }}

{{- if .GenerateProjectScopedObject }}

func ({{ .Receiver }} {{ .Name }}) GetProject() string {
    return {{ .Receiver }}.Metadata.Project
}

func ({{ .Receiver }} {{ .Name }}) SetProject(project string) manifest.Object {
  {{ .Receiver }}.Metadata.Project = project
  return {{ .Receiver }}
}
{{- end }}

{{- if .GenerateV1alphaObjectContext }}

func ({{ .Receiver }} {{ .Name }}) GetOrganization() string {
  return {{ .Receiver }}.Organization
}

func ({{ .Receiver }} {{ .Name }}) SetOrganization(org string) manifest.Object {
  {{ .Receiver }}.Organization = org
  return {{ .Receiver }}
}
{{- end }}
{{- end }}
