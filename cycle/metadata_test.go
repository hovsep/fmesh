package cycle

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCycle_OwnMetadata(t *testing.T) {
	c := New()
	c.Meta().Set("k", "v")
	c.Meta().Set("s", 1.0)
	assert.True(t, c.Meta().ValueIs("k", "v"))
	assert.True(t, c.Meta().ValueIs("s", 1.0))
}

func TestCycle_AllActivatedAreWaiting_NothingActivated(t *testing.T) {
	// An idle cycle is not a stalled one: no activation means no waiting.
	assert.False(t, New().AllActivatedAreWaiting())
}

func TestGroup_OwnMeta(t *testing.T) {
	g := NewGroup()
	g.Meta().Set("s", 2.0)
	assert.True(t, g.Meta().ValueIs("s", 2.0))
}
