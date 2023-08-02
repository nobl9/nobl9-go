package v1alpha

import (
	"time"

	"github.com/nobl9/nobl9-go/manifest"
)

//go:generate go run ../../scripts/generate-object-impl.go AlertSilence

// AlertSilence represents alerts silencing configuration for given SLO and AlertPolicy.
type AlertSilence struct {
	APIVersion string               `json:"apiVersion"`
	Kind       manifest.Kind        `json:"kind"`
	Metadata   AlertSilenceMetadata `json:"metadata"`
	Spec       AlertSilenceSpec     `json:"spec"`
	Status     *AlertSilenceStatus  `json:"status,omitempty"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
}

// AlertSilenceMetadata defines only basic metadata fields - name and project which uniquely identifies
// object on project level.
type AlertSilenceMetadata struct {
	Name    string `json:"name" validate:"required,objectName" example:"name"`
	Project string `json:"project,omitempty" validate:"objectName" example:"default"`
}

// AlertSilenceSpec represents content of AlertSilence's Spec.
type AlertSilenceSpec struct {
	Description string                        `json:"description" validate:"description"`
	Slo         string                        `json:"slo" validate:"required"`
	AlertPolicy AlertSilenceAlertPolicySource `json:"alertPolicy" validate:"required,dive"`
	Period      AlertSilencePeriod            `json:"period" validate:"required,dive"`
}

func (a AlertSilenceSpec) GetParsedDuration() (time.Duration, error) {
	return time.ParseDuration(a.Period.Duration)
}

func (a AlertSilenceSpec) GetParsedStartTimeUTC() (time.Time, error) {
	if a.Period.StartTime == "" {
		return time.Time{}, nil
	}
	startTime, err := time.Parse(time.RFC3339, a.Period.StartTime)
	if err != nil {
		return time.Time{}, err
	}
	return startTime.UTC(), nil
}

func (a AlertSilenceSpec) GetParsedEndTimeUTC() (time.Time, error) {
	if a.Period.EndTime == "" {
		return time.Time{}, nil
	}
	endTime, err := time.Parse(time.RFC3339, a.Period.EndTime)
	if err != nil {
		return time.Time{}, err
	}
	return endTime.UTC(), nil
}

// AlertSilenceAlertPolicySource represents AlertPolicy attached to the SLO.
type AlertSilenceAlertPolicySource struct {
	Name    string `json:"name" validate:"required"`
	Project string `json:"project,omitempty"`
}

// AlertSilencePeriod represents time range configuration for AlertSilence.
type AlertSilencePeriod struct {
	StartTime string `json:"startTime,omitempty"`
	EndTime   string `json:"endTime,omitempty"`
	Duration  string `json:"duration,omitempty"`
}

// AlertSilenceStatus represents content of Status optional for AlertSilence object.
type AlertSilenceStatus struct {
	From      string `json:"from"`
	To        string `json:"to"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}
