package slo

import (
	"strings"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
	"github.com/pkg/errors"
)

// ElasticsearchMetric represents metric from Elasticsearch.
type ElasticsearchMetric struct {
	Index *string `json:"index"`
	Query *string `json:"query"`
}

var elasticsearchValidation = govy.New[ElasticsearchMetric](
	govy.ForPointer(func(e ElasticsearchMetric) *string { return e.Index }).
		WithName("index").
		Required().
		Rules(rules.StringNotEmpty()),
	govy.ForPointer(func(e ElasticsearchMetric) *string { return e.Query }).
		WithName("query").
		Required().
		Cascade(govy.CascadeModeStop).
		Rules(rules.StringNotEmpty()).
		Rules(elasticsearchQueryValidationRule()),
)

func elasticsearchQueryValidationRule() govy.Rule[string] {
	return govy.NewRule(func(s string) error {
		containsBeginEndTime := strings.Contains(s,
			"{{.BeginTime}}") && strings.Contains(s, "{{.EndTime}}")
		containsBeginEndTimeMs := strings.Contains(s,
			"{{.BeginTimeInMilliseconds}}") && strings.Contains(s, "{{.EndTimeInMilliseconds}}")
		if containsBeginEndTime && containsBeginEndTimeMs {
			return errors.New(
				`query must contain either {{.BeginTime}} and {{.EndTime}} or
				{{.BeginTimeInMilliseconds}} and {{.EndTimeInMilliseconds}}, but not both`,
			)
		}
		if !containsBeginEndTime && !containsBeginEndTimeMs {
			return errors.New(
				`query must contain either {{.BeginTime}} and {{.EndTime}} or
				{{.BeginTimeInMilliseconds}} and {{.EndTimeInMilliseconds}}`,
			)
		}
		return nil
	}).
		WithErrorCode(rules.ErrorCodeStringContains)
}
