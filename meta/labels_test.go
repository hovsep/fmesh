package meta

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The generic read surface Labels shares with Scalars is tested once in
// store_test.go; this file covers what the Labels facade adds: chainable
// mutators, Value's error, and the label-oriented set operations.

func TestLabelsCollection_Set(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		collection *Labels
		label      string
		value      string
		assertions func(t *testing.T, result *Labels)
	}{
		{
			name:       "adding to empty collection",
			collection: NewLabels(),
			label:      "l1",
			value:      "v1",
			assertions: func(t *testing.T, result *Labels) {
				assert.Equal(t, 1, result.Len())
				value, err := result.Value("l1")
				require.NoError(t, err)
				assert.Equal(t, "v1", value)
			},
		},
		{
			name:       "adding to non-empty collection",
			collection: NewLabels().Set("l1", "v1"),
			label:      "l2",
			value:      "v2",
			assertions: func(t *testing.T, result *Labels) {
				assert.Equal(t, 2, result.Len())
				assert.True(t, result.Has("l1"))
				assert.True(t, result.Has("l2"))
			},
		},
		{
			name: "overwriting existing label",
			collection: NewLabels().SetMany(map[string]string{
				"l1": "v1",
				"l2": "v2",
			}),
			label: "l2",
			value: "v3",
			assertions: func(t *testing.T, result *Labels) {
				assert.Equal(t, 2, result.Len())
				assert.True(t, result.ValueIs("l2", "v3"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.collection.Set(tt.label, tt.value)
			if tt.assertions != nil {
				tt.assertions(t, result)
			}
		})
	}
}

func TestLabelsCollection_Value(t *testing.T) {
	t.Parallel()
	t.Run("label found", func(t *testing.T) {
		c := NewLabels().SetMany(map[string]string{"l1": "v1", "l2": "v2"})
		got, err := c.Value("l2")
		require.NoError(t, err)
		assert.Equal(t, "v2", got)
	})

	t.Run("missing label errors and names the label", func(t *testing.T) {
		c := NewLabels().Set("l1", "v1")
		got, err := c.Value("l3")
		require.ErrorContains(t, err, "l3")
		assert.Empty(t, got)
	})
}

func TestLabelsCollection_ValueIs_edge_cases(t *testing.T) {
	t.Parallel()
	t.Run("key exists with empty value", func(t *testing.T) {
		c := NewLabels().Set("k", "")
		assert.True(t, c.ValueIs("k", ""))
		assert.False(t, c.ValueIs("k", "anything"))
	})

	t.Run("key absent returns false", func(t *testing.T) {
		c := NewLabels()
		assert.False(t, c.ValueIs("missing", ""))
		assert.False(t, c.ValueIs("missing", "val"))
	})
}

func TestLabelsCollection_HasAll(t *testing.T) {
	t.Parallel()
	c := NewLabels().SetMany(map[string]string{"l1": "v1", "l2": "v2", "l3": "v3"})

	assert.True(t, c.HasAll("l1", "l2"))
	assert.False(t, c.HasAll("l1", "l2", "l4"), "one missing name fails the whole check")
	assert.True(t, c.HasAll(), "empty name list is vacuously true")
	assert.False(t, NewLabels().HasAll("l1"))
}

func TestLabelsCollection_HasAny(t *testing.T) {
	t.Parallel()
	c := NewLabels().SetMany(map[string]string{"l1": "v1", "l2": "v2"})

	assert.True(t, c.HasAny("l1", "l10"))
	assert.False(t, c.HasAny("l10", "l20"))
	assert.False(t, c.HasAny(), "empty name list matches nothing")
	assert.False(t, NewLabels().HasAny("l1"))
}

func TestLabelsCollection_Values(t *testing.T) {
	t.Parallel()
	t.Run("returns values sorted by key", func(t *testing.T) {
		c := NewLabels().SetMany(map[string]string{"b": "beta", "a": "alpha", "c": "gamma"})
		assert.Equal(t, []string{"alpha", "beta", "gamma"}, c.Values())
	})

	t.Run("empty collection", func(t *testing.T) {
		assert.Empty(t, NewLabels().Values())
	})
}

func TestLabelsCollection_Map(t *testing.T) {
	t.Parallel()
	t.Run("empty collection returns empty", func(t *testing.T) {
		result := NewLabels().Map(func(k, v string) (string, string) {
			return strings.ToUpper(k), strings.ToUpper(v)
		})
		assert.Equal(t, 0, result.Len())
	})

	t.Run("transforms both keys and values into a new collection", func(t *testing.T) {
		c := NewLabels().SetMany(map[string]string{"env": "dev", "tier": "frontend"})
		result := c.Map(func(k, v string) (string, string) {
			return "system." + k, "[" + v + "]"
		})

		assert.Equal(t, 2, result.Len())
		assert.True(t, result.ValueIs("system.env", "[dev]"))
		assert.True(t, result.ValueIs("system.tier", "[frontend]"))
		assert.True(t, c.ValueIs("env", "dev"), "original is not modified")
	})
}

func TestLabelsCollection_Filter(t *testing.T) {
	t.Parallel()
	t.Run("keeps only matching labels", func(t *testing.T) {
		c := NewLabels().SetMany(map[string]string{
			"app.env":    "production",
			"app.tier":   "backend",
			"system.cpu": "high",
		})
		result := c.Filter(func(k, v string) bool { return strings.HasPrefix(k, "app.") })

		assert.Equal(t, 2, result.Len())
		assert.True(t, result.Has("app.env"))
		assert.False(t, result.Has("system.cpu"))
		assert.Equal(t, 3, c.Len(), "original is not modified")
	})

	t.Run("no matches returns empty collection", func(t *testing.T) {
		c := NewLabels().SetMany(map[string]string{"l1": "v1", "l2": "v2"})
		assert.Equal(t, 0, c.Filter(func(k, v string) bool { return false }).Len())
	})
}

func TestLabelsCollection_Merge(t *testing.T) {
	t.Parallel()
	t.Run("merges two collections, other wins on conflict", func(t *testing.T) {
		a := NewLabels().SetMany(map[string]string{"x": "1", "y": "2"})
		b := NewLabels().SetMany(map[string]string{"y": "overridden", "z": "3"})
		merged := a.Merge(b)

		assert.Equal(t, 3, merged.Len())
		assert.True(t, merged.ValueIs("x", "1"))
		assert.True(t, merged.ValueIs("y", "overridden"))
		assert.True(t, merged.ValueIs("z", "3"))
	})

	t.Run("neither input is modified", func(t *testing.T) {
		a := NewLabels().Set("k", "v")
		b := NewLabels().Set("k", "other")
		_ = a.Merge(b)
		assert.True(t, a.ValueIs("k", "v"))
		assert.True(t, b.ValueIs("k", "other"))
	})

	t.Run("merge with empty other", func(t *testing.T) {
		merged := NewLabels().Set("k", "v").Merge(NewLabels())
		assert.Equal(t, 1, merged.Len())
		assert.True(t, merged.ValueIs("k", "v"))
	})
}

func TestLabelsCollection_HasAllFrom(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		a    *Labels
		b    *Labels
		want bool
	}{
		{
			name: "both empty → true",
			a:    NewLabels(),
			b:    NewLabels(),
			want: true,
		},
		{
			name: "a contains all labels from b, values ignored",
			a:    NewLabels().SetMany(map[string]string{"x": "1", "y": "2"}),
			b:    NewLabels().SetMany(map[string]string{"x": "ignored"}),
			want: true,
		},
		{
			name: "a missing some labels from b",
			a:    NewLabels().SetMany(map[string]string{"x": "1"}),
			b:    NewLabels().SetMany(map[string]string{"x": "1", "y": "2"}),
			want: false,
		},
		{
			name: "len optimization: b larger than a → false",
			a:    NewLabels().SetMany(map[string]string{"x": "1"}),
			b:    NewLabels().SetMany(map[string]string{"x": "1", "y": "2", "z": "3"}),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.a.HasAllFrom(tt.b))
		})
	}
}

func TestLabelsCollection_HasAnyFrom(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		a    *Labels
		b    *Labels
		want bool
	}{
		{
			name: "both empty → false",
			a:    NewLabels(),
			b:    NewLabels(),
			want: false,
		},
		{
			name: "at least one label matches, values ignored",
			a:    NewLabels().SetMany(map[string]string{"x": "1", "y": "2"}),
			b:    NewLabels().SetMany(map[string]string{"y": "ignored"}),
			want: true,
		},
		{
			name: "no labels match",
			a:    NewLabels().SetMany(map[string]string{"x": "1"}),
			b:    NewLabels().SetMany(map[string]string{"y": "2"}),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.a.HasAnyFrom(tt.b))
		})
	}
}

func TestLabelsCollection_Chainable(t *testing.T) {
	t.Parallel()
	t.Run("chaining multiple operations", func(t *testing.T) {
		lc := NewLabels().
			Set("env", "dev").
			Set("tier", "backend").
			SetMany(map[string]string{
				"region": "us-east",
				"zone":   "1a",
			}).
			Remove("zone")

		assert.Equal(t, 3, lc.Len())
		assert.True(t, lc.Has("env"))
		assert.True(t, lc.Has("tier"))
		assert.True(t, lc.Has("region"))
		assert.False(t, lc.Has("zone"))
	})

	// design.md documents Clear().SetMany(m) as the way to replace every label,
	// so the chain has to survive a Clear in the middle of it.
	t.Run("Clear in the middle of a chain replaces everything", func(t *testing.T) {
		lc := NewLabels().
			Set("old", "value").
			Clear().
			SetMany(map[string]string{"new": "value"})

		assert.Equal(t, 1, lc.Len())
		assert.False(t, lc.Has("old"))
		assert.True(t, lc.ValueIs("new", "value"))
	})

	t.Run("SetMany called twice merges labels", func(t *testing.T) {
		lc := NewLabels().
			SetMany(map[string]string{"k1": "v1", "k2": "v2"}).
			SetMany(map[string]string{"k3": "v3", "k2": "v2-updated"})

		assert.Equal(t, 3, lc.Len())
		assert.True(t, lc.ValueIs("k1", "v1"))
		assert.True(t, lc.ValueIs("k2", "v2-updated"), "should update existing key")
		assert.True(t, lc.ValueIs("k3", "v3"))
	})

	t.Run("Remove called twice removes both sets", func(t *testing.T) {
		lc := NewLabels().SetMany(map[string]string{"k1": "v1", "k2": "v2", "k3": "v3", "k4": "v4"}).
			Remove("k1", "k2").
			Remove("k3")

		assert.Equal(t, 1, lc.Len())
		assert.True(t, lc.ValueIs("k4", "v4"))
		assert.False(t, lc.Has("k1"))
		assert.False(t, lc.Has("k2"))
		assert.False(t, lc.Has("k3"))
	})
}
