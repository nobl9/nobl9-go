package alertpolicy

import (
	"fmt"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// Severity level describe importance of triggered alert
type Severity int16

const (
	SeverityLow Severity = iota + 1
	SeverityMedium
	SeverityHigh
)

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

// ParseSeverity parses string to Severity
func ParseSeverity(value string) (Severity, error) {
	result, ok := getSeverityLevels()[value]
	if !ok {
		return result, fmt.Errorf("'%s' is not valid severity", value)
	}
	return result, nil
}

func severityValidation() govy.Rule[string] {
	return rules.OneOf(SeverityLow.String(), SeverityMedium.String(), SeverityHigh.String())
}
