package slo

import (
	"testing"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestGCM(t *testing.T) {
	t.Run("raw metrics", func(t *testing.T) {
		t.Run("passes", func(t *testing.T) {
			slo := validRawMetricSLO(v1alpha.GCM)
			err := validate(slo)
			testutils.AssertNoError(t, slo, err)
		})
		t.Run("required", func(t *testing.T) {
			t.Run("projectID missing", func(t *testing.T) {
				slo := validRawMetricSLO(v1alpha.GCM)
				slo.Spec.Objectives[0].RawMetric.MetricQuery.GCM = &GCMMetric{
					Query: "123",
				}
				err := validate(slo)
				testutils.AssertContainsErrors(t, slo, err, 1,
					testutils.ExpectedError{
						Prop: "spec.objectives[0].rawMetric.query.gcm.projectId",
						Code: rules.ErrorCodeRequired,
					},
				)
			})
			t.Run("query missing", func(t *testing.T) {
				slo := validRawMetricSLO(v1alpha.GCM)
				slo.Spec.Objectives[0].RawMetric.MetricQuery.GCM = &GCMMetric{
					ProjectID: "123",
				}
				err := validate(slo)
				testutils.AssertContainsErrors(t, slo, err, 1,
					testutils.ExpectedError{
						Prop: "spec.objectives[0].rawMetric.query.gcm",
						Code: rules.ErrorCodeOneOf,
					},
				)
			})
		})
		t.Run("both mql and promql defined", func(t *testing.T) {
			slo := validRawMetricSLO(v1alpha.GCM)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.GCM = &GCMMetric{
				ProjectID: "123",
				Query:     "123",
				PromQL:    "123",
			}
			err := validate(slo)
			testutils.AssertContainsErrors(t, slo, err, 1,
				testutils.ExpectedError{
					Prop: "spec.objectives[0].rawMetric.query.gcm",
					Code: rules.ErrorCodeOneOf,
				},
			)
		})
	})
	t.Run("count metrics", func(t *testing.T) {
		t.Run("passes mql", func(t *testing.T) {
			slo := validCountMetricSLO(v1alpha.GCM)
			err := validate(slo)
			testutils.AssertNoError(t, slo, err)
		})
		t.Run("passes promql", func(t *testing.T) {
			slo := validCountMetricSLO(v1alpha.GCM)
			slo.Spec.Objectives[0].CountMetrics.GoodMetric.GCM = &GCMMetric{
				ProjectID: "123",
				PromQL:    "123",
			}
			slo.Spec.Objectives[0].CountMetrics.TotalMetric.GCM = &GCMMetric{
				ProjectID: "123",
				PromQL:    "123",
			}
			err := validate(slo)
			testutils.AssertNoError(t, slo, err)
		})
		t.Run("good is mql, total is promql", func(t *testing.T) {
			slo := validCountMetricSLO(v1alpha.GCM)
			slo.Spec.Objectives[0].CountMetrics.TotalMetric.GCM = &GCMMetric{
				ProjectID: "123",
				PromQL:    "123",
			}
			err := validate(slo)
			testutils.AssertContainsErrors(t, slo, err, 1,
				testutils.ExpectedError{
					Prop: "spec.objectives[0].countMetrics",
					Code: rules.ErrorCodeNotEqualTo,
				},
			)
		})
		t.Run("good is promql, total is mql", func(t *testing.T) {
			slo := validCountMetricSLO(v1alpha.GCM)
			slo.Spec.Objectives[0].CountMetrics.GoodMetric.GCM = &GCMMetric{
				ProjectID: "123",
				PromQL:    "123",
			}
			err := validate(slo)
			testutils.AssertContainsErrors(t, slo, err, 1,
				testutils.ExpectedError{
					Prop: "spec.objectives[0].countMetrics",
					Code: rules.ErrorCodeNotEqualTo,
				},
			)
		})
	})
}
