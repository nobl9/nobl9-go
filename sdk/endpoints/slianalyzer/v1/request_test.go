package v1

import (
	"strings"
	"testing"

	"github.com/nobl9/govy/pkg/govytest"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/manifest"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
)

func TestCreateAnalysisRequest_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		change func(*CreateAnalysisRequest)
		errors []govytest.ExpectedRuleError
	}{
		{
			name:   "missing display name",
			change: func(r *CreateAnalysisRequest) { r.Metadata.DisplayName = "" },
			errors: []govytest.ExpectedRuleError{{PropertyPath: "metadata.displayName", Code: rules.ErrorCodeRequired}},
		},
		{
			name:   "long display name",
			change: func(r *CreateAnalysisRequest) { r.Metadata.DisplayName = strings.Repeat("x", 254) },
			errors: []govytest.ExpectedRuleError{{PropertyPath: "metadata.displayName", Code: rules.ErrorCodeStringMaxLength}},
		},
		{
			name:   "long name",
			change: func(r *CreateAnalysisRequest) { r.Metadata.Name = strings.Repeat("x", 254) },
			errors: []govytest.ExpectedRuleError{{PropertyPath: "metadata.name", Code: rules.ErrorCodeStringMaxLength}},
		},
		{
			name:   "invalid project",
			change: func(r *CreateAnalysisRequest) { r.Metadata.Project = "Invalid project" },
			errors: []govytest.ExpectedRuleError{{PropertyPath: "metadata.project", Code: "string_name"}},
		},
		{
			name:   "missing project",
			change: func(r *CreateAnalysisRequest) { r.Metadata.Project = "" },
			errors: []govytest.ExpectedRuleError{{PropertyPath: "metadata.project", Code: rules.ErrorCodeRequired}},
		},
		{
			name:   "invalid kind",
			change: func(r *CreateAnalysisRequest) { r.MetricSpec.Kind = manifest.KindSLO },
			errors: []govytest.ExpectedRuleError{{PropertyPath: "metricSpec.kind", Code: rules.ErrorCodeOneOf}},
		},
		{
			name:   "invalid source",
			change: func(r *CreateAnalysisRequest) { r.MetricSpec.MetricSource = "Invalid source" },
			errors: []govytest.ExpectedRuleError{{PropertyPath: "metricSpec.metricSource", Code: "string_name"}},
		},
		{
			name:   "missing source",
			change: func(r *CreateAnalysisRequest) { r.MetricSpec.MetricSource = "" },
			errors: []govytest.ExpectedRuleError{{PropertyPath: "metricSpec.metricSource", Code: rules.ErrorCodeRequired}},
		},
		{
			name:   "missing metric",
			change: func(r *CreateAnalysisRequest) { r.MetricSpec.RawMetric = nil },
			errors: []govytest.ExpectedRuleError{{PropertyPath: "metricSpec", Code: rules.ErrorCodeMutuallyExclusive}},
		},
		{
			name:   "both metric types",
			change: func(r *CreateAnalysisRequest) { r.MetricSpec.CountMetrics = validCountMetrics() },
			errors: []govytest.ExpectedRuleError{{PropertyPath: "metricSpec", Code: rules.ErrorCodeMutuallyExclusive}},
		},
		{
			name:   "missing raw query",
			change: func(r *CreateAnalysisRequest) { r.MetricSpec.RawMetric.Prometheus.PromQL = nil },
			errors: []govytest.ExpectedRuleError{{
				PropertyPath: "metricSpec.rawMetric.prometheus.promql", Code: rules.ErrorCodeRequired,
			}},
		},
		{
			name:   "empty raw query",
			change: func(r *CreateAnalysisRequest) { *r.MetricSpec.RawMetric.Prometheus.PromQL = "" },
			errors: []govytest.ExpectedRuleError{{
				PropertyPath: "metricSpec.rawMetric.prometheus.promql", Code: rules.ErrorCodeStringNotEmpty,
			}},
		},
		{
			name: "missing count query",
			change: func(r *CreateAnalysisRequest) {
				r.MetricSpec.RawMetric = nil
				r.MetricSpec.CountMetrics = validCountMetrics()
				r.MetricSpec.CountMetrics.TotalMetric.Prometheus.PromQL = nil
			},
			errors: []govytest.ExpectedRuleError{{
				PropertyPath: "metricSpec.countMetrics.total.prometheus.promql", Code: rules.ErrorCodeRequired,
			}},
		},
		{
			name: "missing incremental",
			change: func(r *CreateAnalysisRequest) {
				r.MetricSpec.RawMetric = nil
				r.MetricSpec.CountMetrics = validCountMetrics()
				r.MetricSpec.CountMetrics.Incremental = nil
			},
			errors: []govytest.ExpectedRuleError{{
				PropertyPath: "metricSpec.countMetrics.incremental", Code: rules.ErrorCodeRequired,
			}},
		},
		{
			name:   "invalid start",
			change: func(r *CreateAnalysisRequest) { r.Period.StartTime = "2026-09-01T00:00:00Z" },
			errors: []govytest.ExpectedRuleError{{PropertyPath: "period.startTime", Code: rules.ErrorCodeStringDateTime}},
		},
		{
			name:   "missing end",
			change: func(r *CreateAnalysisRequest) { r.Period.EndTime = "" },
			errors: []govytest.ExpectedRuleError{{PropertyPath: "period.endTime", Code: rules.ErrorCodeRequired}},
		},
		{
			name:   "fractional seconds",
			change: func(r *CreateAnalysisRequest) { r.Period.EndTime = "2026-09-01 00:05:00.1" },
			errors: []govytest.ExpectedRuleError{{PropertyPath: "period.endTime", Code: "sli_analysis_timestamp_precision"}},
		},
		{
			name:   "invalid zone",
			change: func(r *CreateAnalysisRequest) { r.Period.TimeZone = "Invalid/Zone" },
			errors: []govytest.ExpectedRuleError{{PropertyPath: "period.timeZone", Code: rules.ErrorCodeStringTimeZone}},
		},
		{
			name:   "missing zone",
			change: func(r *CreateAnalysisRequest) { r.Period.TimeZone = "" },
			errors: []govytest.ExpectedRuleError{{PropertyPath: "period.timeZone", Code: rules.ErrorCodeRequired}},
		},
		{
			name:   "reversed period",
			change: func(r *CreateAnalysisRequest) { r.Period.EndTime = "2026-08-31 23:59:59" },
			errors: []govytest.ExpectedRuleError{{PropertyPath: "period", Code: "sli_analysis_period"}},
		},
		{
			name:   "equal dates",
			change: func(r *CreateAnalysisRequest) { r.Period.EndTime = r.Period.StartTime },
			errors: []govytest.ExpectedRuleError{{PropertyPath: "period", Code: "sli_analysis_period"}},
		},
		{
			name:   "short period",
			change: func(r *CreateAnalysisRequest) { r.Period.EndTime = "2026-09-01 00:04:59" },
			errors: []govytest.ExpectedRuleError{{PropertyPath: "period", Code: "sli_analysis_period"}},
		},
		{
			name: "independent failures",
			change: func(r *CreateAnalysisRequest) {
				r.Metadata.DisplayName = ""
				r.MetricSpec.MetricSource = ""
				r.Period.TimeZone = "Invalid/Zone"
			},
			errors: []govytest.ExpectedRuleError{
				{PropertyPath: "metadata.displayName", Code: rules.ErrorCodeRequired},
				{PropertyPath: "metricSpec.metricSource", Code: rules.ErrorCodeRequired},
				{PropertyPath: "period.timeZone", Code: rules.ErrorCodeStringTimeZone},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			request := validCreateAnalysisRequest()
			tt.change(&request)
			govytest.AssertError(t, request.Validate(), tt.errors...)
		})
	}

	t.Run("valid raw and count metrics at boundaries", func(t *testing.T) {
		t.Parallel()
		request := validCreateAnalysisRequest()
		govytest.AssertNoError(t, request.Validate())
		request.Metadata.Name = strings.Repeat("x", 253)
		request.Metadata.DisplayName = strings.Repeat("x", 253)
		request.Metadata.Project = strings.Repeat("x", 253)
		request.MetricSpec.Kind = manifest.KindDirect
		request.MetricSpec.MetricSource = strings.Repeat("x", 253)
		request.Period.TimeZone = "Europe/Warsaw"
		govytest.AssertNoError(t, request.Validate())
		request.MetricSpec.RawMetric = nil
		request.MetricSpec.CountMetrics = validCountMetrics()
		govytest.AssertNoError(t, request.Validate())
		*request.MetricSpec.CountMetrics.Incremental = false
		govytest.AssertNoError(t, request.Validate())
		request.Period.TimeZone = "Local"
		govytest.AssertNoError(t, request.Validate())
	})
}

func TestUpdateAnalysisRequest_Validate(t *testing.T) {
	t.Parallel()

	govytest.AssertNoError(t, (UpdateAnalysisRequest{Project: "test", DisplayName: "Analysis"}).Validate())
	govytest.AssertNoError(t, (UpdateAnalysisRequest{
		Project: strings.Repeat("x", 253), DisplayName: strings.Repeat("x", 253),
	}).Validate())
	govytest.AssertError(t, (UpdateAnalysisRequest{}).Validate(),
		govytest.ExpectedRuleError{PropertyPath: "project", Code: rules.ErrorCodeRequired},
		govytest.ExpectedRuleError{PropertyPath: "displayName", Code: rules.ErrorCodeRequired},
	)
	govytest.AssertError(t, (UpdateAnalysisRequest{
		Project: "Invalid project", DisplayName: strings.Repeat("x", 254),
	}).Validate(),
		govytest.ExpectedRuleError{PropertyPath: "project", Code: "string_name"},
		govytest.ExpectedRuleError{PropertyPath: "displayName", Code: rules.ErrorCodeStringMaxLength},
	)
}

func TestCreateCalculationRequest_Validate(t *testing.T) {
	t.Parallel()

	request := CreateCalculationRequest{BudgetTarget: 0.99, BudgetingMethod: "Occurrences", Operator: "lte"}
	govytest.AssertNoError(t, request.Validate())
	for _, operator := range []string{"lte", "lt", "gte", "gt"} {
		for _, method := range []string{"Occurrences", "Timeslices"} {
			valid := request
			valid.Operator = operator
			valid.BudgetingMethod = method
			valid.TimeSliceTarget = 0.95
			valid.Value = -1
			govytest.AssertNoError(t, valid.Validate())
		}
	}
	for _, tt := range []struct {
		name   string
		change func(*CreateCalculationRequest)
		error  govytest.ExpectedRuleError
	}{
		{
			"negative target", func(r *CreateCalculationRequest) { r.BudgetTarget = -0.1 },
			govytest.ExpectedRuleError{PropertyPath: "target", Code: rules.ErrorCodeGreaterThan},
		},
		{
			"zero target", func(r *CreateCalculationRequest) { r.BudgetTarget = 0 },
			govytest.ExpectedRuleError{PropertyPath: "target", Code: rules.ErrorCodeRequired},
		},
		{
			"target equals one", func(r *CreateCalculationRequest) { r.BudgetTarget = 1 },
			govytest.ExpectedRuleError{PropertyPath: "target", Code: rules.ErrorCodeLessThan},
		},
		{
			"target exceeds one", func(r *CreateCalculationRequest) { r.BudgetTarget = 1.1 },
			govytest.ExpectedRuleError{PropertyPath: "target", Code: rules.ErrorCodeLessThan},
		},
		{
			"missing method", func(r *CreateCalculationRequest) { r.BudgetingMethod = "" },
			govytest.ExpectedRuleError{PropertyPath: "budgetingMethod", Code: rules.ErrorCodeRequired},
		},
		{
			"invalid method", func(r *CreateCalculationRequest) { r.BudgetingMethod = "invalid" },
			govytest.ExpectedRuleError{PropertyPath: "budgetingMethod", Code: rules.ErrorCodeOneOf},
		},
		{
			"missing operator", func(r *CreateCalculationRequest) { r.Operator = "" },
			govytest.ExpectedRuleError{PropertyPath: "op", Code: rules.ErrorCodeRequired},
		},
		{
			"invalid operator", func(r *CreateCalculationRequest) { r.Operator = "invalid" },
			govytest.ExpectedRuleError{PropertyPath: "op", Code: rules.ErrorCodeOneOf},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			invalid := request
			tt.change(&invalid)
			govytest.AssertError(t, invalid.Validate(), tt.error)
		})
	}
	govytest.AssertError(t, (CreateCalculationRequest{}).Validate(),
		govytest.ExpectedRuleError{PropertyPath: "target", Code: rules.ErrorCodeRequired},
		govytest.ExpectedRuleError{PropertyPath: "budgetingMethod", Code: rules.ErrorCodeRequired},
		govytest.ExpectedRuleError{PropertyPath: "op", Code: rules.ErrorCodeRequired},
	)
}

func validCreateAnalysisRequest() CreateAnalysisRequest {
	query := "vector(1)"
	return CreateAnalysisRequest{
		Metadata: AnalysisMetadata{DisplayName: "Analysis", Project: "test"},
		MetricSpec: AnalysisMetricSpec{
			Kind: manifest.KindAgent, MetricSource: "prometheus",
			RawMetric: &v1alphaSLO.MetricSpec{Prometheus: &v1alphaSLO.PrometheusMetric{PromQL: &query}},
		},
		Period: AnalysisPeriod{
			StartTime: "2026-09-01 00:00:00", EndTime: "2026-09-01 00:05:00", TimeZone: "UTC",
		},
	}
}

func validCountMetrics() *v1alphaSLO.CountMetricsSpec {
	incremental := true
	return &v1alphaSLO.CountMetricsSpec{
		Incremental: &incremental,
		GoodMetric:  validCreateAnalysisRequest().MetricSpec.RawMetric,
		TotalMetric: validCreateAnalysisRequest().MetricSpec.RawMetric,
	}
}
