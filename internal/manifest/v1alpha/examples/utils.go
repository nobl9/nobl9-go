package v1alphaExamples

import (
	"strings"
	"time"
	"unicode"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func toKebabCase(input string) string {
	b := strings.Builder{}
	for i, r := range input {
		if unicode.IsUpper(r) {
			if i != 0 {
				b.WriteRune('-')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func dataSourceTypePrettyName(typ v1alpha.DataSourceType) string {
	switch typ {
	case v1alpha.AppDynamics, v1alpha.ThousandEyes, v1alpha.BigQuery,
		v1alpha.OpenTSDB, v1alpha.CloudWatch, v1alpha.InfluxDB, v1alpha.LogicMonitor:
		return typ.String()
	}
	return splitCamelCase(typ.String())
}

func splitCamelCase(input string) string {
	b := strings.Builder{}
	for i, r := range input {
		if i != 0 && unicode.IsUpper(r) {
			b.WriteRune(' ')
		}
		b.WriteRune(r)
	}
	return b.String()
}
