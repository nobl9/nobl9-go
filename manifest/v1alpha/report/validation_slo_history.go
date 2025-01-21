package report

import (
	"time"

	"github.com/pkg/errors"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

var sloHistoryValidation = govy.New[SLOHistoryConfig](
	govy.For(func(s SLOHistoryConfig) SLOHistoryTimeFrame { return s.TimeFrame }).
		WithName("timeFrame").
		Required().
		Rules(rules.MutuallyExclusive(true, map[string]func(t SLOHistoryTimeFrame) any{
			"rolling":  func(t SLOHistoryTimeFrame) any { return t.Rolling },
			"calendar": func(t SLOHistoryTimeFrame) any { return t.Calendar },
		})).
		Include(govy.New[SLOHistoryTimeFrame](
			govy.For(func(s SLOHistoryTimeFrame) string { return s.TimeZone }).
				WithName("timeZone").
				Required().
				Rules(govy.NewRule(func(v string) error {
					if _, err := time.LoadLocation(v); err != nil {
						return errors.Wrap(err, "not a valid time zone")
					}
					return nil
				})),
			govy.ForPointer(func(s SLOHistoryTimeFrame) *RollingTimeFrame { return s.Rolling }).
				WithName("rolling").
				Include(rollingTimeFrameValidation),
			govy.ForPointer(func(s SLOHistoryTimeFrame) *CalendarTimeFrame { return s.Calendar }).
				WithName("calendar").
				Include(calendarTimeFrameValidation),
		)),
)
