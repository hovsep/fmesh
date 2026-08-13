package meta

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The methods Scalars shares with Labels live on one generic store, and
// labels_test.go exercises them thoroughly. What is worth testing again here is
// the float64 instantiation itself: ValueIs compares with ==, and the aggregates
// below are the only float-specific logic in the package.
func TestScalars_SharedStoreWithFloat64Values(t *testing.T) {
	s := NewScalars()
	assert.True(t, s.IsEmpty())

	s.Set("temp", 36.6).SetMany(map[string]float64{"x": 1, "y": 2})
	assert.Equal(t, 3, s.Len())
	assert.True(t, s.Has("temp"))
	assert.True(t, s.ValueIs("temp", 36.6), "== comparison must work for float64")
	assert.False(t, s.ValueIs("temp", 36.7))

	v, err := s.Value("temp")
	require.NoError(t, err)
	assert.InDelta(t, 36.6, v, 1e-9)

	_, err = s.Value("absent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "scalar", "the error must name what kind of entry is missing")
	assert.InDelta(t, 7.0, s.ValueOrDefault("absent", 7.0), 1e-9)

	assert.Equal(t, []string{"temp", "x", "y"}, s.Keys(), "Keys is sorted")

	copied := s.All()
	copied["x"] = 999
	assert.True(t, s.ValueIs("x", 1), "All returns a defensive copy")

	assert.Equal(t, 2, s.Count(func(_ string, v float64) bool { return v < 3 }))
	assert.True(t, s.Any(func(_ string, v float64) bool { return v > 30 }))
	assert.True(t, s.Every(func(_ string, v float64) bool { return v > 0 }))

	visited := 0
	require.NoError(t, s.ForEach(func(_ string, _ float64) error { visited++; return nil }))
	assert.Equal(t, 3, visited)

	sentinel := errors.New("stop")
	visited = 0
	require.ErrorIs(t, s.ForEach(func(_ string, _ float64) error { visited++; return sentinel }), sentinel)
	assert.Equal(t, 1, visited, "ForEach stops on the first error")

	s.Remove("x")
	assert.False(t, s.Has("x"))
	assert.True(t, s.Clear().IsEmpty())
}

func TestScalars_Merge(t *testing.T) {
	t.Run("merges two stores", func(t *testing.T) {
		a := NewScalars().SetMany(map[string]float64{"x": 1, "y": 2})
		b := NewScalars().SetMany(map[string]float64{"y": 99, "z": 3})
		merged := a.Merge(b)
		assert.Equal(t, 3, merged.Len())
		v, _ := merged.Value("y")
		assert.InDelta(t, 99.0, v, 1e-9, "other wins on conflict")
	})
	t.Run("neither input is modified", func(t *testing.T) {
		a := NewScalars().Set("k", 1.0)
		b := NewScalars().Set("k", 2.0)
		_ = a.Merge(b)
		va, _ := a.Value("k")
		vb, _ := b.Value("k")
		assert.InDelta(t, 1.0, va, 1e-9)
		assert.InDelta(t, 2.0, vb, 1e-9)
	})
}

func TestScalars_Filter(t *testing.T) {
	t.Run("returns matching entries", func(t *testing.T) {
		s := NewScalars().SetMany(map[string]float64{"a": 5, "b": -1, "c": 3})
		positive := s.Filter(func(_ string, v float64) bool { return v > 0 })
		assert.Equal(t, 2, positive.Len())
		assert.True(t, positive.Has("a"))
		assert.True(t, positive.Has("c"))
		assert.False(t, positive.Has("b"))
	})
	t.Run("original is not modified", func(t *testing.T) {
		s := NewScalars().SetMany(map[string]float64{"a": 1, "b": 2})
		_ = s.Filter(func(_ string, v float64) bool { return v > 1 })
		assert.Equal(t, 2, s.Len())
	})
}

func TestScalars_Chainable(t *testing.T) {
	s := NewScalars().
		Set("a", 1.0).
		Set("b", 2.0).
		SetMany(map[string]float64{"c": 3.0, "d": 4.0}).
		Remove("d")

	assert.Equal(t, 3, s.Len())
	v, _ := s.Value("a")
	assert.InDelta(t, 1.0, v, 1e-9)
	assert.False(t, s.Has("d"))
}
