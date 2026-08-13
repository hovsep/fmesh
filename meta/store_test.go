package meta

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The generic store behind Labels and Scalars is tested here once, directly;
// labels_test.go and scalars_test.go only cover what the facades add on top.

func newStore(entries map[string]int) *store[int] {
	s := &store[int]{}
	s.init()
	s.setMany(entries)
	return s
}

func TestStore_Mutators(t *testing.T) {
	t.Parallel()
	t.Run("set upserts", func(t *testing.T) {
		s := newStore(nil)
		s.set("k", 1)
		assert.True(t, s.ValueIs("k", 1))
		s.set("k", 2)
		assert.True(t, s.ValueIs("k", 2))
		assert.Equal(t, 1, s.Len())
	})

	t.Run("setMany merges, incoming wins", func(t *testing.T) {
		s := newStore(map[string]int{"a": 1, "b": 2})
		s.setMany(map[string]int{"b": 20, "c": 3})
		assert.Equal(t, 3, s.Len())
		assert.True(t, s.ValueIs("b", 20))
	})

	t.Run("remove ignores missing names", func(t *testing.T) {
		s := newStore(map[string]int{"a": 1, "b": 2, "c": 3})
		s.remove("a", "missing", "c")
		assert.Equal(t, []string{"b"}, s.Keys())
	})

	t.Run("clear empties the store but keeps it usable", func(t *testing.T) {
		s := newStore(map[string]int{"a": 1})
		s.clear()
		assert.True(t, s.IsEmpty())
		s.set("b", 2)
		assert.Equal(t, 1, s.Len())
	})
}

func TestStore_Lookups(t *testing.T) {
	t.Parallel()
	s := newStore(map[string]int{"a": 1, "b": 2})

	t.Run("lookup reports presence", func(t *testing.T) {
		v, ok := s.lookup("a")
		require.True(t, ok)
		assert.Equal(t, 1, v)

		_, ok = s.lookup("missing")
		assert.False(t, ok)
	})

	t.Run("Has, Len, IsEmpty", func(t *testing.T) {
		assert.True(t, s.Has("a"))
		assert.False(t, s.Has("missing"))
		assert.Equal(t, 2, s.Len())
		assert.False(t, s.IsEmpty())
		assert.True(t, newStore(nil).IsEmpty())
	})

	t.Run("ValueOrDefault", func(t *testing.T) {
		assert.Equal(t, 1, s.ValueOrDefault("a", 99))
		assert.Equal(t, 99, s.ValueOrDefault("missing", 99))
	})

	t.Run("ValueIs is false for a missing name", func(t *testing.T) {
		assert.True(t, s.ValueIs("a", 1))
		assert.False(t, s.ValueIs("a", 2))
		assert.False(t, s.ValueIs("missing", 0), "missing name must not match the zero value")
	})

	t.Run("zero value entry is present, not missing", func(t *testing.T) {
		z := newStore(map[string]int{"k": 0})
		assert.True(t, z.Has("k"))
		assert.True(t, z.ValueIs("k", 0))
		assert.Equal(t, 0, z.ValueOrDefault("k", 99))
	})
}

func TestStore_KeysSortedAndAllCloned(t *testing.T) {
	t.Parallel()
	s := newStore(map[string]int{"b": 2, "c": 3, "a": 1})

	assert.Equal(t, []string{"a", "b", "c"}, s.Keys(), "Keys must be sorted")
	assert.Empty(t, newStore(nil).Keys())

	all := s.All()
	all["a"] = 99
	delete(all, "b")
	assert.True(t, s.ValueIs("a", 1), "All must return a defensive copy")
	assert.True(t, s.Has("b"))
}

func TestStore_Predicates(t *testing.T) {
	t.Parallel()
	even := func(_ string, v int) bool { return v%2 == 0 }
	empty := newStore(nil)
	s := newStore(map[string]int{"a": 1, "b": 2, "c": 4})

	t.Run("Every is vacuously true on empty", func(t *testing.T) {
		assert.True(t, empty.Every(even))
		assert.False(t, s.Every(even))
		assert.True(t, s.Every(func(_ string, v int) bool { return v > 0 }))
	})

	t.Run("Any is false on empty", func(t *testing.T) {
		assert.False(t, empty.Any(even))
		assert.True(t, s.Any(even))
		assert.False(t, s.Any(func(_ string, v int) bool { return v > 100 }))
	})

	t.Run("Count", func(t *testing.T) {
		assert.Equal(t, 0, empty.Count(even))
		assert.Equal(t, 2, s.Count(even))
	})
}

func TestStore_ForEach(t *testing.T) {
	t.Parallel()
	s := newStore(map[string]int{"a": 1, "b": 2, "c": 3})

	visited := 0
	require.NoError(t, s.ForEach(func(string, int) error { visited++; return nil }))
	assert.Equal(t, 3, visited)

	sentinel := errors.New("stop")
	visited = 0
	require.ErrorIs(t, s.ForEach(func(string, int) error { visited++; return sentinel }), sentinel)
	assert.Equal(t, 1, visited, "ForEach stops on the first error")
}

func TestStore_FilterInto(t *testing.T) {
	t.Parallel()
	s := newStore(map[string]int{"a": 1, "b": 2, "c": 4})
	dst := map[string]int{}
	s.filterInto(dst, func(_ string, v int) bool { return v%2 == 0 })

	assert.Equal(t, map[string]int{"b": 2, "c": 4}, dst)
	assert.Equal(t, 3, s.Len(), "source is not modified")
}

func TestStore_MergeInto(t *testing.T) {
	t.Parallel()
	s := newStore(map[string]int{"a": 1, "b": 2})
	other := map[string]int{"b": 20, "c": 3}
	dst := map[string]int{}
	s.mergeInto(dst, other)

	assert.Equal(t, map[string]int{"a": 1, "b": 20, "c": 3}, dst, "other wins on conflict")
	assert.True(t, s.ValueIs("b", 2), "receiver is not modified")
	assert.Equal(t, map[string]int{"b": 20, "c": 3}, other, "other is not modified")
}
