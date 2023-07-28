package v1alpha

import (
	"encoding/json"
	"time"

	"github.com/nobl9/nobl9-go/manifest"
)

type AlertSilencesSlice []AlertSilence

func (alertSilences AlertSilencesSlice) Clone() AlertSilencesSlice {
	clone := make([]AlertSilence, len(alertSilences))
	copy(clone, alertSilences)
	return clone
}

// AlertSilence represents alerts silencing configuration for given SLO and AlertPolicy.
type AlertSilence struct {
	manifest.ObjectInternal
	APIVersion string               `json:"apiVersion" validate:"required" example:"n9/v1alpha"`
	Kind       manifest.Kind        `json:"kind" validate:"required" example:"kind"`
	Metadata   AlertSilenceMetadata `json:"metadata"`
	Spec       AlertSilenceSpec     `json:"spec"`
	Status     AlertSilenceStatus   `json:"status,omitempty"`
}

func (a *AlertSilence) GetAPIVersion() string {
	return a.APIVersion
}

func (a *AlertSilence) GetKind() manifest.Kind {
	return a.Kind
}

func (a *AlertSilence) GetName() string {
	return a.Metadata.Name
}

func (a *AlertSilence) Validate() error {
	//TODO implement me
	panic("implement me")
}

func (a *AlertSilence) GetProject() string {
	return a.Metadata.Project
}

func (a *AlertSilence) SetProject(project string) {
	a.Metadata.Project = project
}

// AlertSilenceSpec represents content of AlertSilence's Spec.
type AlertSilenceSpec struct {
	Description string                        `json:"description" validate:"description"`
	Slo         string                        `json:"slo" validate:"required"`
	AlertPolicy AlertSilenceAlertPolicySource `json:"alertPolicy" validate:"required,dive"`
	Period      AlertSilencePeriod            `json:"period" validate:"required,dive"`
}

// AlertSilenceMetadata defines only basic metadata fields - name and project which uniquely identifies
// object on project level.
type AlertSilenceMetadata struct {
	Name    string `json:"name" validate:"required,objectName" example:"name"`
	Project string `json:"project,omitempty" validate:"objectName" example:"default"`
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

// genericToAlertSilence converts ObjectGeneric to AlertSilence
func genericToAlertSilence(o manifest.ObjectGeneric, v validator, onlyHeader bool) (AlertSilence, error) {
	res := AlertSilence{
		APIVersion: o.ObjectHeader.APIVersion,
		Kind:       o.ObjectHeader.Kind,
		Metadata: AlertSilenceMetadata{
			Name:    o.Metadata.Name,
			Project: o.Metadata.Project,
		},
		ObjectInternal: manifest.ObjectInternal{
			Organization: o.ObjectHeader.Organization,
			ManifestSrc:  o.ObjectHeader.ManifestSrc,
		},
	}
	if onlyHeader {
		return res, nil
	}
	var resSpec AlertSilenceSpec
	if err := json.Unmarshal(o.Spec, &resSpec); err != nil {
		err = manifest.EnhanceError(o, err)
		return res, err
	}
	res.Spec = resSpec
	if err := v.Check(res); err != nil {
		err = manifest.EnhanceError(o, err)
		return res, err
	}
	return res, nil
}
