package v1alphaExamples

import "github.com/nobl9/nobl9-go/manifest/v1alpha"

func Labels() []Example {
	return newExampleSlice(standardExample{
		Object: exampleLabels(),
	})
}

func exampleLabels() v1alpha.Labels {
	return v1alpha.Labels{
		"team":   {"green", "sales"},
		"env":    {"prod", "dev"},
		"region": {"us", "eu"},
		"area":   {"latency", "slow-check"},
	}
}

func exampleCompositeLabels() v1alpha.Labels {
	return v1alpha.Labels{
		"team":   {"green", "ux"},
		"env":    {"prod", "dev"},
		"region": {"us", "eu"},
		"area":   {"user-experience"},
	}
}
