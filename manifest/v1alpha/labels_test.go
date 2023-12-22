package v1alpha

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// nolint: lll
func TestValidateLabels(t *testing.T) {
	for name, test := range map[string]struct {
		Labels Labels
		Error  error
	}{
		"valid: simple strings": {
			Labels: map[labelKey][]labelValue{
				"net":     {"vast", "infinite"},
				"project": {"nobl9"},
			},
		},
		"invalid: empty label key": {
			Labels: map[labelKey][]labelValue{
				"": {"vast", "infinite"},
			},
			Error: errors.New("label key '' length must be between 1 and 63"),
		},
		"valid: one empty label value": {
			Labels: map[labelKey][]labelValue{
				"net": {""},
			},
		},
		"invalid: label value duplicates": {
			Labels: map[labelKey][]labelValue{
				"net": {"same", "same", "same"},
			},
			Error: errors.New("label value 'same' for key 'net' already exists, duplicates are not allowed"),
		},
		"invalid: two empty label values (because duplicates)": {
			Labels: map[labelKey][]labelValue{
				"net": {"", ""},
			},
			Error: errors.New("label value '' for key 'net' already exists, duplicates are not allowed"),
		},
		"valid: no label values for a given key": {
			Labels: map[labelKey][]labelValue{
				"net": {},
			},
		},
		"invalid: label key is too long": {
			Labels: map[labelKey][]labelValue{
				"net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net": {},
			},
			Error: errors.New("label key 'net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net' length must be between 1 and 63"),
		},
		"invalid: label key starts with non letter": {
			Labels: map[labelKey][]labelValue{
				"9net": {},
			},
			Error: errors.New("label key '9net' does not match the regex: ^\\p{L}([_\\-0-9\\p{L}]*[0-9\\p{L}])?$"),
		},
		"invalid: label key ends with non alphanumeric char": {
			Labels: map[labelKey][]labelValue{
				"net_": {},
			},
			Error: errors.New("label key 'net_' does not match the regex: ^\\p{L}([_\\-0-9\\p{L}]*[0-9\\p{L}])?$"),
		},
		"invalid: label key contains uppercase character": {
			Labels: map[labelKey][]labelValue{
				"nEt": {},
			},
			Error: errors.New("label key 'nEt' must not have upper case letters"),
		},
		"invalid: label value is to long (over 200 chars)": {
			Labels: map[labelKey][]labelValue{
				"net": {`
					label-label-label-label-label-label-label-label-label-label-label-label-label-label-label-label-
					label-label-label-label-label-label-label-label-label-label-label-label-label-label-label-label-
					label-label-label-label-label-label-label-label-label-label-label-label-label-label-label-label
				`},
			},
			Error: errors.New("label value '\n\t\t\t\t\tlabel-label-label-label-label-label-label-label-label-label-label-label-label-label-label-label-\n\t\t\t\t\tlabel-label-label-label-label-label-label-label-label-label-label-label-label-label-label-label-\n\t\t\t\t\tlabel-label-label-label-label-label-label-label-label-label-label-label-label-label-label-label\n\t\t\t\t' length for key 'net' must be between 1 and 200"),
		},
		"valid: label value with uppercase characters": {
			Labels: map[labelKey][]labelValue{
				"net": {"THE NET is vast AND INFINITE"},
			},
		},
		"valid: label value with DNS compliant name": {
			Labels: map[labelKey][]labelValue{
				"net": {"the-net-is-vast-and-infinite"},
			},
		},
		"valid: any unicode with rune count 1-200": {
			Labels: map[labelKey][]labelValue{
				"net": {"\uE005[\\\uE006\uE007"},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			err := test.Labels.Validate()
			if test.Error == nil {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.EqualError(t, err, test.Error.Error())
			}
		})
	}
}
