package v1alpha

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

type AWSRegion struct {
	RegionName string `json:"regionName"`
	Code       string `json:"code"`
}

// AWSRegions returns list of all AWS regions. Data is taken from:
// https://docs.aws.amazon.com/general/latest/gr/rande.html
func AWSRegions() []AWSRegion {
	return []AWSRegion{
		{RegionName: "US East (Ohio)", Code: "us-east-2"},
		{RegionName: "US East (N. Virginia)", Code: "us-east-1"},
		{RegionName: "US West (N. California)", Code: "us-west-1"},
		{RegionName: "US West (Oregon)", Code: "us-west-2"},
		{RegionName: "Africa (Cape Town)", Code: "af-south-1"},
		{RegionName: "Asia Pacific (Hong Kong)", Code: "ap-east-1"},
		{RegionName: "Asia Pacific (Hyderabad)", Code: "ap-south-2"},
		{RegionName: "Asia Pacific (Jakarta)", Code: "ap-southeast-3"},
		{RegionName: "Asia Pacific (Malaysia)", Code: "ap-southeast-5"},
		{RegionName: "Asia Pacific (Melbourne)", Code: "ap-southeast-4"},
		{RegionName: "Asia Pacific (Mumbai)", Code: "ap-south-1"},
		{RegionName: "Asia Pacific (New Zealand)", Code: "ap-southeast-6"},
		{RegionName: "Asia Pacific (Osaka)", Code: "ap-northeast-3"},
		{RegionName: "Asia Pacific (Seoul)", Code: "ap-northeast-2"},
		{RegionName: "Asia Pacific (Singapore)", Code: "ap-southeast-1"},
		{RegionName: "Asia Pacific (Sydney)", Code: "ap-southeast-2"},
		{RegionName: "Asia Pacific (Taipei)", Code: "ap-east-2"},
		{RegionName: "Asia Pacific (Thailand)", Code: "ap-southeast-7"},
		{RegionName: "Asia Pacific (Tokyo)", Code: "ap-northeast-1"},
		{RegionName: "Canada (Central)", Code: "ca-central-1"},
		{RegionName: "Canada West (Calgary)", Code: "ca-west-1"},
		{RegionName: "China (Beijing)", Code: "cn-north-1"},
		{RegionName: "China (Ningxia)", Code: "cn-northwest-1"},
		{RegionName: "Europe (Frankfurt)", Code: "eu-central-1"},
		{RegionName: "Europe (Ireland)", Code: "eu-west-1"},
		{RegionName: "Europe (London)", Code: "eu-west-2"},
		{RegionName: "Europe (Milan)", Code: "eu-south-1"},
		{RegionName: "Europe (Paris)", Code: "eu-west-3"},
		{RegionName: "Europe (Spain)", Code: "eu-south-2"},
		{RegionName: "Europe (Stockholm)", Code: "eu-north-1"},
		{RegionName: "Europe (Zurich)", Code: "eu-central-2"},
		{RegionName: "Israel (Tel Aviv)", Code: "il-central-1"},
		{RegionName: "Mexico (Central)", Code: "mx-central-1"},
		{RegionName: "Middle East (Bahrain)", Code: "me-south-1"},
		{RegionName: "Middle East (Saudi Arabia)", Code: "me-west-1"},
		{RegionName: "Middle East (UAE)", Code: "me-central-1"},
		{RegionName: "South America (Chile)", Code: "sa-west-1"},
		{RegionName: "South America (São Paulo)", Code: "sa-east-1"},
	}
}

// AWSRegionCodes returns the codes of [AWSRegions].
func AWSRegionCodes() []string {
	regions := AWSRegions()
	codes := make([]string, 0, len(regions))
	for _, region := range regions {
		codes = append(codes, region.Code)
	}
	return codes
}

// AWSRegionRules validates that a value is one of [AWSRegionCodes].
func AWSRegionRules() govy.RuleSet[string] {
	return govy.NewRuleSet(
		rules.StringMaxLength(255),
		rules.OneOf(AWSRegionCodes()...),
	)
}

func IsValidRegion(code string, regions []AWSRegion) bool {
	for _, region := range regions {
		if region.Code == code {
			return true
		}
	}
	return false
}
