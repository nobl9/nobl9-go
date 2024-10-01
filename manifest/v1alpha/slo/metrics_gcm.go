package slo

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

// GCMMetric represents metric from GCM
type GCMMetric struct {
	Query       string `json:"query,omitempty"`
	ProjectID   string `json:"projectId"`
	PromqlQuery string `json:"promqlQuery,omitempty"`
}

// IsMQLConfiguration returns true if the metric is configured with MQL query.
func (g GCMMetric) IsMQLConfiguration() bool {
	return g.Query != ""
}

// IsPromqlConfiguration returns true if the metric is configured with promql query.
func (g GCMMetric) IsPromqlConfiguration() bool {
	return g.PromqlQuery != ""
}

var gcmValidation = govy.New[GCMMetric](
	govy.For(func(e GCMMetric) string { return e.ProjectID }).
		WithName("projectId").
		Required(),
	govy.For(govy.GetSelf[GCMMetric]()).Rules(
		govy.NewRule(func(c GCMMetric) error {
			var configOptions int
			if c.IsMQLConfiguration() {
				configOptions++
			}
			if c.IsPromqlConfiguration() {
				configOptions++
			}
			if configOptions != 1 {
				return errors.New("exactly one of configuration option is required [query, promqlQuery]")
			}
			return nil
		}).WithErrorCode(rules.ErrorCodeOneOf),
	),
)

var gcmCountMetricsLevelValidation = govy.New[CountMetricsSpec](
	govy.For(govy.GetSelf[CountMetricsSpec]()).Rules(
		govy.NewRule(func(c CountMetricsSpec) error {
			total := c.TotalMetric
			good := c.GoodMetric

			if total == nil {
				return nil
			}

			if good != nil {
				if total.GCM.IsPromqlConfiguration() && !good.GCM.IsPromqlConfiguration() {
					return countMetricsPropertyEqualityError("query format [gcm.promqlQuery]", goodMetric)
				}
				if total.GCM.IsMQLConfiguration() && !good.GCM.IsMQLConfiguration() {
					return countMetricsPropertyEqualityError("query format [gcm.query]", goodMetric)
				}
			}

			return nil
		}).WithErrorCode(rules.ErrorCodeNotEqualTo),
	),
).When(
	whenCountMetricsIs(v1alpha.GCM),
	govy.WhenDescription("countMetrics is GCM"),
)
