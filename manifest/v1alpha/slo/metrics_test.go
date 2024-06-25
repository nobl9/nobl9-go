package slo

import (
	"slices"
	"testing"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/internal/validation"

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

func Test_SingleQueryDisabled(t *testing.T) {
	skippedDataSources := []v1alpha.DataSourceType{
		v1alpha.ThousandEyes, // query is forbidden for this plugin
	}
	for _, src := range v1alpha.DataSourceTypeValues() {
		if slices.Contains(singleQueryGoodOverTotalEnabledSources, src) {
			continue
		}
		if slices.Contains(skippedDataSources, src) {
			continue
		}
		slo := validCountMetricSLO(src)
		slo.Spec.Objectives[0].CountMetrics = &CountMetricsSpec{
			Incremental:     ptr(false),
			GoodTotalMetric: validMetricSpec(src),
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].countMetrics.goodTotal",
			Code: joinErrorCodes(errCodeSingleQueryGoodOverTotalDisabled, validation.ErrorCodeOneOf),
		})
	}
}
