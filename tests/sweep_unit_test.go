//go:build e2e_test

package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsStaleE2EName(t *testing.T) {
	now := time.Now()
	name := func(age time.Duration) string {
		return fmt.Sprintf("sdk-e2e-7-%d", now.Add(-age).UnixNano())
	}
	assert.True(t, isStaleE2EName(name(4*time.Hour), now))
	assert.False(t, isStaleE2EName(name(time.Hour), now))
	assert.False(t, isStaleE2EName("sdk-e2e-default", now))
	assert.False(t, isStaleE2EName("e2e-data-sources", now))
	assert.False(t, isStaleE2EName(fmt.Sprintf("other-7-%d", now.Add(-4*time.Hour).UnixNano()), now))
}
