// Code generated by go-enum DO NOT EDIT.
// Version: 0.6.0
// Revision: 919e61c0174b91303753ee3898569a01abb32c97
// Build Date: 2023-12-18T15:54:43Z
// Built By: goreleaser

package report

import (
	"errors"
	"fmt"
)

const (
	// ReportTypeResourceUsageSummary is a ReportType of type ResourceUsageSummary.
	ReportTypeResourceUsageSummary ReportType = iota + 1
	// ReportTypeSLOHistory is a ReportType of type SLOHistory.
	ReportTypeSLOHistory
	// ReportTypeErrorBudgetStatus is a ReportType of type ErrorBudgetStatus.
	ReportTypeErrorBudgetStatus
	// ReportTypeReliabilityRollup is a ReportType of type ReliabilityRollup.
	ReportTypeReliabilityRollup
	// ReportTypeSystemHealthReview is a ReportType of type SystemHealthReview.
	ReportTypeSystemHealthReview
)

var ErrInvalidReportType = errors.New("not a valid ReportType")

const _ReportTypeName = "ResourceUsageSummarySLOHistoryErrorBudgetStatusReliabilityRollupSystemHealthReview"

// ReportTypeValues returns a list of the values for ReportType
func ReportTypeValues() []ReportType {
	return []ReportType{
		ReportTypeResourceUsageSummary,
		ReportTypeSLOHistory,
		ReportTypeErrorBudgetStatus,
		ReportTypeReliabilityRollup,
		ReportTypeSystemHealthReview,
	}
}

var _ReportTypeMap = map[ReportType]string{
	ReportTypeResourceUsageSummary: _ReportTypeName[0:20],
	ReportTypeSLOHistory:           _ReportTypeName[20:30],
	ReportTypeErrorBudgetStatus:    _ReportTypeName[30:47],
	ReportTypeReliabilityRollup:    _ReportTypeName[47:64],
	ReportTypeSystemHealthReview:   _ReportTypeName[64:82],
}

// String implements the Stringer interface.
func (x ReportType) String() string {
	if str, ok := _ReportTypeMap[x]; ok {
		return str
	}
	return fmt.Sprintf("ReportType(%d)", x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x ReportType) IsValid() bool {
	_, ok := _ReportTypeMap[x]
	return ok
}

var _ReportTypeValue = map[string]ReportType{
	_ReportTypeName[0:20]:  ReportTypeResourceUsageSummary,
	_ReportTypeName[20:30]: ReportTypeSLOHistory,
	_ReportTypeName[30:47]: ReportTypeErrorBudgetStatus,
	_ReportTypeName[47:64]: ReportTypeReliabilityRollup,
	_ReportTypeName[64:82]: ReportTypeSystemHealthReview,
}

// ParseReportType attempts to convert a string to a ReportType.
func ParseReportType(name string) (ReportType, error) {
	if x, ok := _ReportTypeValue[name]; ok {
		return x, nil
	}
	return ReportType(0), fmt.Errorf("%s is %w", name, ErrInvalidReportType)
}