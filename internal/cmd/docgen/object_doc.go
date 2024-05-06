package main

import (
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest"
	v1alphaAgent "github.com/nobl9/nobl9-go/manifest/v1alpha/agent"
	v1alphaAlert "github.com/nobl9/nobl9-go/manifest/v1alpha/alert"
	v1alphaAlertMethod "github.com/nobl9/nobl9-go/manifest/v1alpha/alertmethod"
	v1alphaAlertPolicy "github.com/nobl9/nobl9-go/manifest/v1alpha/alertpolicy"
	v1alphaAlertSilence "github.com/nobl9/nobl9-go/manifest/v1alpha/alertsilence"
	v1alphaAnnotation "github.com/nobl9/nobl9-go/manifest/v1alpha/annotation"
	v1alphaBudgetAdjustment "github.com/nobl9/nobl9-go/manifest/v1alpha/budgetadjustment"
	v1alphaDataExport "github.com/nobl9/nobl9-go/manifest/v1alpha/dataexport"
	v1alphaDirect "github.com/nobl9/nobl9-go/manifest/v1alpha/direct"
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	v1alphaRoleBinding "github.com/nobl9/nobl9-go/manifest/v1alpha/rolebinding"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	v1alphaUserGroup "github.com/nobl9/nobl9-go/manifest/v1alpha/usergroup"
)

func generateObjectDocs() []*ObjectDoc {
	plansToDocs := func(plans []validation.PropertyPlan) []PropertyDoc {
		var docs []PropertyDoc
		for _, plan := range plans {
			docs = append(docs, PropertyDoc{
				Doc:      "TODO",
				Path:     plan.Path,
				Type:     plan.Type,
				Package:  plan.Package,
				Examples: plan.Examples,
				Rules:    plan.Rules,
			})
		}
		return docs
	}
	objects := []*ObjectDoc{
		{
			Kind:                 manifest.KindProject,
			Version:              manifest.VersionV1alpha,
			validationProperties: plansToDocs(validation.Plan(v1alphaProject.Project{}.GetValidator())),
			object:               v1alphaProject.Project{},
		},
		{
			Kind:                 manifest.KindService,
			Version:              manifest.VersionV1alpha,
			validationProperties: plansToDocs(validation.Plan(v1alphaService.Service{}.GetValidator())),
			object:               v1alphaService.Service{},
		},
		{
			Kind:                 manifest.KindSLO,
			Version:              manifest.VersionV1alpha,
			validationProperties: plansToDocs(validation.Plan(v1alphaSLO.SLO{}.GetValidator())),
			object:               v1alphaSLO.SLO{},
		},
		{
			Kind:                 manifest.KindDirect,
			Version:              manifest.VersionV1alpha,
			validationProperties: plansToDocs(validation.Plan(v1alphaDirect.Direct{}.GetValidator())),
			object:               v1alphaDirect.Direct{},
		},
		{
			Kind:                 manifest.KindAgent,
			Version:              manifest.VersionV1alpha,
			validationProperties: plansToDocs(validation.Plan(v1alphaAgent.Agent{}.GetValidator())),
			object:               v1alphaAgent.Agent{},
		},
		{
			Kind:                 manifest.KindAlertMethod,
			Version:              manifest.VersionV1alpha,
			validationProperties: plansToDocs(validation.Plan(v1alphaAlertMethod.AlertMethod{}.GetValidator())),
			object:               v1alphaAlertMethod.AlertMethod{},
		},
		{
			Kind:                 manifest.KindAlertPolicy,
			Version:              manifest.VersionV1alpha,
			validationProperties: plansToDocs(validation.Plan(v1alphaAlertPolicy.AlertPolicy{}.GetValidator())),
			object:               v1alphaAlertPolicy.AlertPolicy{},
		},
		{
			Kind:                 manifest.KindAlertSilence,
			Version:              manifest.VersionV1alpha,
			validationProperties: plansToDocs(validation.Plan(v1alphaAlertSilence.AlertSilence{}.GetValidator())),
			object:               v1alphaAlertSilence.AlertSilence{},
		},
		{
			Kind:                 manifest.KindAlert,
			Version:              manifest.VersionV1alpha,
			validationProperties: plansToDocs(validation.Plan(v1alphaAlert.Alert{}.GetValidator())),
			object:               v1alphaAlert.Alert{},
		},
		{
			Kind:                 manifest.KindAnnotation,
			Version:              manifest.VersionV1alpha,
			validationProperties: plansToDocs(validation.Plan(v1alphaAnnotation.Annotation{}.GetValidator())),
			object:               v1alphaAnnotation.Annotation{},
		},
		{
			Kind:                 manifest.KindBudgetAdjustment,
			Version:              manifest.VersionV1alpha,
			validationProperties: plansToDocs(validation.Plan(v1alphaBudgetAdjustment.BudgetAdjustment{}.GetValidator())),
			object:               v1alphaBudgetAdjustment.BudgetAdjustment{},
		},
		{
			Kind:                 manifest.KindDataExport,
			Version:              manifest.VersionV1alpha,
			validationProperties: plansToDocs(validation.Plan(v1alphaDataExport.DataExport{}.GetValidator())),
			object:               v1alphaDataExport.DataExport{},
		},
		{
			Kind:                 manifest.KindUserGroup,
			Version:              manifest.VersionV1alpha,
			validationProperties: plansToDocs(validation.Plan(v1alphaUserGroup.UserGroup{}.GetValidator())),
			object:               v1alphaUserGroup.UserGroup{},
		},
		{
			Kind:                 manifest.KindRoleBinding,
			Version:              manifest.VersionV1alpha,
			validationProperties: plansToDocs(validation.Plan(v1alphaRoleBinding.RoleBinding{}.GetValidator())),
			object:               v1alphaRoleBinding.RoleBinding{},
		},
	}
	rootPath := testutils.FindModuleRoot()
	// Generate object properties based on reflection.
	for _, object := range objects {
		mapper := newObjectMapper()
		typ := reflect.TypeOf(object.object)
		mapper.Map(typ, "$")
		object.Properties = mapper.Properties
		object.Examples = readObjectExamples(rootPath, typ)
	}
	// Add children paths to properties.
	// The object mapper does not provide this information, but rather returns a flat list of properties.
	for _, object := range objects {
		for i, property := range object.Properties {
			childrenPaths := findPropertyChildrenPaths(property.Path, object.Properties)
			property.ChildrenPaths = childrenPaths
			object.Properties[i] = property
		}
	}
	// Extend properties with validation plan results.
	for _, object := range objects {
		for _, vp := range object.validationProperties {
			found := false
			for i, property := range object.Properties {
				if vp.Path != property.Path {
					continue
				}
				object.Properties[i] = PropertyDoc{
					Path:          property.Path,
					Type:          property.Type,
					Package:       property.Package,
					Examples:      vp.Examples,
					Rules:         vp.Rules,
					ChildrenPaths: property.ChildrenPaths,
				}
				found = true
				break
			}
			if !found && !isValidationInferredProperty(object.Version, object.Kind, vp.Path) {
				log.Panicf("validation property %s not found in object %s", vp.Path, object.Kind)
			}
		}
	}
	return objects
}

func findPropertyChildrenPaths(parent string, properties []PropertyDoc) []string {
	var childrenPaths []string
	for _, property := range properties {
		childRelativePath, found := strings.CutPrefix(property.Path, parent+".")
		if !found {
			continue
		}
		// Not an immediate child.
		if strings.Contains(childRelativePath, ".") {
			continue
		}
		childrenPaths = append(childrenPaths, parent+"."+childRelativePath)
	}
	return childrenPaths
}

func isValidationInferredProperty(version manifest.Version, kind manifest.Kind, path string) bool {
	for _, p := range validationInferredProperties {
		if p.Version == version && p.Kind == kind && strings.HasPrefix(path, p.Path) {
			return true
		}
	}
	return false
}

// validationInferredProperties lists properties which are only available through the validation plan.
// This can be the case for interface{} types which are inferred on runtime.
var validationInferredProperties = []struct {
	Version manifest.Version
	Kind    manifest.Kind
	Path    string
}{
	{
		Version: manifest.VersionV1alpha,
		Kind:    manifest.KindDataExport,
		Path:    "$.spec.spec",
	},
}

func readObjectExamples(root string, typ reflect.Type) []string {
	relPath := strings.TrimPrefix(typ.PkgPath(), moduleRootPath)
	objectPath := filepath.Join(root, relPath, "example.yaml")
	data, err := os.ReadFile(objectPath)
	if err != nil {
		log.Panicf("failed to read examples for object, path: %s, err: %v", objectPath, err)
	}
	return []string{string(data)}
}
