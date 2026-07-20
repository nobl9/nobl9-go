package report

//go:generate ../../../bin/go-enum --nocase --lower --names --values

// ReliabilityScoreType specifies the time range used to calculate reliability scores.
/* ENUM(
SLOTimeWindow
ReportTimeFrame
)*/
type ReliabilityScoreType string
