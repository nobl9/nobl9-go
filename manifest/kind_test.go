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
