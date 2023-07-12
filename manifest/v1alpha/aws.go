package v1alpha

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
		{RegionName: "Asia Pacific (Mumbai)", Code: "ap-south-1"},
		{RegionName: "Asia Pacific (Osaka)", Code: "ap-northeast-3"},
		{RegionName: "Asia Pacific (Seoul)", Code: "ap-northeast-2"},
		{RegionName: "Asia Pacific (Singapore)", Code: "ap-southeast-1"},
		{RegionName: "Asia Pacific (Sydney)", Code: "ap-southeast-2"},
		{RegionName: "Asia Pacific (Tokyo)", Code: "ap-northeast-1"},
		{RegionName: "Canada (Central)", Code: "ca-central-1"},
		{RegionName: "China (Beijing)", Code: "cn-north-1"},
		{RegionName: "China (Ningxia)", Code: "cn-northwest-1"},
		{RegionName: "Europe (Frankfurt)", Code: "eu-central-1"},
		{RegionName: "Europe (Ireland)", Code: "eu-west-1"},
		{RegionName: "Europe (London)", Code: "eu-west-2"},
		{RegionName: "Europe (Milan)", Code: "eu-south-1"},
		{RegionName: "Europe (Paris)", Code: "eu-west-3"},
		{RegionName: "Europe (Stockholm)", Code: "eu-north-1"},
		{RegionName: "Middle East (Bahrain)", Code: "me-south-1"},
		{RegionName: "South America (São Paulo)", Code: "sa-east-1"},
	}
}

func isValidRegion(code string, regions []AWSRegion) bool {
	for _, region := range regions {
		if region.Code == code {
			return true
		}
	}
	return false
}
