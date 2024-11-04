package usergroup

import (
	_ "embed"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest"
)

var validationMessageRegexp = regexp.MustCompile(strings.TrimSpace(`
(?s)Validation for UserGroup '.*' has failed for the following fields:
.*
Manifest source: /home/me/usergroup.yaml
`))

func TestValidate_VersionAndKind(t *testing.T) {
	group := validUserGroup()
	group.APIVersion = "v0.1"
	group.Kind = manifest.KindProject
	group.ManifestSource = "/home/me/usergroup.yaml"
	err := validate(group)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, group, err, 2,
		testutils.ExpectedError{
			Prop: "apiVersion",
			Code: rules.ErrorCodeEqualTo,
		},
		testutils.ExpectedError{
			Prop: "kind",
			Code: rules.ErrorCodeEqualTo,
		},
	)
}

func TestValidate_Metadata(t *testing.T) {
	group := validUserGroup()
	group.Metadata.Name = strings.Repeat("MY GROUP", 20)
	group.ManifestSource = "/home/me/usergroup.yaml"
	err := validate(group)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, group, err, 1,
		testutils.ExpectedError{
			Prop: "metadata.name",
			Code: rules.ErrorCodeStringDNSLabel,
		},
	)
}

func TestValidate_Spec(t *testing.T) {
	t.Run("displayName too long", func(t *testing.T) {
		group := validUserGroup()
		group.Spec.DisplayName = strings.Repeat("MY GROUP", 20)
		err := validate(group)
		testutils.AssertContainsErrors(t, group, err, 1, testutils.ExpectedError{
			Prop: "spec.displayName",
			Code: rules.ErrorCodeStringLength,
		})
	})
}

func validUserGroup() UserGroup {
	return UserGroup{
		APIVersion: manifest.VersionV1alpha,
		Kind:       manifest.KindUserGroup,
		Metadata: Metadata{
			Name: "my-group",
		},
		Spec: Spec{
			DisplayName: "My Group",
		},
	}
}
