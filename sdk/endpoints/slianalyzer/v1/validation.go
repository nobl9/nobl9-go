package v1

import (
	"fmt"
	"time"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/twindow"
)

var createAnalysisValidation = govy.New(
	govy.For(func(r CreateAnalysisRequest) AnalysisMetadata { return r.Metadata }).
		WithName("metadata").Include(analysisMetadataValidation),
	govy.For(func(r CreateAnalysisRequest) AnalysisMetricSpec { return r.MetricSpec }).
		WithName("metricSpec").Include(analysisMetricValidation),
	govy.For(func(r CreateAnalysisRequest) AnalysisPeriod { return r.Period }).
		WithName("period").Include(analysisPeriodValidation),
)

var analysisMetadataValidation = govy.New(
	govy.For(func(m AnalysisMetadata) string { return m.Name }).
		WithName("name").Rules(rules.StringMaxLength(validationV1Alpha.NameMaximumLength)),
	govy.For(func(m AnalysisMetadata) string { return m.DisplayName }).
		WithName("displayName").Required().Rules(rules.StringMaxLength(validationV1Alpha.NameMaximumLength)),
	govy.For(func(m AnalysisMetadata) string { return m.Project }).
		WithName("project").Required().Rules(validationV1Alpha.StringName()),
)

var updateAnalysisValidation = govy.New(
	govy.For(func(r UpdateAnalysisRequest) string { return r.Project }).
		WithName("project").Required().Rules(validationV1Alpha.StringName()),
	govy.For(func(r UpdateAnalysisRequest) string { return r.DisplayName }).
		WithName("displayName").Required().Rules(rules.StringMaxLength(validationV1Alpha.NameMaximumLength)),
)

var analysisMetricValidation = govy.New(
	govy.For(func(m AnalysisMetricSpec) manifest.Kind { return m.Kind }).
		WithName("kind").Rules(rules.OneOf(manifest.KindAgent, manifest.KindDirect)),
	govy.For(func(m AnalysisMetricSpec) string { return m.MetricSource }).
		WithName("metricSource").Required().Rules(validationV1Alpha.StringName()),
	govy.For(govy.GetSelf[AnalysisMetricSpec]()).Rules(
		rules.MutuallyExclusive(true, map[string]func(AnalysisMetricSpec) any{
			"rawMetric":    func(m AnalysisMetricSpec) any { return m.RawMetric },
			"countMetrics": func(m AnalysisMetricSpec) any { return m.CountMetrics },
		}),
	),
	govy.ForPointer(func(m AnalysisMetricSpec) *v1alphaSLO.MetricSpec { return m.RawMetric }).
		WithName("rawMetric").Include(v1alphaSLO.RawMetricQueryValidation),
	govy.ForPointer(func(m AnalysisMetricSpec) *v1alphaSLO.CountMetricsSpec { return m.CountMetrics }).
		WithName("countMetrics").Include(v1alphaSLO.CountMetricsSpecValidation),
)

var analysisPeriodValidation = govy.New(
	govy.For(func(p AnalysisPeriod) string { return p.StartTime }).
		WithName("startTime").Required().Rules(analysisTimestampValidation),
	govy.For(func(p AnalysisPeriod) string { return p.EndTime }).
		WithName("endTime").Required().Rules(analysisTimestampValidation),
	govy.For(func(p AnalysisPeriod) string { return p.TimeZone }).
		WithName("timeZone").Required().
		When(func(p AnalysisPeriod) bool { return p.TimeZone != "Local" }).
		Rules(rules.StringTimeZone()),
	govy.For(govy.GetSelf[AnalysisPeriod]()).Rules(govy.NewRule(func(p AnalysisPeriod) error {
		start, startErr := time.Parse(twindow.IsoDateTimeOnlyLayout, p.StartTime)
		end, endErr := time.Parse(twindow.IsoDateTimeOnlyLayout, p.EndTime)
		if startErr != nil || endErr != nil {
			// The timestamp rules report format errors on their respective fields.
			return nil
		}
		if !end.After(start) {
			return fmt.Errorf("endTime: %s must be after startTime: %s", p.EndTime, p.StartTime)
		}
		if end.Sub(start) < 5*time.Minute {
			return fmt.Errorf("minimum graph time window is 5 minutes")
		}
		return nil
	}).WithErrorCode("sli_analysis_period").WithDescription("endTime must be at least five minutes after startTime")),
)

var analysisTimestampValidation = govy.NewRuleSet(
	rules.StringDateTime(twindow.IsoDateTimeOnlyLayout),
	govy.NewRule(func(value string) error {
		parsed, err := time.Parse(twindow.IsoDateTimeOnlyLayout, value)
		if err != nil {
			return err
		}
		if parsed.Nanosecond() != 0 {
			return fmt.Errorf("timestamp must use whole seconds")
		}
		return nil
	}).WithErrorCode("sli_analysis_timestamp_precision").WithDescription("timestamp must use whole seconds"),
).Cascade(govy.CascadeModeStop)

var createCalculationValidation = govy.New(
	govy.For(func(r CreateCalculationRequest) float64 { return r.BudgetTarget }).
		WithName("target").Required().Rules(rules.GT(0.0), rules.LT(1.0)),
	govy.For(func(r CreateCalculationRequest) string { return r.BudgetingMethod }).
		WithName("budgetingMethod").Required().Rules(rules.OneOf(
		v1alphaSLO.BudgetingMethodOccurrences.String(), v1alphaSLO.BudgetingMethodTimeslices.String(),
	)),
	govy.For(func(r CreateCalculationRequest) string { return r.Operator }).
		WithName("op").Required().Rules(rules.OneOf(v1alpha.OperatorNames()...)),
)
