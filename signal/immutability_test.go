package signal

import (
	"errors"
	"testing"
	"unsafe"

	"github.com/hovsep/fmesh/labels"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Contract tests for github.com/hovsep/fmesh#203 (immutability / non-poisoning).

func TestSignal_immutable_builder_operations(t *testing.T) {
	t.Run("AddLabel_leaves_receiver_unchanged", func(t *testing.T) {
		orig := New(42).WithLabel("a", "1")
		require.Equal(t, 1, orig.Labels().Len())
		require.True(t, orig.Labels().ValueIs("a", "1"))

		next := orig.WithLabel("b", "2")
		require.NotNil(t, next)

		assert.Equal(t, 1, orig.Labels().Len(), "receiver must keep prior labels only")
		assert.True(t, orig.Labels().ValueIs("a", "1"))
		assert.False(t, orig.Labels().Has("b"))

		assert.Equal(t, 2, next.Labels().Len())
		assert.True(t, next.Labels().Has("b"))
	})

	t.Run("AddLabels_leaves_receiver_unchanged", func(t *testing.T) {
		orig := New(1).WithLabel("k", "v")
		next := orig.WithLabels(labels.Map{"x": "y"})

		assert.Equal(t, 1, orig.Labels().Len())
		assert.True(t, orig.Labels().ValueIs("k", "v"))
		assert.False(t, orig.Labels().Has("x"))

		assert.Equal(t, 2, next.Labels().Len())
		assert.True(t, next.Labels().Has("x"))
	})

	t.Run("SetLabels_leaves_receiver_unchanged", func(t *testing.T) {
		orig := New(1).WithLabel("keep", "old")
		next := orig.WithOnlyLabels(labels.Map{"new": "set"})

		assert.True(t, orig.Labels().ValueIs("keep", "old"))
		assert.False(t, orig.Labels().Has("new"))

		assert.Equal(t, 1, next.Labels().Len())
		assert.True(t, next.Labels().ValueIs("new", "set"))
	})

	t.Run("ClearLabels_leaves_receiver_unchanged", func(t *testing.T) {
		orig := New(1).WithLabel("a", "1")
		next := orig.WithNoLabels()

		assert.Equal(t, 1, orig.Labels().Len())
		assert.True(t, orig.Labels().Has("a"))

		assert.Equal(t, 0, next.Labels().Len())
	})

	t.Run("WithoutLabels_leaves_receiver_unchanged", func(t *testing.T) {
		orig := New(1).WithLabels(labels.Map{"a": "1", "b": "2"})
		next := orig.WithoutLabels("b")

		assert.Equal(t, 2, orig.Labels().Len())
		assert.True(t, orig.Labels().Has("b"))

		assert.Equal(t, 1, next.Labels().Len())
		assert.False(t, next.Labels().Has("b"))
	})

	t.Run("WithChainableErr_leaves_receiver_unchanged", func(t *testing.T) {
		orig := New(7).WithLabel("k", "v")
		sentinel := errors.New("sentinel")
		next := orig.WithChainableErr(sentinel)

		assert.False(t, orig.HasChainableErr())
		_, err := orig.Payload()
		require.NoError(t, err)

		assert.True(t, next.HasChainableErr())
		assert.Equal(t, sentinel, next.ChainableErr())
	})
}

func TestSignal_MapPayload_leaves_receiver_unchanged(t *testing.T) {
	orig := New(10).WithLabel("trace", "id")
	next := orig.MapPayload(func(p any) any { return p.(int) * 2 })

	p0, err := orig.Payload()
	require.NoError(t, err)
	assert.Equal(t, 10, p0)
	assert.Equal(t, 1, orig.Labels().Len())
	assert.True(t, orig.Labels().ValueIs("trace", "id"))

	p1, err := next.Payload()
	require.NoError(t, err)
	assert.Equal(t, 20, p1)
	assert.True(t, next.Labels().ValueIs("trace", "id"))
}

func TestGroup_Add_does_not_poison_receiver_on_invalid_signal(t *testing.T) {
	t.Run("nil_signal", func(t *testing.T) {
		g := NewGroup(1)
		_ = g.With(nil)

		assert.False(t, g.HasChainableErr(), "receiver must not retain Add error")
		assert.Equal(t, 1, g.Len())

		g2 := g.With(New(99))
		assert.False(t, g2.HasChainableErr())
		assert.Equal(t, 2, g2.Len())
	})

	t.Run("signal_with_chainable_error", func(t *testing.T) {
		g := NewGroup(1)
		bad := New(2).WithChainableErr(errors.New("bad signal"))
		_ = g.With(bad)

		assert.False(t, g.HasChainableErr())
		assert.Equal(t, 1, g.Len())

		g2 := g.With(New(3))
		assert.False(t, g2.HasChainableErr())
		assert.Equal(t, 2, g2.Len())
	})
}

func TestGroup_ForEach_does_not_poison_receiver_on_error(t *testing.T) {
	g := NewGroup(1, 2, 3)
	_ = g.ForEach(func(*Signal) error { return errors.New("stop") })

	assert.False(t, g.HasChainableErr(), "ForEach must not persist error on original group")
	assert.Equal(t, 3, g.Len())
}

func TestGroup_ForEachIf_does_not_poison_receiver_on_error(t *testing.T) {
	g := NewGroup(1, 2, 3)
	_ = g.ForEachIf(
		func(*Signal) bool { return true },
		func(*Signal) error { return errors.New("stop") },
	)

	assert.False(t, g.HasChainableErr())
	assert.Equal(t, 3, g.Len())
}

func TestGroup_MapIf_non_matching_signals_are_not_shared_pointers(t *testing.T) {
	g := NewGroup(1, 2)
	mapper := func(*Signal) *Signal {
		require.Fail(t, "mapper must not run when predicate matches nothing")
		return nil
	}
	out := g.MapIf(func(*Signal) bool { return false }, mapper)

	outSigs, err := out.All()
	require.NoError(t, err)
	require.Len(t, outSigs, 2)
	require.NotEqual(t, uintptr(unsafe.Pointer(g.First())), uintptr(unsafe.Pointer(outSigs[0])),
		"MapIf pass-through must use cloned signals, not shared pointers (#203)")

	_ = outSigs[0].WithLabel("x", "y")

	assert.False(t, g.First().Labels().Has("x"),
		"mutating output group's signal must not change original group's signal (#203)")
}

func TestGroup_Map_identity_mapper_does_not_alias(t *testing.T) {
	g := NewGroup(1, 2)
	identity := func(s *Signal) *Signal { return s }
	out := g.Map(identity)

	outSigs, err := out.All()
	require.NoError(t, err)
	require.Len(t, outSigs, 2)
	require.NotEqual(t, uintptr(unsafe.Pointer(g.First())), uintptr(unsafe.Pointer(outSigs[0])),
		"Map with identity mapper must clone signals, not share pointers")

	_ = outSigs[0].WithLabel("x", "y")
	assert.False(t, g.First().Labels().Has("x"),
		"mutating output group's signal must not change original group's signal")
}

func TestGroup_MapPayloadsIf_non_matching_signals_are_not_shared_pointers(t *testing.T) {
	g := NewGroup(1, 2)
	out := g.MapPayloadsIf(
		func(*Signal) bool { return false },
		func(any) any {
			require.Fail(t, "mapper must not run when predicate matches nothing")
			return nil
		},
	)

	outSigs, err := out.All()
	require.NoError(t, err)
	require.Len(t, outSigs, 2)
	require.NotEqual(t, uintptr(unsafe.Pointer(g.First())), uintptr(unsafe.Pointer(outSigs[0])))

	_ = outSigs[0].WithLabel("x", "y")

	assert.False(t, g.First().Labels().Has("x"),
		"mutating output group's signal must not change original group's signal (#203)")
}
