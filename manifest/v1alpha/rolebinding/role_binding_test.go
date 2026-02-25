package rolebinding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetType(t *testing.T) {
	tests := map[string]struct {
		projectRef   string
		expectedType Type
	}{
		"returns TypeProject when projectRef is set": {
			projectRef:   "default",
			expectedType: TypeProject,
		},
		"returns TypeOrganization when projectRef is empty": {
			projectRef:   "",
			expectedType: TypeOrganization,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			rb := New(
				Metadata{Name: "my-binding"},
				Spec{
					AccountID:  ptr("123"),
					RoleRef:    "admin",
					ProjectRef: tc.projectRef,
				},
			)

			actualType := rb.GetType()
			assert.Equal(t, tc.expectedType, actualType)
		})
	}
}
