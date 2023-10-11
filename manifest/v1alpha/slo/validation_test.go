package slo

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

//go:embed test_data/expected_error.txt
var expectedError string

func TestValidate_AllErrors(t *testing.T) {
	err := validate(SLO{
		Kind: manifest.KindSLO,
		Metadata: Metadata{
			Name:        strings.Repeat("MY SLO", 20),
			DisplayName: strings.Repeat("my-slo", 10),
			Project:     strings.Repeat("MY PROJECT", 20),
			Labels: v1alpha.Labels{
				"L O L": []string{"dip", "dip"},
			},
		},
		Spec: Spec{
			Description: strings.Repeat("l", 2000),
		},
		ManifestSource: "/home/me/slo.yaml",
	})
	assert.Equal(t, expectedError, err.Error())
}
