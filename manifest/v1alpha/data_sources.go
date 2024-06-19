package v1alpha

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest"
)

//go:generate ../../bin/go-enum  --values --noprefix

// DataSourceType represents the type of data source, either Agent or Direct.
//
// Beware that order of these constants is very important
// existing integrations are saved in db with type = DataSourceType.
// New integrations always have to be added as last item in this list to get new "type id".
//
/* ENUM(
Prometheus = 1
Datadog
NewRelic
AppDynamics
Splunk
Lightstep
SplunkObservability
Dynatrace
ThousandEyes
Graphite
BigQuery
Elasticsearch
OpenTSDB
GrafanaLoki
CloudWatch
Pingdom
AmazonPrometheus
Redshift
SumoLogic
Instana
InfluxDB
GoogleCloudMonitoring
AzureMonitor
Generic
Honeycomb
LogicMonitor
AzurePrometheus
)*/
type DataSourceType int

// GCM aliases GoogleCloudMonitoring.
// Eventually we should solve this inconsistency between the enum name and its string representation.
const GCM = GoogleCloudMonitoring

// HistoricalDataRetrieval represents optional parameters for agent to regard when configuring
// TimeMachine-related SLO properties
type HistoricalDataRetrieval struct {
	MinimumAgentVersion string                      `json:"minimumAgentVersion,omitempty"`
	MaxDuration         HistoricalRetrievalDuration `json:"maxDuration" validate:"required"`
	DefaultDuration     HistoricalRetrievalDuration `json:"defaultDuration" validate:"required"`
}

func HistoricalDataRetrievalValidation() validation.Validator[HistoricalDataRetrieval] {
	return validation.New[HistoricalDataRetrieval](
		validation.For(validation.GetSelf[HistoricalDataRetrieval]()).
			Rules(defaultDataRetrievalDurationValidation),
		validation.For(func(h HistoricalDataRetrieval) HistoricalRetrievalDuration { return h.MaxDuration }).
			WithName("maxDuration").
			Required().
			Include(historicalRetrievalDurationValidation),
		validation.For(func(h HistoricalDataRetrieval) HistoricalRetrievalDuration { return h.DefaultDuration }).
			WithName("defaultDuration").
			Required().
			Include(historicalRetrievalDurationValidation),
	)
}

var historicalRetrievalDurationValidation = validation.New[HistoricalRetrievalDuration](
	validation.ForPointer(func(h HistoricalRetrievalDuration) *int { return h.Value }).
		WithName("value").
		Required().
		Rules(validation.GreaterThanOrEqualTo(0), validation.LessThanOrEqualTo(43200)),
	validation.For(func(h HistoricalRetrievalDuration) HistoricalRetrievalDurationUnit { return h.Unit }).
		WithName("unit").
		Required().
		Rules(validation.OneOf(HRDDay, HRDHour, HRDMinute)),
)

var defaultDataRetrievalDurationValidation = validation.NewSingleRule(
	func(dataRetrieval HistoricalDataRetrieval) error {
		if dataRetrieval.DefaultDuration.BiggerThan(dataRetrieval.MaxDuration) {
			var maxDurationValue int
			if dataRetrieval.MaxDuration.Value != nil {
				maxDurationValue = *dataRetrieval.MaxDuration.Value
			}
			return validation.NewPropertyError(
				"defaultDuration",
				dataRetrieval.DefaultDuration,
				errors.Errorf(
					"must be less than or equal to 'maxDuration' (%d %s)",
					maxDurationValue, dataRetrieval.MaxDuration.Unit))
		}
		return nil
	})

type Interval struct {
	Duration
}

type Jitter struct {
	Duration
}

type Timeout struct {
	Duration
}

type QueryDelay struct {
	MinimumAgentVersion string `json:"minimumAgentVersion,omitempty"`
	Duration
}

var maxQueryDelay = Duration{
	Value: func() *int { v := maxQueryDelayDuration; return &v }(),
	Unit:  maxQueryDelayDurationUnit,
}

func QueryDelayValidation() validation.Validator[QueryDelay] {
	return validation.New[QueryDelay](
		validation.For(func(q QueryDelay) Duration { return q.Duration }).
			Rules(validation.NewSingleRule(func(d Duration) error {
				if d.Duration() > maxQueryDelay.Duration() {
					return errors.Errorf("must be less than or equal to %s", maxQueryDelay)
				}
				return nil
			})),
		// Value's max and min are validated through [GetQueryDelayDefaults] and [maxQueryDelay].
		validation.ForPointer(func(q QueryDelay) *int { return q.Value }).
			WithName("value").
			Required(),
		validation.For(func(q QueryDelay) DurationUnit { return q.Unit }).
			WithName("unit").
			Required().
			Rules(validation.OneOf(Minute, Second)),
	)
}

// HistoricalRetrievalDuration struct was previously called Duration. However, this name was too generic
// since we also needed to introduce a Duration struct for QueryDelay, which allowed for different time units.
// Time travel is allowed for days/hours/minutes, and query delay can be set to minutes/seconds. Separating those two
// structs allows for easier validation logic and avoidance of possible mismatches. Also, later on the database level
// we have time travel duration unit related enum, that's specifically named for data retrieval purposes. Thus,
// it was easier to split those Durations into separate structures.
type HistoricalRetrievalDuration struct {
	Value *int                            `json:"value" validate:"required,min=0,max=43200"`
	Unit  HistoricalRetrievalDurationUnit `json:"unit" validate:"required"`
}

type HistoricalRetrievalDurationUnit string

const (
	HRDDay    HistoricalRetrievalDurationUnit = "Day"
	HRDHour   HistoricalRetrievalDurationUnit = "Hour"
	HRDMinute HistoricalRetrievalDurationUnit = "Minute"
)

const (
	maxQueryDelayDuration     = 1440
	maxQueryDelayDurationUnit = Minute
	secondAlias               = "S"
	minuteAlias               = "M"
)

const MinimalSupportedQueryDelayAgentVersion = "v0.65.0-beta09"

func (hrdu HistoricalRetrievalDurationUnit) IsValid() bool {
	return hrdu == HRDDay || hrdu == HRDHour || hrdu == HRDMinute
}

func (hrdu HistoricalRetrievalDurationUnit) String() string {
	switch hrdu {
	case HRDDay:
		return "Day"
	case HRDHour:
		return "Hour"
	case HRDMinute:
		return "Minute"
	}
	return ""
}

func HistoricalRetrievalDurationUnitFromString(unit string) (HistoricalRetrievalDurationUnit, error) {
	switch cases.Title(language.Und).String(unit) {
	case HRDDay.String():
		return HRDDay, nil
	case HRDHour.String():
		return HRDHour, nil
	case HRDMinute.String():
		return HRDMinute, nil
	}
	return "", errors.Errorf("'%s' is not a valid HistoricalRetrievalDurationUnit", unit)
}

func (d HistoricalRetrievalDuration) BiggerThan(b HistoricalRetrievalDuration) bool {
	return d.duration() > b.duration()
}

func (d HistoricalRetrievalDuration) IsZero() bool {
	return d.Value == nil || *d.Value == 0
}

func (d HistoricalRetrievalDuration) duration() time.Duration {
	if d.Value == nil {
		return time.Duration(0)
	}

	value := time.Duration(*d.Value)

	switch d.Unit {
	case HRDMinute:
		return value * time.Minute
	case HRDHour:
		return value * time.Hour
	case HRDDay:
		return value * time.Hour * 24
	}

	return time.Duration(0)
}

type Duration struct {
	Value *int         `json:"value" validate:"required,min=0,max=86400"`
	Unit  DurationUnit `json:"unit" validate:"required"`
}

type DurationUnit string

const (
	Millisecond DurationUnit = "Millisecond"
	Second      DurationUnit = "Second"
	Minute      DurationUnit = "Minute"
	Hour        DurationUnit = "Hour"
)

func (d Duration) String() string {
	if d.IsZero() {
		return "0s"
	}
	switch d.Unit {
	case Millisecond:
		return fmt.Sprintf("%dms", *d.Value)
	case Second:
		return fmt.Sprintf("%ds", *d.Value)
	case Minute:
		return fmt.Sprintf("%dm", *d.Value)
	case Hour:
		return fmt.Sprintf("%dh", *d.Value)
	default:
		return fmt.Sprintf("%ds", *d.Value)
	}
}

func (d Duration) LessThan(b Duration) bool {
	return d.Duration() < b.Duration()
}

func (d Duration) IsZero() bool {
	return d.Value == nil || *d.Value == 0
}

func (d Duration) Duration() time.Duration {
	if d.Value == nil {
		return time.Duration(0)
	}

	value := time.Duration(*d.Value)
	return value * d.Unit.Duration()
}

func DurationUnitFromString(unit string) (DurationUnit, error) {
	switch cases.Title(language.Und).String(unit) {
	case string(Minute), minuteAlias:
		return Minute, nil
	case string(Second), secondAlias:
		return Second, nil
	}
	return Second, errors.Errorf("'%s' is not a valid DurationUnit", unit)
}

func (d DurationUnit) Duration() time.Duration {
	switch d {
	case Millisecond:
		return time.Millisecond
	case Second:
		return time.Second
	case Minute:
		return time.Minute
	case Hour:
		return time.Hour
	}
	return time.Duration(0)
}

func (d DurationUnit) String() string {
	return string(d)
}

var agentDataRetrievalMaxDuration = map[DataSourceType]HistoricalRetrievalDuration{
	Datadog:               {Value: ptr(30), Unit: HRDDay},
	Prometheus:            {Value: ptr(30), Unit: HRDDay},
	AmazonPrometheus:      {Value: ptr(30), Unit: HRDDay},
	NewRelic:              {Value: ptr(30), Unit: HRDDay},
	Splunk:                {Value: ptr(30), Unit: HRDDay},
	Graphite:              {Value: ptr(30), Unit: HRDDay},
	Lightstep:             {Value: ptr(30), Unit: HRDDay},
	CloudWatch:            {Value: ptr(15), Unit: HRDDay},
	Dynatrace:             {Value: ptr(28), Unit: HRDDay},
	AppDynamics:           {Value: ptr(30), Unit: HRDDay},
	AzureMonitor:          {Value: ptr(30), Unit: HRDDay},
	Honeycomb:             {Value: ptr(7), Unit: HRDDay},
	GoogleCloudMonitoring: {Value: ptr(30), Unit: HRDDay},
	AzurePrometheus:       {Value: ptr(30), Unit: HRDDay},
}

var directDataRetrievalMaxDuration = map[DataSourceType]HistoricalRetrievalDuration{
	Datadog:               {Value: ptr(30), Unit: HRDDay},
	NewRelic:              {Value: ptr(30), Unit: HRDDay},
	Splunk:                {Value: ptr(30), Unit: HRDDay},
	Lightstep:             {Value: ptr(30), Unit: HRDDay},
	CloudWatch:            {Value: ptr(15), Unit: HRDDay},
	Dynatrace:             {Value: ptr(28), Unit: HRDDay},
	AppDynamics:           {Value: ptr(30), Unit: HRDDay},
	AzureMonitor:          {Value: ptr(30), Unit: HRDDay},
	Honeycomb:             {Value: ptr(7), Unit: HRDDay},
	GoogleCloudMonitoring: {Value: ptr(30), Unit: HRDDay},
}

func GetDataRetrievalMaxDuration(kind manifest.Kind, typ DataSourceType) (HistoricalRetrievalDuration, error) {
	//nolint: exhaustive
	switch kind {
	case manifest.KindAgent:
		if hrd, ok := agentDataRetrievalMaxDuration[typ]; ok {
			return hrd, nil
		}
	case manifest.KindDirect:
		if hrd, ok := directDataRetrievalMaxDuration[typ]; ok {
			return hrd, nil
		}
	}
	return HistoricalRetrievalDuration{},
		errors.Errorf("historical data retrieval is not supported for %s %s", typ, kind)
}

type QueryDelayDefaults map[DataSourceType]Duration

func (q QueryDelayDefaults) Get(at DataSourceType) string {
	return q[at].String()
}

// GetQueryDelayDefaults serves an exported, single source of truth map that is now a part of v1alpha contract.
// Its entries are used in two places: in one of internal endpoints serving Query Delay defaults,
// and in internal telegraf intake configuration, where it is passed to plugins as Query Delay defaults.
//
// WARNING: All string values of this map must satisfy the "customDuration" regex pattern.
func GetQueryDelayDefaults() QueryDelayDefaults {
	return QueryDelayDefaults{
		AmazonPrometheus: {
			Value: ptr(0),
			Unit:  Second,
		},
		Prometheus: {
			Value: ptr(0),
			Unit:  Second,
		},
		AppDynamics: {
			Value: ptr(1),
			Unit:  Minute,
		},
		BigQuery: {
			Value: ptr(0),
			Unit:  Second,
		},
		CloudWatch: {
			Value: ptr(1),
			Unit:  Minute,
		},
		Datadog: {
			Value: ptr(1),
			Unit:  Minute,
		},
		Dynatrace: {
			Value: ptr(2),
			Unit:  Minute,
		},
		Elasticsearch: {
			Value: ptr(1),
			Unit:  Minute,
		},
		GCM: {
			Value: ptr(2),
			Unit:  Minute,
		},
		GrafanaLoki: {
			Value: ptr(1),
			Unit:  Minute,
		},
		Graphite: {
			Value: ptr(1),
			Unit:  Minute,
		},
		InfluxDB: {
			Value: ptr(1),
			Unit:  Minute,
		},
		Instana: {
			Value: ptr(1),
			Unit:  Minute,
		},
		Lightstep: {
			Value: ptr(2),
			Unit:  Minute,
		},
		NewRelic: {
			Value: ptr(1),
			Unit:  Minute,
		},
		OpenTSDB: {
			Value: ptr(1),
			Unit:  Minute,
		},
		Pingdom: {
			Value: ptr(1),
			Unit:  Minute,
		},
		Redshift: {
			Value: ptr(30),
			Unit:  Second,
		},
		Splunk: {
			Value: ptr(5),
			Unit:  Minute,
		},
		SplunkObservability: {
			Value: ptr(5),
			Unit:  Minute,
		},
		SumoLogic: {
			Value: ptr(4),
			Unit:  Minute,
		},
		ThousandEyes: {
			Value: ptr(1),
			Unit:  Minute,
		},
		AzureMonitor: {
			Value: ptr(5),
			Unit:  Minute,
		},
		Generic: {
			Value: ptr(0),
			Unit:  Second,
		},
		Honeycomb: {
			Value: ptr(5),
			Unit:  Minute,
		},
		LogicMonitor: {
			Value: ptr(2),
			Unit:  Minute,
		},
		AzurePrometheus: {
			Value: ptr(0),
			Unit:  Second,
		},
	}
}

func DataDogSiteValidationRule() validation.SingleRule[string] {
	return validation.OneOf(
		"eu",
		"com",
		"datadoghq.com",
		"us3.datadoghq.com",
		"us5.datadoghq.com",
		"datadoghq.eu",
		"ddog-gov.com",
		"ap1.datadoghq.com")
}

func ptr[T interface{}](val T) *T {
	return &val
}
