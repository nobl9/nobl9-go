package slo

import (
	"time"

	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/twindow"
)

// TimeWindow represents content of time window
type TimeWindow struct {
	Unit      string    `json:"unit"`
	Count     int       `json:"count"`
	IsRolling bool      `json:"isRolling"`
	Calendar  *Calendar `json:"calendar,omitempty"`

	// Period is only returned in `/get/slo` requests it is ignored for `/apply`
	Period *Period `json:"period,omitempty"`
}

// GetType returns value of twindow.TimeWindowTypeEnum for given time window>
func (tw TimeWindow) GetType() twindow.TimeWindowTypeEnum {
	if tw.isCalendar() {
		return twindow.Calendar
	}
	return twindow.Rolling
}

func (tw TimeWindow) isCalendar() bool {
	return tw.Calendar != nil
}

// Calendar struct represents calendar time window
type Calendar struct {
	StartTime string `json:"startTime"`
	TimeZone  string `json:"timeZone"`
}

// Period represents period of time
type Period struct {
	Begin string `json:"begin"`
	End   string `json:"end"`
}

// Values used to validate time window size
const (
	minimumRollingTimeWindowSize  = 5 * time.Minute
	maximumRollingTimeWindowSize  = 31 * 24 * time.Hour  // 31 days
	maximumCalendarTimeWindowSize = 366 * 24 * time.Hour // 366 days
)

var timeWindowsValidation = validation.New[TimeWindow](
	validation.For(func(t TimeWindow) string { return t.Unit }).
		WithName("unit").
		Required().
		Rules(twindow.ValidationRuleTimeUnit()),
	validation.For(func(t TimeWindow) int { return t.Count }).
		WithName("count").
		Rules(validation.GreaterThan(0)),
	validation.ForPointer(func(t TimeWindow) *Calendar { return t.Calendar }).
		WithName("calendar").
		Include(validation.New[Calendar](
			validation.For(func(c Calendar) string { return c.StartTime }).
				WithName("startTime").
				Required().
				Rules(calendarStartTimeValidationRule()),
			validation.For(func(c Calendar) string { return c.TimeZone }).
				WithName("timeZone").
				Required().
				Rules(validation.NewSingleRule(func(v string) error {
					if _, err := time.LoadLocation(v); err != nil {
						return errors.Wrap(err, "not a valid time zone")
					}
					return nil
				}))),
		),
)

func timeWindowValidationRule() validation.SingleRule[TimeWindow] {
	return validation.NewSingleRule(func(v TimeWindow) error {
		if err := validateTimeWindowAmbiguity(v); err != nil {
			return err
		}
		if err := validateTimeUnitForTimeWindowType(v); err != nil {
			return err
		}
		switch v.GetType() {
		case twindow.Rolling:
			return rollingWindowSizeValidation(v)
		case twindow.Calendar:
			return calendarWindowSizeValidation(v)
		}
		return nil
	})
}

func rollingWindowSizeValidation(timeWindow TimeWindow) error {
	rollingWindowTimeUnitEnum := twindow.GetTimeUnitEnum(twindow.Rolling, timeWindow.Unit)
	var timeWindowSize time.Duration
	switch rollingWindowTimeUnitEnum {
	case twindow.Minute:
		timeWindowSize = time.Duration(timeWindow.Count) * time.Minute
	case twindow.Hour:
		timeWindowSize = time.Duration(timeWindow.Count) * time.Hour
	case twindow.Day:
		timeWindowSize = time.Duration(timeWindow.Count) * 24 * time.Hour
	default:
		return errors.New("valid window type for time unit required")
	}
	switch {
	case timeWindowSize > maximumRollingTimeWindowSize:
		return errors.Errorf(
			"rolling time window size must be less than or equal to %s",
			maximumRollingTimeWindowSize)
	case timeWindowSize < minimumRollingTimeWindowSize:
		return errors.Errorf(
			"rolling time window size must be greater than or equal to %s",
			minimumRollingTimeWindowSize)
	}
	return nil
}

func calendarWindowSizeValidation(timeWindow TimeWindow) error {
	tw, err := twindow.NewCalendarTimeWindow(
		twindow.MustParseTimeUnit(timeWindow.Unit),
		uint32(timeWindow.Count),
		time.UTC,
		time.Now().UTC(),
	)
	if err != nil {
		return err
	}
	timeWindowSize := tw.GetTimePeriod(time.Now().UTC()).Duration()
	if timeWindowSize > maximumCalendarTimeWindowSize {
		return errors.Errorf("calendar time window size must be less than %s", maximumCalendarTimeWindowSize)
	}
	return nil
}

func validateTimeWindowAmbiguity(timeWindow TimeWindow) error {
	if timeWindow.IsRolling && timeWindow.isCalendar() {
		return errors.New(
			"if 'isRolling' property is true, 'calendar' property must be omitted")
	}
	if !timeWindow.IsRolling && !timeWindow.isCalendar() {
		return errors.New(
			"if 'isRolling' property is false or not set, 'calendar' property must be provided")
	}
	return nil
}

func validateTimeUnitForTimeWindowType(tw TimeWindow) error {
	var err error
	typ := tw.GetType()
	switch typ {
	case twindow.Rolling:
		err = twindow.ValidateRollingWindowTimeUnit(tw.Unit)
	case twindow.Calendar:
		err = twindow.ValidateCalendarAlignedTimeUnit(tw.Unit)
	}
	if err != nil {
		return errors.Wrapf(err, "invalid time window unit for %s window type", typ)
	}
	return nil
}

func calendarStartTimeValidationRule() validation.SingleRule[string] {
	return validation.NewSingleRule(func(v string) error {
		date, err := twindow.ParseStartDate(v)
		if err != nil {
			return err
		}
		minStartDate := twindow.GetMinStartDate()
		if date.Before(minStartDate) {
			return errors.Errorf("date must be after or equal to %s", minStartDate.Format(time.RFC3339))
		}
		if date.Nanosecond() != 0 {
			return errors.New(
				"setting nanoseconds or milliseconds in time are forbidden to be set")
		}
		return nil
	})
}
