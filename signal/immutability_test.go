package signal

import (
	"slices"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Contract tests for github.com/hovsep/fmesh#203 (immutability / non-poisoning).

func TestSignal_immutable_builder_operations(t *testing.T) {
	t.Run("WithMeta_leaves_receiver_unchanged", func(t *testing.T) {
		orig := New(42).WithMeta("a", "1")
		require.Equal(t, 1, orig.Meta().Len())
		require.True(t, orig.Meta().ValueIs("a", "1"))

		next := orig.WithMeta("b", "2")
		require.NotNil(t, next)

		assert.Equal(t, 1, orig.Meta().Len(), "receiver must keep prior entries only")
		assert.True(t, orig.Meta().ValueIs("a", "1"))
		assert.False(t, orig.Meta().Has("b"))

		assert.Equal(t, 2, next.Meta().Len())
		assert.True(t, next.Meta().Has("b"))
	})

	t.Run("WithMetaMany_leaves_receiver_unchanged", func(t *testing.T) {
		orig := New(1).WithMeta("k", "v")
		next := orig.WithMetaMany(map[string]string{"x": "y"})

		assert.Equal(t, 1, orig.Meta().Len())
		assert.True(t, orig.Meta().ValueIs("k", "v"))
		assert.False(t, orig.Meta().Has("x"))

		assert.Equal(t, 2, next.Meta().Len())
		assert.True(t, next.Meta().Has("x"))
	})

	t.Run("WithoutMeta_leaves_receiver_unchanged", func(t *testing.T) {
		orig := New(1).WithMetaMany(map[string]string{"a": "1", "b": "2"})
		next := orig.WithoutMeta("b")

		assert.Equal(t, 2, orig.Meta().Len())
		assert.True(t, orig.Meta().Has("b"))

		assert.Equal(t, 1, next.Meta().Len())
		assert.False(t, next.Meta().Has("b"))
	})
}

func TestSignal_MapPayload_leaves_receiver_unchanged(t *testing.T) {
	orig := New(10).WithMeta("trace", "id")
	next := orig.MapPayload(func(p any) any { return p.(int) * 2 })

	p0 := orig.Payload()
	assert.Equal(t, 10, p0)
	assert.Equal(t, 1, orig.Meta().Len())
	assert.True(t, orig.Meta().ValueIs("trace", "id"))

	p1 := next.Payload()
	assert.Equal(t, 20, p1)
	assert.True(t, next.Meta().ValueIs("trace", "id"))
}

func TestGroup_Add_does_not_poison_receiver_on_nil_signal(t *testing.T) {
	g := NewGroup(1)
	_ = g.With(nil)

	assert.Equal(t, 1, g.Len(), "receiver must not change after nil add")

	g2 := g.With(New(99))
	assert.Equal(t, 2, g2.Len())
}

func TestGroup_MapIf_non_matching_signals_are_not_shared_pointers(t *testing.T) {
	g := NewGroup(1, 2)
	mapper := func(*Signal) *Signal {
		require.Fail(t, "mapper must not run when predicate matches nothing")
		return nil
	}
	out := g.MapIf(func(*Signal) bool { return false }, mapper)

	outSigs := slices.Collect(out.All())
	require.Len(t, outSigs, 2)
	assert.NotSame(t, g.First(), outSigs[0],
		"MapIf pass-through must use cloned signals, not shared pointers (#203)")

	_ = outSigs[0].WithMeta("x", "y")

	assert.False(t, g.First().Meta().Has("x"),
		"mutating output group's signal must not change original group's signal (#203)")
}

func TestGroup_Meta_returns_a_copy(t *testing.T) {
	// The one back door through CoW used to be Meta() handing out
	// the live store; mutating the returned store must not touch the group.
	g := NewGroup(1).WithMeta("k", "v").WithMeta("s", 1.0)

	g.Meta().Set("k", "changed")
	g.Meta().Set("s", 2.0)

	assert.True(t, g.Meta().ValueIs("k", "v"))
	v, err := g.Meta().Value[float64]("s")
	require.NoError(t, err)
	assert.InDelta(t, 1.0, v, 1e-9)
}

func TestGroup_Map_identity_mapper_does_not_alias(t *testing.T) {
	g := NewGroup(1, 2)
	identity := func(s *Signal) *Signal { return s }
	out := g.Map(identity)

	outSigs := slices.Collect(out.All())
	require.Len(t, outSigs, 2)
	assert.NotSame(t, g.First(), outSigs[0],
		"Map with identity mapper must clone signals, not share pointers")

	_ = outSigs[0].WithMeta("x", "y")
	assert.False(t, g.First().Meta().Has("x"),
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

	outSigs := slices.Collect(out.All())
	require.Len(t, outSigs, 2)
	assert.NotSame(t, g.First(), outSigs[0])

	_ = outSigs[0].WithMeta("x", "y")

	assert.False(t, g.First().Meta().Has("x"),
		"mutating output group's signal must not change original group's signal (#203)")
}

// TestSignal_concurrent_CoW_is_race_free verifies that multiple goroutines
// simultaneously calling CoW methods on the same shared *Signal do not cause
// a data race. This mirrors the fan-out scenario: the same *Signal pointer
// lands in N input ports and N components activate concurrently, each
// deriving their own annotated copy.
//
// Run with: go test -race ./signal/...
func TestSignal_concurrent_CoW_is_race_free(t *testing.T) {
	const goroutines = 50

	// Shared signal — simulates a fanned-out signal sitting in multiple ports.
	shared := New("payload").
		WithMeta("origin", "sensor-1").
		WithMeta("temp", 36.6)

	var wg sync.WaitGroup
	results := make([]*Signal, goroutines)

	for i := range goroutines {
		wg.Go(func() {
			// Each goroutine acts like a component: reads the shared signal and
			// produces its own annotated copy without touching the original.
			results[i] = shared.
				WithMeta("processed-by", "component").
				WithMeta("adjusted", shared.Meta().ValueOrDefault("temp", 0.0)+float64(i))
		})
	}
	wg.Wait()

	// Original must be completely unchanged.
	assert.Equal(t, 2, shared.Meta().Len(), "shared signal metadata must not grow")
	assert.True(t, shared.Meta().ValueIs("origin", "sensor-1"))
	assert.False(t, shared.Meta().Has("processed-by"))

	v, err := shared.Meta().Value[float64]("temp")
	require.NoError(t, err)
	assert.InDelta(t, 36.6, v, 1e-9)
	assert.False(t, shared.Meta().Has("adjusted"))

	// Every derived signal must have both the inherited and the new metadata.
	for i, s := range results {
		assert.True(t, s.Meta().ValueIs("origin", "sensor-1"),
			"goroutine %d: inherited entry must be present", i)
		assert.True(t, s.Meta().Has("processed-by"),
			"goroutine %d: own string entry must be present", i)
		assert.True(t, s.Meta().Has("adjusted"), "goroutine %d: own float entry must be present", i)
	}
}

// TestGroup_concurrent_WithMeta_is_race_free verifies that multiple goroutines
// calling WithMeta on the same *Group concurrently do not race. Each call
// returns a new group; the original must be unmodified.
func TestGroup_concurrent_WithMeta_is_race_free(t *testing.T) {
	const goroutines = 50

	shared := NewGroup(1, 2, 3).WithMeta("batch", "A")

	var wg sync.WaitGroup
	results := make([]*Group, goroutines)

	for i := range goroutines {
		wg.Go(func() {
			results[i] = shared.WithMeta("worker", "x")
		})
	}
	wg.Wait()

	// Original group must still have only its own entry.
	assert.Equal(t, 1, shared.Meta().Len())
	assert.True(t, shared.Meta().ValueIs("batch", "A"))
	assert.False(t, shared.Meta().Has("worker"))

	for i, g := range results {
		assert.True(t, g.Meta().Has("worker"),
			"goroutine %d: derived group must have added entry", i)
		assert.True(t, g.Meta().ValueIs("batch", "A"),
			"goroutine %d: derived group must inherit original entry", i)
	}
}

// TestGroup_concurrent_WithMetaOnEach_is_race_free verifies that multiple
// goroutines stamping entries on copies of the same group do not race.
func TestGroup_concurrent_WithMetaOnEach_is_race_free(t *testing.T) {
	const goroutines = 50

	shared := NewGroup(1, 2, 3)

	var wg sync.WaitGroup
	results := make([]*Group, goroutines)

	for i := range goroutines {
		wg.Go(func() {
			results[i] = shared.WithMetaOnEach("priority", float64(i))
		})
	}
	wg.Wait()

	// Signals inside the original group must have no metadata.
	sigs := slices.Collect(shared.All())
	for _, s := range sigs {
		assert.False(t, s.Meta().Has("priority"),
			"original group's signals must not be affected")
	}

	// Each result group's signals must have the entry.
	for i, g := range results {
		outSigs := slices.Collect(g.All())
		for _, s := range outSigs {
			assert.True(t, s.Meta().Has("priority"),
				"goroutine %d: derived signal must have the entry", i)
		}
	}
}
