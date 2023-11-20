package usergroup

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/manifest"
)

//go:embed test_data/expected_error.txt
var expectedError string

func TestValidate_AllErrors(t *testing.T) {
	err := validate(UserGroup{
		Kind: manifest.KindUserGroup,
		Metadata: Metadata{
			Name: strings.Repeat("MY GROUP", 20),
		},
		Spec: Spec{
			DisplayName: strings.Repeat("my-group", 10),
		},
		ManifestSource: "/home/me/usergroup.yaml",
	})
	assert.Equal(t, strings.TrimSuffix(expectedError, "\n"), err.Error())
}
