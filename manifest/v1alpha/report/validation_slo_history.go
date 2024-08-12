package report

import (
	"time"

	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/internal/validation"
)

var sloHistoryValidation = validation.New[SLOHistoryConfig](
	validation.For(func(s SLOHistoryConfig) string { return s.TimeFrame.TimeZone }).
		WithName("timeZone").
		Required().
		Rules(validation.NewSingleRule(func(v string) error {
			if _, err := time.LoadLocation(v); err != nil {
				return errors.Wrap(err, "not a valid time zone")
			}
			return nil
		})),
	validation.For(func(s SLOHistoryConfig) SLOHistoryTimeFrame { return s.TimeFrame }).
		WithName("timeFrame").
		Required().
		Rules(validation.MutuallyExclusive(true, map[string]func(t SLOHistoryTimeFrame) any{
			"rolling":  func(t SLOHistoryTimeFrame) any { return t.Rolling },
			"calendar": func(t SLOHistoryTimeFrame) any { return t.Calendar },
		})),
	validation.ForPointer(func(s SLOHistoryConfig) *RollingTimeFrame { return s.TimeFrame.Rolling }).
		WithName("rolling").
		Include(rollingTimeFrameValidation),
	validation.ForPointer(func(s SLOHistoryConfig) *CalendarTimeFrame { return s.TimeFrame.Calendar }).
		WithName("calendar").
		Include(calendarTimeFrameValidation),
)

var rollingTimeFrameValidation = validation.New[RollingTimeFrame](
	validation.ForPointer(func(t RollingTimeFrame) *string { return t.Unit }).
		WithName("unit").
		Required(),
	validation.ForPointer(func(t RollingTimeFrame) *int { return t.Count }).
		WithName("count").
		Required(),
)

var calendarTimeFrameValidation = validation.New[CalendarTimeFrame](
	validation.For(validation.GetSelf[CalendarTimeFrame]()).
		Rules(
			validation.NewSingleRule(func(t CalendarTimeFrame) error {
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
