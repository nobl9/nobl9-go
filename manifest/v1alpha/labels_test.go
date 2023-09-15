package v1alpha

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateLabels(t *testing.T) {
	testCases := []struct {
		desc    string
		labels  Labels
		isValid bool
	}{
		{
			desc: "valid: simple strings",
			labels: map[string][]string{
				"net":     {"vast", "infinite"},
				"project": {"nobl9"},
			},
			isValid: true,
		},
		{
			desc: "invalid: empty label key",
			labels: map[string][]string{
				"": {"vast", "infinite"},
			},
			isValid: false,
		},
		{
			desc: "valid: one empty label value",
			labels: map[string][]string{
				"net": {""},
			},
			isValid: true,
		},
		{
			desc: "invalid: two empty label values (because duplicates)",
			labels: map[string][]string{
				"net": {"", ""},
			},
			isValid: false,
		},
		{
			desc: "valid: no label values for a given key",
			labels: map[string][]string{
				"net": {},
			},
			isValid: true,
		},
		{
			desc: "valid: no label values for a given key",
			labels: map[string][]string{
				"net": {},
			},
			isValid: true,
		},
		{
			desc: "invalid: label key is too long (over 63 chars)",
			labels: map[string][]string{
				"netnetnetnetnetnetnetnetnetnetnetnetnetnetnetnetnetnetnetnetnetnetnet": {},
			},
			isValid: false,
		},
		{
			desc: "invalid: label key starts with non letter",
			labels: map[string][]string{
				"9net": {},
			},
			isValid: false,
		},
		{
			desc: "invalid: label key ends with non alphanumeric char",
			labels: map[string][]string{
				"net_": {},
			},
			isValid: false,
		},
		{
			desc: "invalid: label key contains uppercase character",
			labels: map[string][]string{
				"nEt": {},
			},
			isValid: false,
		},
		{
			desc: "invalid: label value is to long (over 200 chars)",
			labels: map[string][]string{
				"net": {`
					labellabellabellabellabellabellabellabellabellabellabellabellabellabellabellabellabellabellabel
					labellabellabellabellabellabellabellabellabellabellabellabellabellabellabellabellabellabellabel
					labellabellabellabellabellabellabellabellabellabellabellabellabellabellabellabellabellabel
				`},
			},
			isValid: false,
		},
		{
			desc: "valid: label value with uppercase characters",
			labels: map[string][]string{
				"net": {"THE NET is vast AND INFINITE"},
			},
			isValid: true,
		},
		{
			desc: "valid: label value with uppercase characters",
			labels: map[string][]string{
				"net": {"the-net-is-vast-and-infinite"},
			},
			isValid: true,
		},
		{
			desc: "valid: any unicode with rune count 1-200",
			labels: map[string][]string{
				"net": {"\uE005[\\\uE006\uE007"},
			},
			isValid: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			err := tC.labels.Validate()
			assert.Equal(t, tC.isValid, err == nil)
		})
	}
}
