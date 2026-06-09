package report

type SLOHistoryConfig struct {
	TimeFrame SLOHistoryTimeFrame `json:"timeFrame"`
}

type SLOHistoryTimeFrame struct {
	Rolling  *RollingTimeFrame  `json:"rolling,omitempty"`
	Calendar *CalendarTimeFrame `json:"calendar,omitempty"`
	TimeZone string             `json:"timeZone"`
}
