package slo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSLO_CompositeSLOFunctions(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validCompositeSLO()

		assert.Equal(t, true, slo.Spec.HasCompositeObjectives())
		assert.Equal(t, 2, slo.Spec.CompositeObjectivesComponentCount())
	})

	t.Run("passes - 4 composite components", func(t *testing.T) {
		slo := validCompositeSLO()

		slo.Spec.Objectives[0].Composite.Objectives = append(
			slo.Spec.Objectives[0].Composite.Objectives,
			slo.Spec.Objectives[0].Composite.Objectives[0],
			slo.Spec.Objectives[0].Composite.Objectives[1],
		)

		assert.Equal(t, true, slo.Spec.HasCompositeObjectives())
		assert.Equal(t, 4, slo.Spec.CompositeObjectivesComponentCount())
	})
}
