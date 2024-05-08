package main

import (
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest"
)

func generateObjectDocs(objectNames []string) []*ObjectDoc {
	objects := make([]*ObjectDoc, 0, len(objectNames))
	for _, objectName := range objectNames {
		found := false
		for _, object := range objectsRegistry {
			if object.Version.String()+"/"+object.Kind.String() == objectName {
				objects = append(objects, object)
				found = true
				break
			}
		}
		if !found {
			log.Panicf("object %s was not found", objectName)
		}
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
					IsOptional:    vp.IsOptional,
					IsSecret:      vp.IsSecret,
					originalType:  property.originalType,
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
	childrenPaths := make([]string, 0, len(properties))
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
	// #nosec G304
	data, err := os.ReadFile(objectPath)
	if err != nil {
		log.Panicf("failed to read examples for object, path: %s, err: %v", objectPath, err)
	}
	return []string{string(data)}
}
