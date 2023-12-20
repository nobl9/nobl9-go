package rolebinding

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest"
)

//go:embed test_data/expected_error.txt
var expectedError string

func TestValidate_AllErrors(t *testing.T) {
	err := validate(RoleBinding{
		Kind: manifest.KindRoleBinding,
		Metadata: Metadata{
			Name: strings.Repeat("MY BINDING", 20),
		},
		Spec: Spec{
			RoleRef:    "",
			User:       ptr("123"),
			ProjectRef: strings.Repeat("MY PROJECT", 20),
		},
		ManifestSource: "/home/me/rolebinding.yaml",
	})
	assert.Equal(t, strings.TrimSuffix(expectedError, "\n"), err.Error())
}

func TestSpec(t *testing.T) {
	t.Run("fields mutual exclusion", func(t *testing.T) {
		tests := map[string]Spec{
			"both user and roleRef": {
				User:     ptr("123"),
				GroupRef: ptr("project-editor"),
				RoleRef:  "my-role",
			},
			"no user or roleRef": {
				User:     nil,
				GroupRef: nil,
				RoleRef:  "my-role",
			},
		}
		for name, spec := range tests {
			t.Run(name, func(t *testing.T) {
				rb := New(Metadata{Name: "my-name"}, spec)
				err := validate(rb)
				testutils.AssertContainsErrors(t, rb, err, 1, testutils.ExpectedError{
					Prop: "spec",
					Code: validation.ErrorCodeMutuallyExclusive,
				})
			})
		}
	})
}

func ptr[T any](v T) *T { return &v }
