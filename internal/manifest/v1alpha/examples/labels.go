package v1alphaExamples

import "github.com/nobl9/nobl9-go/manifest/v1alpha"

func Labels() v1alpha.Labels {
	return v1alpha.Labels{
		"team":   {"green", "sales"},
		"env":    {"prod", "dev"},
		"region": {"us", "eu"},
		"area":   {"latency", "slow-check"},
	}
}
