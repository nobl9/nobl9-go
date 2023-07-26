package twindow

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetTimePeriod(t *testing.T) {
	t.Parallel()

	t.Run("rolling time window", func(t *testing.T) {
		t.Parallel()
		testCases := []struct {
			name          string
			unit          TimeUnitEnum
			count         int
			timestamp     string
			expectedBegin string
			expectedEnd   string
		}{
			{
				name:          "1 day",
				unit:          Day,
				count:         1,
				timestamp:     "2006-01-04T15:04:05Z",
				expectedBegin: "2006-01-03T15:04:05Z",
				expectedEnd:   "2006-01-04T15:04:05Z",
			},
			{
				name:          "7 days",
				unit:          Day,
				count:         7,
				timestamp:     "2006-01-04T15:04:05Z",
				expectedBegin: "2005-12-28T15:04:05Z",
				expectedEnd:   "2006-01-04T15:04:05Z",
			},
			{
				name:          "1 hour",
				unit:          Hour,
				count:         1,
				timestamp:     "2006-01-04T15:04:05Z",
				expectedBegin: "2006-01-04T14:04:05Z",
				expectedEnd:   "2006-01-04T15:04:05Z",
			},
			{
				name:          "24 hours",
				unit:          Hour,
				count:         24,
				timestamp:     "2006-01-04T15:04:05Z",
				expectedBegin: "2006-01-03T15:04:05Z",
				expectedEnd:   "2006-01-04T15:04:05Z",
			},
			{
				name:          "60 minutes",
				unit:          Minute,
				count:         60,
				timestamp:     "2006-01-04T15:04:05Z",
				expectedBegin: "2006-01-04T14:04:05Z",
				expectedEnd:   "2006-01-04T15:04:05Z",
			},
		}

		for _, tc := range testCases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				rollingWindow, err := NewRollingTimeWindow(tc.unit, uint32(tc.count))
				if err != nil {
					assert.FailNow(t, "could not create new rolling time window")
				}

				timestamp, _ := time.Parse(time.RFC3339, tc.timestamp)
				expectedBegin, _ := time.Parse(time.RFC3339, tc.expectedBegin)
				expectedEnd, _ := time.Parse(time.RFC3339, tc.expectedEnd)

				period := rollingWindow.GetTimePeriod(timestamp)

				assert.Equal(t, expectedBegin.UTC(), period.Start.UTC(), "wrong time period start date")
				assert.Equal(t, expectedEnd.UTC(), period.End.UTC(), "wrong time period end date")

				assert.False(t, period.Start.After(timestamp))
				assert.Equal(
					t,
					period.End.UTC().String(),
					timestamp.UTC().String(),
					"timestamp must define end of the period",
				)
			})
		}
	})

	t.Run("calendar time window", func(t *testing.T) {
		t.Parallel()
		testCases := []struct {
			name          string
			unit          TimeUnitEnum
			count         int
			startTime     string
			timeZone      string
			timestamp     string
			expectedBegin string
			expectedEnd   string
		}{
			{
				name:          "1 year, after threshold",
				unit:          Year,
				count:         1,
				startTime:     "2004-01-02 12:00:00",
				timeZone:      "America/New_York",
				timestamp:     "2006-01-04T15:04:05Z",
				expectedBegin: "2006-01-02T17:00:00Z",
				expectedEnd:   "2007-01-02T17:00:00Z",
			},
			{
				name:          "1 year, before threshold",
				unit:          Year,
				count:         1,
				startTime:     "2004-01-02 12:00:00",
				timeZone:      "America/New_York",
				timestamp:     "2006-01-02T11:04:05Z",
				expectedBegin: "2005-01-02T17:00:00Z",
				expectedEnd:   "2006-01-02T17:00:00Z",
			},
			{
				name:          "1 year, right before start time",
				unit:          Year,
				count:         1,
				startTime:     "2004-01-02 12:00:00",
				timeZone:      "America/New_York",
				timestamp:     "2004-01-02T11:04:05Z",
				expectedBegin: "2003-01-02T17:00:00Z",
				expectedEnd:   "2004-01-02T17:00:00Z",
			},
			{
				name:          "1 year, year before start time",
				unit:          Year,
				count:         1,
				startTime:     "2004-01-02 12:00:00",
				timeZone:      "America/New_York",
				timestamp:     "2003-01-02T11:04:05Z",
				expectedBegin: "2002-01-02T17:00:00Z",
				expectedEnd:   "2003-01-02T17:00:00Z",
			},
			{
				name:          "3 years, after threshold",
				unit:          Year,
				count:         3,
				startTime:     "2004-01-02 12:00:00",
				timeZone:      "America/New_York",
				timestamp:     "2006-01-04T15:04:05Z",
				expectedBegin: "2004-01-02T17:00:00Z",
				expectedEnd:   "2007-01-02T17:00:00Z",
			},
			{
				name:          "1 quarter, after threshold",
				unit:          Quarter,
				count:         1,
				startTime:     "2004-01-02 12:00:00",
				timeZone:      "America/New_York",
				timestamp:     "2006-01-04T15:04:05Z",
				expectedBegin: "2006-01-02T17:00:00Z",
				expectedEnd:   "2006-04-02T16:00:00Z",
			},
			{
				name:          "1 quarter, before threshold",
				unit:          Quarter,
				count:         1,
				startTime:     "2004-01-02 12:00:00",
				timeZone:      "America/New_York",
				timestamp:     "2006-01-02T11:04:05Z",
				expectedBegin: "2005-10-02T16:00:00Z",
				expectedEnd:   "2006-01-02T17:00:00Z",
			},
			{
				name:          "1 quarter, right before start time",
				unit:          Quarter,
				count:         1,
				startTime:     "2004-01-02 12:00:00",
				timeZone:      "America/New_York",
				timestamp:     "2004-01-02T11:04:05Z",
				expectedBegin: "2003-10-02T16:00:00Z",
				expectedEnd:   "2004-01-02T17:00:00Z",
			},
			{
				name:          "3 quarters, after threshold",
				unit:          Quarter,
				count:         3,
				startTime:     "2004-01-02 12:00:00",
				timeZone:      "America/New_York",
				timestamp:     "2006-01-04T15:04:05Z",
				expectedBegin: "2005-07-02T16:00:00Z",
				expectedEnd:   "2006-04-02T16:00:00Z",
			},
			{
				name:          "1 month, after threshold",
				unit:          Month,
				count:         1,
				startTime:     "2004-01-02 12:00:00",
				timeZone:      "America/New_York",
				timestamp:     "2006-01-04T15:04:05Z",
				expectedBegin: "2006-01-02T17:00:00Z",
				expectedEnd:   "2006-02-02T17:00:00Z",
			},
			{
				name:          "1 month, before threshold",
				unit:          Month,
				count:         1,
				startTime:     "2004-01-02 12:00:00",
				timeZone:      "America/New_York",
				timestamp:     "2006-01-02T11:04:05Z",
				expectedBegin: "2005-12-02T17:00:00Z",
				expectedEnd:   "2006-01-02T17:00:00Z",
			},
			{
				name:          "1 month, right before start time",
				unit:          Month,
				count:         1,
				startTime:     "2004-01-02 12:00:00",
				timeZone:      "America/New_York",
				timestamp:     "2004-01-02T11:04:05Z",
				expectedBegin: "2003-12-02T17:00:00Z",
				expectedEnd:   "2004-01-02T17:00:00Z",
			},
			{
				name:          "1 month, month before start time",
				unit:          Month,
				count:         1,
				startTime:     "2004-01-02 12:00:00",
				timeZone:      "America/New_York",
				timestamp:     "2003-12-02T16:55:05Z",
				expectedBegin: "2003-11-02T17:00:00Z",
				expectedEnd:   "2003-12-02T17:00:00Z",
			},
			{
				name:          "2 month, after threshold",
				unit:          Month,
				count:         2,
				startTime:     "2004-01-02 12:00:00",
				timeZone:      "America/New_York",
				timestamp:     "2006-01-04T15:04:05Z",
				expectedBegin: "2006-01-02T17:00:00Z",
				expectedEnd:   "2006-03-02T17:00:00Z",
			},
			{
				name:          "1 week, after threshold",
				unit:          Week,
				count:         1,
				startTime:     "2004-01-02 12:00:00",
				timeZone:      "America/New_York",
				timestamp:     "2006-01-04T15:04:05Z",
				expectedBegin: "2005-12-30T17:00:00Z",
				expectedEnd:   "2006-01-06T17:00:00Z",
			},
			{
				name:          "1 week, before threshold",
				unit:          Week,
				count:         1,
				startTime:     "2004-01-02 12:00:00",
				timeZone:      "America/New_York",
				timestamp:     "2006-01-02T11:04:05Z",
				expectedBegin: "2005-12-30T17:00:00Z",
				expectedEnd:   "2006-01-06T17:00:00Z",
			},
			{
				name:          "1 week, before start time",
				unit:          Week,
				count:         1,
				startTime:     "2004-01-02 12:00:00",
				timeZone:      "America/New_York",
				timestamp:     "2004-01-02T11:04:05Z",
				expectedBegin: "2003-12-26T17:00:00Z",
				expectedEnd:   "2004-01-02T17:00:00Z",
			},
			{
				name:          "3 weeks, after threshold",
				unit:          Week,
				count:         3,
				startTime:     "2004-01-02 12:00:00",
				timeZone:      "America/New_York",
				timestamp:     "2004-01-04T15:04:05Z",
				expectedBegin: "2004-01-02T17:00:00Z",
				expectedEnd:   "2004-01-23T17:00:00Z",
			},

			{
				name:          "1 day, after threshold",
				unit:          Day,
				count:         1,
				startTime:     "2004-01-02 12:00:00",
				timeZone:      "America/New_York",
				timestamp:     "2006-01-02T17:05:05Z",
				expectedBegin: "2006-01-02T17:00:00Z",
				expectedEnd:   "2006-01-03T17:00:00Z",
			},
			{
				name:          "1 day, before threshold",
				unit:          Day,
				count:         1,
				startTime:     "2004-01-02 12:00:00",
				timeZone:      "America/New_York",
				timestamp:     "2006-01-02T16:55:05Z",
				expectedBegin: "2006-01-01T17:00:00Z",
				expectedEnd:   "2006-01-02T17:00:00Z",
			},
			{
				name:          "1 day, right before start time",
				unit:          Day,
				count:         1,
				startTime:     "2004-01-02 12:00:00",
				timeZone:      "America/New_York",
				timestamp:     "2004-01-02T16:55:05Z",
				expectedBegin: "2004-01-01T17:00:00Z",
				expectedEnd:   "2004-01-02T17:00:00Z",
			},
			{
				name:          "1 day, day before start time",
				unit:          Day,
				count:         1,
				startTime:     "2004-01-02 12:00:00",
				timeZone:      "America/New_York",
				timestamp:     "2004-01-01T16:55:05Z",
				expectedBegin: "2003-12-31T17:00:00Z",
				expectedEnd:   "2004-01-01T17:00:00Z",
			},

			{
				name:          "1 day, timestamp before start time on same day",
				unit:          Day,
				count:         1,
				startTime:     "2004-01-02 17:00:00",
				timeZone:      "UTC",
				timestamp:     "2004-01-02T16:55:00Z",
				expectedBegin: "2004-01-01T17:00:00Z",
				expectedEnd:   "2004-01-02T17:00:00Z",
			},
			{
				name:          "1 day, timestamp on day before start date but after start clock time",
				unit:          Day,
				count:         1,
				startTime:     "2004-01-02 17:00:00",
				timeZone:      "UTC",
				timestamp:     "2004-01-01T23:55:05Z",
				expectedBegin: "2004-01-01T17:00:00Z",
				expectedEnd:   "2004-01-02T17:00:00Z",
			},
			{
				name:          "1 day, timestamp on day before start date but before start clock time",
				unit:          Day,
				count:         1,
				startTime:     "2004-01-02 17:00:00",
				timeZone:      "UTC",
				timestamp:     "2004-01-01T14:00:00Z",
				expectedBegin: "2003-12-31T17:00:00Z",
				expectedEnd:   "2004-01-01T17:00:00Z",
			},
			{
				name:          "2 days, timestamp after start time, at start of the second period",
				unit:          Day,
				count:         2,
				startTime:     "2004-01-02 17:00:00",
				timeZone:      "UTC",
				timestamp:     "2004-01-04T17:00:00Z",
				expectedBegin: "2004-01-04T17:00:00Z",
				expectedEnd:   "2004-01-06T17:00:00Z",
			},
			{
				name:          "2 days, timestamp 2 days after start time, at start clock time",
				unit:          Day,
				count:         2,
				startTime:     "2004-01-02 17:00:00",
				timeZone:      "UTC",
				timestamp:     "2004-01-04T17:00:00Z",
				expectedBegin: "2004-01-04T17:00:00Z",
				expectedEnd:   "2004-01-06T17:00:00Z",
			},
			{
				name:          "2 days, timestamp 2 days after start time, before start clock time",
				unit:          Day,
				count:         2,
				startTime:     "2004-01-02 17:00:00",
				timeZone:      "UTC",
				timestamp:     "2004-01-04T16:55:00Z",
				expectedBegin: "2004-01-02T17:00:00Z",
				expectedEnd:   "2004-01-04T17:00:00Z",
			},
			{
				name:          "2 days, timestamp 1 day after start date, after start clock time",
				unit:          Day,
				count:         2,
				startTime:     "2004-01-02 17:00:00",
				timeZone:      "UTC",
				timestamp:     "2004-01-03T17:05:00Z",
				expectedBegin: "2004-01-02T17:00:00Z",
				expectedEnd:   "2004-01-04T17:00:00Z",
			},
			{
				name:          "2 days, timestamp 1 day after start date, at start clock time",
				unit:          Day,
				count:         2,
				startTime:     "2004-01-02 17:00:00",
				timeZone:      "UTC",
				timestamp:     "2004-01-03T17:00:00Z",
				expectedBegin: "2004-01-02T17:00:00Z",
				expectedEnd:   "2004-01-04T17:00:00Z",
			},
			{
				name:          "2 days, timestamp 1 day after start date, before clock time",
				unit:          Day,
				count:         2,
				startTime:     "2004-01-02 17:00:00",
				timeZone:      "UTC",
				timestamp:     "2004-01-03T16:55:00Z",
				expectedBegin: "2004-01-02T17:00:00Z",
				expectedEnd:   "2004-01-04T17:00:00Z",
			},
			{
				name:          "2 days, timestamp on start date, after start time",
				unit:          Day,
				count:         2,
				startTime:     "2004-01-02 17:00:00",
				timeZone:      "UTC",
				timestamp:     "2004-01-02T17:05:00Z",
				expectedBegin: "2004-01-02T17:00:00Z",
				expectedEnd:   "2004-01-04T17:00:00Z",
			},
			{
				name:          "2 days, timestamp at start date",
				unit:          Day,
				count:         2,
				startTime:     "2004-01-02 17:00:00",
				timeZone:      "UTC",
				timestamp:     "2004-01-02T17:00:00Z",
				expectedBegin: "2004-01-02T17:00:00Z",
				expectedEnd:   "2004-01-04T17:00:00Z",
			},
			{
				name:          "2 days, timestamp before start time, at start date",
				unit:          Day,
				count:         2,
				startTime:     "2004-01-02 17:00:00",
				timeZone:      "UTC",
				timestamp:     "2004-01-02T16:55:00Z",
				expectedBegin: "2003-12-31T17:00:00Z",
				expectedEnd:   "2004-01-02T17:00:00Z",
			},
			{
				name:          "2 days, timestamp 1 day before start time, after start clock time",
				unit:          Day,
				count:         2,
				startTime:     "2004-01-02 17:00:00",
				timeZone:      "UTC",
				timestamp:     "2004-01-01T17:05:00Z",
				expectedBegin: "2003-12-31T17:00:00Z",
				expectedEnd:   "2004-01-02T17:00:00Z",
			},
			{
				name:          "2 days, timestamp 1 day before start time, at start clock time",
				unit:          Day,
				count:         2,
				startTime:     "2004-01-02 17:00:00",
				timeZone:      "UTC",
				timestamp:     "2004-01-01T17:00:00Z",
				expectedBegin: "2003-12-31T17:00:00Z",
				expectedEnd:   "2004-01-02T17:00:00Z",
			},
			{
				name:          "2 days, timestamp 1 day before start time, before clock time",
				unit:          Day,
				count:         2,
				startTime:     "2004-01-02 17:00:00",
				timeZone:      "UTC",
				timestamp:     "2004-01-01T16:55:00Z",
				expectedBegin: "2003-12-31T17:00:00Z",
				expectedEnd:   "2004-01-02T17:00:00Z",
			},
			{
				name:          "2 days, timestamp 2 days before start date, after start clock time",
				unit:          Day,
				count:         2,
				startTime:     "2004-01-02 17:00:00",
				timeZone:      "UTC",
				timestamp:     "2003-12-31T17:05:00Z",
				expectedBegin: "2003-12-31T17:00:00Z",
				expectedEnd:   "2004-01-02T17:00:00Z",
			},
			{
				name:          "2 days, timestamp 2 days before start date, at start clock time",
				unit:          Day,
				count:         2,
				startTime:     "2004-01-02 17:00:00",
				timeZone:      "UTC",
				timestamp:     "2003-12-31T17:00:00Z",
				expectedBegin: "2003-12-31T17:00:00Z",
				expectedEnd:   "2004-01-02T17:00:00Z",
			},
			{
				name:          "2 days, timestamp 2 days before start date, at before clock time",
				unit:          Day,
				count:         2,
				startTime:     "2004-01-02 17:00:00",
				timeZone:      "UTC",
				timestamp:     "2003-12-31T16:55:00Z",
				expectedBegin: "2003-12-29T17:00:00Z",
				expectedEnd:   "2003-12-31T17:00:00Z",
			},
			{
				name:          "bug PC-1218",
				unit:          Day,
				count:         1,
				startTime:     "2020-11-04 11:10:00",
				timeZone:      "Europe/Warsaw",
				timestamp:     "2020-11-05T11:08:40+01:00",
				expectedBegin: "2020-11-04T11:10:00+01:00",
				expectedEnd:   "2020-11-05T11:10:00+01:00",
			},
			{
				name:          "task PC-1401",
				unit:          Day,
				count:         1,
				startTime:     "1939-09-03 5:45:00",
				timeZone:      "Europe/Warsaw",
				timestamp:     "1939-09-03T09:00:00+01:00",
				expectedBegin: "1939-09-03T05:45:00+01:00",
				expectedEnd:   "1939-09-04T05:45:00+01:00",
			},
			{
				name:          "story PC-1895",
				unit:          Week,
				count:         1,
				startTime:     "2021-06-06 00:00:00",
				timeZone:      "UTC",
				timestamp:     "2021-06-04T00:00:00.000Z",
				expectedBegin: "2021-05-30T00:00:00.000Z",
				expectedEnd:   "2021-06-06T00:00:00Z",
			},
			{
				name:          "timestamp has different dates in UTC and in time window time zone",
				unit:          Day,
				count:         3,
				startTime:     "2022-01-03 22:00:00",
				timeZone:      "America/Anchorage", // Alaska UTC-9
				timestamp:     "2021-12-23T01:00:00Z",
				expectedBegin: "2021-12-20T07:00:00Z",
				expectedEnd:   "2021-12-23T07:00:00Z",
			},
			{
				name:          "timestamp has same date in UTC and in time window time zone",
				unit:          Day,
				count:         3,
				startTime:     "2022-01-03 22:00:00",
				timeZone:      "America/Anchorage", // Alaska UTC-9
				timestamp:     "2021-12-23T06:00:00+05:00",
				expectedBegin: "2021-12-20T07:00:00Z",
				expectedEnd:   "2021-12-23T07:00:00Z",
			},
			{
				name:          "timestamp has earlier date in UTC than in time window time zone",
				unit:          Day,
				count:         3,
				startTime:     "2022-01-03 22:00:00",
				timeZone:      "America/Anchorage", // Alaska UTC-9
				timestamp:     "2021-12-22T22:00:00-08:00",
				expectedBegin: "2021-12-20T07:00:00Z",
				expectedEnd:   "2021-12-23T07:00:00Z",
			},
		}

		for _, tc := range testCases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				location, _ := time.LoadLocation(tc.timeZone)
				startTime, _ := time.ParseInLocation(IsoDateTimeOnlyLayout, tc.startTime, location)
				calendarWindow, _ := NewCalendarTimeWindow(tc.unit, uint32(tc.count), location, startTime)

				timestamp, _ := time.Parse(time.RFC3339, tc.timestamp)
				expectedBegin, _ := time.Parse(time.RFC3339, tc.expectedBegin)
				expectedEnd, _ := time.Parse(time.RFC3339, tc.expectedEnd)

				period := calendarWindow.GetTimePeriod(timestamp)

				assert.Equal(t, expectedBegin.UTC().String(), period.Start.UTC().String())
				assert.Equal(t, expectedEnd.UTC().String(), period.End.UTC().String())

				// calendar-aligned time window periods have inclusive Start and exclusive End
				assert.False(t, period.Start.After(timestamp))
				assert.True(t, period.End.After(timestamp))
			})
		}
	})
}
