package signal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The CoW contract on metadata: every With* returns a new signal and leaves
// the receiver untouched, and the two value types share one store.

func TestSignal_WithMetaMany_floats(t *testing.T) {
	t.Parallel()
	base := New(1).WithMeta("a", 1.0)

	next := base.WithMetaMany(map[string]float64{"b": 2, "c": 3})

	assert.Equal(t, 3, next.Meta().Len())
	assert.True(t, next.Meta().ValueIs("b", 2.0))
	assert.Equal(t, 1, base.Meta().Len(), "receiver must not change")
}

func TestCloneSignal_NilIsNil(t *testing.T) {
	t.Parallel()
	assert.Nil(t, cloneSignal(nil))
}

func TestZeroValueSignal_CoWStillWorks(t *testing.T) {
	t.Parallel()
	// A Signal{} built without New has nil stores; the CoW methods must treat
	// that as empty rather than dereference it.
	var zero Signal
	next := zero.WithMeta("k", "v").WithMeta("s", 1.0)
	assert.True(t, next.Meta().ValueIs("k", "v"))
	assert.True(t, next.Meta().ValueIs("s", 1.0))
	assert.Nil(t, next.Payload())
}

func TestGroup_WithMetaOnEach(t *testing.T) {
	t.Parallel()
	g := NewGroup(1, 2).WithMeta("batch", "A")

	next := g.WithMetaOnEach("seen", "yes")

	for s := range next.All() {
		assert.True(t, s.Meta().ValueIs("seen", "yes"))
	}
	for s := range g.All() {
		assert.False(t, s.Meta().Has("seen"), "receiver's signals must not change")
	}
	assert.True(t, next.Meta().ValueIs("batch", "A"), "the group's own metadata is preserved")
}

func TestGroup_WithoutMetaOnEach(t *testing.T) {
	t.Parallel()
	g := NewGroup().With(
		New(1).WithMeta("k", "v").WithMeta("keep", "me"),
		New(2).WithMeta("k", "v"),
	)

	next := g.WithoutMetaOnEach("k")

	require.Equal(t, 2, next.Len())
	for s := range next.All() {
		assert.False(t, s.Meta().Has("k"))
	}
	assert.True(t, next.First().Meta().ValueIs("keep", "me"))
	assert.True(t, g.First().Meta().Has("k"), "receiver's signals must not change")
}
