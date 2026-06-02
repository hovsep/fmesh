package signal

import (
	"errors"
	"sync"
	"testing"
	"unsafe"

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
		next := orig.WithLabels(map[string]string{"x": "y"})

		assert.Equal(t, 1, orig.Labels().Len())
		assert.True(t, orig.Labels().ValueIs("k", "v"))
		assert.False(t, orig.Labels().Has("x"))

		assert.Equal(t, 2, next.Labels().Len())
		assert.True(t, next.Labels().Has("x"))
	})

	t.Run("SetLabels_leaves_receiver_unchanged", func(t *testing.T) {
		orig := New(1).WithLabel("keep", "old")
		next := orig.WithOnlyLabels(map[string]string{"new": "set"})

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
		orig := New(1).WithLabels(map[string]string{"a": "1", "b": "2"})
		next := orig.WithoutLabels("b")

		assert.Equal(t, 2, orig.Labels().Len())
		assert.True(t, orig.Labels().Has("b"))

		assert.Equal(t, 1, next.Labels().Len())
		assert.False(t, next.Labels().Has("b"))
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

func TestGroup_Add_does_not_poison_receiver_on_nil_signal(t *testing.T) {
	g := NewGroup(1)
	_ = g.With(nil)

	assert.Equal(t, 1, g.Len(), "receiver must not change after nil add")

	g2 := g.With(New(99))
	assert.Equal(t, 2, g2.Len())
}

func TestGroup_ForEach_does_not_poison_receiver_on_error(t *testing.T) {
	g := NewGroup(1, 2, 3)
	_ = g.ForEach(func(*Signal) error { return errors.New("stop") })

	assert.Equal(t, 3, g.Len())
}

func TestGroup_ForEachIf_does_not_poison_receiver_on_error(t *testing.T) {
	g := NewGroup(1, 2, 3)
	_ = g.ForEachIf(
		func(*Signal) bool { return true },
		func(*Signal) error { return errors.New("stop") },
	)

	assert.Equal(t, 3, g.Len())
}

func TestGroup_MapIf_non_matching_signals_are_not_shared_pointers(t *testing.T) {
	g := NewGroup(1, 2)
	mapper := func(*Signal) *Signal {
		require.Fail(t, "mapper must not run when predicate matches nothing")
		return nil
	}
	out := g.MapIf(func(*Signal) bool { return false }, mapper)

	outSigs := out.All()
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

	outSigs := out.All()
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

	outSigs := out.All()
	require.Len(t, outSigs, 2)
	require.NotEqual(t, uintptr(unsafe.Pointer(g.First())), uintptr(unsafe.Pointer(outSigs[0])))

	_ = outSigs[0].WithLabel("x", "y")

	assert.False(t, g.First().Labels().Has("x"),
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
		WithLabel("origin", "sensor-1").
		WithScalar("temp", 36.6)

	var wg sync.WaitGroup
	results := make([]*Signal, goroutines)

	for i := range goroutines {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			// Each goroutine acts like a component: reads the shared signal and
			// produces its own annotated copy without touching the original.
			results[idx] = shared.
				WithLabel("processed-by", "component").
				WithScalar("adjusted", shared.Scalars().GetOrDefault("temp", 0)+float64(idx))
		}(i)
	}
	wg.Wait()

	// Original must be completely unchanged.
	assert.Equal(t, 1, shared.Labels().Len(), "shared signal labels must not grow")
	assert.True(t, shared.Labels().ValueIs("origin", "sensor-1"))
	assert.False(t, shared.Labels().Has("processed-by"))

	assert.Equal(t, 1, shared.Scalars().Len(), "shared signal scalars must not grow")
	v, ok := shared.Scalars().Get("temp")
	assert.True(t, ok)
	assert.InDelta(t, 36.6, v, 1e-9)
	assert.False(t, shared.Scalars().Has("adjusted"))

	// Every derived signal must have both the inherited and the new metadata.
	for i, s := range results {
		assert.True(t, s.Labels().ValueIs("origin", "sensor-1"),
			"goroutine %d: inherited label must be present", i)
		assert.True(t, s.Labels().Has("processed-by"),
			"goroutine %d: own label must be present", i)

		_, hasScalar := s.Scalars().Get("adjusted")
		assert.True(t, hasScalar, "goroutine %d: own scalar must be present", i)
	}
}

// TestGroup_concurrent_WithLabel_is_race_free verifies that multiple goroutines
// calling WithLabel on the same *Group concurrently do not race. Each call
// returns a new group; the original must be unmodified.
func TestGroup_concurrent_WithLabel_is_race_free(t *testing.T) {
	const goroutines = 50

	shared := NewGroup(1, 2, 3).WithLabel("batch", "A")

	var wg sync.WaitGroup
	results := make([]*Group, goroutines)

	for i := range goroutines {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = shared.WithLabel("worker", "x")
		}(i)
	}
	wg.Wait()

	// Original group must still have only its own label.
	assert.Equal(t, 1, shared.Labels().Len())
	assert.True(t, shared.Labels().ValueIs("batch", "A"))
	assert.False(t, shared.Labels().Has("worker"))

	for i, g := range results {
		assert.True(t, g.Labels().Has("worker"),
			"goroutine %d: derived group must have added label", i)
		assert.True(t, g.Labels().ValueIs("batch", "A"),
			"goroutine %d: derived group must inherit original label", i)
	}
}

// TestGroup_concurrent_WithScalarOnEach_is_race_free verifies that multiple
// goroutines stamping scalars on copies of the same group do not race.
func TestGroup_concurrent_WithScalarOnEach_is_race_free(t *testing.T) {
	const goroutines = 50

	shared := NewGroup(1, 2, 3)

	var wg sync.WaitGroup
	results := make([]*Group, goroutines)

	for i := range goroutines {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = shared.WithScalarOnEach("priority", float64(idx))
		}(i)
	}
	wg.Wait()

	// Signals inside the original group must have no scalars.
	sigs := shared.All()
	for _, s := range sigs {
		assert.False(t, s.Scalars().Has("priority"),
			"original group's signals must not be affected")
	}

	// Each result group's signals must have the scalar.
	for i, g := range results {
		outSigs := g.All()
		for _, s := range outSigs {
			assert.True(t, s.Scalars().Has("priority"),
				"goroutine %d: derived signal must have scalar", i)
		}
	}
}
