package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"

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
	IsDeprecated  bool                  `json:"isDeprecated,omitempty"`
	IsOptional    bool                  `json:"isOptional,omitempty"`
	IsSecret      bool                  `json:"isSecret,omitempty"`
	Examples      []string              `json:"examples,omitempty"`
	Rules         []validation.RulePlan `json:"rules,omitempty"`
	ChildrenPaths []string              `json:"childrenPaths,omitempty"`

	// originalType holds the original type alias info while the Type field holds resolved type name.
	originalType typeInfo
}

// TODO:
// - Merge Doc and FieldDoc into a single, well formatted doc (maybe?).
// - Consider stopping at RuleSet level if a description was provided (instead of using SingleRule descriptions).
//
// Docs improvements:
// - Fill out documentation gaps.
// - Provide more examples.
func main() {
	outputFilePathFlag := flag.String("o", "docs.yaml", "Output plan file path")
	objectsFlag := flag.String("objects", strings.Join(
		getAllObjectNames(),
		",",
	), "Comma separated list of objects to generate docs for in the form of <VERSION>/<KIND>, e.g. v1alpha/Service")
	flag.Parse()

	objectNames := strings.Split(*objectsFlag, ",")
	run(*outputFilePathFlag, objectNames)
}

func run(outputFilePath string, objectNames []string) {
	docs := generateObjectDocs(objectNames)
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
			if property.originalType.Package == "" {
				continue
			}
			key := filepath.Join(property.originalType.Package, property.originalType.Name)
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

func getAllObjectNames() []string {
	allObjects := make([]string, 0, len(objectsRegistry))
	for _, object := range objectsRegistry {
		allObjects = append(allObjects, object.Version.String()+"/"+object.Kind.String())
	}
	return allObjects
}
