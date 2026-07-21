package report

//go:generate ../../../bin/go-enum --nocase --lower --names --values

// ReliabilityScoreType identifies a reliability scoring mechanism used by the
// Reliability Rollup Report.
/* ENUM(
SLOTimeWindow
ReportTimeFrame
)*/
type ReliabilityScoreType string
