package signal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The scalar half of the CoW contract, mirroring the label tests: every With*
// returns a new signal and leaves the receiver untouched.

func TestSignal_WithScalars(t *testing.T) {
	t.Parallel()
	base := New(1).WithScalar("a", 1)

	next := base.WithScalars(map[string]float64{"b": 2, "c": 3})

	assert.Equal(t, 3, next.Scalars().Len())
	assert.True(t, next.Scalars().ValueIs("b", 2))
	assert.Equal(t, 1, base.Scalars().Len(), "receiver must not change")
}

func TestSignal_WithOnlyScalars(t *testing.T) {
	t.Parallel()
	base := New(1).WithScalar("a", 1)

	next := base.WithOnlyScalars(map[string]float64{"b": 2})

	assert.False(t, next.Scalars().Has("a"), "existing scalars are replaced")
	assert.True(t, next.Scalars().ValueIs("b", 2))
	assert.True(t, base.Scalars().ValueIs("a", 1), "receiver must not change")
}

func TestSignal_WithNoScalars(t *testing.T) {
	t.Parallel()
	base := New(1).WithScalar("a", 1)

	next := base.WithNoScalars()

	assert.True(t, next.Scalars().IsEmpty())
	assert.True(t, base.Scalars().ValueIs("a", 1), "receiver must not change")
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
	next := zero.WithLabel("k", "v").WithScalar("s", 1)
	assert.True(t, next.Labels().ValueIs("k", "v"))
	assert.True(t, next.Scalars().ValueIs("s", 1))
	assert.Nil(t, next.Payload())
}

func TestGroup_WithLabelOnEach(t *testing.T) {
	t.Parallel()
	g := NewGroup(1, 2).WithLabel("batch", "A")

	next := g.WithLabelOnEach("seen", "yes")

	for _, s := range next.All() {
		assert.True(t, s.Labels().ValueIs("seen", "yes"))
	}
	for _, s := range g.All() {
		assert.False(t, s.Labels().Has("seen"), "receiver's signals must not change")
	}
	assert.True(t, next.Labels().ValueIs("batch", "A"), "the group's own metadata is preserved")
}

func TestGroup_RemoveLabelOnEach(t *testing.T) {
	t.Parallel()
	g := NewGroup().With(
		New(1).WithLabel("k", "v").WithLabel("keep", "me"),
		New(2).WithLabel("k", "v"),
	)

	next := g.RemoveLabelOnEach("k")

	require.Equal(t, 2, next.Len())
	for _, s := range next.All() {
		assert.False(t, s.Labels().Has("k"))
	}
	assert.True(t, next.First().Labels().ValueIs("keep", "me"))
	assert.True(t, g.First().Labels().Has("k"), "receiver's signals must not change")
}
