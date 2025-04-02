// Package twindow provides enums and functions to operate with resources related to Time Windows
package twindow

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

const (
	// IsoDateTimeOnlyLayout is date and time only (without time zone) of iso layout
	IsoDateTimeOnlyLayout string = "2006-01-02 15:04:05"
)

const (
	HoursInDay      uint32 = 24
	daysInWeek      uint32 = 7
	monthsInQuarter uint32 = 3
	monthsInYear    uint32 = 12
)

// TimePeriod represents a period in time with fixed start and end
type TimePeriod struct {
	Start time.Time
	End   time.Time
}

// Duration returns duration from Start to End of TimePeriod
func (tp TimePeriod) Duration() time.Duration {
	return tp.End.Sub(tp.Start)
}

// TimeWindow is an algorithm that assigns a TimePeriod for every point in time
type TimeWindow interface {
	GetTimePeriod(t time.Time) TimePeriod
	String() string
	Type() TimeWindowTypeEnum
	MarshalJSON() ([]byte, error)
}

// Date represents day in calendar
type Date interface {
	Date() (year int, month time.Month, day int)
}

// TimeOfDay represents valid time of day
type TimeOfDay interface {
	Clock() (hour, minute, sec int)
}

// DateWithTime represents day in calendar with time of day
type DateWithTime interface {
	Date
	TimeOfDay
}

type unitCount struct {
	Unit  TimeUnitEnum
	Count uint32
}

func (uc unitCount) String() string {
	return fmt.Sprintf("%d %s", uc.Count, uc.Unit.String())
}

type CalendarTimeWindow struct {
	unitCount
	TimeZone     *time.Location
	DateWithTime DateWithTime
}

func (timeWindow CalendarTimeWindow) GetStartTime() time.Time {
	year, month, day := timeWindow.DateWithTime.Date()
	hour, minute, second := timeWindow.DateWithTime.Clock()
	return time.Date(year, month, day, hour, minute, second, 0, timeWindow.TimeZone)
}

type periodCalculator interface {
	// periodsCountDiff expects that if time.Time is used as Date implementation then it should be in the same time zone
	// as in CalendarTimeWindow.
	periodsCountDiff(CalendarTimeWindow, Date) int
	// periodsThresholdAtDate expects that if time.Time is used as Date implementation then it should be in the same
	// time zone as in CalendarTimeWindow.
	periodsThresholdAtDate(CalendarTimeWindow, Date) time.Time
	addPeriods(time.Time, int) time.Time
}

type yearPeriodCalculator struct{}

func (yearPeriodCalculator) periodsCountDiff(timeWindow CalendarTimeWindow, date Date) int {
	timeWindowYear, _, _ := timeWindow.DateWithTime.Date()
	timestampYear, _, _ := date.Date()
	return timestampYear - timeWindowYear
}

func (yearPeriodCalculator) periodsThresholdAtDate(timeWindow CalendarTimeWindow, date Date) time.Time {
	year, _, _ := date.Date()
	_, month, day := timeWindow.DateWithTime.Date()
	hour, minute, second := timeWindow.DateWithTime.Clock()
	return time.Date(year, month, day, hour, minute, second, 0, timeWindow.TimeZone)
}

func (yearPeriodCalculator) addPeriods(timestamp time.Time, numberOfPeriods int) time.Time {
	return timestamp.AddDate(numberOfPeriods, 0, 0)
}

type monthPeriodCalculator struct{}

func (monthPeriodCalculator) periodsCountDiff(timeWindow CalendarTimeWindow, date Date) int {
	timeWindowYear, timeWindowMonth, timeWindowDay := timeWindow.DateWithTime.Date()
	timestampYear, timestampMonth, _ := date.Date()

	return (timestampYear-timeWindowYear)*int(monthsInYear) +
		int(timestampMonth) - int(timeWindowMonth) +
		normalizedDateCorrection(timeWindowDay, date)
}

// normalizedDateCorrection corrects for the situation when starting date of monthly calendar time window is at a day
// that doesn't  occur in certain months. For example, if the starting date is January 31st and the timestamp is
// March 1st then the difference in months should be 1 and not 2. This mirrors the way AddDate normalizes dates, where
// February 31st is normalized to March 3rd.
func normalizedDateCorrection(timeWindowStartDay int, date Date) int {
	timestampYear, timestampMonth, timestampDay := date.Date()
	daysInPreviousMonth := daysIn(timestampMonth-1, timestampYear)
	isNormalized := timeWindowStartDay > daysInPreviousMonth && timestampDay <= timeWindowStartDay-daysInPreviousMonth
	if isNormalized {
		return -1
	}
	return 0
}

func daysIn(month time.Month, year int) int {
	t := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC)
	return t.Day()
}

func (monthPeriodCalculator) periodsThresholdAtDate(timeWindow CalendarTimeWindow, date Date) time.Time {
	year, month, _ := date.Date()
	_, _, day := timeWindow.DateWithTime.Date()
	hour, minute, second := timeWindow.DateWithTime.Clock()
	month += time.Month(normalizedDateCorrection(day, date))
	return time.Date(year, month, day, hour, minute, second, 0, timeWindow.TimeZone)
}

func (monthPeriodCalculator) addPeriods(timestamp time.Time, numberOfPeriods int) time.Time {
	return timestamp.AddDate(0, numberOfPeriods, 0)
}

type dayPeriodCalculator struct{}

func (dayPeriodCalculator) periodsCountDiff(timeWindow CalendarTimeWindow, date Date) int {
	return daysBetweenDates(timeWindow.GetStartTime(), date)
}

// daysBetweenDates expects that if time.Time is used as Date implementation then both arguments should be in the same
// time zone.
func daysBetweenDates(fromDate, toDate Date) int {
	fromYear, fromMonth, fromDay := fromDate.Date()
	fromTime := time.Date(fromYear, fromMonth, fromDay, 0, 0, 0, 0, time.UTC)

	toYear, toMonth, toDay := toDate.Date()
	toTime := time.Date(toYear, toMonth, toDay, 0, 0, 0, 0, time.UTC)

	return int(toTime.Sub(fromTime).Hours()) / int(HoursInDay)
}

func (dayPeriodCalculator) periodsThresholdAtDate(timeWindow CalendarTimeWindow, date Date) time.Time {
	year, month, day := date.Date()
	hour, minute, second := timeWindow.DateWithTime.Clock()
	return time.Date(year, month, day, hour, minute, second, 0, timeWindow.TimeZone)
}

func (dayPeriodCalculator) addPeriods(timestamp time.Time, numberOfPeriods int) time.Time {
	return timestamp.AddDate(0, 0, numberOfPeriods)
}

// getTimePeriod calculates time period that contains Sometimes start and end of the period have different clock times
// in UTC. It is an intended behavior which arises from changes in daylight savings in a given Location.
// Calendar-aligned time window will calculate periods such as the clock time set in start time will not change
// regardless of the daylight savings. That means that 2 dates in a year don't have 24h, one has 25 and one has 23.
// Intention for that behavior was that calendar-aligned time windows are set around business processes operating with
// the same logic. Example: if a business operating in Australia want's their monthly SLOs starting at 1st of each
// month at midnight, they would expect that each period will correspond to their calendar and their clock instead of
// starting 6 months at midnight and 6 months at 1am. If however user desires different behavior they can always set
// time zone to UTC.
func getTimePeriod(calculator periodCalculator, timeWindow CalendarTimeWindow, timestamp time.Time) TimePeriod {
	timestampInWindowTimeZone := timestamp.In(timeWindow.TimeZone)
	delta := calculator.periodsCountDiff(timeWindow, timestampInWindowTimeZone)

	var correction int
	modulo := delta % int(timeWindow.Count)
	if modulo < 0 ||
		(modulo == 0 && timestamp.Before(calculator.periodsThresholdAtDate(timeWindow, timestampInWindowTimeZone))) {
		correction = -1
	}

	periods := delta/int(timeWindow.Count) + correction

	startTime := timeWindow.GetStartTime()

	return TimePeriod{
		Start: calculator.addPeriods(startTime, periods*int(timeWindow.Count)),
		End:   calculator.addPeriods(startTime, (periods+1)*int(timeWindow.Count)),
	}
}

func (timeWindow CalendarTimeWindow) GetTimePeriod(timestamp time.Time) TimePeriod {
	switch timeWindow.Unit {
	case Year:
		return getTimePeriod(yearPeriodCalculator{}, timeWindow, timestamp)
	case Quarter:
		normalizedWindow := CalendarTimeWindow{
			unitCount: unitCount{
				Unit:  Month,
				Count: timeWindow.Count * monthsInQuarter,
			},
			DateWithTime: timeWindow.DateWithTime,
			TimeZone:     timeWindow.TimeZone,
		}
		return getTimePeriod(monthPeriodCalculator{}, normalizedWindow, timestamp)
	case Month:
		return getTimePeriod(monthPeriodCalculator{}, timeWindow, timestamp)
	case Week:
		normalizedWindow := CalendarTimeWindow{
			unitCount: unitCount{
				Unit:  Day,
				Count: timeWindow.Count * daysInWeek,
			},
			DateWithTime: timeWindow.DateWithTime,
			TimeZone:     timeWindow.TimeZone,
		}
		return getTimePeriod(dayPeriodCalculator{}, normalizedWindow, timestamp)
	case Day:
		return getTimePeriod(dayPeriodCalculator{}, timeWindow, timestamp)
	default:
		return TimePeriod{}
	}
}

func (timeWindow CalendarTimeWindow) String() string {
	return fmt.Sprintf(
		"Calendar %s %s %s",
		timeWindow.unitCount,
		timeWindow.TimeZone,
		timeWindow.GetStartTime().Format(time.RFC3339))
}

func (timeWindow CalendarTimeWindow) Type() TimeWindowTypeEnum {
	return Calendar
}

func (timeWindow CalendarTimeWindow) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"type":       "calendar",
		"unit":       timeWindow.Unit.String(),
		"count":      strconv.Itoa(int(timeWindow.Count)),
		"start_date": timeWindow.GetStartTime().Format(time.RFC3339),
		"time_zone":  timeWindow.TimeZone.String(),
	})
}

type RollingTimeWindow struct {
	unitCount
}

func (r RollingTimeWindow) GetTimePeriod(timestamp time.Time) TimePeriod {
	return TimePeriod{
		Start: timestamp.Add(-r.Unit.Duration() * time.Duration(r.Count)),
		End:   timestamp,
	}
}

func (r RollingTimeWindow) String() string {
	return fmt.Sprintf("Rolling %s", r.unitCount)
}

func (r RollingTimeWindow) Type() TimeWindowTypeEnum {
	return Rolling
}

func (r RollingTimeWindow) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"type":  "rolling",
		"unit":  r.Unit.String(),
		"count": strconv.Itoa(int(r.Count)),
	})
}

// Duration returns time.Duration value of fixed duration TimeUnitEnum
func (tu TimeUnitEnum) Duration() time.Duration {
	switch tu {
	case Second:
		return time.Second
	case Minute:
		return time.Minute
	case Hour:
		return time.Hour
	case Day:
		return time.Hour * time.Duration(HoursInDay)
	case Week:
		return time.Hour * time.Duration(HoursInDay) * time.Duration(daysInWeek)
	default:
		return time.Duration(0)
	}
}

// NewCalendarTimeWindow creates new calendar time window
func NewCalendarTimeWindow(
	unit TimeUnitEnum,
	count uint32,
	timeZone *time.Location,
	dateWithTime DateWithTime,
) (TimeWindow, error) {
	if !containsTimeUnitEnum(calendarWindowTimeUnits, unit) {
		return nil, fmt.Errorf("unit '%s' is not valid for calendar time window", unit.String())
	}
	tw := CalendarTimeWindow{
		unitCount: unitCount{
			Unit:  unit,
			Count: count,
		},
		TimeZone:     timeZone,
		DateWithTime: dateWithTime,
	}
	return tw, nil
}

// NewRollingTimeWindow creates new rolling time window
func NewRollingTimeWindow(unit TimeUnitEnum, count uint32) (TimeWindow, error) {
	if !containsTimeUnitEnum(rollingWindowTimeUnits, unit) {
		return nil, fmt.Errorf("unit '%s' is not valid for rolling time window", unit.String())
	}
	tw := RollingTimeWindow{
		unitCount: unitCount{
			Unit:  unit,
			Count: count,
		},
	}
	return tw, nil
}

// TimeWindowTypeEnum represents enum for time window types
type TimeWindowTypeEnum int16

// Rolling is value of valid time windows
const (
	Rolling TimeWindowTypeEnum = iota + 1
	Calendar
)

func (s TimeWindowTypeEnum) String() string {
	switch s {
	case Rolling:
		return "Rolling"
	case Calendar:
		return "Calendar"
	}
	return ""
}

// TimeUnitEnum represents enum for time unit types
type TimeUnitEnum int16

// Second is value of valid time units
const (
	Second TimeUnitEnum = iota + 1
	Minute
	Hour
	Day
	Week
	Month
	Quarter
	Year
)

const (
	second  = "Second"
	minute  = "Minute"
	hour    = "Hour"
	day     = "Day"
	week    = "Week"
	month   = "Month"
	quarter = "Quarter"
	year    = "Year"
)

var (
	timeUnits = map[string]TimeUnitEnum{
		second:  Second,
		minute:  Minute,
		hour:    Hour,
		day:     Day,
		week:    Week,
		month:   Month,
		quarter: Quarter,
		year:    Year,
	}
	rollingWindowTimeUnits = map[string]TimeUnitEnum{
		minute: Minute,
		hour:   Hour,
		day:    Day,
	}
	calendarWindowTimeUnits = map[string]TimeUnitEnum{
		day:     Day,
		week:    Week,
		month:   Month,
		quarter: Quarter,
		year:    Year,
	}

	timeUnitsList               = []string{"Second", "Minute", "Hour", "Day", "Week", "Month", "Quarter", "Year"}
	rollingWindowTimeUnitsList  = []string{"Minute", "Hour", "Day"}
	calendarWindowTimeUnitsList = []string{"Day", "Week", "Month", "Quarter", "Year"}
)

// containsTimeUnitEnum checks if time unit is contained in a provided enum string map
func containsTimeUnitEnum(timeUnits map[string]TimeUnitEnum, timeUnit TimeUnitEnum) bool {
	for _, value := range timeUnits {
		if value == timeUnit {
			return true
		}
	}
	return false
}

func GetTimeUnitEnum(typ TimeWindowTypeEnum, timeUnit string) (timeUnitEnum TimeUnitEnum) {
	switch typ {
	case Rolling:
		timeUnitEnum = rollingWindowTimeUnits[timeUnit]
	case Calendar:
		timeUnitEnum = calendarWindowTimeUnits[timeUnit]
	}
	return
}

func IsTimeUnit(timeUnit string) bool {
	_, ok := timeUnits[timeUnit]
	return ok
}

func ValidateCalendarAlignedTimeUnit(timeUnit string) error {
	return rules.OneOf[string](calendarWindowTimeUnitsList...).Validate(timeUnit)
}

func ValidateRollingWindowTimeUnit(timeUnit string) error {
	return rules.OneOf[string](rollingWindowTimeUnitsList...).Validate(timeUnit)
}

func (tu TimeUnitEnum) String() string {
	for key, value := range timeUnits {
		if value == tu {
			return key
		}
	}
	return "UNKNOWN"
}

// MustParseTimeUnit parses passed time unit
func MustParseTimeUnit(timeUnit string) TimeUnitEnum {
	result, ok := timeUnits[timeUnit]
	if !ok {
		panic(fmt.Sprintf("'%s' is not valid time unit", timeUnit))
	}
	return result
}

func GetMinStartDate() time.Time {
	return time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
}

// ParseStartDate parses passed string containing start date
func ParseStartDate(startDateStr string) (time.Time, error) {
	startDate, err := time.Parse(IsoDateTimeOnlyLayout, startDateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing date: %w", err)
	}
	return startDate, nil
}

func ValidationRuleTimeUnit() govy.Rule[string] {
	return rules.OneOf[string](timeUnitsList...)
}
