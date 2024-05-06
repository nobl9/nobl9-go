package main

import (
	"flag"
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
	Examples   []string         `yaml:"examples,omitempty"`
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
	IsDeprecated  bool                  `json:"isDeprecated,omitempty"`
}

// TODO:
// - Merge Doc and FieldDoc into a single, well formatted doc (maybe?).
// - Figure out how to handle maps (keys vs values vs items validation).
//
// Docs improvements:
// - Fill out documentation gaps.
// - Provide more examples.
func main() {
	outputFilePath := flag.String("o", "validation_plan.yaml", "Output plan file path")
	flag.Parse()

	run(*outputFilePath)
}

func run(outputFilePath string) {
	docs := generateObjectDocs()
	goDocs := parseGoDocs()

	mergeDocs(docs, goDocs)

	postProcessProperties(docs,
		removeEnumDeclaration,
		extractDeprecatedInformation,
		removeTrailingWhitespace,
	)

	out, err := os.OpenFile(outputFilePath, os.O_CREATE|os.O_WRONLY, 0o600)
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
