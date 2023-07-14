// Package v1alpha represents objects available in API n9/v1alpha
package v1alpha

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/nobl9/nobl9-go/manifest"
)

// APIVersion is a value of valid apiVersions
const (
	APIVersion = "n9/v1alpha"
)

// HiddenValue can be used as a value of a secret field and is ignored during saving
const HiddenValue = "[hidden]"

// Kind groups Objects describing a specific type. Use this alias to increase code readability.
type Kind = string

// Possible values of field kind for valid Objects.
const (
	KindSLO          Kind = "SLO"
	KindService      Kind = "Service"
	KindAgent        Kind = "Agent"
	KindProject      Kind = "Project"
	KindAlertPolicy  Kind = "AlertPolicy"
	KindAlertSilence Kind = "AlertSilence"
	KindAlert        Kind = "Alert"
	KindAlertMethod  Kind = "AlertMethod"
	KindDirect       Kind = "Direct"
	KindDataExport   Kind = "DataExport"
	KindRoleBinding  Kind = "RoleBinding"
	KindAnnotation   Kind = "Annotation"
	KindUserGroup    Kind = "UserGroup"
)

const DatasourceStableChannel = "stable"

type AgentsSlice []Agent
type SLOsSlice []SLO
type ServicesSlice []Service
type AlertPoliciesSlice []AlertPolicy
type AlertSilencesSlice []AlertSilence
type AlertsSlice []Alert
type AlertMethodsSlice []AlertMethod
type DirectsSlice []Direct
type DataExportsSlice []DataExport
type ProjectsSlice []Project
type RoleBindingsSlice []RoleBinding
type AnnotationsSlice []Annotation
type UserGroupsSlice []UserGroup

func KindFromString(kindString string) Kind {
	for _, kind := range []Kind{
		KindSLO, KindService, KindAgent, KindProject, KindAlertPolicy, KindAlertSilence, KindAlert, KindAlertMethod,
		KindDirect, KindDataExport, KindRoleBinding, KindAnnotation,
	} {
		if strings.EqualFold(kindString, kind) {
			return kind
		}
	}
	return ""
}

func (agents AgentsSlice) Clone() AgentsSlice {
	clone := make([]Agent, len(agents))
	copy(clone, agents)
	return clone
}

func (slos SLOsSlice) Clone() SLOsSlice {
	clone := make([]SLO, len(slos))
	copy(clone, slos)
	return clone
}

func (services ServicesSlice) Clone() ServicesSlice {
	clone := make([]Service, len(services))
	copy(clone, services)
	return clone
}
func (alertPolicies AlertPoliciesSlice) Clone() AlertPoliciesSlice {
	clone := make([]AlertPolicy, len(alertPolicies))
	copy(clone, alertPolicies)
	return clone
}

func (alertSilences AlertSilencesSlice) Clone() AlertSilencesSlice {
	clone := make([]AlertSilence, len(alertSilences))
	copy(clone, alertSilences)
	return clone
}

func (alerts AlertsSlice) Clone() AlertsSlice {
	clone := make([]Alert, len(alerts))
	copy(clone, alerts)
	return clone
}

func (alertMethods AlertMethodsSlice) Clone() AlertMethodsSlice {
	clone := make([]AlertMethod, len(alertMethods))
	copy(clone, alertMethods)
	return clone
}

func (directs DirectsSlice) Clone() DirectsSlice {
	clone := make([]Direct, len(directs))
	copy(clone, directs)
	return clone
}

func (dataExports DataExportsSlice) Clone() DataExportsSlice {
	clone := make([]DataExport, len(dataExports))
	copy(clone, dataExports)
	return clone
}

func (projects ProjectsSlice) Clone() ProjectsSlice {
	clone := make([]Project, len(projects))
	copy(clone, projects)
	return clone
}

func (roleBindings RoleBindingsSlice) Clone() RoleBindingsSlice {
	clone := make([]RoleBinding, len(roleBindings))
	copy(clone, roleBindings)
	return clone
}

func (annotations AnnotationsSlice) Clone() AnnotationsSlice {
	clone := make([]Annotation, len(annotations))
	copy(clone, annotations)
	return clone
}

func (u UserGroupsSlice) Clone() UserGroupsSlice {
	clone := make([]UserGroup, len(u))
	copy(clone, u)
	return clone
}

// APIObjects - all Objects available for this version of API
// Sorted in order of applying
type APIObjects struct {
	SLOs          SLOsSlice          `json:"slos,omitempty"`
	Services      ServicesSlice      `json:"services,omitempty"`
	Agents        AgentsSlice        `json:"agents,omitempty"`
	AlertPolicies AlertPoliciesSlice `json:"alertpolicies,omitempty"`
	AlertSilences AlertSilencesSlice `json:"alertsilences,omitempty"`
	Alerts        AlertsSlice        `json:"alerts,omitempty"`
	AlertMethods  AlertMethodsSlice  `json:"alertmethods,omitempty"`
	Directs       DirectsSlice       `json:"directs,omitempty"`
	DataExports   DataExportsSlice   `json:"dataexports,omitempty"`
	Projects      ProjectsSlice      `json:"projects,omitempty"`
	RoleBindings  RoleBindingsSlice  `json:"rolebindings,omitempty"`
	Annotations   AnnotationsSlice   `json:"annotations,omitempty"`
	UserGroups    UserGroupsSlice    `json:"usergroups,omitempty"`
}

func (o APIObjects) Clone() APIObjects {
	return APIObjects{
		SLOs:          o.SLOs.Clone(),
		Services:      o.Services.Clone(),
		Agents:        o.Agents.Clone(),
		AlertPolicies: o.AlertPolicies.Clone(),
		AlertSilences: o.AlertSilences.Clone(),
		Alerts:        o.Alerts.Clone(),
		AlertMethods:  o.AlertMethods.Clone(),
		Directs:       o.Directs.Clone(),
		DataExports:   o.DataExports.Clone(),
		Projects:      o.Projects.Clone(),
		RoleBindings:  o.RoleBindings.Clone(),
		Annotations:   o.Annotations.Clone(),
	}
}

func (o APIObjects) Len() int {
	return len(o.SLOs) +
		len(o.Services) +
		len(o.Agents) +
		len(o.AlertPolicies) +
		len(o.AlertSilences) +
		len(o.Alerts) +
		len(o.AlertMethods) +
		len(o.Directs) +
		len(o.DataExports) +
		len(o.Projects) +
		len(o.RoleBindings) +
		len(o.Annotations)
}

// CountMetricsSpec represents set of two time series of good and total counts
type CountMetricsSpec struct {
	Incremental *bool       `json:"incremental" validate:"required"`
	GoodMetric  *MetricSpec `json:"good,omitempty"`
	BadMetric   *MetricSpec `json:"bad,omitempty"`
	TotalMetric *MetricSpec `json:"total" validate:"required"`
}

// RawMetricSpec represents integration with a metric source for a particular threshold
type RawMetricSpec struct {
	MetricQuery *MetricSpec `json:"query" validate:"required"`
}

// MetricSpec defines single time series obtained from data source
type MetricSpec struct {
	Prometheus          *PrometheusMetric          `json:"prometheus,omitempty"`
	Datadog             *DatadogMetric             `json:"datadog,omitempty"`
	NewRelic            *NewRelicMetric            `json:"newRelic,omitempty"`
	AppDynamics         *AppDynamicsMetric         `json:"appDynamics,omitempty"`
	Splunk              *SplunkMetric              `json:"splunk,omitempty"`
	Lightstep           *LightstepMetric           `json:"lightstep,omitempty"`
	SplunkObservability *SplunkObservabilityMetric `json:"splunkObservability,omitempty"`
	Dynatrace           *DynatraceMetric           `json:"dynatrace,omitempty"`
	Elasticsearch       *ElasticsearchMetric       `json:"elasticsearch,omitempty"`
	ThousandEyes        *ThousandEyesMetric        `json:"thousandEyes,omitempty"`
	Graphite            *GraphiteMetric            `json:"graphite,omitempty"`
	BigQuery            *BigQueryMetric            `json:"bigQuery,omitempty"`
	OpenTSDB            *OpenTSDBMetric            `json:"opentsdb,omitempty"`
	GrafanaLoki         *GrafanaLokiMetric         `json:"grafanaLoki,omitempty"`
	CloudWatch          *CloudWatchMetric          `json:"cloudWatch,omitempty"`
	Pingdom             *PingdomMetric             `json:"pingdom,omitempty"`
	AmazonPrometheus    *AmazonPrometheusMetric    `json:"amazonPrometheus,omitempty"`
	Redshift            *RedshiftMetric            `json:"redshift,omitempty"`
	SumoLogic           *SumoLogicMetric           `json:"sumoLogic,omitempty"`
	Instana             *InstanaMetric             `json:"instana,omitempty"`
	InfluxDB            *InfluxDBMetric            `json:"influxdb,omitempty"`
	GCM                 *GCMMetric                 `json:"gcm,omitempty"`
}

// PrometheusMetric represents metric from Prometheus
type PrometheusMetric struct {
	PromQL *string `json:"promql" validate:"required" example:"cpu_usage_user{cpu=\"cpu-total\"}"`
}

// AmazonPrometheusMetric represents metric from Amazon Managed Prometheus
type AmazonPrometheusMetric struct {
	PromQL *string `json:"promql" validate:"required" example:"cpu_usage_user{cpu=\"cpu-total\"}"`
}

// DatadogMetric represents metric from Datadog
type DatadogMetric struct {
	Query *string `json:"query" validate:"required"`
}

// NewRelicMetric represents metric from NewRelic
type NewRelicMetric struct {
	NRQL *string `json:"nrql" validate:"required,noSinceOrUntil"`
}

// ThousandEyesMetric represents metric from ThousandEyes
type ThousandEyesMetric struct {
	TestID   *int64  `json:"testID" validate:"required,gte=0"`
	TestType *string `json:"testType" validate:"supportedThousandEyesTestType"`
}

// AppDynamicsMetric represents metric from AppDynamics
type AppDynamicsMetric struct {
	ApplicationName *string `json:"applicationName" validate:"required,notEmpty"`
	MetricPath      *string `json:"metricPath" validate:"required,unambiguousAppDynamicMetricPath"`
}

// SplunkMetric represents metric from Splunk
type SplunkMetric struct {
	Query *string `json:"query" validate:"required,notEmpty,splunkQueryValid"`
}

// LightstepMetric represents metric from Lightstep
type LightstepMetric struct {
	StreamID   *string  `json:"streamId,omitempty"`
	TypeOfData *string  `json:"typeOfData" validate:"required,oneof=latency error_rate good total metric"`
	Percentile *float64 `json:"percentile,omitempty"`
	UQL        *string  `json:"uql,omitempty"`
}

// SplunkObservabilityMetric represents metric from SplunkObservability
type SplunkObservabilityMetric struct {
	Program *string `json:"program" validate:"required"`
}

// DynatraceMetric represents metric from Dynatrace.
type DynatraceMetric struct {
	MetricSelector *string `json:"metricSelector" validate:"required"`
}

// ElasticsearchMetric represents metric from Elasticsearch.
type ElasticsearchMetric struct {
	Index *string `json:"index" validate:"required"`
	Query *string `json:"query" validate:"required,elasticsearchBeginEndTimeRequired"`
}

// CloudWatchMetric represents metric from CloudWatch.
type CloudWatchMetric struct {
	Region     *string                     `json:"region" validate:"required,max=255"`
	Namespace  *string                     `json:"namespace,omitempty"`
	MetricName *string                     `json:"metricName,omitempty"`
	Stat       *string                     `json:"stat,omitempty"`
	Dimensions []CloudWatchMetricDimension `json:"dimensions,omitempty" validate:"max=10,uniqueDimensionNames,dive"`
	SQL        *string                     `json:"sql,omitempty"`
	JSON       *string                     `json:"json,omitempty"`
}

// RedshiftMetric represents metric from Redshift.
type RedshiftMetric struct {
	Region       *string `json:"region" validate:"required,max=255"`
	ClusterID    *string `json:"clusterId" validate:"required"`
	DatabaseName *string `json:"databaseName" validate:"required"`
	Query        *string `json:"query" validate:"required,redshiftRequiredColumns"`
}

// SumoLogicMetric represents metric from Sumo Logic.
type SumoLogicMetric struct {
	Type         *string `json:"type" validate:"required"`
	Query        *string `json:"query" validate:"required"`
	Quantization *string `json:"quantization,omitempty"`
	Rollup       *string `json:"rollup,omitempty"`
	// For struct level validation refer to sumoLogicStructValidation in pkg/manifest/v1alpha/validator.go
}

// InstanaMetric represents metric from Redshift.
type InstanaMetric struct {
	MetricType     string                           `json:"metricType" validate:"required,oneof=infrastructure application"` //nolint:lll
	Infrastructure *InstanaInfrastructureMetricType `json:"infrastructure,omitempty"`
	Application    *InstanaApplicationMetricType    `json:"application,omitempty"`
}

// InfluxDBMetric represents metric from InfluxDB
type InfluxDBMetric struct {
	Query *string `json:"query" validate:"required,influxDBRequiredPlaceholders"`
}

// GCMMetric represents metric from GCM
type GCMMetric struct {
	Query     string `json:"query" validate:"required"`
	ProjectID string `json:"projectId" validate:"required"`
}

type InstanaInfrastructureMetricType struct {
	MetricRetrievalMethod string  `json:"metricRetrievalMethod" validate:"required,oneof=query snapshot"`
	Query                 *string `json:"query,omitempty"`
	SnapshotID            *string `json:"snapshotId,omitempty"`
	MetricID              string  `json:"metricId" validate:"required"`
	PluginID              string  `json:"pluginId" validate:"required"`
}

type InstanaApplicationMetricType struct {
	MetricID         string                          `json:"metricId" validate:"required,oneof=calls erroneousCalls errors latency"` //nolint:lll
	Aggregation      string                          `json:"aggregation" validate:"required"`
	GroupBy          InstanaApplicationMetricGroupBy `json:"groupBy" validate:"required"`
	APIQuery         string                          `json:"apiQuery" validate:"required,json"`
	IncludeInternal  bool                            `json:"includeInternal,omitempty"`
	IncludeSynthetic bool                            `json:"includeSynthetic,omitempty"`
}

type InstanaApplicationMetricGroupBy struct {
	Tag               string  `json:"tag" validate:"required"`
	TagEntity         string  `json:"tagEntity" validate:"required,oneof=DESTINATION SOURCE NOT_APPLICABLE"`
	TagSecondLevelKey *string `json:"tagSecondLevelKey,omitempty"`
}

// IsStandardConfiguration returns true if the struct represents CloudWatch standard configuration.
func (c CloudWatchMetric) IsStandardConfiguration() bool {
	return c.Stat != nil || c.Dimensions != nil || c.MetricName != nil || c.Namespace != nil
}

// IsSQLConfiguration returns true if the struct represents CloudWatch SQL configuration.
func (c CloudWatchMetric) IsSQLConfiguration() bool {
	return c.SQL != nil
}

// IsJSONConfiguration returns true if the struct represents CloudWatch JSON configuration.
func (c CloudWatchMetric) IsJSONConfiguration() bool {
	return c.JSON != nil
}

// CloudWatchMetricDimension represents name/value pair that is part of the identity of a metric.
type CloudWatchMetricDimension struct {
	Name  *string `json:"name" validate:"required,max=255,ascii,notBlank"`
	Value *string `json:"value" validate:"required,max=255,ascii,notBlank"`
}

// PingdomMetric represents metric from Pingdom.
type PingdomMetric struct {
	CheckID   *string `json:"checkId" validate:"required,notBlank,numeric" example:"1234567"`
	CheckType *string `json:"checkType" validate:"required,pingdomCheckTypeFieldValid" example:"uptime"`
	Status    *string `json:"status,omitempty" validate:"omitempty,pingdomStatusValid" example:"up,down"`
}

// GraphiteMetric represents metric from Graphite.
type GraphiteMetric struct {
	MetricPath *string `json:"metricPath" validate:"required,metricPathGraphite"`
}

// BigQueryMetric represents metric from BigQuery
type BigQueryMetric struct {
	Query     string `json:"query" validate:"required,bigQueryRequiredColumns"`
	ProjectID string `json:"projectId" validate:"required"`
	Location  string `json:"location" validate:"required"`
}

// OpenTSDBMetric represents metric from OpenTSDB.
type OpenTSDBMetric struct {
	Query *string `json:"query" validate:"required"`
}

// GrafanaLokiMetric represents metric from GrafanaLokiMetric.
type GrafanaLokiMetric struct {
	Logql *string `json:"logql" validate:"required"`
}

// ThresholdBase base structure representing a threshold
type ThresholdBase struct {
	DisplayName string  `json:"displayName" validate:"omitempty,min=0,max=63" example:"Good"`
	Value       float64 `json:"value" validate:"numeric" example:"100"`
	Name        string  `json:"name" validate:"omitempty,objectName"`
	NameChanged bool    `json:"-"`
}

// Threshold represents single threshold for SLO, for internal usage
type Threshold struct {
	ThresholdBase
	// <!-- Go struct field and type names renaming budgetTarget to target has been postponed after GA as requested
	// in PC-1240. -->
	BudgetTarget *float64 `json:"target" validate:"required,numeric,gte=0,lt=1" example:"0.9"`
	// <!-- Go struct field and type names renaming thresholds to objectives has been postponed after GA as requested
	// in PC-1240. -->
	TimeSliceTarget *float64          `json:"timeSliceTarget,omitempty" example:"0.9"`
	CountMetrics    *CountMetricsSpec `json:"countMetrics,omitempty"`
	RawMetric       *RawMetricSpec    `json:"rawMetric,omitempty"`
	Operator        *string           `json:"op,omitempty" example:"lte"`
}

// Indicator represents integration with metric source can be. e.g. Prometheus, Datadog, for internal usage
type Indicator struct {
	MetricSource *MetricSourceSpec `json:"metricSource" validate:"required"`
	RawMetric    *MetricSpec       `json:"rawMetric,omitempty"`
}

type MetricSourceSpec struct {
	Project string `json:"project,omitempty" validate:"omitempty,objectName" example:"default"`
	Name    string `json:"name" validate:"required,objectName" example:"prometheus-source"`
	Kind    string `json:"kind" validate:"omitempty,metricSourceKind" example:"Agent"`
}

// Composite represents configuration for Composite SLO.
type Composite struct {
	BudgetTarget      float64                     `json:"target" validate:"required,numeric,gte=0,lt=1" example:"0.9"`
	BurnRateCondition *CompositeBurnRateCondition `json:"burnRateCondition,omitempty"`
}

// CompositeVersion represents composite version history stored for restoring process.
type CompositeVersion struct {
	Version      int32
	Created      string
	Dependencies []string
}

// CompositeBurnRateCondition represents configuration for Composite SLO  with occurrences budgeting method.
type CompositeBurnRateCondition struct {
	Value    float64 `json:"value" validate:"numeric,gte=0,lte=1000" example:"2"`
	Operator string  `json:"op" validate:"required,oneof=gt" example:"gt"`
}

// AnomalyConfig represents relationship between anomaly type and selected notification methods.
// This will be removed (moved into Anomaly Policy) in PC-8502
type AnomalyConfig struct {
	NoData *AnomalyConfigNoData `json:"noData" validate:"omitempty"`
}

// AnomalyConfigNoData contains alertMethods used for No Data anomaly type.
type AnomalyConfigNoData struct {
	AlertMethods []AnomalyConfigAlertMethod `json:"alertMethods" validate:"required"`
}

// AnomalyConfigAlertMethod represents a single alert method used in AnomalyConfig
// defined by name and project.
type AnomalyConfigAlertMethod struct {
	Name    string `json:"name" validate:"required,objectName" example:"slack-monitoring-channel"`
	Project string `json:"project,omitempty" validate:"objectName" example:"default"`
}

// SLOSpec represents content of Spec typical for SLO Object
type SLOSpec struct {
	Description     string         `json:"description" validate:"description" example:"Total count of server requests"` //nolint:lll
	Indicator       Indicator      `json:"indicator"`
	BudgetingMethod string         `json:"budgetingMethod" validate:"required,budgetingMethod" example:"Occurrences"`
	Thresholds      []Threshold    `json:"objectives" validate:"required,dive"`
	Service         string         `json:"service" validate:"required,objectName" example:"webapp-service"`
	TimeWindows     []TimeWindow   `json:"timeWindows" validate:"required,len=1,dive"`
	AlertPolicies   []string       `json:"alertPolicies" validate:"omitempty"`
	Attachments     []Attachment   `json:"attachments,omitempty" validate:"omitempty,max=20,dive"`
	CreatedAt       string         `json:"createdAt,omitempty"`
	Composite       *Composite     `json:"composite,omitempty" validate:"omitempty"`
	AnomalyConfig   *AnomalyConfig `json:"anomalyConfig,omitempty" validate:"omitempty"`
}

// SLO struct which mapped one to one with kind: slo yaml definition, external usage
type SLO struct {
	manifest.ObjectHeader
	Spec   SLOSpec    `json:"spec"`
	Status *SLOStatus `json:"status,omitempty"`
}

// getUniqueIdentifiers returns uniqueIdentifiers used to check
// potential conflicts between simultaneously applied objects.
func (s SLO) getUniqueIdentifiers() uniqueIdentifiers {
	return uniqueIdentifiers{Name: s.Metadata.Name, Project: s.Metadata.Project}
}

type SLOStatus struct {
	TimeTravelStatus *TimeTravelStatus `json:"timeTravel,omitempty"`
}

// genericToSLO converts ObjectGeneric to Object SLO
func genericToSLO(o manifest.ObjectGeneric, v validator, onlyHeader bool) (SLO, error) {
	res := SLO{
		ObjectHeader: o.ObjectHeader,
	}
	if onlyHeader {
		return res, nil
	}
	var resSpec SLOSpec
	if err := json.Unmarshal(o.Spec, &resSpec); err != nil {
		return res, manifest.EnhanceError(o, err)
	}
	res.Spec = resSpec

	// to keep BC with the ThousandEyes initial implementation (that did not support passing TestType),
	// we default `res.Spec.Indicator.RawMetrics.ThousandEyes.TestType` to a value that, until now, was implicitly assumed
	setThousandEyesDefaults(&res)

	if err := v.Check(res); err != nil {
		return res, manifest.EnhanceError(o, err)
	}

	if res.Spec.Indicator.MetricSource.Project == "" {
		res.Spec.Indicator.MetricSource.Project = res.Metadata.Project
	}
	if res.Spec.Indicator.MetricSource.Kind == "" {
		res.Spec.Indicator.MetricSource.Kind = KindAgent
	}

	// we're moving towards the version where raw metrics are defined on each objective, but for now,
	// we have to make sure that old contract (with indicator defined directly on the SLO's spec) is also supported
	if res.Spec.Indicator.RawMetric != nil {
		for i := range res.Spec.Thresholds {
			res.Spec.Thresholds[i].RawMetric = &RawMetricSpec{
				MetricQuery: res.Spec.Indicator.RawMetric,
			}
		}
	}

	// AnomalyConfig will be moved into Anomaly Rules in PC-8502.
	// Set the default value of all alert methods defined in anomaly config to the same project
	// that is used by SLO.
	if res.Spec.AnomalyConfig != nil && res.Spec.AnomalyConfig.NoData != nil {
		for i := 0; i < len(res.Spec.AnomalyConfig.NoData.AlertMethods); i++ {
			if res.Spec.AnomalyConfig.NoData.AlertMethods[i].Project == "" {
				res.Spec.AnomalyConfig.NoData.AlertMethods[i].Project = res.Metadata.Project
			}
		}
	}

	return res, nil
}

func setThousandEyesDefaults(slo *SLO) {
	if slo.Spec.Indicator.RawMetric != nil &&
		slo.Spec.Indicator.RawMetric.ThousandEyes != nil &&
		slo.Spec.Indicator.RawMetric.ThousandEyes.TestType == nil {
		metricType := ThousandEyesNetLatency
		slo.Spec.Indicator.RawMetric.ThousandEyes.TestType = &metricType
	}

	for i, threshold := range slo.Spec.Thresholds {
		if threshold.RawMetric != nil &&
			threshold.RawMetric.MetricQuery != nil &&
			threshold.RawMetric.MetricQuery.ThousandEyes != nil &&
			threshold.RawMetric.MetricQuery.ThousandEyes.TestType == nil {
			metricType := ThousandEyesNetLatency
			slo.Spec.Thresholds[i].RawMetric.MetricQuery.ThousandEyes.TestType = &metricType
		}
	}
}

// Calendar struct represents calendar time window
type Calendar struct {
	StartTime string `json:"startTime" validate:"required,dateWithTime,minDateTime" example:"2020-01-21 12:30:00"`
	TimeZone  string `json:"timeZone" validate:"required,timeZone" example:"America/New_York"`
}

// Period represents period of time
type Period struct {
	Begin string `json:"begin"`
	End   string `json:"end"`
}

// TimeWindow represents content of time window
type TimeWindow struct {
	Unit      string    `json:"unit" validate:"required,timeUnit" example:"Week"`
	Count     int       `json:"count" validate:"required,gt=0" example:"1"`
	IsRolling bool      `json:"isRolling" example:"true"`
	Calendar  *Calendar `json:"calendar,omitempty"`

	// Period is only returned in `/get/slo` requests it is ignored for `/apply`
	Period *Period `json:"period,omitempty"`
}

// Attachment represents user defined URL attached to SLO
type Attachment struct {
	URL         string  `json:"url" validate:"required,url"`
	DisplayName *string `json:"displayName,omitempty" validate:"max=63"`
}

// PrometheusAgentConfig represents content of Prometheus Configuration typical for Agent Object.
type PrometheusAgentConfig struct {
	URL    *string `json:"url,omitempty" example:"http://prometheus-service.monitoring:8080"`
	Region string  `json:"region,omitempty" example:"eu-cental-1"`
}

// FilterEntry represents single metric label to be matched against value
type FilterEntry struct {
	Label string `json:"label" validate:"required,prometheusLabelName"`
	Value string `json:"value" validate:"required"`
}

// DatadogAgentConfig represents content of Datadog Configuration typical for Agent Object.
type DatadogAgentConfig struct {
	Site string `json:"site,omitempty" validate:"site" example:"eu,us3.datadoghq.com"`
}

// DatadogDirectConfig represents content of Datadog Configuration typical for Direct Object.
type DatadogDirectConfig struct {
	Site           string `json:"site,omitempty" validate:"site" example:"eu,us3.datadoghq.com"`
	APIKey         string `json:"apiKey" example:"secret"`
	ApplicationKey string `json:"applicationKey" example:"secret"`
}

// PublicDatadogDirectConfig represents content of Datadog Configuration typical for Direct Object without secrets.
type PublicDatadogDirectConfig struct {
	Site                 string `json:"site,omitempty" example:"eu,us3.datadoghq.com"`
	HiddenAPIKey         string `json:"apiKey" example:"[hidden]"`
	HiddenApplicationKey string `json:"applicationKey" example:"[hidden]"`
}

// NewRelicAgentConfig represents content of NewRelic Configuration typical for Agent Object.
type NewRelicAgentConfig struct {
	AccountID json.Number `json:"accountId,omitempty" example:"123654"`
}

// NewRelicDirectConfig represents content of NewRelic Configuration typical for Direct Object.
type NewRelicDirectConfig struct {
	AccountID        json.Number `json:"accountId" validate:"required" example:"123654"`
	InsightsQueryKey string      `json:"insightsQueryKey" validate:"newRelicApiKey" example:"secret"`
}

// PublicNewRelicDirectConfig represents content of NewRelic Configuration typical for Direct Object without secrets.
type PublicNewRelicDirectConfig struct {
	AccountID              json.Number `json:"accountId,omitempty" example:"123654"`
	HiddenInsightsQueryKey string      `json:"insightsQueryKey" example:"[hidden]"`
}

// AppDynamicsAgentConfig represents content of AppDynamics Configuration typical for Agent Object.
type AppDynamicsAgentConfig struct {
	URL string `json:"url,omitempty" example:"https://nobl9.saas.appdynamics.com"`
}

// PublicAppDynamicsDirectConfig represents public content of AppDynamics Configuration
// typical for Direct Object without secrets.
type PublicAppDynamicsDirectConfig struct {
	URL                string `json:"url,omitempty" example:"https://nobl9.saas.appdynamics.com"`
	ClientID           string `json:"clientID,omitempty" example:"apiClientID@accountID"`
	ClientName         string `json:"clientName,omitempty" example:"apiClientID"`
	AccountName        string `json:"accountName,omitempty" example:"accountID"`
	HiddenClientSecret string `json:"clientSecret,omitempty" example:"[hidden]"`
}

// GenerateMissingFields checks if there is no ClientName and AccountName
//
//	then separates ClientID into ClientName and AccountName.
func (a *PublicAppDynamicsDirectConfig) GenerateMissingFields() {
	if a.ClientName == "" && a.AccountName == "" {
		at := strings.LastIndex(a.ClientID, "@")
		if at >= 0 {
			a.ClientName, a.AccountName = a.ClientID[:at], a.ClientID[at+1:]
		}
	}
}

// AppDynamicsDirectConfig represents content of AppDynamics Configuration typical for Direct Object.
type AppDynamicsDirectConfig struct {
	URL          string `json:"url,omitempty" validate:"httpsURL" example:"https://nobl9.saas.appdynamics.com"`
	ClientID     string `json:"clientID,omitempty" example:"apiClientID@accountID"`
	ClientName   string `json:"clientName,omitempty" example:"apiClientID"`
	AccountName  string `json:"accountName,omitempty" example:"accountID"`
	ClientSecret string `json:"clientSecret,omitempty" example:"secret"`
}

// GenerateMissingFields - this function is responsible for generating ClientID from AccountName and ClientName
// when provided with new, also it generates AccountName and ClientName for old already existing configs.
func (a *AppDynamicsDirectConfig) GenerateMissingFields() {
	if a.AccountName != "" && a.ClientName != "" {
		a.ClientID = fmt.Sprintf("%s@%s", a.ClientName, a.AccountName)
	} else if a.ClientID != "" {
		at := strings.LastIndex(a.ClientID, "@")
		if at >= 0 {
			a.ClientName, a.AccountName = a.ClientID[:at], a.ClientID[at+1:]
		}
	}
}

// SplunkAgentConfig represents content of Splunk Configuration typical for Agent Object.
type SplunkAgentConfig struct {
	URL string `json:"url,omitempty" example:"https://localhost:8089/servicesNS/admin/"`
}

// SplunkDirectConfig represents content of Splunk Configuration typical for Direct Object.
type SplunkDirectConfig struct {
	URL         string `json:"url,omitempty" validate:"httpsURL" example:"https://api.eu0.signalfx.com"`
	AccessToken string `json:"accessToken,omitempty"`
}

// PublicSplunkDirectConfig represents content of Splunk Configuration typical for Direct Object.
type PublicSplunkDirectConfig struct {
	URL               string `json:"url,omitempty" example:"https://api.eu0.signalfx.com"`
	HiddenAccessToken string `json:"accessToken,omitempty"`
}

// LightstepAgentConfig represents content of Lightstep Configuration typical for Agent Object.
type LightstepAgentConfig struct {
	Organization string `json:"organization,omitempty" example:"LightStep-Play"`
	Project      string `json:"project,omitempty" example:"play"`
}

type LightstepDirectConfig struct {
	Organization string `json:"organization,omitempty" validate:"required" example:"LightStep-Play"`
	Project      string `json:"project,omitempty" validate:"required" example:"play"`
	AppToken     string `json:"appToken"`
}

type PublicLightstepDirectConfig struct {
	Organization   string `json:"organization,omitempty" example:"LightStep-Play"`
	Project        string `json:"project,omitempty" example:"play"`
	HiddenAppToken string `json:"appToken"`
}

// SplunkObservabilityAgentConfig represents content of SplunkObservability Configuration typical for Agent Object.
type SplunkObservabilityAgentConfig struct {
	Realm string `json:"realm,omitempty" validate:"required"  example:"us1"`
}

// SplunkObservabilityDirectConfig represents content of SplunkObservability Configuration typical for Direct Object.
type SplunkObservabilityDirectConfig struct {
	Realm       string `json:"realm,omitempty" validate:"required" example:"us1"`
	AccessToken string `json:"accessToken,omitempty"`
}

// PublicSplunkObservabilityDirectConfig represents content of SplunkObservability
// Configuration typical for Direct Object.
type PublicSplunkObservabilityDirectConfig struct {
	Realm             string `json:"realm,omitempty" example:"us1"`
	HiddenAccessToken string `json:"accessToken,omitempty"`
}

// ThousandEyesAgentConfig represents content of ThousandEyes Configuration typical for Agent Object.
type ThousandEyesAgentConfig struct {
	// ThousandEyes agent doesn't require any additional parameters.
}

// ThousandEyesDirectConfig represents content of ThousandEyes Configuration typical for Direct Object.
type ThousandEyesDirectConfig struct {
	OauthBearerToken string `json:"oauthBearerToken,omitempty"`
}

// PublicThousandEyesDirectConfig content of ThousandEyes
// Configuration typical for Direct Object
type PublicThousandEyesDirectConfig struct {
	HiddenOauthBearerToken string `json:"oauthBearerToken,omitempty"`
}

// DynatraceAgentConfig represents content of Dynatrace Configuration typical for Agent Object.
type DynatraceAgentConfig struct {
	URL string `json:"url,omitempty" validate:"required,url,urlDynatrace" example:"https://{your-environment-id}.live.dynatrace.com or https://{your-domain}/e/{your-environment-id}"` //nolint: lll
}

// ElasticsearchAgentConfig represents content of Elasticsearch Configuration typical for Agent Object.
type ElasticsearchAgentConfig struct {
	URL string `json:"url,omitempty" validate:"required,url,urlElasticsearch" example:"https://observability-deployment-946814.es.eu-central-1.aws.cloud.es.io:9243"` //nolint: lll
}

// GraphiteAgentConfig represents content of Graphite Configuration typical for Agent Object.
type GraphiteAgentConfig struct {
	URL string `json:"url,omitempty" validate:"required,url" example:"http://graphite.example.com"`
}

// BigQueryAgentConfig represents content of BigQuery configuration.
// Since the agent does not require additional configuration this is just a marker struct.
type BigQueryAgentConfig struct {
}

// BigQueryDirectConfig represents content of BigQuery configuration typical for Direct Object.
type BigQueryDirectConfig struct {
	ServiceAccountKey string `json:"serviceAccountKey,omitempty"`
}

// PublicBigQueryDirectConfig represents content of BigQuery configuration typical for Direct Object without secrets.
type PublicBigQueryDirectConfig struct {
	HiddenServiceAccountKey string `json:"serviceAccountKey,omitempty"`
}

// GCMAgentConfig represents content of GCM configuration.
// Since the agent does not require additional configuration this is just a marker struct.
type GCMAgentConfig struct {
}

// GCMDirectConfig represents content of GCM configuration typical for Direct Object.
type GCMDirectConfig struct {
	ServiceAccountKey string `json:"serviceAccountKey,omitempty"`
}

// PublicGCMDirectConfig represents content of GCM configuration typical for Direct Object without secrets.
type PublicGCMDirectConfig struct {
	HiddenServiceAccountKey string `json:"serviceAccountKey,omitempty"`
}

// DynatraceDirectConfig represents content of Dynatrace configuration typical for Direct Object.
type DynatraceDirectConfig struct {
	URL            string `json:"url,omitempty" validate:"required,url,httpsURL" example:"https://{your-environment-id}.live.dynatrace.com or https://{your-domain}/e/{your-environment-id}"` //nolint: lll
	DynatraceToken string `json:"dynatraceToken,omitempty"`
}

// PublicDynatraceDirectConfig represents content of Dynatrace configuration typical for Direct Object without secrets.
type PublicDynatraceDirectConfig struct {
	URL                  string `json:"url,omitempty" validate:"required,url,httpsURL" example:"https://{your-environment-id}.live.dynatrace.com or https://{your-domain}/e/{your-environment-id}"` //nolint: lll
	HiddenDynatraceToken string `json:"dynatraceToken,omitempty"`
}

// OpenTSDBAgentConfig represents content of OpenTSDB Configuration typical for Agent Object.
type OpenTSDBAgentConfig struct {
	URL string `json:"url,omitempty" validate:"required,url" example:"example of OpenTSDB cluster URL"` //nolint: lll
}

// GrafanaLokiAgentConfig represents content of GrafanaLoki Configuration typical for Agent Object.
type GrafanaLokiAgentConfig struct {
	URL string `json:"url,omitempty" validate:"required,url" example:"example of GrafanaLoki cluster URL"` //nolint: lll
}

// CloudWatchAgentConfig represents content of CloudWatch Configuration typical for Agent Object.
type CloudWatchAgentConfig struct {
	// CloudWatch agent doesn't require any additional parameters.
}

// CloudWatchDirectConfig represents content of CloudWatch Configuration typical for Direct Object.
type CloudWatchDirectConfig struct {
	AccessKeyID     string `json:"accessKeyID"`
	SecretAccessKey string `json:"secretAccessKey"`
}

// PublicCloudWatchDirectConfig represents content of CloudWatch Configuration typical for Direct Object
// without secrets.
type PublicCloudWatchDirectConfig struct {
	HiddenAccessKeyID     string `json:"accessKeyID"`
	HiddenSecretAccessKey string `json:"secretAccessKey"`
}

// PingdomAgentConfig represents content of Pingdom Configuration typical for Agent Object.
type PingdomAgentConfig struct {
	// Pingdom agent doesn't require any additional parameter
}

// PingdomDirectConfig represents content of Pingdom Configuration typical for Direct Object.
type PingdomDirectConfig struct {
	APIToken string `json:"apiToken"`
}

type PublicPingdomDirectConfig struct {
	HiddenAPIToken string `json:"apiToken"`
}

// AmazonPrometheusAgentConfig represents content of Amazon Managed Service Configuration typical for Agent Object.
type AmazonPrometheusAgentConfig struct {
	URL    string `json:"url" validate:"required,url"`
	Region string `json:"region" validate:"required,max=255"`
}

// RedshiftAgentConfig represents content of Redshift configuration typical for Agent Object
// Since the agent does not require additional configuration this is just a marker struct.
type RedshiftAgentConfig struct {
}

// SumoLogicAgentConfig represents content of Sumo Logic configuration typical for Agent Object.
type SumoLogicAgentConfig struct {
	URL string `json:"url" validate:"required,url"`
}

// InstanaAgentConfig represents content of Instana configuration typical for Agent Object
type InstanaAgentConfig struct {
	URL string `json:"url" validate:"required,url"`
}

// InfluxDBAgentConfig represents content of InfluxDB configuration typical fo Agent Object
type InfluxDBAgentConfig struct {
	URL string `json:"url" validate:"required,url"`
}

// InstanaDirectConfig represents content of Instana configuration typical for Direct Object.
type InstanaDirectConfig struct {
	APIToken string `json:"apiToken"`
	URL      string `json:"url" validate:"required,url,httpsURL"`
}

// PublicInstanaDirectConfig represents content of Instana configuration typical for Direct Object without secrets.
type PublicInstanaDirectConfig struct {
	HiddenAPIToken string `json:"apiToken"`
	URL            string `json:"url"`
}

// InfluxDBDirectConfig represents content of InfluxDB configuration typical for Direct Object.
type InfluxDBDirectConfig struct {
	URL            string `json:"url" validate:"required,url"`
	APIToken       string `json:"apiToken"`
	OrganizationID string `json:"organizationID"`
}

// PublicInfluxDBDirectConfig represents content of InfluxDB configuration typical for Direct Object without secrets.
type PublicInfluxDBDirectConfig struct {
	URL                  string `json:"url"`
	HiddenAPIToken       string `json:"apiToken"`
	HiddenOrganizationID string `json:"organizationID"`
}

// RedshiftDirectConfig represents content of Redshift configuration typical for Direct Object.
type RedshiftDirectConfig struct {
	AccessKeyID     string `json:"accessKeyID"`
	SecretAccessKey string `json:"secretAccessKey"`
	SecretARN       string `json:"secretARN"`
}

// PublicRedshiftDirectConfig represents content of Redshift configuration typical for Direct Object without secrets.
type PublicRedshiftDirectConfig struct {
	HiddenAccessKeyID     string `json:"accessKeyID"`
	HiddenSecretAccessKey string `json:"secretAccessKey"`
	SecretARN             string `json:"secretARN"`
}

// SumoLogicDirectConfig represents content of SumoLogic configuration typical for Direct Object.
type SumoLogicDirectConfig struct {
	AccessID  string `json:"accessID"`
	AccessKey string `json:"accessKey"`
	URL       string `json:"url" validate:"required,url"`
}

// PublicSumoLogicDirectConfig represents content of SumoLogic configuration typical for Direct Object without secrets.
type PublicSumoLogicDirectConfig struct {
	HiddenAccessID  string `json:"accessID"`
	HiddenAccessKey string `json:"accessKey"`
	URL             string `json:"url"`
}

// genericToAgent converts ObjectGeneric to ObjectAgent
func genericToAgent(o manifest.ObjectGeneric, v validator, onlyHeader bool) (Agent, error) {
	res := Agent{
		ObjectHeader: o.ObjectHeader,
	}
	if onlyHeader {
		return res, nil
	}
	var resSpec AgentSpec
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

// Agent struct which mapped one to one with kind: Agent yaml definition
type Agent struct {
	manifest.ObjectHeader
	Spec   AgentSpec   `json:"spec"`
	Status AgentStatus `json:"status"`
}

// getUniqueIdentifiers returns uniqueIdentifiers used to check
// potential conflicts between simultaneously applied objects.
func (a Agent) getUniqueIdentifiers() uniqueIdentifiers {
	return uniqueIdentifiers{Name: a.Metadata.Name, Project: a.Metadata.Project}
}

// AgentWithSLOs struct which mapped one to one with kind: agent and slo yaml definition
type AgentWithSLOs struct {
	Agent Agent `json:"agent"`
	SLOs  []SLO `json:"slos"`
}

// AgentStatus represents content of Status optional for Agent Object
type AgentStatus struct {
	AgentType      string `json:"agentType" example:"Prometheus"`
	AgentVersion   string `json:"agentVersion,omitempty" example:"0.0.9"`
	LastConnection string `json:"lastConnection,omitempty" example:"2020-08-31T14:26:13Z"`
}

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

// AgentSpec represents content of Spec typical for Agent Object
type AgentSpec struct {
	Description             string                          `json:"description,omitempty" validate:"description" example:"Prometheus description"` //nolint:lll
	SourceOf                []string                        `json:"sourceOf" example:"Metrics,Services"`
	ReleaseChannel          string                          `json:"releaseChannel,omitempty" example:"beta,stable"`
	Prometheus              *PrometheusAgentConfig          `json:"prometheus,omitempty"`
	Datadog                 *DatadogAgentConfig             `json:"datadog,omitempty"`
	NewRelic                *NewRelicAgentConfig            `json:"newRelic,omitempty"`
	AppDynamics             *AppDynamicsAgentConfig         `json:"appDynamics,omitempty"`
	Splunk                  *SplunkAgentConfig              `json:"splunk,omitempty"`
	Lightstep               *LightstepAgentConfig           `json:"lightstep,omitempty"`
	SplunkObservability     *SplunkObservabilityAgentConfig `json:"splunkObservability,omitempty"`
	Dynatrace               *DynatraceAgentConfig           `json:"dynatrace,omitempty"`
	Elasticsearch           *ElasticsearchAgentConfig       `json:"elasticsearch,omitempty"`
	ThousandEyes            *ThousandEyesAgentConfig        `json:"thousandEyes,omitempty"`
	Graphite                *GraphiteAgentConfig            `json:"graphite,omitempty"`
	BigQuery                *BigQueryAgentConfig            `json:"bigQuery,omitempty"`
	OpenTSDB                *OpenTSDBAgentConfig            `json:"opentsdb,omitempty"`
	GrafanaLoki             *GrafanaLokiAgentConfig         `json:"grafanaLoki,omitempty"`
	CloudWatch              *CloudWatchAgentConfig          `json:"cloudWatch,omitempty"`
	Pingdom                 *PingdomAgentConfig             `json:"pingdom,omitempty"`
	AmazonPrometheus        *AmazonPrometheusAgentConfig    `json:"amazonPrometheus,omitempty"`
	Redshift                *RedshiftAgentConfig            `json:"redshift,omitempty"`
	SumoLogic               *SumoLogicAgentConfig           `json:"sumoLogic,omitempty"`
	Instana                 *InstanaAgentConfig             `json:"instana,omitempty"`
	InfluxDB                *InfluxDBAgentConfig            `json:"influxdb,omitempty"`
	GCM                     *GCMAgentConfig                 `json:"gcm,omitempty"`
	HistoricalDataRetrieval *HistoricalDataRetrieval        `json:"historicalDataRetrieval,omitempty"`
	QueryDelay              *QueryDelay                     `json:"queryDelay,omitempty"`
}

func (spec AgentSpec) GetType() (DataSourceType, error) {
	switch {
	case spec.Prometheus != nil:
		return Prometheus, nil
	case spec.Datadog != nil:
		return Datadog, nil
	case spec.NewRelic != nil:
		return NewRelic, nil
	case spec.AppDynamics != nil:
		return AppDynamics, nil
	case spec.Splunk != nil:
		return Splunk, nil
	case spec.Lightstep != nil:
		return Lightstep, nil
	case spec.SplunkObservability != nil:
		return SplunkObservability, nil
	case spec.Dynatrace != nil:
		return Dynatrace, nil
	case spec.Elasticsearch != nil:
		return Elasticsearch, nil
	case spec.ThousandEyes != nil:
		return ThousandEyes, nil
	case spec.Graphite != nil:
		return Graphite, nil
	case spec.BigQuery != nil:
		return BigQuery, nil
	case spec.OpenTSDB != nil:
		return OpenTSDB, nil
	case spec.GrafanaLoki != nil:
		return GrafanaLoki, nil
	case spec.CloudWatch != nil:
		return CloudWatch, nil
	case spec.Pingdom != nil:
		return Pingdom, nil
	case spec.AmazonPrometheus != nil:
		return AmazonPrometheus, nil
	case spec.Redshift != nil:
		return Redshift, nil
	case spec.SumoLogic != nil:
		return SumoLogic, nil
	case spec.Instana != nil:
		return Instana, nil
	case spec.InfluxDB != nil:
		return InfluxDB, nil
	case spec.GCM != nil:
		return GCM, nil
	}
	return 0, errors.New("unknown agent type")
}

// genericToDirect converts ObjectGeneric to ObjectDirect
func genericToDirect(o manifest.ObjectGeneric, v validator, onlyHeader bool) (Direct, error) {
	res := Direct{
		ObjectHeader: o.ObjectHeader,
	}
	if onlyHeader {
		return res, nil
	}
	var resSpec DirectSpec
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

// Direct struct which mapped one to one with kind: Direct yaml definition
type Direct struct {
	manifest.ObjectHeader
	Spec   DirectSpec   `json:"spec"`
	Status DirectStatus `json:"status"`
}

// getUniqueIdentifiers returns uniqueIdentifiers used to check
// potential conflicts between simultaneously applied objects.
func (d Direct) getUniqueIdentifiers() uniqueIdentifiers {
	return uniqueIdentifiers{Name: d.Metadata.Name, Project: d.Metadata.Project}
}

// PublicDirect struct which mapped one to one with kind: Direct yaml definition without secrets
type PublicDirect struct {
	manifest.ObjectHeader
	Spec   PublicDirectSpec `json:"spec"`
	Status DirectStatus     `json:"status"`
}

// PublicDirectWithSLOs struct which mapped one to one with kind: direct and slo yaml definition
type PublicDirectWithSLOs struct {
	Direct PublicDirect `json:"direct"`
	SLOs   []SLO        `json:"slos"`
}

// DirectSpec represents content of Spec typical for Direct Object
type DirectSpec struct {
	Description             string                           `json:"description,omitempty" validate:"description" example:"Datadog description"` //nolint:lll
	SourceOf                []string                         `json:"sourceOf" example:"Metrics,Services"`
	Datadog                 *DatadogDirectConfig             `json:"datadog,omitempty"`
	LogCollectionEnabled    *bool                            `json:"logCollectionEnabled,omitempty"`
	NewRelic                *NewRelicDirectConfig            `json:"newRelic,omitempty"`
	AppDynamics             *AppDynamicsDirectConfig         `json:"appDynamics,omitempty"`
	SplunkObservability     *SplunkObservabilityDirectConfig `json:"splunkObservability,omitempty"`
	ThousandEyes            *ThousandEyesDirectConfig        `json:"thousandEyes,omitempty"`
	BigQuery                *BigQueryDirectConfig            `json:"bigQuery,omitempty"`
	Splunk                  *SplunkDirectConfig              `json:"splunk,omitempty"`
	CloudWatch              *CloudWatchDirectConfig          `json:"cloudWatch,omitempty"`
	Pingdom                 *PingdomDirectConfig             `json:"pingdom,omitempty"`
	Redshift                *RedshiftDirectConfig            `json:"redshift,omitempty"`
	SumoLogic               *SumoLogicDirectConfig           `json:"sumoLogic,omitempty"`
	Instana                 *InstanaDirectConfig             `json:"instana,omitempty"`
	InfluxDB                *InfluxDBDirectConfig            `json:"influxdb,omitempty"`
	GCM                     *GCMDirectConfig                 `json:"gcm,omitempty"`
	Lightstep               *LightstepDirectConfig           `json:"lightstep,omitempty"`
	Dynatrace               *DynatraceDirectConfig           `json:"dynatrace,omitempty"`
	HistoricalDataRetrieval *HistoricalDataRetrieval         `json:"historicalDataRetrieval,omitempty"`
	QueryDelay              *QueryDelay                      `json:"queryDelay,omitempty"`
}

var allDirectTypes = map[string]struct{}{
	Datadog.String():             {},
	NewRelic.String():            {},
	SplunkObservability.String(): {},
	AppDynamics.String():         {},
	ThousandEyes.String():        {},
	BigQuery.String():            {},
	Splunk.String():              {},
	CloudWatch.String():          {},
	Pingdom.String():             {},
	Redshift.String():            {},
	SumoLogic.String():           {},
	Instana.String():             {},
	InfluxDB.String():            {},
	GCM.String():                 {},
	Lightstep.String():           {},
	Dynatrace.String():           {},
}

func IsValidDirectType(directType string) bool {
	_, isValid := allDirectTypes[directType]
	return isValid
}

func (spec DirectSpec) GetType() (string, error) {
	switch {
	case spec.Datadog != nil:
		return Datadog.String(), nil
	case spec.NewRelic != nil:
		return NewRelic.String(), nil
	case spec.SplunkObservability != nil:
		return SplunkObservability.String(), nil
	case spec.AppDynamics != nil:
		return AppDynamics.String(), nil
	case spec.ThousandEyes != nil:
		return ThousandEyes.String(), nil
	case spec.BigQuery != nil:
		return BigQuery.String(), nil
	case spec.Splunk != nil:
		return Splunk.String(), nil
	case spec.CloudWatch != nil:
		return CloudWatch.String(), nil
	case spec.Pingdom != nil:
		return Pingdom.String(), nil
	case spec.Redshift != nil:
		return Redshift.String(), nil
	case spec.SumoLogic != nil:
		return SumoLogic.String(), nil
	case spec.Instana != nil:
		return Instana.String(), nil
	case spec.InfluxDB != nil:
		return InfluxDB.String(), nil
	case spec.GCM != nil:
		return GCM.String(), nil
	case spec.Lightstep != nil:
		return Lightstep.String(), nil
	case spec.Dynatrace != nil:
		return Dynatrace.String(), nil
	}
	return "", errors.New("unknown direct type")
}

// PublicDirectSpec represents content of Spec typical for Direct Object without secrets
type PublicDirectSpec struct {
	Description             string                                 `json:"description,omitempty" validate:"description" example:"Datadog description"` //nolint:lll
	SourceOf                []string                               `json:"sourceOf" example:"Metrics,Services"`
	ReleaseChannel          string                                 `json:"releaseChannel,omitempty" example:"beta,stable"`
	LogCollectionEnabled    bool                                   `json:"logCollectionEnabled,omitempty"`
	Datadog                 *PublicDatadogDirectConfig             `json:"datadog,omitempty"`
	NewRelic                *PublicNewRelicDirectConfig            `json:"newRelic,omitempty"`
	SplunkObservability     *PublicSplunkObservabilityDirectConfig `json:"splunkObservability,omitempty"`
	AppDynamics             *PublicAppDynamicsDirectConfig         `json:"appDynamics,omitempty"`
	ThousandEyes            *PublicThousandEyesDirectConfig        `json:"thousandEyes,omitempty"`
	BigQuery                *PublicBigQueryDirectConfig            `json:"bigQuery,omitempty"`
	Splunk                  *PublicSplunkDirectConfig              `json:"splunk,omitempty"`
	CloudWatch              *PublicCloudWatchDirectConfig          `json:"cloudWatch,omitempty"`
	Pingdom                 *PublicPingdomDirectConfig             `json:"pingdom,omitempty"`
	Redshift                *PublicRedshiftDirectConfig            `json:"redshift,omitempty"`
	SumoLogic               *PublicSumoLogicDirectConfig           `json:"sumoLogic,omitempty"`
	Instana                 *PublicInstanaDirectConfig             `json:"instana,omitempty"`
	InfluxDB                *PublicInfluxDBDirectConfig            `json:"influxdb,omitempty"`
	GCM                     *PublicGCMDirectConfig                 `json:"gcm,omitempty"`
	Lightstep               *PublicLightstepDirectConfig           `json:"lightstep,omitempty"`
	Dynatrace               *PublicDynatraceDirectConfig           `json:"dynatrace,omitempty"`
	HistoricalDataRetrieval *HistoricalDataRetrieval               `json:"historicalDataRetrieval,omitempty"`
	QueryDelay              *QueryDelay                            `json:"queryDelay,omitempty"`
}

// DirectStatus represents content of Status optional for Direct Object
type DirectStatus struct {
	DirectType string `json:"directType" example:"Datadog"`
}

// Service struct which mapped one to one with kind: service yaml definition
type Service struct {
	manifest.ObjectHeader
	Spec   ServiceSpec    `json:"spec"`
	Status *ServiceStatus `json:"status,omitempty"`
}

// getUniqueIdentifiers returns uniqueIdentifiers used to check
// potential conflicts between simultaneously applied objects.
func (s Service) getUniqueIdentifiers() uniqueIdentifiers {
	return uniqueIdentifiers{Name: s.Metadata.Name, Project: s.Metadata.Project}
}

// ServiceWithSLOs struct which mapped one to one with kind: service and slo yaml definition.
type ServiceWithSLOs struct {
	Service Service `json:"service"`
	SLOs    []SLO   `json:"slos"`
}

// ServiceStatus represents content of Status optional for Service Object.
type ServiceStatus struct {
	SloCount int `json:"sloCount"`
}

// ServiceSpec represents content of Spec typical for Service Object.
type ServiceSpec struct {
	Description string `json:"description" validate:"description" example:"Bleeding edge web app"`
}

// Project struct which mapped one to one with kind: project yaml definition.
type Project struct {
	manifest.ObjectInternal
	APIVersion string                   `json:"apiVersion" validate:"required" example:"n9/v1alpha"`
	Kind       string                   `json:"kind" validate:"required" example:"kind"`
	Metadata   manifest.ProjectMetadata `json:"metadata"`
	Spec       ProjectSpec              `json:"spec"`
}

// getUniqueIdentifiers returns uniqueIdentifiers used to check
// potential conflicts between simultaneously applied objects.
func (p Project) getUniqueIdentifiers() uniqueIdentifiers {
	return uniqueIdentifiers{Name: p.Metadata.Name}
}

// ProjectSpec represents content of Spec typical for Project Object.
type ProjectSpec struct {
	Description string `json:"description" validate:"description" example:"Bleeding edge web app"`
}

// genericToService converts ObjectGeneric to Object Service.
func genericToService(o manifest.ObjectGeneric, v validator, onlyHeader bool) (Service, error) {
	res := Service{
		ObjectHeader: o.ObjectHeader,
	}
	if onlyHeader {
		return res, nil
	}

	var resSpec ServiceSpec
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

// AlertPolicy represents a set of conditions that can trigger an alert.
type AlertPolicy struct {
	manifest.ObjectHeader
	Spec AlertPolicySpec `json:"spec"`
}

// getUniqueIdentifiers returns uniqueIdentifiers used to check
// potential conflicts between simultaneously applied objects.
func (a AlertPolicy) getUniqueIdentifiers() uniqueIdentifiers {
	return uniqueIdentifiers{Project: a.Metadata.Project, Name: a.Metadata.Name}
}

// AlertPolicySpec represents content of AlertPolicy's Spec.
type AlertPolicySpec struct {
	Description      string              `json:"description" validate:"description" example:"Error budget is at risk"`
	Severity         string              `json:"severity" validate:"required,severity" example:"High"`
	CoolDownDuration string              `json:"coolDown,omitempty" validate:"omitempty,validDuration,nonNegativeDuration,durationAtLeast=5m" example:"5m"` //nolint:lll
	Conditions       []AlertCondition    `json:"conditions" validate:"required,min=1,dive"`
	AlertMethods     []PublicAlertMethod `json:"alertMethods"`
}

func (spec AlertPolicySpec) GetAlertMethods() []PublicAlertMethod {
	return spec.AlertMethods
}

// AlertCondition represents a condition to meet to trigger an alert.
type AlertCondition struct {
	Measurement      string      `json:"measurement" validate:"required,alertPolicyMeasurement" example:"BurnedBudget"`
	Value            interface{} `json:"value" validate:"required" example:"0.97"`
	AlertingWindow   string      `json:"alertingWindow,omitempty" validate:"omitempty,validDuration,nonNegativeDuration" example:"30m"` //nolint:lll
	LastsForDuration string      `json:"lastsFor,omitempty" validate:"omitempty,validDuration,nonNegativeDuration" example:"15m"`       //nolint:lll
	Operator         string      `json:"op,omitempty" validate:"omitempty,operator" example:"lt"`
}

// AlertPolicyWithSLOs struct which mapped one to one with kind: alert policy and slo yaml definition
type AlertPolicyWithSLOs struct {
	AlertPolicy AlertPolicy `json:"alertPolicy"`
	SLOs        []SLO       `json:"slos"`
}

// AlertMethodWithAlertPolicy represents an AlertPolicies assigned to AlertMethod.
type AlertMethodWithAlertPolicy struct {
	AlertMethod   PublicAlertMethod `json:"alertMethod"`
	AlertPolicies []AlertPolicy     `json:"alertPolicies"`
}

// genericToAlertPolicy converts ObjectGeneric to ObjectAlertPolicy
func genericToAlertPolicy(o manifest.ObjectGeneric, v validator, onlyHeader bool) (AlertPolicy, error) {
	res := AlertPolicy{
		ObjectHeader: o.ObjectHeader,
	}
	if onlyHeader {
		return res, nil
	}
	var resSpec AlertPolicySpec
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

// Alert represents triggered alert
type Alert struct {
	manifest.ObjectHeader
	Spec AlertSpec `json:"spec"`
}

// AlertSpec represents content of Alert's Spec
type AlertSpec struct {
	AlertPolicy         manifest.Metadata `json:"alertPolicy"`
	SLO                 manifest.Metadata `json:"slo"`
	Service             manifest.Metadata `json:"service"`
	Threshold           AlertThreshold    `json:"objective"`
	Severity            string            `json:"severity" validate:"required,severity" example:"High"`
	Status              string            `json:"status" example:"Resolved"`
	TriggeredMetricTime string            `json:"triggeredMetricTime"`
	TriggeredClockTime  string            `json:"triggeredClockTime"`
	ResolvedClockTime   *string           `json:"resolvedClockTime,omitempty"`
	ResolvedMetricTime  *string           `json:"resolvedMetricTime,omitempty"`
	CoolDown            string            `json:"coolDown"`
	Conditions          []AlertCondition  `json:"conditions"`
}

type AlertThreshold struct {
	Value       float64 `json:"value" example:"100"`
	Name        string  `json:"name" validate:"omitempty"`
	DisplayName string  `json:"displayName" validate:"omitempty"`
}

// AlertMethod represents the configuration required to send a notification to an external service
// when an alert is triggered.
type AlertMethod struct {
	manifest.ObjectHeader
	Spec AlertMethodSpec `json:"spec"`
}

// getUniqueIdentifiers returns uniqueIdentifiers used to check
// potential conflicts between simultaneously applied objects.
func (a AlertMethod) getUniqueIdentifiers() uniqueIdentifiers {
	return uniqueIdentifiers{Project: a.Metadata.Project, Name: a.Metadata.Name}
}

// PublicAlertMethod represents the configuration required to send a notification to an external service
// when an alert is triggered.
type PublicAlertMethod struct {
	manifest.ObjectHeader
	Spec   PublicAlertMethodSpec    `json:"spec"`
	Status *PublicAlertMethodStatus `json:"status,omitempty"`
}

// PublicAlertMethodStatus represents content of Status optional for PublicAlertMethod Object
type PublicAlertMethodStatus struct {
	LastTestDate       string `json:"lastTestDate,omitempty" example:"2021-02-09T10:43:07Z"`
	NextTestPossibleAt string `json:"nextTestPossibleAt,omitempty" example:"2021-02-09T10:43:07Z"`
}

// AlertMethodSpec represents content of AlertMethod's Spec.
type AlertMethodSpec struct {
	Description string                 `json:"description" validate:"description" example:"Sends notification"`
	Webhook     *WebhookAlertMethod    `json:"webhook,omitempty" validate:"omitempty,dive"`
	PagerDuty   *PagerDutyAlertMethod  `json:"pagerduty,omitempty"`
	Slack       *SlackAlertMethod      `json:"slack,omitempty"`
	Discord     *DiscordAlertMethod    `json:"discord,omitempty"`
	Opsgenie    *OpsgenieAlertMethod   `json:"opsgenie,omitempty"`
	ServiceNow  *ServiceNowAlertMethod `json:"servicenow,omitempty"`
	Jira        *JiraAlertMethod       `json:"jira,omitempty"`
	Teams       *TeamsAlertMethod      `json:"msteams,omitempty"`
	Email       *EmailAlertMethod      `json:"email,omitempty"`
}

// PublicAlertMethodSpec represents content of AlertMethod's Spec without secrets.
type PublicAlertMethodSpec struct {
	Description string                       `json:"description" validate:"description" example:"Sends notification"`
	Webhook     *PublicWebhookAlertMethod    `json:"webhook,omitempty"`
	PagerDuty   *PublicPagerDutyAlertMethod  `json:"pagerduty,omitempty"`
	Slack       *PublicSlackAlertMethod      `json:"slack,omitempty"`
	Discord     *PublicDiscordAlertMethod    `json:"discord,omitempty"`
	Opsgenie    *PublicOpsgenieAlertMethod   `json:"opsgenie,omitempty"`
	ServiceNow  *PublicServiceNowAlertMethod `json:"servicenow,omitempty"`
	Jira        *PublicJiraAlertMethod       `json:"jira,omitempty"`
	Teams       *PublicTeamsAlertMethod      `json:"msteams,omitempty"`
	Email       *EmailAlertMethod            `json:"email,omitempty"`
}

// WebhookAlertMethod represents a set of properties required to send a webhook request.
type WebhookAlertMethod struct {
	URL            string          `json:"url" validate:"optionalURL"` // Field required when AlertMethod is created.
	Template       *string         `json:"template,omitempty" validate:"omitempty,allowedWebhookTemplateFields"`
	TemplateFields []string        `json:"templateFields,omitempty" validate:"omitempty,min=1,allowedWebhookTemplateFields"` //nolint:lll
	Headers        []WebhookHeader `json:"headers,omitempty" validate:"omitempty,max=10,dive"`
}
type WebhookHeader struct {
	Name     string `json:"name" validate:"required,headerName"`
	Value    string `json:"value" validate:"required"`
	IsSecret bool   `json:"isSecret"`
}

// PublicWebhookAlertMethod represents a set of properties required to send a webhook request without secrets.
type PublicWebhookAlertMethod struct {
	HiddenURL      string          `json:"url"`
	Template       *string         `json:"template,omitempty" validate:"omitempty,allowedWebhookTemplateFields"`
	TemplateFields []string        `json:"templateFields,omitempty" validate:"omitempty,min=1,allowedWebhookTemplateFields"` //nolint:lll
	Headers        []WebhookHeader `json:"headers,omitempty"`
}

// SendResolution If user set SendResolution, then “Send a notification after the cooldown period is over"
type SendResolution struct {
	Message *string `json:"message"`
}

// PagerDutyAlertMethod represents a set of properties required to open an Incident in PagerDuty.
type PagerDutyAlertMethod struct {
	IntegrationKey string          `json:"integrationKey" validate:"pagerDutyIntegrationKey"`
	SendResolution *SendResolution `json:"sendResolution,omitempty"`
}

// PublicPagerDutyAlertMethod represents a set of properties required to open an Incident in PagerDuty without secrets.
type PublicPagerDutyAlertMethod struct {
	HiddenIntegrationKey string          `json:"integrationKey"`
	SendResolution       *SendResolution `json:"sendResolution,omitempty"`
}

// SlackAlertMethod represents a set of properties required to send message to Slack.
type SlackAlertMethod struct {
	URL string `json:"url" validate:"optionalURL"` // Required when AlertMethod is created.
}

// PublicSlackAlertMethod represents a set of properties required to send message to Slack without secrets.
type PublicSlackAlertMethod struct {
	HiddenURL string `json:"url"`
}

// OpsgenieAlertMethod represents a set of properties required to send message to Opsgenie.
type OpsgenieAlertMethod struct {
	Auth string `json:"auth" validate:"opsgenieApiKey"` // Field required when AlertMethod is created.
	URL  string `json:"url" validate:"optionalURL"`
}

// PublicOpsgenieAlertMethod represents a set of properties required to send message to Opsgenie without secrets.
type PublicOpsgenieAlertMethod struct {
	HiddenAuth string `json:"auth"`
	URL        string `json:"url" validate:"required,url"`
}

// ServiceNowAlertMethod represents a set of properties required to send message to ServiceNow.
type ServiceNowAlertMethod struct {
	Username     string `json:"username" validate:"required"`
	Password     string `json:"password"` // Field required when AlertMethod is created.
	InstanceName string `json:"instanceName" validate:"required"`
}

// PublicServiceNowAlertMethod represents a set of properties required to send message to ServiceNow without secrets.
type PublicServiceNowAlertMethod struct {
	Username       string `json:"username" validate:"required"`
	InstanceName   string `json:"instanceName" validate:"required"`
	HiddenPassword string `json:"password"`
}

// DiscordAlertMethod represents a set of properties required to send message to Discord.
type DiscordAlertMethod struct {
	URL string `json:"url" validate:"urlDiscord"` // Field required when AlertMethod is created.
}

// PublicDiscordAlertMethod represents a set of properties required to send message to Discord without secrets.
type PublicDiscordAlertMethod struct {
	HiddenURL string `json:"url"`
}

// JiraAlertMethod represents a set of properties required create tickets in Jira.
type JiraAlertMethod struct {
	URL        string `json:"url" validate:"required,httpsURL,url"`
	Username   string `json:"username" validate:"required"`
	APIToken   string `json:"apiToken"` // Field required when AlertMethod is created.
	ProjectKey string `json:"projectKey" validate:"required"`
}

// PublicJiraAlertMethod represents a set of properties required create tickets in Jira without secrets.
type PublicJiraAlertMethod struct {
	URL            string `json:"url" validate:"required,httpsURL,url"`
	Username       string `json:"username" validate:"required"`
	ProjectKey     string `json:"projectKey" validate:"required"`
	HiddenAPIToken string `json:"apiToken"`
}

// TeamsAlertMethod represents a set of properties required create Microsoft Teams notifications.
type TeamsAlertMethod struct {
	URL string `json:"url" validate:"httpsURL"`
}

// PublicTeamsAlertMethod represents a set of properties required create Microsoft Teams notifications.
type PublicTeamsAlertMethod struct {
	HiddenURL string `json:"url"`
}

// EmailAlertMethod represents a set of properties required to send an email.
type EmailAlertMethod struct {
	To      []string `json:"to,omitempty" validate:"omitempty,max=10,emails"`
	Cc      []string `json:"cc,omitempty" validate:"omitempty,max=10,emails"`
	Bcc     []string `json:"bcc,omitempty" validate:"omitempty,max=10,emails"`
	Subject string   `json:"subject" validate:"required,max=90,allowedAlertMethodEmailSubjectFields"`
	Body    string   `json:"body" validate:"required,max=2000,allowedAlertMethodEmailBodyFields"`
}

// genericToAlertMethod converts ObjectGeneric to ObjectAlertMethod
func genericToAlertMethod(o manifest.ObjectGeneric, v validator, onlyHeader bool) (AlertMethod, error) {
	res := AlertMethod{
		ObjectHeader: o.ObjectHeader,
	}
	if onlyHeader {
		return res, nil
	}
	var resSpec AlertMethodSpec
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

// BurnedBudget represents content of burned budget for a given threshold.
type BurnedBudget struct {
	Value *float64 `json:"burnedBudget,omitempty"`
}

// DataExport struct which mapped one to one with kind: DataExport yaml definition
type DataExport struct {
	manifest.ObjectHeader
	Spec   DataExportSpec   `json:"spec"`
	Status DataExportStatus `json:"status"`
}

// getUniqueIdentifiers returns uniqueIdentifiers used to check
// potential conflicts between simultaneously applied objects.
func (d DataExport) getUniqueIdentifiers() uniqueIdentifiers {
	return uniqueIdentifiers{Project: d.Metadata.Project, Name: d.Metadata.Name}
}

// DataExportSpec represents content of DataExport's Spec
type DataExportSpec struct {
	ExportType string      `json:"exportType" validate:"required,exportType" example:"Snowflake"`
	Spec       interface{} `json:"spec" validate:"required"`
}

const (
	DataExportTypeS3        string = "S3"
	DataExportTypeSnowflake string = "Snowflake"
	DataExportTypeGCS       string = "GCS"
)

// S3DataExportSpec represents content of Amazon S3 export type spec.
type S3DataExportSpec struct {
	BucketName string `json:"bucketName" validate:"required,min=3,max=63,s3BucketName" example:"examplebucket"`
	RoleARN    string `json:"roleArn" validate:"required,min=20,max=2048,roleARN" example:"arn:aws:iam::12345/role/n9-access"` //nolint:lll
}

// GCSDataExportSpec represents content of GCP Cloud Storage export type spec.
type GCSDataExportSpec struct {
	BucketName string `json:"bucketName" validate:"required,min=3,max=222,gcsBucketName" example:"example-bucket.org.com"`
}

// DataExportStatus represents content of Status optional for DataExport Object
type DataExportStatus struct {
	ExportJob     DataExportStatusJob `json:"exportJob"`
	AWSExternalID *string             `json:"awsExternalID,omitempty"`
}

// DataExportStatusJob represents content of ExportJob status
type DataExportStatusJob struct {
	Timestamp string `json:"timestamp,omitempty" example:"2021-02-09T10:43:07Z"`
	State     string `json:"state" example:"finished"`
}

// dataExportGeneric represents struct to which every DataExport is parsable.
// Specific types of DataExport have different structures as Spec.
type dataExportGeneric struct {
	ExportType string          `json:"exportType" validate:"required,exportType" example:"Snowflake"`
	Spec       json.RawMessage `json:"spec"`
}

// genericToDataExport converts ObjectGeneric to ObjectDataExport
func genericToDataExport(o manifest.ObjectGeneric, v validator, onlyHeader bool) (DataExport, error) {
	res := DataExport{
		ObjectHeader: o.ObjectHeader,
	}
	if onlyHeader {
		return res, nil
	}
	deg := dataExportGeneric{}
	if err := json.Unmarshal(o.Spec, &deg); err != nil {
		err = manifest.EnhanceError(o, err)
		return res, err
	}

	resSpec := DataExportSpec{ExportType: deg.ExportType}
	switch resSpec.ExportType {
	case DataExportTypeS3, DataExportTypeSnowflake:
		resSpec.Spec = &S3DataExportSpec{}
	case DataExportTypeGCS:
		resSpec.Spec = &GCSDataExportSpec{}
	}
	if deg.Spec != nil {
		if err := json.Unmarshal(deg.Spec, &resSpec.Spec); err != nil {
			err = manifest.EnhanceError(o, err)
			return res, err
		}
	}
	res.Spec = resSpec
	if err := v.Check(res); err != nil {
		err = manifest.EnhanceError(o, err)
		return res, err
	}
	return res, nil
}

// genericToProject converts ObjectGeneric to Project
func genericToProject(o manifest.ObjectGeneric, v validator, onlyHeader bool) (Project, error) {
	res := Project{
		APIVersion: o.ObjectHeader.APIVersion,
		Kind:       o.ObjectHeader.Kind,
		Metadata: manifest.ProjectMetadata{
			Name:        o.Metadata.Name,
			DisplayName: o.Metadata.DisplayName,
			Labels:      o.Metadata.Labels,
		},
		ObjectInternal: manifest.ObjectInternal{
			Organization: o.ObjectHeader.Organization,
			ManifestSrc:  o.ObjectHeader.ManifestSrc,
		},
	}
	if onlyHeader {
		return res, nil
	}

	var resSpec ProjectSpec
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

// RoleBinding represents relation of User and Role
type RoleBinding struct {
	manifest.ObjectInternal
	APIVersion string                       `json:"apiVersion" validate:"required" example:"n9/v1alpha"`
	Kind       string                       `json:"kind" validate:"required" example:"kind"`
	Metadata   manifest.RoleBindingMetadata `json:"metadata"`
	Spec       RoleBindingSpec              `json:"spec"`
}

// getUniqueIdentifiers returns uniqueIdentifiers used to check
// potential conflicts between simultaneously applied objects.
func (r RoleBinding) getUniqueIdentifiers() uniqueIdentifiers {
	return uniqueIdentifiers{Name: r.Metadata.Name}
}

type RoleBindingSpec struct {
	User       *string `json:"user,omitempty" validate:"required_without=GroupRef"`
	GroupRef   *string `json:"groupRef,omitempty" validate:"required_without=User"`
	RoleRef    string  `json:"roleRef" validate:"required"`
	ProjectRef string  `json:"projectRef,omitempty"`
}

type OrganizationInformation struct {
	ID          string  `json:"id"`
	DisplayName *string `json:"displayName"`
}

// genericToRoleBinding converts ObjectGeneric to ObjectRoleBinding
// onlyHeader parameter is not supported for RoleBinding since ProjectRef is defined on Spec section.
func genericToRoleBinding(o manifest.ObjectGeneric, v validator) (RoleBinding, error) {
	res := RoleBinding{
		APIVersion: o.ObjectHeader.APIVersion,
		Kind:       o.ObjectHeader.Kind,
		Metadata: manifest.RoleBindingMetadata{
			Name: o.Metadata.Name,
		},
		ObjectInternal: manifest.ObjectInternal{
			Organization: o.ObjectHeader.Organization,
			ManifestSrc:  o.ObjectHeader.ManifestSrc,
		},
	}
	var resSpec RoleBindingSpec
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

// Applying multiple Agents at once can cause timeout for whole sloctl apply command.
// This is caused by long request to Okta to create client credentials app.
// The same case is applicable for delete command.
const allowedAgentsToModify = 1

// Parse takes care of all Object supported by n9/v1alpha apiVersion
func Parse(o manifest.ObjectGeneric, parsedObjects *APIObjects, onlyHeaders bool) (err error) {
	v := NewValidator()
	switch o.Kind {
	case KindSLO:
		var slo SLO
		slo, err = genericToSLO(o, v, onlyHeaders)
		parsedObjects.SLOs = append(parsedObjects.SLOs, slo)
	case KindService:
		var service Service
		service, err = genericToService(o, v, onlyHeaders)
		parsedObjects.Services = append(parsedObjects.Services, service)
	case KindAgent:
		var agent Agent
		if len(parsedObjects.Agents) >= allowedAgentsToModify {
			err = manifest.EnhanceError(o, errors.New("only one Agent can be defined in this configuration"))
		} else {
			agent, err = genericToAgent(o, v, onlyHeaders)
			parsedObjects.Agents = append(parsedObjects.Agents, agent)
		}
	case KindAlertPolicy:
		var alertPolicy AlertPolicy
		alertPolicy, err = genericToAlertPolicy(o, v, onlyHeaders)
		parsedObjects.AlertPolicies = append(parsedObjects.AlertPolicies, alertPolicy)
	case KindAlertSilence:
		var alertSilence AlertSilence
		alertSilence, err = genericToAlertSilence(o, v, onlyHeaders)
		parsedObjects.AlertSilences = append(parsedObjects.AlertSilences, alertSilence)
	case KindAlertMethod:
		var alertMethod AlertMethod
		alertMethod, err = genericToAlertMethod(o, v, onlyHeaders)
		parsedObjects.AlertMethods = append(parsedObjects.AlertMethods, alertMethod)
	case KindDirect:
		var direct Direct
		direct, err = genericToDirect(o, v, onlyHeaders)
		parsedObjects.Directs = append(parsedObjects.Directs, direct)
	case KindDataExport:
		var dataExport DataExport
		dataExport, err = genericToDataExport(o, v, onlyHeaders)
		parsedObjects.DataExports = append(parsedObjects.DataExports, dataExport)
	case KindProject:
		var project Project
		project, err = genericToProject(o, v, onlyHeaders)
		parsedObjects.Projects = append(parsedObjects.Projects, project)
	case KindRoleBinding:
		var roleBinding RoleBinding
		roleBinding, err = genericToRoleBinding(o, v)
		parsedObjects.RoleBindings = append(parsedObjects.RoleBindings, roleBinding)
	case KindAnnotation:
		var annotation Annotation
		annotation, err = genericToAnnotation(o, v)
		parsedObjects.Annotations = append(parsedObjects.Annotations, annotation)
	case KindUserGroup:
		var group UserGroup
		group, err = genericToUserGroup(o)
		parsedObjects.UserGroups = append(parsedObjects.UserGroups, group)
	// catching invalid kinds of objects for this apiVersion
	default:
		err = manifest.UnsupportedKindErr(o)
	}
	return err
}

// Validate performs validation of parsed APIObjects.
func (o APIObjects) Validate() (err error) {
	var errs []error
	if err = validateUniquenessConstraints(KindSLO, o.SLOs); err != nil {
		errs = append(errs, err)
	}
	if err = validateUniquenessConstraints(KindService, o.Services); err != nil {
		errs = append(errs, err)
	}
	if err = validateUniquenessConstraints(KindProject, o.Projects); err != nil {
		errs = append(errs, err)
	}
	if err = validateUniquenessConstraints(KindAgent, o.Agents); err != nil {
		errs = append(errs, err)
	}
	if err = validateUniquenessConstraints(KindDirect, o.Directs); err != nil {
		errs = append(errs, err)
	}
	if err = validateUniquenessConstraints(KindAlertMethod, o.AlertMethods); err != nil {
		errs = append(errs, err)
	}
	if err = validateUniquenessConstraints(KindAlertPolicy, o.AlertPolicies); err != nil {
		errs = append(errs, err)
	}
	if err = validateUniquenessConstraints(KindAlertSilence, o.AlertSilences); err != nil {
		errs = append(errs, err)
	}
	if err = validateUniquenessConstraints(KindDataExport, o.DataExports); err != nil {
		errs = append(errs, err)
	}
	if err = validateUniquenessConstraints(KindRoleBinding, o.RoleBindings); err != nil {
		errs = append(errs, err)
	}
	if err = validateUniquenessConstraints(KindAnnotation, o.Annotations); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		builder := strings.Builder{}
		for i, err := range errs {
			builder.WriteString(err.Error())
			if i < len(errs)-1 {
				builder.WriteString("; ")
			}
		}
		return errors.New(builder.String())
	}
	return nil
}

// uniqueIdentifiers holds metadata used to uniquely identify an object across a single organization.
// While Name is required, Project might not apply to all objects.
type uniqueIdentifiers struct {
	Name    string
	Project string
}

// uniqueIdentifiersGetter allows generics to be used when iterating
// over all Kind slices like SLOsSlice or ServicesSlice.
type uniqueIdentifiersGetter interface {
	getUniqueIdentifiers() uniqueIdentifiers
}

// validateUniquenessConstraints finds conflicting objects in a Kind slice.
// It returns an error if any conflicts were encountered.
// The error informs about the cause and lists ALL conflicts.
func validateUniquenessConstraints[T uniqueIdentifiersGetter](kind Kind, slice []T) error {
	unique := make(map[string]struct{}, len(slice))
	var details []string
	for i := range slice {
		uid := slice[i].getUniqueIdentifiers()
		key := uid.Project + uid.Name
		if _, conflicts := unique[key]; conflicts {
			details = append(details, conflictDetails(kind, uid))
			continue
		}
		unique[key] = struct{}{}
	}
	if len(details) > 0 {
		return conflictError(kind, details)
	}
	return nil
}

// conflictDetails creates a formatted string identifying a single conflict between two objects.
func conflictDetails(kind Kind, uid uniqueIdentifiers) string {
	switch kind {
	case KindProject, KindRoleBinding:
		return fmt.Sprintf(`"%s"`, uid.Name)
	default:
		return fmt.Sprintf(`{"Project": "%s", "%s": "%s"}`, uid.Project, kind, uid.Name)
	}
}

// conflictError formats an error returned for a specific Kind with all it's conflicts listed as a JSON array.
// nolint: stylecheck
func conflictError(kind Kind, details []string) error {
	return fmt.Errorf(`Constraint "%s" was violated due to the following conflicts: [%s]`,
		constraintDetails(kind), strings.Join(details, ", "))
}

// constraintDetails creates a formatted string specifying the constraint which was broken.
func constraintDetails(kind Kind) string {
	switch kind {
	case KindProject, KindRoleBinding:
		return fmt.Sprintf(`%s.metadata.name has to be unique`, kind)
	default:
		return fmt.Sprintf(`%s.metadata.name has to be unique across a single Project`, kind)
	}
}
