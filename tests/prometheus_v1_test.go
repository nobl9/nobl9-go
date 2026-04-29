//go:build e2e_test

package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Prometheus_V1_Buildinfo(t *testing.T) {
	t.Parallel()

	buildInfo, err := client.Prometheus().V1().Buildinfo(t.Context())
	require.NoError(t, err)
	assert.NotEmpty(t, buildInfo.Version)
	assert.NotEmpty(t, buildInfo.Revision)
	assert.NotEmpty(t, buildInfo.Branch)
	assert.NotEmpty(t, buildInfo.GoVersion)
}
