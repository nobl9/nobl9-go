package slo

import (
	"regexp"
	"strings"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
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

var pingdomCountMetricsLevelValidation = validation.New[CountMetricsSpec](
	validation.For(validation.GetSelf[CountMetricsSpec]()).
		Rules(
			validation.NewSingleRule(func(c CountMetricsSpec) error {
				if c.GoodMetric.Pingdom.CheckID == nil || c.TotalMetric.Pingdom.CheckID == nil {
					return nil
				}
				if *c.GoodMetric.Pingdom.CheckID != *c.TotalMetric.Pingdom.CheckID {
					return countMetricsPropertyEqualityError("pingdom.checkId", goodMetric)
				}
				return nil
			}).WithErrorCode(validation.ErrorCodeEqualTo),
			validation.NewSingleRule(func(c CountMetricsSpec) error {
				if c.GoodMetric.Pingdom.CheckType == nil || c.TotalMetric.Pingdom.CheckType == nil {
					return nil
				}
				if *c.GoodMetric.Pingdom.CheckType != *c.TotalMetric.Pingdom.CheckType {
					return countMetricsPropertyEqualityError("pingdom.checkType", goodMetric)
				}
				return nil
			}).WithErrorCode(validation.ErrorCodeEqualTo),
		),
).When(whenCountMetricsIs(v1alpha.Pingdom))

// createPingdomMetricSpecValidation constructs a new MetriSpec level validation for Pingdom.
func createPingdomMetricSpecValidation(
	include validation.Validator[PingdomMetric],
) validation.Validator[MetricSpec] {
	return validation.New[MetricSpec](
		validation.ForPointer(func(m MetricSpec) *PingdomMetric { return m.Pingdom }).
			WithName("pingdom").
			Include(include))
}

var pingdomRawMetricValidation = createPingdomMetricSpecValidation(validation.New[PingdomMetric](
	validation.ForPointer(func(p PingdomMetric) *string { return p.CheckType }).
		WithName("checkType").
		Required().
		Rules(validation.EqualTo(PingdomTypeUptime)),
))

var pingdomCountMetricsValidation = createPingdomMetricSpecValidation(validation.New[PingdomMetric](
	validation.ForPointer(func(p PingdomMetric) *string { return p.CheckType }).
		WithName("checkType").
		Required().
		Rules(validation.OneOf(PingdomTypeUptime, PingdomTypeTransaction)),
))

var pingdomValidation = validation.New[PingdomMetric](
	validation.For(validation.GetSelf[PingdomMetric]()).
		Include(pingdomUptimeCheckTypeValidation).
		Include(pingdomTransactionCheckTypeValidation),
	validation.ForPointer(func(p PingdomMetric) *string { return p.CheckID }).
		WithName("checkId").
		Required().
		Rules(
			validation.StringNotEmpty(),
			// This regexp is crafted in order to not interweave with StringNotEmpty validation.
			validation.StringMatchRegexp(regexp.MustCompile(`^(?:|\d+)$`))), // nolint: gocritic
)

var pingdomUptimeCheckTypeValidation = validation.New[PingdomMetric](
	validation.ForPointer(func(m PingdomMetric) *string { return m.Status }).
		WithName("status").
		Required().
		Rules(validation.NewSingleRule(func(s string) error {
			for _, status := range strings.Split(s, ",") {
				if err := validation.OneOf(
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
	When(func(m PingdomMetric) bool {
		return m.CheckType != nil && *m.CheckType == PingdomTypeUptime
	})

var pingdomTransactionCheckTypeValidation = validation.New[PingdomMetric](
	validation.ForPointer(func(m PingdomMetric) *string { return m.Status }).
		WithName("status").
		Rules(validation.Forbidden[string]()),
).
	When(func(m PingdomMetric) bool {
		return m.CheckType != nil && *m.CheckType == PingdomTypeTransaction
	})
