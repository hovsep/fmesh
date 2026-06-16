package meta

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScalars_New(t *testing.T) {
	s := NewScalars()
	assert.NotNil(t, s)
	assert.Equal(t, 0, s.Len())
	assert.True(t, s.IsEmpty())
}

func TestScalars_Set(t *testing.T) {
	tests := []struct {
		name       string
		store      *Scalars
		setName    string
		setValue   float64
		assertions func(t *testing.T, result *Scalars)
	}{
		{
			name:     "add to empty store",
			store:    NewScalars(),
			setName:  "temp",
			setValue: 36.6,
			assertions: func(t *testing.T, result *Scalars) {
				assert.Equal(t, 1, result.Len())
				v, err := result.Value("temp")
				require.NoError(t, err)
				assert.InDelta(t, 36.6, v, 1e-9)
			},
		},
		{
			name:     "add to non-empty store",
			store:    NewScalars().Set("x", 1.0),
			setName:  "y",
			setValue: 2.0,
			assertions: func(t *testing.T, result *Scalars) {
				assert.Equal(t, 2, result.Len())
				assert.True(t, result.Has("x"))
				assert.True(t, result.Has("y"))
			},
		},
		{
			name:     "overwrite existing scalar",
			store:    NewScalars().Set("x", 1.0),
			setName:  "x",
			setValue: 99.0,
			assertions: func(t *testing.T, result *Scalars) {
				assert.Equal(t, 1, result.Len())
				v, err := result.Value("x")
				require.NoError(t, err)
				assert.InDelta(t, 99.0, v, 1e-9)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.store.Set(tt.setName, tt.setValue)
			if tt.assertions != nil {
				tt.assertions(t, result)
			}
		})
	}
}

func TestScalars_SetMany(t *testing.T) {
	t.Run("adds multiple scalars", func(t *testing.T) {
		s := NewScalars().SetMany(map[string]float64{"a": 1.0, "b": 2.0})
		assert.Equal(t, 2, s.Len())
	})
	t.Run("upsert semantics", func(t *testing.T) {
		s := NewScalars().Set("a", 1.0).SetMany(map[string]float64{"a": 100.0, "b": 2.0})
		assert.Equal(t, 2, s.Len())
		v, err := s.Value("a")
		require.NoError(t, err)
		assert.InDelta(t, 100.0, v, 1e-9)
	})
}

func TestScalars_Get(t *testing.T) {
	s := NewScalars().Set("x", 42.0)

	t.Run("found", func(t *testing.T) {
		v, err := s.Value("x")
		require.NoError(t, err)
		assert.InDelta(t, 42.0, v, 1e-9)
	})
	t.Run("not found returns zero and false", func(t *testing.T) {
		v, err := s.Value("missing")
		require.Error(t, err)
		assert.InDelta(t, 0.0, v, 1e-9)
	})
}

func TestScalars_GetOrDefault(t *testing.T) {
	s := NewScalars().Set("x", 7.0)

	t.Run("found returns actual value", func(t *testing.T) {
		assert.InDelta(t, 7.0, s.ValueOrDefault("x", -1.0), 1e-9)
	})
	t.Run("not found returns default", func(t *testing.T) {
		assert.InDelta(t, -1.0, s.ValueOrDefault("missing", -1.0), 1e-9)
	})
}

func TestScalars_Has(t *testing.T) {
	s := NewScalars().Set("a", 0.0) // zero value is still "has"
	assert.True(t, s.Has("a"))
	assert.False(t, s.Has("b"))
}

func TestScalars_Remove(t *testing.T) {
	t.Run("removes existing", func(t *testing.T) {
		s := NewScalars().SetMany(map[string]float64{"a": 1, "b": 2}).Remove("a")
		assert.Equal(t, 1, s.Len())
		assert.False(t, s.Has("a"))
	})
	t.Run("missing name is no-op", func(t *testing.T) {
		s := NewScalars().Set("a", 1).Remove("zzz")
		assert.Equal(t, 1, s.Len())
	})
	t.Run("remove multiple", func(t *testing.T) {
		s := NewScalars().SetMany(map[string]float64{"a": 1, "b": 2, "c": 3}).Remove("a", "c")
		assert.Equal(t, 1, s.Len())
		assert.True(t, s.Has("b"))
	})
}

func TestScalars_Clear(t *testing.T) {
	s := NewScalars().SetMany(map[string]float64{"a": 1, "b": 2}).Clear()
	assert.Equal(t, 0, s.Len())
	assert.True(t, s.IsEmpty())
}

func TestScalars_Keys(t *testing.T) {
	t.Run("returns sorted keys", func(t *testing.T) {
		s := NewScalars().SetMany(map[string]float64{"c": 3, "a": 1, "b": 2})
		assert.Equal(t, []string{"a", "b", "c"}, s.Keys())
	})
	t.Run("empty store returns empty slice", func(t *testing.T) {
		assert.Empty(t, NewScalars().Keys())
	})
}

func TestScalars_All(t *testing.T) {
	t.Run("returns defensive copy", func(t *testing.T) {
		s := NewScalars().Set("x", 5.0)
		got := s.All()
		got["x"] = 999.0 // mutate the copy
		v, _ := s.Value("x")
		assert.InDelta(t, 5.0, v, 1e-9, "original must not be affected")
	})
}

func TestScalars_Min(t *testing.T) {
	t.Run("empty store returns ok=false", func(t *testing.T) {
		_, _, ok := NewScalars().Min()
		assert.False(t, ok)
	})
	t.Run("single entry", func(t *testing.T) {
		name, v, ok := NewScalars().Set("only", 3.14).Min()
		require.True(t, ok)
		assert.Equal(t, "only", name)
		assert.InDelta(t, 3.14, v, 1e-9)
	})
	t.Run("multiple entries", func(t *testing.T) {
		s := NewScalars().SetMany(map[string]float64{"a": 5, "b": 2, "c": 8})
		name, v, ok := s.Min()
		require.True(t, ok)
		assert.Equal(t, "b", name)
		assert.InDelta(t, 2.0, v, 1e-9)
	})
	t.Run("negative values", func(t *testing.T) {
		s := NewScalars().SetMany(map[string]float64{"x": -100, "y": 0, "z": 50})
		name, v, ok := s.Min()
		require.True(t, ok)
		assert.Equal(t, "x", name)
		assert.InDelta(t, -100.0, v, 1e-9)
	})
}

func TestScalars_Max(t *testing.T) {
	t.Run("empty store returns ok=false", func(t *testing.T) {
		_, _, ok := NewScalars().Max()
		assert.False(t, ok)
	})
	t.Run("multiple entries", func(t *testing.T) {
		s := NewScalars().SetMany(map[string]float64{"a": 5, "b": 2, "c": 8})
		name, v, ok := s.Max()
		require.True(t, ok)
		assert.Equal(t, "c", name)
		assert.InDelta(t, 8.0, v, 1e-9)
	})
	t.Run("all negative values", func(t *testing.T) {
		s := NewScalars().SetMany(map[string]float64{"x": -10, "y": -3, "z": -50})
		name, v, ok := s.Max()
		require.True(t, ok)
		assert.Equal(t, "y", name)
		assert.InDelta(t, -3.0, v, 1e-9)
	})
}

func TestScalars_Sum(t *testing.T) {
	s := NewScalars().SetMany(map[string]float64{"a": 1, "b": 2, "c": 3})

	t.Run("sum all when no names given", func(t *testing.T) {
		assert.InDelta(t, 6.0, s.Sum(), 1e-9)
	})
	t.Run("sum named subset", func(t *testing.T) {
		assert.InDelta(t, 3.0, s.Sum("a", "b"), 1e-9)
	})
	t.Run("missing name contributes 0", func(t *testing.T) {
		assert.InDelta(t, 1.0, s.Sum("a", "missing"), 1e-9)
	})
	t.Run("empty store sum = 0", func(t *testing.T) {
		assert.InDelta(t, 0.0, NewScalars().Sum(), 1e-9)
	})
}

func TestScalars_Average(t *testing.T) {
	s := NewScalars().SetMany(map[string]float64{"a": 1, "b": 3})

	t.Run("average all", func(t *testing.T) {
		avg, ok := s.Average()
		require.True(t, ok)
		assert.InDelta(t, 2.0, avg, 1e-9)
	})
	t.Run("average named subset", func(t *testing.T) {
		avg, ok := s.Average("a", "b")
		require.True(t, ok)
		assert.InDelta(t, 2.0, avg, 1e-9)
	})
	t.Run("empty store returns ok=false", func(t *testing.T) {
		_, ok := NewScalars().Average()
		assert.False(t, ok)
	})
}

func TestScalars_Scale(t *testing.T) {
	t.Run("scales existing entry", func(t *testing.T) {
		s := NewScalars().Set("x", 5.0).Scale("x", 3.0)
		v, _ := s.Value("x")
		assert.InDelta(t, 15.0, v, 1e-9)
	})
	t.Run("missing name is no-op", func(t *testing.T) {
		s := NewScalars().Set("x", 5.0).Scale("missing", 100.0)
		assert.Equal(t, 1, s.Len())
	})
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

func TestScalars_Every(t *testing.T) {
	t.Run("empty store returns true (vacuous truth)", func(t *testing.T) {
		assert.True(t, NewScalars().Every(func(_ string, _ float64) bool { return false }))
	})
	t.Run("all pass predicate", func(t *testing.T) {
		s := NewScalars().SetMany(map[string]float64{"a": 1, "b": 2})
		assert.True(t, s.Every(func(_ string, v float64) bool { return v > 0 }))
	})
	t.Run("not all pass", func(t *testing.T) {
		s := NewScalars().SetMany(map[string]float64{"a": 1, "b": -1})
		assert.False(t, s.Every(func(_ string, v float64) bool { return v > 0 }))
	})
}

func TestScalars_Any(t *testing.T) {
	t.Run("empty store returns false", func(t *testing.T) {
		assert.False(t, NewScalars().Any(func(_ string, _ float64) bool { return true }))
	})
	t.Run("at least one passes", func(t *testing.T) {
		s := NewScalars().SetMany(map[string]float64{"a": 1, "b": -5})
		assert.True(t, s.Any(func(_ string, v float64) bool { return v < 0 }))
	})
	t.Run("none pass", func(t *testing.T) {
		s := NewScalars().SetMany(map[string]float64{"a": 1, "b": 2})
		assert.False(t, s.Any(func(_ string, v float64) bool { return v < 0 }))
	})
}

func TestScalars_Count(t *testing.T) {
	s := NewScalars().SetMany(map[string]float64{"a": 1, "b": 2, "c": -1})

	t.Run("count matching", func(t *testing.T) {
		assert.Equal(t, 2, s.Count(func(_ string, v float64) bool { return v > 0 }))
	})
	t.Run("count all", func(t *testing.T) {
		assert.Equal(t, 3, s.Count(func(_ string, _ float64) bool { return true }))
	})
	t.Run("count none", func(t *testing.T) {
		assert.Equal(t, 0, s.Count(func(_ string, v float64) bool { return v > 100 }))
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
	t.Run("no matches returns empty", func(t *testing.T) {
		s := NewScalars().Set("x", 1.0)
		result := s.Filter(func(_ string, _ float64) bool { return false })
		assert.True(t, result.IsEmpty())
	})
	t.Run("original is not modified", func(t *testing.T) {
		s := NewScalars().SetMany(map[string]float64{"a": 1, "b": 2})
		_ = s.Filter(func(_ string, v float64) bool { return v > 1 })
		assert.Equal(t, 2, s.Len())
	})
}

func TestScalars_ForEach(t *testing.T) {
	t.Run("visits all entries", func(t *testing.T) {
		s := NewScalars().SetMany(map[string]float64{"a": 1, "b": 2, "c": 3})
		count := 0
		err := s.ForEach(func(_ string, _ float64) error {
			count++
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, 3, count)
	})
	t.Run("stops on first error", func(t *testing.T) {
		s := NewScalars().SetMany(map[string]float64{"a": 1, "b": 2, "c": 3})
		sentinel := errors.New("stop")
		count := 0
		err := s.ForEach(func(_ string, _ float64) error {
			count++
			return sentinel
		})
		require.ErrorIs(t, err, sentinel)
		assert.Equal(t, 1, count)
	})
	t.Run("empty store returns nil", func(t *testing.T) {
		err := NewScalars().ForEach(func(_ string, _ float64) error { return nil })
		require.NoError(t, err)
	})
}

func TestScalars_Chainable(t *testing.T) {
	t.Run("chaining multiple operations", func(t *testing.T) {
		s := NewScalars().
			Set("a", 1.0).
			Set("b", 2.0).
			SetMany(map[string]float64{"c": 3.0, "d": 4.0}).
			Remove("d").
			Scale("a", 10.0)

		assert.Equal(t, 3, s.Len())
		v, _ := s.Value("a")
		assert.InDelta(t, 10.0, v, 1e-9)
		assert.False(t, s.Has("d"))
	})
}
