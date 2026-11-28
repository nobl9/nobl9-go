package manifest

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKind_ProjectScoped(t *testing.T) {
	for _, kind := range KindValues() {
		isProjectScoped := slices.Contains(ProjectScopedKinds(), kind)
		assert.Equal(t, kind.ProjectScoped(), isProjectScoped)
	}
}

func TestKind_Applicable(t *testing.T) {
	assert.True(t, KindUserGroup.Applicable())
	assert.Contains(t, ApplicableKinds(), KindUserGroup)
	assert.False(t, KindAlert.Applicable())
	assert.NotContains(t, ApplicableKinds(), KindAlert)
}
