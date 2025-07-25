package v1

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

type GetBudgetAdjustmentRequest struct {
	Names            []string
	SLOName, Project string
}

func (p GetBudgetAdjustmentRequest) Validate() error {
	return validator.Validate(p)
}

var validator = govy.New(
	govy.For(func(p GetBudgetAdjustmentRequest) GetBudgetAdjustmentRequest { return p }).
		Rules(
			govy.NewRule(func(v GetBudgetAdjustmentRequest) error {
				// Check if SLOName is set when Project is set
				if v.Project != "" && v.SLOName == "" {
					return govy.NewPropertyError(
						QueryKeySLOName,
						v.SLOName,
						govy.NewRuleError("SLO is required when Project is set", "required"),
					)
				}

				// Check if Project is set when SLOName is set
				if v.SLOName != "" && v.Project == "" {
					return govy.NewPropertyError(
						QueryKeySLOProjectName,
						v.Project,
						govy.NewRuleError("Project is required when SLO is set", "required"),
					)
				}
				return nil
			}),
		),
	govy.For(func(p GetBudgetAdjustmentRequest) string { return p.Project }).
		WithName(QueryKeySLOProjectName).
		OmitEmpty().
		Rules(rules.StringDNSLabel()),
	govy.For(func(p GetBudgetAdjustmentRequest) string { return p.SLOName }).
		WithName(QueryKeySLOName).
		OmitEmpty().
		Rules(rules.StringDNSLabel()),
	govy.ForSlice(func(p GetBudgetAdjustmentRequest) []string { return p.Names }).
		WithName(QueryKeyName).
		RulesForEach(rules.StringDNSLabel()),
)
