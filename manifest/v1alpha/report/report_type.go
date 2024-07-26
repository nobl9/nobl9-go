package report

//go:generate ../../../bin/go-enum  --values

// ReportType represents the specific type of [Report].
//
/* ENUM(
ResourceUsageSummary = 1
SLOHistory
ErrorBudgetStatus
ReliabilityRollup
SystemHealthReview
)*/
type ReportType int
