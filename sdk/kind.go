package sdk

//go:generate ../bin/go-enum --nocomments --nocase

import "strings"

// Kind represents all the object kinds available in the API to perform operations on.
/* ENUM(
SLO
Service
Agent
AlertPolicy
AlertSilence
Alert
Project
AlertMethod
MetricSource
Direct
DataExport
UsageSummary
RoleBinding
SLOErrorBudgetStatus
Annotation
)*/
type Kind int

func (k Kind) ToLower() string {
	return strings.ToLower(k.String())
}
