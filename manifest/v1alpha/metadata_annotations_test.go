package v1alpha

import (
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// nolint: lll
func TestValidateMetadataAnnotations(t *testing.T) {
	for name, test := range map[string]struct {
		Annotations MetadataAnnotations
		Error       error
	}{
		"valid: empty annotations map": {
			Annotations: nil,
		},
		"valid: simple strings": {
			Annotations: MetadataAnnotations{
				"net":     "vast",
				"project": "nobl9",
			},
		},
		"invalid: empty key": {
			Annotations: MetadataAnnotations{
				"": "vast",
			},
			Error: errors.New("Validation has failed for the following properties:\n  - 'key':\n    - property is required but was empty"),
		},
		"invalid: empty value": {
			Annotations: MetadataAnnotations{
				"net": "",
			},
			Error: errors.New("Validation has failed for the following properties:\n  - 'value':\n    - property is required but was empty"),
		},
		"invalid: key is too long": {
			Annotations: MetadataAnnotations{
				"net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net": "",
			},
			Error: errors.New("Validation has failed for the following properties:\n  - 'key' with value 'net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net-net':\n    - length must be between 1 and 63"),
		},
		"invalid: key starts with non letter": {
			Annotations: MetadataAnnotations{
				"9net": "",
			},
			Error: errors.New("Validation has failed for the following properties:\n  - 'key' with value '9net':\n    - string does not match regular expression: '^\\p{L}([_\\-0-9\\p{L}]*[0-9\\p{L}])?$'"),
		},
		"invalid: key ends with non alphanumeric char": {
			Annotations: MetadataAnnotations{
				"net_": "",
			},
			Error: errors.New("Validation has failed for the following properties:\n  - 'key' with value 'net_':\n    - string does not match regular expression: '^\\p{L}([_\\-0-9\\p{L}]*[0-9\\p{L}])?$'"),
		},
		"invalid: key contains uppercase character": {
			Annotations: MetadataAnnotations{
				"nEt": "",
			},
			Error: errors.New("Validation has failed for the following properties:\n  - 'key' with value 'nEt':\n    - string must not match regular expression: '[A-Z]+'"),
		},
		"invalid: value is too long (over 1050 chars)": {
			Annotations: MetadataAnnotations{
				"net": strings.Repeat("l", 1051),
			},
			Error: errors.New("Validation has failed for the following properties:\n  - 'value' with value 'llllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllll...':\n    - length must be between 1 and 1050"),
		},
		"valid: value with uppercase characters": {
			Annotations: MetadataAnnotations{
				"net": "THE NET is vast AND INFINITE",
			},
		},
		"valid: value with DNS compliant name": {
			Annotations: MetadataAnnotations{
				"net": "the-net-is-vast-and-infinite",
			},
		},
		"valid: any unicode": {
			Annotations: MetadataAnnotations{
				"net": "\uE005[\\\uE006\uE007",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			err := ValidationRuleMetadataAnnotations().Validate(test.Annotations)
			if test.Error == nil {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.EqualError(t, err, test.Error.Error())
			}
		})
	}
}
