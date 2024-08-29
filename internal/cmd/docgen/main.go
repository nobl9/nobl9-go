package main

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/nobl9/go-yaml"

	"github.com/nobl9/govy/pkg/govy"

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
	Path          string          `json:"path"`
	Type          string          `json:"type"`
	Package       string          `json:"package,omitempty"`
	Doc           string          `yaml:"doc,omitempty"`
	IsDeprecated  bool            `json:"isDeprecated,omitempty"`
	IsOptional    bool            `json:"isOptional,omitempty"`
	IsSecret      bool            `json:"isSecret,omitempty"`
	Examples      []string        `json:"examples,omitempty"`
	Rules         []govy.RulePlan `json:"rules,omitempty"`
	ChildrenPaths []string        `json:"childrenPaths,omitempty"`
	// Compute-only fields.
	// fieldDoc holds the documentation which was provided on the struct field level.
	fieldDoc string
	// originalType holds the original type alias info while the Type field holds resolved type name.
	originalType typeInfo
}

func main() {
	outputFilePathFlag := flag.String("f", "docs.yaml", "Output plan file path")
	outputFileFormatFlag := flag.String("o", "yaml", "Output plan file format")
	objectsFlag := flag.String("objects", strings.Join(
		getAllObjectNames(),
		",",
	), "Comma separated list of objects to generate docs for in the form of <VERSION>/<KIND>, e.g. v1alpha/Service")
	flag.Parse()

	objectNames := strings.Split(*objectsFlag, ",")
	run(*outputFilePathFlag, *outputFileFormatFlag, objectNames)
}

func run(outputFilePath, outputFileFormat string, objectNames []string) {
	docs := generateObjectDocs(objectNames)
	goDocs := parseGoDocs()

	mergeDocs(docs, goDocs)

	postProcessProperties(docs,
		mergeFieldDocIntoDoc,
		removeEnumDeclaration,
		extractDeprecatedInformation,
		removeTrailingWhitespace,
	)

	// #nosec G304
	out, err := os.OpenFile(outputFilePath, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		panic(err)
	}
	var enc interface{ Encode(v any) error }
	switch outputFileFormat {
	case "json":
		enc = json.NewEncoder(out)
	case "yaml":
		enc = yaml.NewEncoder(out,
			yaml.Indent(2),
			yaml.UseLiteralStyleIfMultiline(true))
	default:
		panic("unsupported output format: " + outputFileFormat)
	}
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
						objectDoc.Properties[j].fieldDoc = field.Doc
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
