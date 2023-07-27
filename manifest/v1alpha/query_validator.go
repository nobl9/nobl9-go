package v1alpha

import (
	"time"
)

type QueryValidationStatus struct {
	ValidationStatus `json:"validationStatus"`
}

type ValidationStatus struct {
	GoodMetricValidation  *ValidationDetails `json:"goodMetric,omitempty"`
	BadMetricValidation   *ValidationDetails `json:"badMetric,omitempty"`
	TotalMetricValidation *ValidationDetails `json:"totalMetric,omitempty"`
	RawMetricValidation   *ValidationDetails `json:"rawMetric,omitempty"`
}

type ErrorDetails struct {
	Message          *string    `json:"message"`
	ValidationResult string     `json:"validationResult"`
	LogTimestamp     *time.Time `json:"logTimestamp"`
	HTTPStatusCode   *int       `json:"httpStatusCode"`
	Query            string     `json:"query"`
}

type ValidationDetails struct {
	*ErrorDetails
	*MetricSpec
}
