package v1alpha

import (
	"testing"

	v "github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestValidateURLDynatrace(t *testing.T) {
	testCases := []struct {
		desc    string
		url     string
		isValid bool
	}{
		{
			desc:    "valid SaaS",
			url:     "https://test.live.dynatrace.com",
			isValid: true,
		},
		{
			desc:    "valid SaaS with port explicit speciefed",
			url:     "https://test.live.dynatrace.com:433",
			isValid: true,
		},
		{
			desc:    "valid SaaS multiple trailing /",
			url:     "https://test.live.dynatrace.com///",
			isValid: true,
		},
		{
			desc:    "invalid SaaS lack of https",
			url:     "http://test.live.dynatrace.com",
			isValid: false,
		},
		{
			desc:    "valid Managed/Environment ActiveGate lack of https",
			url:     "http://test.com/e/environment-id",
			isValid: true,
		},
		{
			desc:    "valid Managed/Environment ActiveGate wrong environment-id",
			url:     "https://test.com/e/environment-id",
			isValid: true,
		},
		{
			desc:    "valid Managed/Environment ActiveGate IP",
			url:     "https://127.0.0.1/e/environment-id",
			isValid: true,
		},
		{
			desc:    "valid Managed/Environment ActiveGate wrong environment-id",
			url:     "https://test.com/some-devops-path/e/environment-id",
			isValid: true,
		},
		{
			desc:    "valid Managed/Environment ActiveGate wrong environment-id, multiple /",
			url:     "https://test.com///some-devops-path///e///environment-id///",
			isValid: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			assert.Equal(t, tC.isValid, validateURLDynatrace(tC.url))
		})
	}
}

func TestAlertSilencePeriodValidation(t *testing.T) {
	validate := v.New()
	validate.RegisterStructValidation(alertSilencePeriodValidation, AlertSilencePeriod{})

	testCases := []struct {
		desc    string
		spec    AlertSilencePeriod
		isValid bool
	}{
		{
			desc: "endTime before starTime",
			spec: AlertSilencePeriod{
				StartTime: "2006-01-02T17:04:05Z",
				EndTime:   "2006-01-02T17:00:05Z",
			},
			isValid: false,
		},
		{
			desc: "endTime equals starTime",
			spec: AlertSilencePeriod{
				StartTime: "2006-01-02T17:00:05Z",
				EndTime:   "2006-01-02T17:00:05Z",
			},
			isValid: false,
		},
		{
			desc: "endTime after starTime",
			spec: AlertSilencePeriod{
				StartTime: "2006-01-02T17:00:05Z",
				EndTime:   "2006-01-02T17:04:05Z",
			},
			isValid: true,
		},
		{
			desc: "both endTime and duration are provided",
			spec: AlertSilencePeriod{
				EndTime:  "2006-01-02T17:04:05Z",
				Duration: "1h",
			},
			isValid: false,
		},
		{
			desc: "both endTime and duration are missing",
			spec: AlertSilencePeriod{
				StartTime: "2006-01-02T17:04:05Z",
			},
			isValid: false,
		},
		{
			desc: "negative value for duration",
			spec: AlertSilencePeriod{
				Duration: "-1h",
			},
			isValid: false,
		},
		{
			desc: "zero value for duration",
			spec: AlertSilencePeriod{
				Duration: "0",
			},
			isValid: false,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			err := validate.Struct(tC.spec)
			if tC.isValid {
				assert.Nil(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestIsReleaseChannelValid(t *testing.T) {
	for name, test := range map[string]struct {
		ReleaseChannel ReleaseChannel
		IsValid        bool
	}{
		"unset release channel, valid": {IsValid: true},
		"beta channel, valid":          {ReleaseChannel: ReleaseChannelBeta, IsValid: true},
		"stable channel, valid":        {ReleaseChannel: ReleaseChannelStable, IsValid: true},
		"alpha channel, invalid":       {ReleaseChannel: ReleaseChannelAlpha},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.IsValid, isValidReleaseChannel(test.ReleaseChannel))
		})
	}
}
