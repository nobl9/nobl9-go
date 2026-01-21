package slo

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

// CountMetricsSpec represents set of two time series of good and total counts
type CountMetricsSpec struct {
	Incremental *bool       `json:"incremental"`
	GoodMetric  *MetricSpec `json:"good,omitempty"`
	BadMetric   *MetricSpec `json:"bad,omitempty"`
	TotalMetric *MetricSpec `json:"total,omitempty"`
	// Experimental: Splunk and Honeycomb only.
	// Single query returning both good and total counts.
	GoodTotalMetric *MetricSpec `json:"goodTotal,omitempty"`
}

// RawMetricSpec represents integration with a metric source for a particular objective.
type RawMetricSpec struct {
	MetricQuery *MetricSpec `json:"query"`
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
	AzureMonitor        *AzureMonitorMetric        `json:"azureMonitor,omitempty"`
	Generic             *GenericMetric             `json:"generic,omitempty"`
	Honeycomb           *HoneycombMetric           `json:"honeycomb,omitempty"`
	LogicMonitor        *LogicMonitorMetric        `json:"logicMonitor,omitempty"`
	AzurePrometheus     *AzurePrometheusMetric     `json:"azurePrometheus,omitempty"`
	Coralogix           *CoralogixMetric           `json:"coralogix,omitempty"`
	Atlas               *AtlasMetric               `json:"atlas,omitempty"`
}

func (s *Spec) containsIndicatorRawMetric() bool {
	return s.Indicator != nil && s.Indicator.RawMetric != nil
}

// IsComposite returns true if SLOSpec contains composite type.
//
// Deprecated: this implementation of Composite will be removed and replaced with new CompositeSpec
// use HasCompositeObjectives instead for new implementation
func (s *Spec) IsComposite() bool {
	return s.Composite != nil
}

// HasCompositeObjectives returns true if any SLOSpec Objective is of composite type.
func (s *Spec) HasCompositeObjectives() bool {
	for _, obj := range s.Objectives {
		if obj.IsComposite() {
			return true
		}
	}
	return false
}

// HasRawMetric returns true if SLOSpec has raw metric.
func (s *Spec) HasRawMetric() bool {
	if s.containsIndicatorRawMetric() {
		return true
	}
	for _, objective := range s.Objectives {
		if objective.HasRawMetricQuery() {
			return true
		}
	}
	return false
}

// RawMetrics returns raw metric spec.
func (s *Spec) RawMetrics() []*MetricSpec {
	if s.containsIndicatorRawMetric() {
		return []*MetricSpec{s.Indicator.RawMetric}
	}
	rawMetrics := make([]*MetricSpec, 0, s.ObjectivesRawMetricsCount())
	for _, objective := range s.Objectives {
		if objective.RawMetric != nil {
			rawMetrics = append(rawMetrics, objective.RawMetric.MetricQuery)
		}
	}
	return rawMetrics
}

// ObjectivesRawMetricsCount returns total number of all raw metrics defined in this SLO Spec's objectives.
func (s *Spec) ObjectivesRawMetricsCount() int {
	var count int
	for _, objective := range s.Objectives {
		if objective.HasRawMetricQuery() {
			count++
		}
	}
	return count
}

// HasCountMetrics returns true if SLOSpec has count metrics.
func (s *Spec) HasCountMetrics() bool {
	for _, objective := range s.Objectives {
		if objective.HasCountMetrics() {
			return true
		}
	}
	return false
}

// CountMetricsCount returns total number of all count metrics defined in this SLOSpec's objectives.
func (s *Spec) CountMetricsCount() int {
	var count int
	for _, objective := range s.Objectives {
		if objective.CountMetrics != nil {
			if objective.CountMetrics.GoodMetric != nil {
				count++
			}
			if objective.CountMetrics.TotalMetric != nil {
				count++
			}
			if objective.CountMetrics.BadMetric != nil {
				count++
			}
			if objective.CountMetrics.GoodTotalMetric != nil {
				count++
			}
		}
	}
	return count
}

// CountMetrics returns a flat slice of all count metrics defined in this SLOSpec's objectives.
func (s *Spec) CountMetrics() []*MetricSpec {
	countMetrics := make([]*MetricSpec, s.CountMetricsCount())
	var i int
	for _, objective := range s.Objectives {
		if objective.CountMetrics == nil {
			continue
		}
		if objective.CountMetrics.GoodMetric != nil {
			countMetrics[i] = objective.CountMetrics.GoodMetric
			i++
		}
		if objective.CountMetrics.TotalMetric != nil {
			countMetrics[i] = objective.CountMetrics.TotalMetric
			i++
		}
		if objective.CountMetrics.BadMetric != nil {
			countMetrics[i] = objective.CountMetrics.BadMetric
			i++
		}
		if objective.CountMetrics.GoodTotalMetric != nil {
			countMetrics[i] = objective.CountMetrics.GoodTotalMetric
			i++
		}
	}
	return countMetrics
}

// CountMetricPairs returns a slice of all count metrics defined in this SLOSpec's objectives.
func (s *Spec) CountMetricPairs() []*CountMetricsSpec {
	countMetrics := make([]*CountMetricsSpec, s.CountMetricsCount())
	var i int
	for _, objective := range s.Objectives {
		if objective.CountMetrics == nil {
			continue
		}
		if objective.CountMetrics.GoodMetric != nil && objective.CountMetrics.TotalMetric != nil {
			countMetrics[i] = objective.CountMetrics
			i++
		}
	}
	return countMetrics
}

func (s *Spec) GoodTotalCountMetrics() (good, total []*MetricSpec) {
	for _, objective := range s.Objectives {
		if objective.CountMetrics == nil {
			continue
		}
		if objective.CountMetrics.GoodMetric != nil {
			good = append(good, objective.CountMetrics.GoodMetric)
		}
		if objective.CountMetrics.TotalMetric != nil {
			total = append(total, objective.CountMetrics.TotalMetric)
		}
	}
	return good, total
}

// AllMetricSpecs returns slice of all metrics defined in SLO regardless of their type.
func (s *Spec) AllMetricSpecs() []*MetricSpec {
	var metrics []*MetricSpec
	metrics = append(metrics, s.RawMetrics()...)
	metrics = append(metrics, s.CountMetrics()...)
	return metrics
}

// DataSourceType returns a type of data source.
func (m *MetricSpec) DataSourceType() v1alpha.DataSourceType {
	switch {
	case m.Prometheus != nil:
		return v1alpha.Prometheus
	case m.Datadog != nil:
		return v1alpha.Datadog
	case m.NewRelic != nil:
		return v1alpha.NewRelic
	case m.AppDynamics != nil:
		return v1alpha.AppDynamics
	case m.Splunk != nil:
		return v1alpha.Splunk
	case m.Lightstep != nil:
		return v1alpha.Lightstep
	case m.SplunkObservability != nil:
		return v1alpha.SplunkObservability
	case m.Dynatrace != nil:
		return v1alpha.Dynatrace
	case m.Elasticsearch != nil:
		return v1alpha.Elasticsearch
	case m.ThousandEyes != nil:
		return v1alpha.ThousandEyes
	case m.Graphite != nil:
		return v1alpha.Graphite
	case m.BigQuery != nil:
		return v1alpha.BigQuery
	case m.OpenTSDB != nil:
		return v1alpha.OpenTSDB
	case m.GrafanaLoki != nil:
		return v1alpha.GrafanaLoki
	case m.CloudWatch != nil:
		return v1alpha.CloudWatch
	case m.Pingdom != nil:
		return v1alpha.Pingdom
	case m.AmazonPrometheus != nil:
		return v1alpha.AmazonPrometheus
	case m.Redshift != nil:
		return v1alpha.Redshift
	case m.SumoLogic != nil:
		return v1alpha.SumoLogic
	case m.Instana != nil:
		return v1alpha.Instana
	case m.InfluxDB != nil:
		return v1alpha.InfluxDB
	case m.GCM != nil:
		return v1alpha.GCM
	case m.AzureMonitor != nil:
		return v1alpha.AzureMonitor
	case m.Generic != nil:
		return v1alpha.Generic
	case m.Honeycomb != nil:
		return v1alpha.Honeycomb
	case m.LogicMonitor != nil:
		return v1alpha.LogicMonitor
	case m.AzurePrometheus != nil:
		return v1alpha.AzurePrometheus
	case m.Coralogix != nil:
		return v1alpha.Coralogix
	case m.Atlas != nil:
		return v1alpha.Atlas
	default:
		return 0
	}
}

// Query returns interface containing metric query for this MetricSpec.
// nolint: gocyclo
func (m *MetricSpec) Query() interface{} {
	switch m.DataSourceType() {
	case v1alpha.Prometheus:
		return m.Prometheus
	case v1alpha.Datadog:
		return m.Datadog
	case v1alpha.NewRelic:
		return m.NewRelic
	case v1alpha.AppDynamics:
		return m.AppDynamics
	case v1alpha.Splunk:
		return m.Splunk
	case v1alpha.Lightstep:
		return m.Lightstep
	case v1alpha.SplunkObservability:
		return m.SplunkObservability
	case v1alpha.Dynatrace:
		return m.Dynatrace
	case v1alpha.Elasticsearch:
		return m.Elasticsearch
	case v1alpha.ThousandEyes:
		return m.ThousandEyes
	case v1alpha.Graphite:
		return m.Graphite
	case v1alpha.BigQuery:
		return m.BigQuery
	case v1alpha.OpenTSDB:
		return m.OpenTSDB
	case v1alpha.GrafanaLoki:
		return m.GrafanaLoki
	case v1alpha.CloudWatch:
		// To be clean, entire metric spec is copied so that original value is not mutated.
		var cloudWatchCopy CloudWatchMetric
		cloudWatchCopy = *m.CloudWatch
		// Dimension list is optional. This is done so that during upsert empty slice and nil slice are treated equally.
		if cloudWatchCopy.Dimensions == nil {
			cloudWatchCopy.Dimensions = []CloudWatchMetricDimension{}
		}
		// Dimensions are sorted so that metric_query = '...':jsonb comparison was insensitive to the order in slice.
		// It assumes that all dimensions' names are unique (ensured by validation).
		sort.Slice(cloudWatchCopy.Dimensions, func(i, j int) bool {
			return *cloudWatchCopy.Dimensions[i].Name < *cloudWatchCopy.Dimensions[j].Name
		})
		return cloudWatchCopy
	case v1alpha.Pingdom:
		return m.Pingdom
	case v1alpha.AmazonPrometheus:
		return m.AmazonPrometheus
	case v1alpha.Redshift:
		return m.Redshift
	case v1alpha.SumoLogic:
		return m.SumoLogic
	case v1alpha.Instana:
		return m.Instana
	case v1alpha.InfluxDB:
		return m.InfluxDB
	case v1alpha.GCM:
		return m.GCM
	case v1alpha.AzureMonitor:
		// To be clean, entire metric spec is copied so that original value is not mutated.
		var azureMonitorCopy AzureMonitorMetric
		azureMonitorCopy = *m.AzureMonitor
		// Dimension list is optional. This is done so that during upsert empty slice and nil slice are treated equally.
		if azureMonitorCopy.Dimensions == nil {
			azureMonitorCopy.Dimensions = []AzureMonitorMetricDimension{}
		}
		// Dimensions are sorted so that metric_query = '...':jsonb comparison was insensitive to the order in slice.
		// It assumes that all dimensions' names are unique (ensured by validation).
		sort.Slice(azureMonitorCopy.Dimensions, func(i, j int) bool {
			return *azureMonitorCopy.Dimensions[i].Name < *azureMonitorCopy.Dimensions[j].Name
		})
		return azureMonitorCopy
	case v1alpha.Generic:
		return m.Generic
	case v1alpha.Honeycomb:
		return m.Honeycomb
	case v1alpha.LogicMonitor:
		return m.LogicMonitor
	case v1alpha.AzurePrometheus:
		return m.AzurePrometheus
	case v1alpha.Coralogix:
		return m.Coralogix
	case v1alpha.Atlas:
		return m.Atlas
	default:
		return nil
	}
}

func (m *MetricSpec) FormatQuery(query json.RawMessage) *string {
	switch m.DataSourceType() {
	case v1alpha.Dynatrace:
		return m.Dynatrace.MetricSelector
	case v1alpha.Datadog:
		return m.Datadog.Query
	case v1alpha.NewRelic:
		return m.NewRelic.NRQL
	case v1alpha.SplunkObservability:
		return m.SplunkObservability.Program
	case v1alpha.SumoLogic:
		return m.SumoLogic.Query
	default:
		var formattedQuery string
		if len(query) > 0 {
			formattedQuery = formatRawJSONMetricQueryToString(query)
		}
		return &formattedQuery
	}
}

func formatRawJSONMetricQueryToString(queryAsJSON []byte) string {
	var formatJSONToString func(jsonObjAny any, prefix string) string
	var queryObj map[string]any

	err := json.Unmarshal(queryAsJSON, &queryObj)
	if err != nil {
		return ""
	}

	toTitle := func(s string) string {
		return cases.Title(language.English).String(s)
	}

	formatJSONToString = func(jsonObjAny any, prefix string) string {
		var sb strings.Builder

		switch jsonObj := jsonObjAny.(type) {
		case string:
			sb.WriteString(fmt.Sprintf("%s\n", jsonObj))
		case float64:
			number := jsonObj
			if math.Floor(number) == number {
				sb.WriteString(fmt.Sprintf("%d\n", int64(number)))
			} else {
				sb.WriteString(fmt.Sprintf("%f\n", number))
			}
		case map[string]any:
			keys := make([]string, 0, len(jsonObj))
			for k := range jsonObj {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				sb.WriteString(
					fmt.Sprintf("%s%s: %s", prefix, toTitle(k), formatJSONToString(jsonObj[k], prefix)),
				)
			}
		case []any:
			sb.WriteString("\n")
			prefix += " "
			for i, val := range jsonObj {
				sb.WriteString(
					fmt.Sprintf("%s%d:\n%s", prefix, i+1, formatJSONToString(val, prefix+" ")),
				)
			}
		default:
			sb.WriteString("")
		}

		return sb.String()
	}

	return formatJSONToString(queryObj, "")
}
