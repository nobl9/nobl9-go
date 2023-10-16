package v1alpha

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/twindow"
)

type DataSourceType int

// Beware that order of these constants is very important
// existing integrations are saved in db with type = DataSourceType.
// New integrations always have to be added as last item in this list to get new "type id".
const (
	Prometheus DataSourceType = iota + 1
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
	GCM
	AzureMonitor
)

const DatasourceStableChannel = "stable"

// HistoricalDataRetrieval represents optional parameters for agent to regard when configuring
// TimeMachine-related SLO properties
type HistoricalDataRetrieval struct {
	MinimumAgentVersion string                      `json:"minimumAgentVersion,omitempty" example:"0.0.9"`
	MaxDuration         HistoricalRetrievalDuration `json:"maxDuration" validate:"required"`
	DefaultDuration     HistoricalRetrievalDuration `json:"defaultDuration" validate:"required"`
}

type QueryDelay struct {
	MinimumAgentVersion string `json:"minimumAgentVersion,omitempty" example:"0.0.9"`
	QueryDelayDuration
}

type SourceOf int

const (
	SourceOfServices SourceOf = iota + 1
	SourceOfMetrics
)

const (
	sourceOfServicesStr = "Services"
	sourceOfMetricsStr  = "Metrics"
)

func getSourceOfNames() map[string]SourceOf {
	return map[string]SourceOf{
		sourceOfServicesStr: SourceOfServices,
		sourceOfMetricsStr:  SourceOfMetrics,
	}
}

func MustParseSourceOf(sourceOf string) SourceOf {
	result, ok := getSourceOfNames()[sourceOf]
	if !ok {
		panic(fmt.Sprintf("'%s' is not valid source of", sourceOf))
	}
	return result
}

func SourceOfToStringSlice(isMetrics, isServices bool) []string {
	var sourceOf []string
	if isMetrics {
		sourceOf = append(sourceOf, sourceOfMetricsStr)
	}
	if isServices {
		sourceOf = append(sourceOf, sourceOfServicesStr)
	}
	return sourceOf
}

func IsValidSourceOf(sourceOf string) bool {
	_, ok := getSourceOfNames()[sourceOf]
	return ok
}

var agentTypeToName = map[DataSourceType]string{
	Prometheus:          "Prometheus",
	Datadog:             "Datadog",
	NewRelic:            "NewRelic",
	AppDynamics:         "AppDynamics",
	Splunk:              "Splunk",
	Lightstep:           "Lightstep",
	SplunkObservability: "SplunkObservability",
	Dynatrace:           "Dynatrace",
	Elasticsearch:       "Elasticsearch",
	ThousandEyes:        "ThousandEyes",
	Graphite:            "Graphite",
	BigQuery:            "BigQuery",
	OpenTSDB:            "OpenTSDB",
	GrafanaLoki:         "GrafanaLoki",
	CloudWatch:          "CloudWatch",
	Pingdom:             "Pingdom",
	AmazonPrometheus:    "AmazonPrometheus",
	Redshift:            "Redshift",
	SumoLogic:           "SumoLogic",
	Instana:             "Instana",
	InfluxDB:            "InfluxDB",
	GCM:                 "GoogleCloudMonitoring",
	AzureMonitor:        "AzureMonitor",
}

func (dst DataSourceType) String() string {
	if key, ok := agentTypeToName[dst]; ok {
		return key
	}
	//nolint: goconst
	return "Unknown"
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

type QueryDelayDuration struct {
	Value *int                 `json:"value" validate:"required,min=0,max=86400"`
	Unit  twindow.TimeUnitEnum `json:"unit" validate:"required"`
}

type QueryIntervalDuration struct {
	Value *int                 `json:"value" validate:"required,min=0,max=86400"`
	Unit  twindow.TimeUnitEnum `json:"unit" validate:"required"`
}

type CollectionJitterDuration struct {
	Value *int                 `json:"value" validate:"required,min=0,max=86400"`
	Unit  twindow.TimeUnitEnum `json:"unit" validate:"required"`
}

type TimeoutDuration struct {
	Value *int                 `json:"value" validate:"required,min=0,max=86400"`
	Unit  twindow.TimeUnitEnum `json:"unit" validate:"required"`
}

type HistoricalRetrievalDurationUnit string

const (
	HRDDay    HistoricalRetrievalDurationUnit = "Day"
	HRDHour   HistoricalRetrievalDurationUnit = "Hour"
	HRDMinute HistoricalRetrievalDurationUnit = "Minute"
)

const (
	maxQueryDelayDuration     = 1440
	maxQueryDelayDurationUnit = twindow.Minute
	SecondAlias               = "S"
	MinuteAlias               = "M"
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

func (qdd QueryDelayDuration) IsValid() bool {
	return isMinuteOrSecond(qdd.Unit)
}

func (qid QueryIntervalDuration) IsValid() bool {
	return isMinuteOrSecond(qid.Unit)
}

func (td TimeoutDuration) IsValid() bool {
	return td.Unit == twindow.Second
}

func (qdd QueryDelayDuration) String() string {
	return fmt.Sprintf("%d%s", *qdd.Value, formatTimeUnit(qdd.Unit))
}

func (qid QueryIntervalDuration) String() string {
	return fmt.Sprintf("%d%s", *qid.Value, formatTimeUnit(qid.Unit))
}

func (cjd CollectionJitterDuration) String() string {
	return fmt.Sprintf("%d%s", *cjd.Value, formatTimeUnit(cjd.Unit))
}

func (td TimeoutDuration) String() string {
	return fmt.Sprintf("%d%s", *td.Value, formatTimeUnit(td.Unit))
}

func (qdd QueryDelayDuration) BiggerThanMax() bool {
	maxQueryDelayDurationInt := maxQueryDelayDuration
	max := QueryDelayDuration{
		Value: &maxQueryDelayDurationInt,
		Unit:  maxQueryDelayDurationUnit,
	}
	return qdd.Duration() > max.Duration()
}

func (qdd QueryDelayDuration) LesserThan(b QueryDelayDuration) bool {
	return qdd.Duration() < b.Duration()
}

func (qdd QueryDelayDuration) IsZero() bool {
	return qdd.Value == nil || *qdd.Value == 0
}

func (qdd QueryDelayDuration) Duration() time.Duration {
	if qdd.Value == nil {
		return time.Duration(0)
	}

	value := time.Duration(*qdd.Value)
	return value * qdd.Duration()
}

func QueryDelayDurationUnitFromString(unit string) (twindow.TimeUnitEnum, error) {
	switch cases.Title(language.Und).String(unit) {
	case twindow.Minute.String(), MinuteAlias:
		return twindow.Minute, nil
	case twindow.Second.String(), SecondAlias:
		return twindow.Second, nil
	}
	return twindow.Second, errors.Errorf("'%s' is not a valid QueryDelayDurationUnit", unit)
}

func QueryIntervalDurationUnitFromString(unit string) (twindow.TimeUnitEnum, error) {
	switch cases.Title(language.Und).String(unit) {
	case twindow.Minute.String(), MinuteAlias:
		return twindow.Minute, nil
	case twindow.Second.String(), SecondAlias:
		return twindow.Second, nil
	}
	return twindow.Second, errors.Errorf("'%s' is not a valid QueryIntervalDurationUnit", unit)
}

func formatTimeUnit(unit twindow.TimeUnitEnum) string {
	switch unit {
	case twindow.Second:
		return "s"
	case twindow.Minute:
		return "m"
	case twindow.Hour:
		return "h"
	case twindow.Day:
		return "d"
	default:
		return "UNDEFINED"
	}
}

func isMinuteOrSecond(unit twindow.TimeUnitEnum) bool {
	return unit == twindow.Second || unit == twindow.Minute
}

var agentDataRetrievalMaxDuration = map[string]HistoricalRetrievalDuration{
	Datadog.String():          {Value: ptr(30), Unit: HRDDay},
	Prometheus.String():       {Value: ptr(30), Unit: HRDDay},
	AmazonPrometheus.String(): {Value: ptr(30), Unit: HRDDay},
	NewRelic.String():         {Value: ptr(30), Unit: HRDDay},
	Splunk.String():           {Value: ptr(30), Unit: HRDDay},
	Graphite.String():         {Value: ptr(30), Unit: HRDDay},
	Lightstep.String():        {Value: ptr(30), Unit: HRDDay},
	CloudWatch.String():       {Value: ptr(15), Unit: HRDDay},
	Dynatrace.String():        {Value: ptr(28), Unit: HRDDay},
	AppDynamics.String():      {Value: ptr(30), Unit: HRDDay},
	AzureMonitor.String():     {Value: ptr(30), Unit: HRDDay},
}

var directDataRetrievalMaxDuration = map[string]HistoricalRetrievalDuration{
	Datadog.String():      {Value: ptr(30), Unit: HRDDay},
	NewRelic.String():     {Value: ptr(30), Unit: HRDDay},
	Splunk.String():       {Value: ptr(30), Unit: HRDDay},
	Lightstep.String():    {Value: ptr(30), Unit: HRDDay},
	CloudWatch.String():   {Value: ptr(15), Unit: HRDDay},
	Dynatrace.String():    {Value: ptr(28), Unit: HRDDay},
	AppDynamics.String():  {Value: ptr(30), Unit: HRDDay},
	AzureMonitor.String(): {Value: ptr(30), Unit: HRDDay},
}

func GetDataRetrievalMaxDuration(kind manifest.Kind, typeName string) (HistoricalRetrievalDuration, error) {
	//nolint: exhaustive
	switch kind {
	case manifest.KindAgent:
		if hrd, ok := agentDataRetrievalMaxDuration[typeName]; ok {
			return hrd, nil
		}
	case manifest.KindDirect:
		if hrd, ok := directDataRetrievalMaxDuration[typeName]; ok {
			return hrd, nil
		}
	}
	return HistoricalRetrievalDuration{},
		errors.Errorf("historical data retrieval is not supported for %s %s", typeName, kind)
}

type QueryDelayDefaults map[string]QueryDelayDuration

func (q QueryDelayDefaults) GetByName(name string) string {
	return q[name].String()
}

func (q QueryDelayDefaults) GetByType(at DataSourceType) string {
	return q[at.String()].String()
}

// GetQueryDelayDefaults serves an exported, single source of truth map that is now a part of v1alpha contract.
// Its entries are used in two places: in one of internal endpoints serving Query Delay defaults,
// and in internal telegraf intake configuration, where it is passed to plugins as Query Delay defaults.
//
// WARNING: All string values of this map must satisfy the "customDuration" regex pattern.
func GetQueryDelayDefaults() QueryDelayDefaults {
	return QueryDelayDefaults{
		AmazonPrometheus.String(): {
			Value: ptr(0),
			Unit:  twindow.Second,
		},
		Prometheus.String(): {
			Value: ptr(0),
			Unit:  twindow.Second,
		},
		AppDynamics.String(): {
			Value: ptr(1),
			Unit:  twindow.Minute,
		},
		BigQuery.String(): {
			Value: ptr(0),
			Unit:  twindow.Second,
		},
		CloudWatch.String(): {
			Value: ptr(1),
			Unit:  twindow.Minute,
		},
		Datadog.String(): {
			Value: ptr(1),
			Unit:  twindow.Minute,
		},
		Dynatrace.String(): {
			Value: ptr(2),
			Unit:  twindow.Minute,
		},
		Elasticsearch.String(): {
			Value: ptr(1),
			Unit:  twindow.Minute,
		},
		GCM.String(): {
			Value: ptr(2),
			Unit:  twindow.Minute,
		},
		GrafanaLoki.String(): {
			Value: ptr(1),
			Unit:  twindow.Minute,
		},
		Graphite.String(): {
			Value: ptr(1),
			Unit:  twindow.Minute,
		},
		InfluxDB.String(): {
			Value: ptr(1),
			Unit:  twindow.Minute,
		},
		Instana.String(): {
			Value: ptr(1),
			Unit:  twindow.Minute,
		},
		Lightstep.String(): {
			Value: ptr(2),
			Unit:  twindow.Minute,
		},
		NewRelic.String(): {
			Value: ptr(1),
			Unit:  twindow.Minute,
		},
		OpenTSDB.String(): {
			Value: ptr(1),
			Unit:  twindow.Minute,
		},
		Pingdom.String(): {
			Value: ptr(1),
			Unit:  twindow.Minute,
		},
		Redshift.String(): {
			Value: ptr(30),
			Unit:  twindow.Second,
		},
		Splunk.String(): {
			Value: ptr(5),
			Unit:  twindow.Minute,
		},
		SplunkObservability.String(): {
			Value: ptr(5),
			Unit:  twindow.Minute,
		},
		SumoLogic.String(): {
			Value: ptr(4),
			Unit:  twindow.Minute,
		},
		ThousandEyes.String(): {
			Value: ptr(1),
			Unit:  twindow.Minute,
		},
		AzureMonitor.String(): {
			Value: ptr(5),
			Unit:  twindow.Minute,
		},
	}
}

func ptr[T interface{}](val T) *T {
	return &val
}
