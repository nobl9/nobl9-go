package v1alpha

// FailoverState is the failover state of a data source.
type FailoverState string

const (
	FailoverStatePrimary  FailoverState = "primary"
	FailoverStateFailover FailoverState = "failover"
)

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
