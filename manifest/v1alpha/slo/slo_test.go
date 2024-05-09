package slo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSLO_SpecHasCompositeObjectives(t *testing.T) {
	t.Run("passes - composite SLO", func(t *testing.T) {
		slo := validCompositeSLO()

		assert.Equal(t, true, slo.Spec.HasCompositeObjectives())
	})

	t.Run("passes - normal SLO", func(t *testing.T) {
		slo := validSLO()

		assert.Equal(t, false, slo.Spec.HasCompositeObjectives())
	})
}
