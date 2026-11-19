package v1alpha

//go:generate ../../bin/go-enum --values

// FailoverState is the failover state of a data source.
//
// ENUM(primary, failover)
type FailoverState string

// FailoverTarget is the place that a data source with failover reads from.
// Each data source type sets its own keys. CloudWatch sets Region.
type FailoverTarget struct {
	Region string `json:"region,omitempty"`
}

// FailoverStatus is the read-only failover state of a data source.
type FailoverStatus struct {
	State              FailoverState  `json:"state"`
	LastEventID        string         `json:"lastEventId,omitempty"`
	LastEventTimestamp string         `json:"lastEventTimestamp,omitempty"`
	Active             FailoverTarget `json:"active"`
}
