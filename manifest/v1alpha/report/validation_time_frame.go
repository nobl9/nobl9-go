package report

import (
	"time"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
	"github.com/pkg/errors"
)

const (
	week    = "Week"
	month   = "Month"
	quarter = "Quarter"
	year    = "Year"
)

func GetMinDate() time.Time {
	return time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
}

const (
	// IsoDateOnlyLayout is date only (without time zone) of iso layout
	IsoDateOnlyLayout string = "2006-01-02"
)

var rollingTimeFrameValidation = govy.New[RollingTimeFrame](
	govy.ForPointer(func(t RollingTimeFrame) *string { return t.Unit }).
		WithName("unit").
		Required().
		Rules(validationRuleTimeUnit()),
	govy.ForPointer(func(t RollingTimeFrame) *int { return t.Count }).
		WithName("count").
		Required(),
	govy.For(govy.GetSelf[RollingTimeFrame]()).
		Rules(rollingTimeFrameCountAndUnitValidation),
)

var rollingTimeFrameCountAndUnitValidation = govy.NewRule(func(t RollingTimeFrame) error {
	if t.Count != nil && t.Unit != nil &&
		(((*t.Count == 1 || *t.Count == 2 || *t.Count == 4) && *t.Unit == week) ||
			(*t.Count == 1 && (*t.Unit == month || *t.Unit == quarter || *t.Unit == year))) {
		return nil
	}
	return errors.New(
		"valid 'unit' and 'count' pairs are: 1 week, 2 weeks, 4 weeks, 1 month, 1 quarter, 1 year",
	)
})

func validationRuleTimeUnit() govy.Rule[string] {
	return rules.OneOf[string](week, month, quarter, year)
}

var calendarTimeFrameValidation = govy.New[CalendarTimeFrame](
	govy.For(govy.GetSelf[CalendarTimeFrame]()).
		Rules(
			govy.NewRule(func(t CalendarTimeFrame) error {
				allFieldsSet := t.Count != nil && t.Unit != nil && t.From != nil && t.To != nil
				noFieldSet := t.Count == nil && t.Unit == nil && t.From == nil && t.To == nil
				onlyCountSet := t.Count != nil && t.Unit == nil
				onlyUnitSet := t.Count == nil && t.Unit != nil
				onlyFromSet := t.From != nil && t.To == nil
				onlyToSet := t.From == nil && t.To != nil
				if allFieldsSet || noFieldSet || onlyCountSet || onlyUnitSet || onlyFromSet || onlyToSet {
					return errors.New("must contain either 'unit' and 'count' pair or 'from' and 'to' pair")
				}
				return nil
			}),
			govy.NewRule(func(t CalendarTimeFrame) error {
				if t.From != nil && t.To != nil {
					if *t.From >= *t.To {
						return errors.New("'from' must be before 'to'")
					}

					from, err := time.Parse(IsoDateOnlyLayout, *t.From)
					if err != nil {
						return errors.New("error parsing 'from' date")
					}

					to, err := time.Parse(IsoDateOnlyLayout, *t.To)
					if err != nil {
						return errors.New("error parsing 'to' date")
					}

					now := time.Now()
					if now.Before(from) || now.Before(to) {
						return errors.New("dates must be in the past")
					}

					minDate := GetMinDate()
					if from.Before(minDate) || to.Before(minDate) {
						return errors.Errorf("date must be after or equal to %s", minDate.Format(time.DateOnly))
					}
				}
				return nil
			}),
			govy.NewRule(func(t CalendarTimeFrame) error {
				if t.Count != nil && t.Unit != nil &&
					!(*t.Count == 1 && (*t.Unit == week || *t.Unit == month || *t.Unit == quarter || *t.Unit == year)) {
					return errors.New("valid 'unit' and 'count' pairs are: 1 week, 1 month, 1 quarter, 1 year")
				}
				return nil
			})),
	govy.ForPointer(func(t CalendarTimeFrame) *string { return t.Unit }).
		WithName("unit").
		Rules(validationRuleTimeUnit()),
)
