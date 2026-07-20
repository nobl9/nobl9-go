package report

//go:generate ../../../bin/go-enum --nocase --lower --names --values

// ReliabilityScoreType identifies a type of Reliability Rollup Report.
/* ENUM(
SLOTimeWindow
ReportTimeFrame
)*/
type ReliabilityScoreType string
