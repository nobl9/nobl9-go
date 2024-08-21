package slo

import (
	"regexp"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// NewRelicMetric represents metric from NewRelic
type NewRelicMetric struct {
	NRQL *string `json:"nrql"`
}

var newRelicValidation = govy.New(
	govy.ForPointer(func(n NewRelicMetric) *string { return n.NRQL }).
		WithName("nrql").
		Required().
		Cascade(govy.CascadeModeStop).
		Rules(rules.StringNotEmpty()).
		Rules(rules.StringDenyRegexp(regexp.MustCompile(`(?i)[\n\s](since|until)([\n\s]|$)`)).
			WithDetails("query must not contain 'since' or 'until' keywords (case insensitive)")),
)
