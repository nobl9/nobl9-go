package slo

import (
	"strings"
	"text/template"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
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
		Rules(xor()),
)

func xor() govy.Rule[string] {
	return govy.NewRule(func(s string) error {
		containsBeginEndTime := strings.Contains(s,
			"{{.BeginTime}}") && strings.Contains(s, "{{.EndTime}}")
		containsBeginEndTimeMs := strings.Contains(s,
			"{{.BeginTimeInMilliseconds}}") && strings.Contains(s, "{{.EndTimeInMilliseconds}}")
		if containsBeginEndTime && containsBeginEndTimeMs {
			return govy.NewRuleErrorTemplate(govy.TemplateVars{
				PropertyValue: s,
			})
		}
		if !containsBeginEndTime && !containsBeginEndTimeMs {
			return govy.NewRuleErrorTemplate(govy.TemplateVars{
				PropertyValue: s,
			})
		}
		return nil
	}).
		WithErrorCode(rules.ErrorCodeStringContains).
		WithMessageTemplate(template.Must(template.New("msg").
			Parse("Query must contain either BeginTime/EndTime or BeginTimeInMilliseconds/EndTimeInMilliseconds, but not both")))
}
