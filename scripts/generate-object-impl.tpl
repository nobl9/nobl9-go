// Code generated by "{{ .ProgramInvocation }}"; DO NOT EDIT.

package {{ .Package }}

import "github.com/nobl9/nobl9-go/manifest"

// Ensure interfaces are implemented.
var _ Object = {{ .StructName }}{}
{{- if .GenerateProjectScopedObject }}
var _ ProjectScopedObject = {{ .StructName }}{}
{{- end }}

{{- if .GenerateObject }}

func ({{ .Receiver }} {{ .StructName }}) GetVersion() string {
  return {{ .Receiver }}.APIVersion
}

func ({{ .Receiver }} {{ .StructName }}) GetKind() manifest.Kind {
  return {{ .Receiver }}.Kind
}

func ({{ .Receiver }} {{ .StructName }}) GetName() string {
  return {{ .Receiver }}.Metadata.Name
}

func ({{ .Receiver }} {{ .StructName }}) Validate() error {
  return validator.Check({{ .Receiver }})
}
{{- end }}

{{- if .GenerateProjectScopedObject }}

func ({{ .Receiver }} {{ .StructName }}) GetProject() string {
    return {{ .Receiver }}.Metadata.Project
}

func ({{ .Receiver }} {{ .StructName }}) SetProject(project string) manifest.Object {
  {{ .Receiver }}.Metadata.Project = project
  return {{ .Receiver }}
}
{{- end }}

{{- if .GenerateV1alphaObjectContext }}

func ({{ .Receiver }} {{ .StructName }}) GetOrganization() string {
  return {{ .Receiver }}.Organization
}

func ({{ .Receiver }} {{ .StructName }}) SetOrganization(org string) manifest.Object {
  {{ .Receiver }}.Organization = org
  return {{ .Receiver }}
}

func ({{ .Receiver }} {{ .StructName }}) GetManifestSource() string {
  return {{ .Receiver }}.ManifestSource
}

func ({{ .Receiver }} {{ .StructName }}) SetManifestSource(src string) manifest.Object {
  {{ .Receiver }}.ManifestSource = src
  return {{ .Receiver }}
}
{{- end }}
