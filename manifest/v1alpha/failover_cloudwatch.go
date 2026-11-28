package v1alpha

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// CloudWatchFailoverConfig arms region failover for a CloudWatch data source.
// Only a failover event switches the active region.
type CloudWatchFailoverConfig struct {
	PrimaryRegion  string `json:"primaryRegion"`
	FailoverRegion string `json:"failoverRegion"`
}

func CloudWatchFailoverValidation() govy.Validator[CloudWatchFailoverConfig] {
	return cloudWatchFailoverValidation
}

var cloudWatchFailoverValidation = govy.New[CloudWatchFailoverConfig](
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
		Rules(AWSRegionRules()),
	govy.For(func(c CloudWatchFailoverConfig) string { return c.FailoverRegion }).
		WithName("failoverRegion").
		Required().
		Rules(AWSRegionRules()),
)
