package report

type SLOHistoryConfig struct {
	TimeFrame SLOHistoryTimeFrame `json:"timeFrame" validate:"required"`
}

type SLOHistoryTimeFrame struct {
	Rolling  *RollingTimeFrame  `json:"rolling,omitempty"`
	Calendar *CalendarTimeFrame `json:"calendar,omitempty"`
	TimeZone string             `json:"timeZone" validate:"required,timeZone" example:"America/New_York"`
}
