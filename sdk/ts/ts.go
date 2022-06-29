// Package ts provides basic types work with time series in SDK
package ts

import (
	"fmt"
)

// TimeSeries represents a type of possible time series defined over an object kind
type TimeSeries int

// Possible time series that can be retrieved
const (
	InstantaneousBurnRate TimeSeries = iota + 1
	Counts
	BurnDown
	Percentiles
)

func getNamesToTimeSeriesMap() map[string]TimeSeries {
	return map[string]TimeSeries{
		"instantaneousBurnRate": InstantaneousBurnRate,
		"counts":                Counts,
		"burnDown":              BurnDown,
		"percentiles":           Percentiles,
	}
}

// Parse converts string to TimeSeries
func Parse(val string) (TimeSeries, error) {
	ts, ok := getNamesToTimeSeriesMap()[val]
	if !ok {
		return TimeSeries(0), fmt.Errorf("'%s' is not a valid time series", val)
	}
	return ts, nil
}

func (ts TimeSeries) String() string {
	for k, v := range getNamesToTimeSeriesMap() {
		if v == ts {
			return k
		}
	}
	return "UNKNOWN"
}
