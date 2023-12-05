package alertpolicy

import (
	"github.com/nobl9/nobl9-go/validation"
)

// Severity level describe importance of triggered alert
type Severity int16

const (
	SeverityLow Severity = iota + 1
	SeverityMedium
	SeverityHigh
)

const ErrorCodeSeverity validation.ErrorCode = "severity"

func getSeverityLevels() map[string]Severity {
	return map[string]Severity{
		"Low":    SeverityLow,
		"Medium": SeverityMedium,
		"High":   SeverityHigh,
	}
}

func (m Severity) String() string {
	for key, val := range getSeverityLevels() {
		if val == m {
			return key
		}
	}
	return "Unknown"
}

func SeverityValidation() validation.SingleRule[string] {
	return validation.OneOf(SeverityLow.String(), SeverityMedium.String(), SeverityHigh.String()).
		WithErrorCode(ErrorCodeSeverity)
}
