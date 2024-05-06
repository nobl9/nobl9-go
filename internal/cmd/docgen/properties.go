package main

import (
	"regexp"
	"strings"
)

func postProcessProperties(docs []*ObjectDoc, formatters ...propertyPostProcessor) {
	for _, doc := range docs {
		for i := range doc.Properties {
			for _, formatter := range formatters {
				doc.Properties[i] = formatter(doc.Properties[i])
			}
		}
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
	doc.FieldDoc = strings.TrimSpace(doc.FieldDoc)
	return doc
}

// extractDeprecatedInformation extracts deprecated information from the docs
// and sets PropertyDoc.IsDeprecated accordingly.
func extractDeprecatedInformation(doc PropertyDoc) PropertyDoc {
	doc.IsDeprecated = deprecatedRegex.MatchString(doc.Doc) || deprecatedRegex.MatchString(doc.FieldDoc)
	return doc
}
