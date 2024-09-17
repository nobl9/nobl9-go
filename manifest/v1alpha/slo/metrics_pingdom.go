package slo

import (
	"regexp"
	"strings"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

// PingdomMetric represents metric from Pingdom.
type PingdomMetric struct {
	CheckID   *string `json:"checkId"`
	CheckType *string `json:"checkType"`
	Status    *string `json:"status,omitempty"`
}

const (
	PingdomTypeUptime      = "uptime"
	PingdomTypeTransaction = "transaction"
)

const (
	pingdomStatusUp          = "up"
	pingdomStatusDown        = "down"
	pingdomStatusUnconfirmed = "unconfirmed"
	pingdomStatusUnknown     = "unknown"
)

var pingdomCountMetricsLevelValidation = govy.New[CountMetricsSpec](
	govy.For(govy.GetSelf[CountMetricsSpec]()).
		Rules(
			govy.NewRule(func(c CountMetricsSpec) error {
				if c.GoodMetric.Pingdom.CheckID == nil || c.TotalMetric.Pingdom.CheckID == nil {
					return nil
				}
				if *c.GoodMetric.Pingdom.CheckID != *c.TotalMetric.Pingdom.CheckID {
					return countMetricsPropertyEqualityError("pingdom.checkId", goodMetric)
				}
				return nil
			}).WithErrorCode(rules.ErrorCodeEqualTo),
			govy.NewRule(func(c CountMetricsSpec) error {
				if c.GoodMetric.Pingdom.CheckType == nil || c.TotalMetric.Pingdom.CheckType == nil {
					return nil
				}
				if *c.GoodMetric.Pingdom.CheckType != *c.TotalMetric.Pingdom.CheckType {
					return countMetricsPropertyEqualityError("pingdom.checkType", goodMetric)
				}
				return nil
			}).WithErrorCode(rules.ErrorCodeEqualTo),
		),
).When(
	whenCountMetricsIs(v1alpha.Pingdom),
	govy.WhenDescription("countMetrics is pingdom"),
)

// createPingdomMetricSpecValidation constructs a new MetricSpec level validation for Pingdom.
func createPingdomMetricSpecValidation(
	include govy.Validator[PingdomMetric],
) govy.Validator[MetricSpec] {
	return govy.New[MetricSpec](
		govy.ForPointer(func(m MetricSpec) *PingdomMetric { return m.Pingdom }).
			WithName("pingdom").
			Include(include))
}

var pingdomRawMetricValidation = createPingdomMetricSpecValidation(govy.New[PingdomMetric](
	govy.ForPointer(func(p PingdomMetric) *string { return p.CheckType }).
		WithName("checkType").
		Required().
		Rules(rules.EQ(PingdomTypeUptime)),
))

var pingdomCountMetricsValidation = createPingdomMetricSpecValidation(govy.New[PingdomMetric](
	govy.ForPointer(func(p PingdomMetric) *string { return p.CheckType }).
		WithName("checkType").
		Required().
		Rules(rules.OneOf(PingdomTypeUptime, PingdomTypeTransaction)),
	govy.For(func(p PingdomMetric) *string { return p.Status }).
		WithName("status").
		When(
			func(p PingdomMetric) bool { return p.CheckType != nil && *p.CheckType == PingdomTypeUptime },
			govy.WhenDescription("checkType is equal to '%s'", PingdomTypeUptime),
		).
		Rules(rules.Required[*string]()),
))

var pingdomValidation = govy.New[PingdomMetric](
	govy.For(govy.GetSelf[PingdomMetric]()).
		Include(pingdomUptimeCheckTypeValidation).
		Include(pingdomTransactionCheckTypeValidation),
	govy.ForPointer(func(p PingdomMetric) *string { return p.CheckID }).
		WithName("checkId").
		Required().
		Rules(
			rules.StringNotEmpty(),
			// This regexp is crafted in order to not interweave with StringNotEmpty govy.
			rules.StringMatchRegexp(regexp.MustCompile(`^(?:|\d+)$`))), // nolint: gocritic
)

var pingdomUptimeCheckTypeValidation = govy.New[PingdomMetric](
	govy.ForPointer(func(m PingdomMetric) *string { return m.Status }).
		WithName("status").
		Rules(govy.NewRule(func(s string) error {
			for _, status := range strings.Split(s, ",") {
				if err := rules.OneOf(
					pingdomStatusUp,
					pingdomStatusDown,
					pingdomStatusUnconfirmed,
					pingdomStatusUnknown).Validate(status); err != nil {
					return err
				}
			}
			return nil
		})),
).
	When(
		func(m PingdomMetric) bool { return m.CheckType != nil && *m.CheckType == PingdomTypeUptime },
		govy.WhenDescription("checkType is equal to '%s'", PingdomTypeUptime),
	)

var pingdomTransactionCheckTypeValidation = govy.New[PingdomMetric](
	govy.ForPointer(func(m PingdomMetric) *string { return m.Status }).
		WithName("status").
		Rules(rules.Forbidden[string]()),
).
	When(
		func(m PingdomMetric) bool { return m.CheckType != nil && *m.CheckType == PingdomTypeTransaction },
		govy.WhenDescription("checkType is equal to '%s'", PingdomTypeTransaction),
	)
