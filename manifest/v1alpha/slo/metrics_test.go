package slo

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestDataSourceType(t *testing.T) {
	for _, src := range v1alpha.DataSourceTypeValues() {
		typ := validMetricSpec(src).DataSourceType()
		assert.Equal(t, src.String(), typ.String())
	}
}

func TestQuery(t *testing.T) {
	for _, src := range v1alpha.DataSourceTypeValues() {
		spec := validMetricSpec(src).Query()
		assert.NotEmpty(t, spec)
	}
}
