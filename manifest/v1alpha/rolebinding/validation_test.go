package rolebinding

import (
	_ "embed"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/govy/pkg/rules"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest"
)

var validationMessageRegexp = regexp.MustCompile(strings.TrimSpace(`
(?s)Validation for RoleBinding '.*' has failed for the following fields:
.*
Manifest source: /home/me/rolebinding.yaml
`))

func TestValidate_VersionAndKind(t *testing.T) {
	rb := validRoleBinding()
	rb.APIVersion = "v0.1"
	rb.Kind = manifest.KindProject
	rb.ManifestSource = "/home/me/rolebinding.yaml"
	err := validate(rb)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, rb, err, 2,
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
	rb := RoleBinding{
		APIVersion: manifest.VersionV1alpha,
		Kind:       manifest.KindRoleBinding,
		Metadata: Metadata{
			Name: strings.Repeat("MY BINDING", 20),
		},
		Spec: Spec{
			RoleRef:    "admin",
			User:       ptr("123"),
			ProjectRef: "default",
		},
		ManifestSource: "/home/me/rolebinding.yaml",
	}
	err := validate(rb)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, rb, err, 1,
		testutils.ExpectedError{
			Prop: "metadata.name",
			Code: validationV1Alpha.ErrorCodeStringName,
		},
	)
}

func TestSpec(t *testing.T) {
	t.Run("required roleRef", func(t *testing.T) {
		rb := validRoleBinding()
		rb.Spec.RoleRef = ""
		err := validate(rb)
		testutils.AssertContainsErrors(t, rb, err, 1,
			testutils.ExpectedError{
				Prop: "spec.roleRef",
				Code: rules.ErrorCodeRequired,
			},
		)
	})
	t.Run("invalid projectRef", func(t *testing.T) {
		rb := validRoleBinding()
		rb.Spec.ProjectRef = strings.Repeat("MY PROJECT", 20)
		err := validate(rb)
		testutils.AssertContainsErrors(t, rb, err, 1,
			testutils.ExpectedError{
				Prop: "spec.projectRef",
				Code: validationV1Alpha.ErrorCodeStringName,
			},
		)
	})
	t.Run("fields mutual exclusion", func(t *testing.T) {
		tests := map[string]Spec{
			"both user and groupRef": {
				User:     ptr("123"),
				GroupRef: ptr("project-editor"),
				RoleRef:  "my-role",
			},
			"both accountId and groupRef": {
				AccountID: ptr("123"),
				GroupRef:  ptr("project-editor"),
				RoleRef:   "my-role",
			},
			"both user and accountId": {
				User:      ptr("123"),
				AccountID: ptr("456"),
				RoleRef:   "my-role",
			},
			"all three set": {
				User:      ptr("123"),
				AccountID: ptr("456"),
				GroupRef:  ptr("project-editor"),
				RoleRef:   "my-role",
			},
			"none set": {
				User:      nil,
				AccountID: nil,
				GroupRef:  nil,
				RoleRef:   "my-role",
			},
		}
		for name, spec := range tests {
			t.Run(name, func(t *testing.T) {
				rb := New(Metadata{Name: "my-name"}, spec)
				err := validate(rb)
				testutils.AssertContainsErrors(t, rb, err, 1, testutils.ExpectedError{
					Prop: "spec",
					Code: rules.ErrorCodeMutuallyExclusive,
				})
			})
		}
	})
	t.Run("valid with accountId", func(t *testing.T) {
		rb := New(Metadata{Name: "my-name"}, Spec{
			AccountID:  ptr("123"),
			RoleRef:    "my-role",
			ProjectRef: "default",
		})
		err := validate(rb)
		assert.Nil(t, err)
	})
}

func validRoleBinding() RoleBinding {
	return RoleBinding{
		APIVersion: manifest.VersionV1alpha,
		Kind:       manifest.KindRoleBinding,
		Metadata: Metadata{
			Name: "my-binding",
		},
		Spec: Spec{
			RoleRef:    "admin",
			User:       ptr("123"),
			ProjectRef: "default",
		},
	}
}

func ptr[T any](v T) *T { return &v }
