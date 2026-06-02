package report

import (
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
				Rules(timeZoneValidationRule),
			govy.ForPointer(func(t ReliabilityRollupTimeFrame) *RollingTimeFrame { return t.Rolling }).
				WithName("rolling").
				Include(rollingTimeFrameValidation),
			govy.ForPointer(func(t ReliabilityRollupTimeFrame) *CalendarTimeFrame { return t.Calendar }).
				WithName("calendar").
				Include(calendarTimeFrameValidation),
		)),
	govy.ForSlice(func(c ReliabilityRollupConfig) []HierarchyFolder { return c.CustomHierarchy }).
		WithName("customHierarchy").
		Rules(govy.NewRule(func(folders []HierarchyFolder) error {
			if hierarchyDepthExceeds(folders, MaxHierarchyDepth) {
				return errors.Errorf(
					"customHierarchy depth exceeds maximum allowed (%d)",
					MaxHierarchyDepth,
				)
			}
			return nil
		})).
		Cascade(govy.CascadeModeStop).
		IncludeForEach(hierarchyFolderValidation),
)

// hierarchyDepthExceeds reports whether any branch of folders is deeper than
// maxDepth. The walk is iterative with an explicit stack so a pathological
// input cannot exhaust the goroutine stack.
func hierarchyDepthExceeds(folders []HierarchyFolder, maxDepth int) bool {
	type frame struct {
		folder HierarchyFolder
		depth  int
	}
	stack := make([]frame, 0, len(folders))
	for _, f := range folders {
		stack = append(stack, frame{folder: f, depth: 1})
	}
	for len(stack) > 0 {
		n := len(stack) - 1
		cur := stack[n]
		stack = stack[:n]
		if cur.depth > maxDepth {
			return true
		}
		for _, c := range cur.folder.Children {
			stack = append(stack, frame{folder: c, depth: cur.depth + 1})
		}
	}
	return false
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
