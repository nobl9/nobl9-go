package main

import (
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"

	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest"
)

type ObjectDoc struct {
	Kind       manifest.Kind    `yaml:"kind"`
	Version    manifest.Version `yaml:"version"`
	Properties []PropertyDoc    `yaml:"properties"`
	// Compute-only fields.
	object               manifest.Object
	validationProperties []PropertyDoc
}

type PropertyDoc struct {
	Path          string                `json:"path"`
	Type          string                `json:"type"`
	Package       string                `json:"package,omitempty"`
	Doc           string                `yaml:"doc,omitempty"`
	FieldDoc      string                `yaml:"fieldDoc,omitempty"`
	Examples      []string              `json:"examples,omitempty"`
	Rules         []validation.RulePlan `json:"rules,omitempty"`
	ChildrenPaths []string              `json:"childrenPaths,omitempty"`
}

func main() {
	docs := generateObjectDocs()
	goDocs := parseGoDocs()

	mergeDocs(docs, goDocs)

	out, err := os.OpenFile("validation_plan.yaml", os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		panic(err)
	}
	enc := yaml.NewEncoder(out,
		yaml.Indent(2),
		yaml.UseLiteralStyleIfMultiline(true))
	if err = enc.Encode(docs); err != nil {
		panic(err)
	}
}

func mergeDocs(docs []*ObjectDoc, goDocs map[string]goTypeDoc) {
	for _, objectDoc := range docs {
		for i, property := range objectDoc.Properties {
			// Builtin type.
			if property.Package == "" {
				continue
			}
			key := filepath.Join(property.Package, property.Type)
			goDoc, found := goDocs[key]
			if !found {
				continue
			}
			objectDoc.Properties[i].Doc = goDoc.Doc
			for name, field := range goDoc.StructFields {
				fieldPath := property.Path + "." + name
				for j, p := range objectDoc.Properties {
					if fieldPath == p.Path {
						objectDoc.Properties[j].FieldDoc = field.Doc
						break
					}
				}
			}
		}
	}
}
