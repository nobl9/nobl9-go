package report

//go:generate ../../../bin/go-enum --nocase --lower --names --values

// ReliabilityScoreType identifies a reliability scoring mechanism used by a
// Reliability Rollup Report.
/* ENUM(
SLOTimeWindow
ReportTimeFrame
)*/
type ReliabilityScoreType string
