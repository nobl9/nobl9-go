package main

import (
	"regexp"
	"slices"
	"strings"
)

// filterProperties is a list of property paths that should be filtered out from the documentation.
var filterProperties = []string{
	"$.organization",
}

func postProcessProperties(docs []*ObjectDoc, formatters ...propertyPostProcessor) {
	for _, doc := range docs {
		properties := make([]PropertyDoc, 0, len(doc.Properties))
		for _, property := range doc.Properties {
			if slices.Contains(filterProperties, property.Path) {
				continue
			}
			for _, formatter := range formatters {
				property = formatter(property)
			}
			properties = append(properties, property)
		}
		doc.Properties = properties
	}
}

// propertyPostProcessor is a function type that post-processes PropertyDoc.
// It can be used to apply additional formatting to the property documentation or add more details to the doc.
type propertyPostProcessor func(doc PropertyDoc) PropertyDoc

var (
	enumDeclarationRegex = regexp.MustCompile(`(?s)ENUM(.*)`)
	deprecatedRegex      = regexp.MustCompile(`^Deprecated:\s`)
)

// removeEnumDeclaration removes ENUM (used with go-enum generator) declarations from the code docs.
func removeEnumDeclaration(doc PropertyDoc) PropertyDoc {
	doc.Doc = enumDeclarationRegex.ReplaceAllString(doc.Doc, "")
	return doc
}

// removeTrailingWhitespace removes trailing whitespace from the docs.
func removeTrailingWhitespace(doc PropertyDoc) PropertyDoc {
	doc.Doc = strings.TrimSpace(doc.Doc)
	doc.fieldDoc = strings.TrimSpace(doc.fieldDoc)
	return doc
}

// extractDeprecatedInformation extracts deprecated information from the docs
// and sets PropertyDoc.IsDeprecated accordingly.
func extractDeprecatedInformation(doc PropertyDoc) PropertyDoc {
	doc.IsDeprecated = deprecatedRegex.MatchString(doc.Doc) || deprecatedRegex.MatchString(doc.fieldDoc)
	return doc
}

// mergeFieldDocIntoDoc merges the PropertyDoc.fieldDoc into PropertyDoc.Doc.
// It inserts the field documentation at the beginning of the property documentation.
func mergeFieldDocIntoDoc(doc PropertyDoc) PropertyDoc {
	if doc.fieldDoc == "" {
		return doc
	}
	if doc.Doc == "" {
		doc.Doc = doc.fieldDoc
		return doc
	}
	doc.Doc = strings.TrimSpace(doc.fieldDoc) + "\n\n" + doc.Doc
	return doc
}
