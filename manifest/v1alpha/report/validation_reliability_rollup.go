package report

import (
	"time"

	"github.com/pkg/errors"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
)

const MaxHierarchyDepth = 8

var reliabilityRollupValidation = govy.New[ReliabilityRollupConfig](
	govy.For(func(c ReliabilityRollupConfig) ReliabilityRollupTimeFrame { return c.TimeFrame }).
		WithName("timeFrame").
		Required().
		Rules(rules.MutuallyExclusive(true, map[string]func(t ReliabilityRollupTimeFrame) any{
			"rolling":  func(t ReliabilityRollupTimeFrame) any { return t.Rolling },
			"calendar": func(t ReliabilityRollupTimeFrame) any { return t.Calendar },
		})).
		Include(govy.New[ReliabilityRollupTimeFrame](
			govy.For(func(t ReliabilityRollupTimeFrame) string { return t.TimeZone }).
				WithName("timeZone").
				Required().
				Rules(govy.NewRule(func(v string) error {
					if _, err := time.LoadLocation(v); err != nil {
						return errors.Wrap(err, "not a valid time zone")
					}
					return nil
				})),
			govy.ForPointer(func(t ReliabilityRollupTimeFrame) *RollingTimeFrame { return t.Rolling }).
				WithName("rolling").
				Include(rollingTimeFrameValidation),
			govy.ForPointer(func(t ReliabilityRollupTimeFrame) *CalendarTimeFrame { return t.Calendar }).
				WithName("calendar").
				Include(calendarTimeFrameValidation),
		)),
	govy.For(func(c ReliabilityRollupConfig) []HierarchyFolder { return c.CustomHierarchy }).
		WithName("customHierarchy").
		Rules(govy.NewRule(func(folders []HierarchyFolder) error {
			if depth := maxHierarchyDepth(folders); depth > MaxHierarchyDepth {
				return errors.Errorf(
					"customHierarchy depth %d exceeds maximum allowed (%d)",
					depth, MaxHierarchyDepth,
				)
			}
			return nil
		})),
	govy.ForSlice(func(c ReliabilityRollupConfig) []HierarchyFolder { return c.CustomHierarchy }).
		WithName("customHierarchy").
		IncludeForEach(hierarchyFolderValidation),
)

func maxHierarchyDepth(folders []HierarchyFolder) int {
	deepest := 0
	for _, f := range folders {
		d := 1 + maxHierarchyDepth(f.Children)
		if d > deepest {
			deepest = d
		}
	}
	return deepest
}

// hierarchyFolderValidation validates a single folder in the reliability rollup
// custom hierarchy. The builder pattern lets the recursive children rule capture
// the validator before the package-level variable is assigned.
var hierarchyFolderValidation = buildHierarchyFolderValidation()

func buildHierarchyFolderValidation() govy.Validator[HierarchyFolder] {
	var folderValidator govy.Validator[HierarchyFolder]
	folderValidator = govy.New[HierarchyFolder](
		govy.For(func(f HierarchyFolder) string { return f.DisplayName }).
			WithName("displayName").
			Required().
			Rules(rules.StringMaxLength(validationV1Alpha.NameMaximumLength)),
		govy.For(govy.GetSelf[HierarchyFolder]()).
			Rules(govy.NewRule(func(f HierarchyFolder) error {
				if len(f.Children) == 0 && len(f.SLOs) == 0 {
					return errors.New("folder must contain at least one child folder or slo")
				}
				return nil
			})),
		govy.ForSlice(func(f HierarchyFolder) []HierarchyFolder { return f.Children }).
			WithName("children").
			RulesForEach(govy.NewRule(func(child HierarchyFolder) error { //nolint:gocritic
				return folderValidator.Validate(child)
			})),
		govy.ForSlice(func(f HierarchyFolder) []HierarchySLORef { return f.SLOs }).
			WithName("slos").
			IncludeForEach(hierarchySLORefValidation),
	)
	return folderValidator
}

var hierarchySLORefValidation = govy.New[HierarchySLORef](
	govy.For(func(s HierarchySLORef) string { return s.Project }).
		WithName("project").
		Include(requiredNameValidation),
	govy.For(func(s HierarchySLORef) string { return s.Name }).
		WithName("name").
		Include(requiredNameValidation),
	govy.For(func(s HierarchySLORef) string { return s.DisplayName }).
		WithName("displayName").
		OmitEmpty().
		Rules(rules.StringMaxLength(validationV1Alpha.NameMaximumLength)),
)
