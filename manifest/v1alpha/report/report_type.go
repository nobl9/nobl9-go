package report

//go:generate ../../../bin/go-enum  --values

// ReportType represents the specific type of report.
//
/* ENUM(
ResourceUsageSummary = 1
SLOHistory
ErrorBudgetStatus
ReliabilityRollup
SystemHealthReview
)*/
type ReportType int
