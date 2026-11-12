package v1alpha

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

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

// CloudWatchFailoverConfig arms region failover for a CloudWatch data source.
// Only a failover event switches the active region.
type CloudWatchFailoverConfig struct {
	PrimaryRegion  string `json:"primaryRegion"`
	FailoverRegion string `json:"failoverRegion"`
}

func CloudWatchFailoverValidation() govy.Validator[CloudWatchFailoverConfig] {
	return govy.New[CloudWatchFailoverConfig](
		govy.For(govy.GetSelf[CloudWatchFailoverConfig]()).
			Rules(rules.UniqueProperties(rules.HashFuncSelf[string](), map[string]func(c CloudWatchFailoverConfig) string{
				"primaryRegion":  func(c CloudWatchFailoverConfig) string { return c.PrimaryRegion },
				"failoverRegion": func(c CloudWatchFailoverConfig) string { return c.FailoverRegion },
			})).
			When(
				func(c CloudWatchFailoverConfig) bool { return c.PrimaryRegion != "" && c.FailoverRegion != "" },
				govy.WhenDescription("both regions are set"),
			),
		govy.For(func(c CloudWatchFailoverConfig) string { return c.PrimaryRegion }).
			WithName("primaryRegion").
			Required().
			Rules(rules.OneOf(AWSRegionCodes()...)),
		govy.For(func(c CloudWatchFailoverConfig) string { return c.FailoverRegion }).
			WithName("failoverRegion").
			Required().
			Rules(rules.OneOf(AWSRegionCodes()...)),
	)
}
