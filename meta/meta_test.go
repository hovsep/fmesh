package meta

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMeta_Set(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		meta       *Meta
		key        string
		value      string
		assertions func(t *testing.T, m *Meta)
	}{
		{
			name:  "adding to empty store",
			meta:  New(),
			key:   "k1",
			value: "v1",
			assertions: func(t *testing.T, m *Meta) {
				assert.Equal(t, 1, m.Len())
				assert.True(t, m.ValueIs("k1", "v1"))
			},
		},
		{
			name:  "adding next to an entry of the other type",
			meta:  New().Set("n", 1.5),
			key:   "k1",
			value: "v1",
			assertions: func(t *testing.T, m *Meta) {
				assert.Equal(t, 2, m.Len())
				assert.True(t, m.ValueIs("n", 1.5))
				assert.True(t, m.ValueIs("k1", "v1"))
			},
		},
		{
			name:  "overwriting an existing key, even across types",
			meta:  New().Set("k1", 2.0),
			key:   "k1",
			value: "v1",
			assertions: func(t *testing.T, m *Meta) {
				assert.Equal(t, 1, m.Len())
				assert.True(t, m.ValueIs("k1", "v1"))
				assert.False(t, m.ValueIs("k1", 2.0))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertions(t, tt.meta.Set(tt.key, tt.value))
		})
	}
}

func TestMeta_Value(t *testing.T) {
	t.Parallel()
	m := New().Set("s", "text").Set("f", 1.5).Set("i", 3).Set("b", true).Set("u", uint8(7))

	t.Run("string found", func(t *testing.T) {
		got, err := m.Value[string]("s")
		require.NoError(t, err)
		assert.Equal(t, "text", got)
	})

	t.Run("float found", func(t *testing.T) {
		got, err := m.Value[float64]("f")
		require.NoError(t, err)
		assert.InDelta(t, 1.5, got, 1e-9)
	})

	t.Run("int, bool and sized ints found", func(t *testing.T) {
		i, err := m.Value[int]("i")
		require.NoError(t, err)
		assert.Equal(t, 3, i)
		b, err := m.Value[bool]("b")
		require.NoError(t, err)
		assert.True(t, b)
		u, err := m.Value[uint8]("u")
		require.NoError(t, err)
		assert.Equal(t, uint8(7), u)
	})

	// Numbers are not converted: an untyped 3 is stored as int and reads back
	// only as int.
	t.Run("another number type errors", func(t *testing.T) {
		_, err := m.Value[float64]("i")
		require.ErrorContains(t, err, "holds int, not float64")
		_, err = m.Value[int64]("i")
		require.ErrorContains(t, err, "holds int, not int64")
	})

	t.Run("named type reads only as itself", func(t *testing.T) {
		type celsius float64
		named := New().Set("t", celsius(36.6))
		got, err := named.Value[celsius]("t")
		require.NoError(t, err)
		assert.InDelta(t, 36.6, float64(got), 1e-9)
		_, err = named.Value[float64]("t")
		require.ErrorContains(t, err, "not float64")
	})

	t.Run("missing key errors and names the key", func(t *testing.T) {
		got, err := m.Value[string]("zz")
		require.ErrorContains(t, err, "zz")
		assert.Empty(t, got)
	})

	// The one failure mode a split label/scalar store never had: the key exists
	// but holds the other type. It must be an error that names both types.
	t.Run("wrong type errors and names both types", func(t *testing.T) {
		got, err := m.Value[float64]("s")
		require.ErrorContains(t, err, `"s"`)
		require.ErrorContains(t, err, "string")
		require.ErrorContains(t, err, "float64")
		assert.Zero(t, got)
	})
}

func TestMeta_ValueOrDefault(t *testing.T) {
	t.Parallel()
	m := New().Set("s", "text").Set("f", 1.5).Set("i", 3).Set("b", true)

	assert.Equal(t, "text", m.ValueOrDefault("s", "dflt"))
	assert.InDelta(t, 1.5, m.ValueOrDefault("f", 0.0), 1e-9)
	assert.Equal(t, 3, m.ValueOrDefault("i", 0))
	assert.True(t, m.ValueOrDefault("b", false))
	assert.InDelta(t, 9.0, m.ValueOrDefault("i", 9.0), 1e-9, "int entry read as float64 reads as absent")
	assert.Equal(t, "dflt", m.ValueOrDefault("missing", "dflt"))
	assert.InDelta(t, 9.0, m.ValueOrDefault("s", 9.0), 1e-9, "wrong type reads as absent")
}

func TestMeta_ValueIs(t *testing.T) {
	t.Parallel()
	t.Run("key exists with empty value", func(t *testing.T) {
		m := New().Set("k", "")
		assert.True(t, m.ValueIs("k", ""))
		assert.False(t, m.ValueIs("k", "anything"))
	})

	t.Run("key absent returns false", func(t *testing.T) {
		m := New()
		assert.False(t, m.ValueIs("missing", ""))
		assert.False(t, m.ValueIs("missing", 0.0))
	})

	t.Run("same key, other type, is not a match", func(t *testing.T) {
		m := New().Set("k", "1")
		assert.False(t, m.ValueIs("k", 1.0))
	})

	t.Run("int and bool match only their own type", func(t *testing.T) {
		m := New().Set("n", 1).Set("ok", true)
		assert.True(t, m.ValueIs("n", 1))
		assert.False(t, m.ValueIs("n", 1.0))
		assert.True(t, m.ValueIs("ok", true))
		assert.False(t, m.ValueIs("ok", false))
	})
}

func TestMeta_Has(t *testing.T) {
	t.Parallel()
	m := New().SetMany(map[string]string{"a": "1", "b": "2"}).Set("c", 3.0)

	assert.True(t, m.Has("a"))
	assert.True(t, m.Has("a", "c"), "keys of both types count")
	assert.False(t, m.Has("a", "zz"), "one missing key fails the whole check")
	assert.True(t, m.Has(), "no keys is vacuously true")
	assert.False(t, New().Has("a"))
}

func TestMeta_HasAny(t *testing.T) {
	t.Parallel()
	m := New().SetMany(map[string]string{"a": "1", "b": "2"})

	assert.True(t, m.HasAny("a", "zz"))
	assert.False(t, m.HasAny("zz", "yy"))
	assert.False(t, m.HasAny(), "no keys matches nothing")
	assert.False(t, New().HasAny("a"))
}

func TestMeta_Keys_sorted(t *testing.T) {
	t.Parallel()
	m := New().SetMany(map[string]string{"b": "beta", "a": "alpha"}).Set("c", 1.0)
	assert.Equal(t, []string{"a", "b", "c"}, m.Keys())
	assert.Empty(t, New().Keys())
}

func TestMeta_Len_IsEmpty(t *testing.T) {
	t.Parallel()
	m := New()
	assert.True(t, m.IsEmpty())
	assert.Equal(t, 0, m.Len())

	m.Set("k", "v")
	assert.False(t, m.IsEmpty())
	assert.Equal(t, 1, m.Len())
}

func TestMeta_Chainable(t *testing.T) {
	t.Parallel()
	t.Run("chaining mixed mutators", func(t *testing.T) {
		m := New().
			Set("env", "dev").
			Set("weight", 0.5).
			SetMany(map[string]string{"region": "us-east", "zone": "1a"}).
			SetMany(map[string]float64{"retries": 3}).
			Remove("zone")

		assert.Equal(t, 4, m.Len())
		assert.True(t, m.Has("env", "weight", "region", "retries"))
		assert.False(t, m.Has("zone"))
	})

	// design.md documents Clear().SetMany(m) as the way to replace everything,
	// so the chain has to survive a Clear in the middle of it.
	t.Run("Clear in the middle of a chain replaces everything", func(t *testing.T) {
		m := New().
			Set("old", "value").
			Clear().
			SetMany(map[string]string{"new": "value"})

		assert.Equal(t, 1, m.Len())
		assert.False(t, m.Has("old"))
		assert.True(t, m.ValueIs("new", "value"))
	})

	t.Run("SetMany called twice merges entries", func(t *testing.T) {
		m := New().
			SetMany(map[string]string{"k1": "v1", "k2": "v2"}).
			SetMany(map[string]string{"k3": "v3", "k2": "v2-updated"})

		assert.Equal(t, 3, m.Len())
		assert.True(t, m.ValueIs("k2", "v2-updated"), "should update existing key")
	})

	t.Run("Remove ignores missing keys", func(t *testing.T) {
		m := New().Set("k1", "v1").Remove("k1", "nope")
		assert.True(t, m.IsEmpty())
	})
}

// All must return a defensive copy of the internal map, not a live reference.
// The store itself is mutable by design; this guards only against callers
// reaching inside and corrupting internal state (#203).
func TestMeta_All_returnsDefensiveCopy(t *testing.T) {
	t.Parallel()
	m := New().Set("k", "v").Set("n", 1.0)
	all := m.All()

	all["k"] = "mutated"
	delete(all, "n")

	assert.True(t, m.ValueIs("k", "v"))
	assert.True(t, m.ValueIs("n", 1.0))
	assert.Equal(t, map[string]any{"k": "v", "n": 1.0}, m.All())
}

func TestMeta_Clone(t *testing.T) {
	t.Parallel()
	t.Run("copy is independent in both directions", func(t *testing.T) {
		orig := New().Set("k", "v")
		c := orig.Clone()

		c.Set("k", "changed").Set("extra", 1.0)
		orig.Set("only-orig", "x")

		assert.True(t, orig.ValueIs("k", "v"))
		assert.False(t, orig.Has("extra"))
		assert.True(t, c.ValueIs("k", "changed"))
		assert.False(t, c.Has("only-orig"))
	})

	// A zero-value Signal has no store; cloneSignal relies on this being safe.
	t.Run("nil receiver clones to an empty store", func(t *testing.T) {
		var m *Meta
		c := m.Clone()
		require.NotNil(t, c)
		assert.True(t, c.IsEmpty())
		assert.True(t, c.Set("k", "v").ValueIs("k", "v"), "the result is a usable store")
	})
}

func TestMeta_Filter(t *testing.T) {
	t.Parallel()
	t.Run("keeps only matching entries, original untouched", func(t *testing.T) {
		m := New().
			SetMany(map[string]string{"app.env": "production", "app.tier": "backend"}).
			Set("system.cpu", 0.9)
		result := m.Filter(func(k string, _ any) bool { return strings.HasPrefix(k, "app.") })

		assert.Equal(t, 2, result.Len())
		assert.True(t, result.Has("app.env", "app.tier"))
		assert.False(t, result.Has("system.cpu"))
		assert.Equal(t, 3, m.Len(), "original is not modified")
	})

	t.Run("predicate sees the typed value", func(t *testing.T) {
		m := New().Set("s", "text").Set("f", 1.5)
		floats := m.Filter(func(_ string, v any) bool { _, ok := v.(float64); return ok })
		assert.Equal(t, []string{"f"}, floats.Keys())
	})

	t.Run("no matches returns empty store", func(t *testing.T) {
		m := New().Set("k", "v")
		assert.True(t, m.Filter(func(string, any) bool { return false }).IsEmpty())
	})
}
