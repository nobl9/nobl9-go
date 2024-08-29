package report

import (
	"time"

	"github.com/pkg/errors"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

var sloHistoryValidation = govy.New[SLOHistoryConfig](
	govy.For(func(s SLOHistoryConfig) string { return s.TimeFrame.TimeZone }).
		WithName("timeZone").
		Required().
		Rules(govy.NewRule(func(v string) error {
			if _, err := time.LoadLocation(v); err != nil {
				return errors.Wrap(err, "not a valid time zone")
			}
			return nil
		})),
	govy.For(func(s SLOHistoryConfig) SLOHistoryTimeFrame { return s.TimeFrame }).
		WithName("timeFrame").
		Required().
		Rules(rules.MutuallyExclusive(true, map[string]func(t SLOHistoryTimeFrame) any{
			"rolling":  func(t SLOHistoryTimeFrame) any { return t.Rolling },
			"calendar": func(t SLOHistoryTimeFrame) any { return t.Calendar },
		})),
	govy.ForPointer(func(s SLOHistoryConfig) *RollingTimeFrame { return s.TimeFrame.Rolling }).
		WithName("rolling").
		Include(rollingTimeFrameValidation),
	govy.ForPointer(func(s SLOHistoryConfig) *CalendarTimeFrame { return s.TimeFrame.Calendar }).
		WithName("calendar").
		Include(calendarTimeFrameValidation),
)

var rollingTimeFrameValidation = govy.New[RollingTimeFrame](
	govy.ForPointer(func(t RollingTimeFrame) *string { return t.Unit }).
		WithName("unit").
		Required(),
	govy.ForPointer(func(t RollingTimeFrame) *int { return t.Count }).
		WithName("count").
		Required(),
)

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
					return errors.New("must contain either unit and count pair or from and to pair")
				}
				return nil
			})),
)
