package cycle

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCycle_OwnMetadata(t *testing.T) {
	c := New()
	c.Labels().Set("k", "v")
	c.Scalars().Set("s", 1)
	assert.True(t, c.Labels().ValueIs("k", "v"))
	assert.True(t, c.Scalars().ValueIs("s", 1))
}

func TestCycle_AllActivatedAreWaiting_NothingActivated(t *testing.T) {
	// An idle cycle is not a stalled one: no activation means no waiting.
	assert.False(t, New().AllActivatedAreWaiting())
}

func TestGroup_OwnScalars(t *testing.T) {
	g := NewGroup()
	g.Scalars().Set("s", 2)
	assert.True(t, g.Scalars().ValueIs("s", 2))
}
