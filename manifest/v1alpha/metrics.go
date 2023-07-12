// Package v1alpha represents objects available in API n9/v1alpha
package v1alpha

import "sort"

func (slo *SLOSpec) containsIndicatorRawMetric() bool {
	return slo.Indicator.RawMetric != nil
}

// IsComposite returns true if SLOSpec contains composite type.
func (slo *SLOSpec) IsComposite() bool {
	return slo.Composite != nil
}

// HasRawMetric returns true if SLOSpec has raw metric.
func (slo *SLOSpec) HasRawMetric() bool {
	if slo.containsIndicatorRawMetric() {
		return true
	}
	for _, t := range slo.Thresholds {
		if t.HasRawMetricQuery() {
			return true
		}
	}
	return false
}

// RawMetrics returns raw metric spec.
func (slo *SLOSpec) RawMetrics() []*MetricSpec {
	if slo.containsIndicatorRawMetric() {
		return []*MetricSpec{slo.Indicator.RawMetric}
	}
	rawMetrics := make([]*MetricSpec, 0, slo.ObjectivesRawMetricsCount())
	for _, thresh := range slo.Thresholds {
		if thresh.RawMetric != nil {
			rawMetrics = append(rawMetrics, thresh.RawMetric.MetricQuery)
		}
	}
	return rawMetrics
}

// HasRawMetricQuery returns true if Threshold has raw metric with query set.
func (t *Threshold) HasRawMetricQuery() bool {
	return t.RawMetric != nil && t.RawMetric.MetricQuery != nil
}

// ObjectivesRawMetricsCount returns total number of all raw metrics defined in this SLO Spec's thresholds.
func (slo *SLOSpec) ObjectivesRawMetricsCount() int {
	var count int
	for _, thresh := range slo.Thresholds {
		if thresh.HasRawMetricQuery() {
			count++
		}
	}
	return count
}

// HasCountMetrics returns true if SLOSpec has count metrics.
func (slo *SLOSpec) HasCountMetrics() bool {
	for _, t := range slo.Thresholds {
		if t.HasCountMetrics() {
			return true
		}
	}
	return false
}

// HasCountMetrics returns true if Threshold has count metrics.
func (t *Threshold) HasCountMetrics() bool {
	return t.CountMetrics != nil
}

// CountMetricsCount returns total number of all count metrics defined in this SLOSpec's thresholds.
func (slo *SLOSpec) CountMetricsCount() int {
	var count int
	for _, thresh := range slo.Thresholds {
		if thresh.CountMetrics != nil {
			if thresh.CountMetrics.GoodMetric != nil {
				count++
			}
			if thresh.CountMetrics.TotalMetric != nil {
				count++
			}
			if thresh.CountMetrics.BadMetric != nil && isBadOverTotalEnabledForDataSourceType(thresh) {
				count++
			}
		}
	}
	return count
}

// CountMetrics returns a flat slice of all count metrics defined in this SLOSpec's thresholds.
func (slo *SLOSpec) CountMetrics() []*MetricSpec {
	countMetrics := make([]*MetricSpec, slo.CountMetricsCount())
	var i int
	for _, thresh := range slo.Thresholds {
		if thresh.CountMetrics == nil {
			continue
		}
		if thresh.CountMetrics.GoodMetric != nil {
			countMetrics[i] = thresh.CountMetrics.GoodMetric
			i++
		}
		if thresh.CountMetrics.TotalMetric != nil {
			countMetrics[i] = thresh.CountMetrics.TotalMetric
			i++
		}
		if thresh.CountMetrics.BadMetric != nil && isBadOverTotalEnabledForDataSourceType(thresh) {
			countMetrics[i] = thresh.CountMetrics.BadMetric
			i++
		}
	}
	return countMetrics
}

// CountMetricPairs returns a slice of all count metrics defined in this SLOSpec's thresholds.
func (slo *SLOSpec) CountMetricPairs() []*CountMetricsSpec {
	countMetrics := make([]*CountMetricsSpec, slo.CountMetricsCount())
	var i int
	for _, thresh := range slo.Thresholds {
		if thresh.CountMetrics == nil {
			continue
		}
		if thresh.CountMetrics.GoodMetric != nil && thresh.CountMetrics.TotalMetric != nil {
			countMetrics[i] = thresh.CountMetrics
			i++
		}
	}
	return countMetrics
}

func (slo *SLOSpec) GoodTotalCountMetrics() (good, total []*MetricSpec) {
	for _, thresh := range slo.Thresholds {
		if thresh.CountMetrics == nil {
			continue
		}
		if thresh.CountMetrics.GoodMetric != nil {
			good = append(good, thresh.CountMetrics.GoodMetric)
		}
		if thresh.CountMetrics.TotalMetric != nil {
			total = append(total, thresh.CountMetrics.TotalMetric)
		}
	}
	return
}

// AllMetricSpecs returns slice of all metrics defined in SLO regardless of their type.
func (slo *SLOSpec) AllMetricSpecs() []*MetricSpec {
	var metrics []*MetricSpec
	metrics = append(metrics, slo.RawMetrics()...)
	metrics = append(metrics, slo.CountMetrics()...)
	return metrics
}

// DataSourceType returns a type of data source.
func (m *MetricSpec) DataSourceType() DataSourceType {
	switch {
	case m.Prometheus != nil:
		return Prometheus
	case m.Datadog != nil:
		return Datadog
	case m.NewRelic != nil:
		return NewRelic
	case m.AppDynamics != nil:
		return AppDynamics
	case m.Splunk != nil:
		return Splunk
	case m.Lightstep != nil:
		return Lightstep
	case m.SplunkObservability != nil:
		return SplunkObservability
	case m.Dynatrace != nil:
		return Dynatrace
	case m.Elasticsearch != nil:
		return Elasticsearch
	case m.ThousandEyes != nil:
		return ThousandEyes
	case m.Graphite != nil:
		return Graphite
	case m.BigQuery != nil:
		return BigQuery
	case m.OpenTSDB != nil:
		return OpenTSDB
	case m.GrafanaLoki != nil:
		return GrafanaLoki
	case m.CloudWatch != nil:
		return CloudWatch
	case m.Pingdom != nil:
		return Pingdom
	case m.AmazonPrometheus != nil:
		return AmazonPrometheus
	case m.Redshift != nil:
		return Redshift
	case m.SumoLogic != nil:
		return SumoLogic
	case m.Instana != nil:
		return Instana
	case m.InfluxDB != nil:
		return InfluxDB
	case m.GCM != nil:
		return GCM
	default:
		return 0
	}
}

// Query returns interface containing metric query for this MetricSpec.
func (m *MetricSpec) Query() interface{} {
	switch m.DataSourceType() {
	case Prometheus:
		return m.Prometheus
	case Datadog:
		return m.Datadog
	case NewRelic:
		return m.NewRelic
	case AppDynamics:
		return m.AppDynamics
	case Splunk:
		return m.Splunk
	case Lightstep:
		return m.Lightstep
	case SplunkObservability:
		return m.SplunkObservability
	case Dynatrace:
		return m.Dynatrace
	case Elasticsearch:
		return m.Elasticsearch
	case ThousandEyes:
		return m.ThousandEyes
	case Graphite:
		return m.Graphite
	case BigQuery:
		return m.BigQuery
	case OpenTSDB:
		return m.OpenTSDB
	case GrafanaLoki:
		return m.GrafanaLoki
	case CloudWatch:
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
	case Pingdom:
		return m.Pingdom
	case AmazonPrometheus:
		return m.AmazonPrometheus
	case Redshift:
		return m.Redshift
	case SumoLogic:
		return m.SumoLogic
	case Instana:
		return m.Instana
	case InfluxDB:
		return m.InfluxDB
	case GCM:
		return m.GCM
	default:
		return nil
	}
}
