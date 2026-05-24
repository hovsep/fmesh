package meta

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLabelsCollection_New(t *testing.T) {
	c := NewLabels()
	assert.NotNil(t, c)
	assert.Equal(t, 0, c.Len())
}

func TestLabelsCollection_Add(t *testing.T) {
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

func TestLabelsCollection_AddMany(t *testing.T) {
	tests := []struct {
		name       string
		collection *Labels
		labels     map[string]string
		assertions func(t *testing.T, result *Labels)
	}{
		{
			name:       "adding to empty collection",
			collection: NewLabels(),
			labels: map[string]string{
				"l1": "v1",
				"l2": "v2",
			},
			assertions: func(t *testing.T, result *Labels) {
				assert.Equal(t, 2, result.Len())
			},
		},
		{
			name: "adding to non-empty collection with overwrite",
			collection: NewLabels().SetMany(map[string]string{
				"l1": "v1",
				"l2": "v2",
				"l3": "v3",
			}),
			labels: map[string]string{
				"l3": "v100",
				"l4": "v4",
				"l5": "v5",
			},
			assertions: func(t *testing.T, result *Labels) {
				assert.Equal(t, 5, result.Len())
				assert.True(t, result.ValueIs("l3", "v100"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.collection.SetMany(tt.labels)
			if tt.assertions != nil {
				tt.assertions(t, result)
			}
		})
	}
}

func TestLabelsCollection_Remove(t *testing.T) {
	tests := []struct {
		name       string
		collection *Labels
		labels     []string
		assertions func(t *testing.T, result *Labels)
	}{
		{
			name: "label found and deleted",
			collection: NewLabels().SetMany(map[string]string{
				"l1": "v1",
				"l2": "v2",
			}),
			labels: []string{"l1"},
			assertions: func(t *testing.T, result *Labels) {
				assert.Equal(t, 1, result.Len())
				assert.False(t, result.Has("l1"))
				assert.True(t, result.Has("l2"))
			},
		},
		{
			name: "label not found is no-op",
			collection: NewLabels().SetMany(map[string]string{
				"l1": "v1",
				"l2": "v2",
			}),
			labels: []string{"l3"},
			assertions: func(t *testing.T, result *Labels) {
				assert.Equal(t, 2, result.Len())
			},
		},
		{
			name: "delete multiple labels",
			collection: NewLabels().SetMany(map[string]string{
				"l1": "v1",
				"l2": "v2",
				"l3": "v3",
			}),
			labels: []string{"l1", "l3"},
			assertions: func(t *testing.T, result *Labels) {
				assert.Equal(t, 1, result.Len())
				assert.False(t, result.Has("l1"))
				assert.True(t, result.Has("l2"))
				assert.False(t, result.Has("l3"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.collection.Remove(tt.labels...)
			if tt.assertions != nil {
				tt.assertions(t, result)
			}
		})
	}
}

func TestLabelsCollection_Has(t *testing.T) {
	tests := []struct {
		name       string
		collection *Labels
		label      string
		want       bool
	}{
		{
			name:       "empty collection",
			collection: NewLabels(),
			label:      "l1",
			want:       false,
		},
		{
			name: "label exists",
			collection: NewLabels().SetMany(map[string]string{
				"l1": "v1",
				"l2": "v2",
			}),
			label: "l1",
			want:  true,
		},
		{
			name: "label does not exist",
			collection: NewLabels().SetMany(map[string]string{
				"l1": "v1",
				"l2": "v2",
			}),
			label: "l3",
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.collection.Has(tt.label))
		})
	}
}

func TestLabelsCollection_HasAll(t *testing.T) {
	tests := []struct {
		name       string
		collection *Labels
		labels     []string
		want       bool
	}{
		{
			name:       "empty collection",
			collection: NewLabels(),
			labels:     []string{"l1"},
			want:       false,
		},
		{
			name: "has all labels",
			collection: NewLabels().SetMany(map[string]string{
				"l1": "v1",
				"l2": "v2",
				"l3": "v3",
			}),
			labels: []string{"l1", "l2"},
			want:   true,
		},
		{
			name: "missing some labels",
			collection: NewLabels().SetMany(map[string]string{
				"l1": "v1",
				"l2": "v2",
				"l3": "v3",
			}),
			labels: []string{"l1", "l2", "l4"},
			want:   false,
		},
		{
			name: "empty labels list returns true",
			collection: NewLabels().SetMany(map[string]string{
				"l1": "v1",
			}),
			labels: []string{},
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.collection.HasAll(tt.labels...))
		})
	}
}

func TestLabelsCollection_HasAny(t *testing.T) {
	tests := []struct {
		name       string
		collection *Labels
		labels     []string
		want       bool
	}{
		{
			name:       "empty collection",
			collection: NewLabels(),
			labels:     []string{"l1"},
			want:       false,
		},
		{
			name: "has at least one label",
			collection: NewLabels().SetMany(map[string]string{
				"l1": "v1",
				"l2": "v2",
				"l3": "v3",
			}),
			labels: []string{"l1", "l10"},
			want:   true,
		},
		{
			name: "has none of the labels",
			collection: NewLabels().SetMany(map[string]string{
				"l1": "v1",
				"l2": "v2",
				"l3": "v3",
			}),
			labels: []string{"l10", "l20", "l4"},
			want:   false,
		},
		{
			name: "empty labels list returns false",
			collection: NewLabels().SetMany(map[string]string{
				"l1": "v1",
			}),
			labels: []string{},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.collection.HasAny(tt.labels...))
		})
	}
}

func TestLabelsCollection_Value(t *testing.T) {
	tests := []struct {
		name       string
		collection *Labels
		label      string
		want       string
		wantErr    bool
	}{
		{
			name:       "label not found in empty collection",
			collection: NewLabels(),
			label:      "l1",
			want:       "",
			wantErr:    true,
		},
		{
			name: "label found",
			collection: NewLabels().SetMany(map[string]string{
				"l1": "v1",
				"l2": "v2",
			}),
			label:   "l2",
			want:    "v2",
			wantErr: false,
		},
		{
			name: "label not found in non-empty collection",
			collection: NewLabels().SetMany(map[string]string{
				"l1": "v1",
				"l2": "v2",
			}),
			label:   "l3",
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.collection.Value(tt.label)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLabelsCollection_ValueOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		collection   *Labels
		label        string
		defaultValue string
		want         string
	}{
		{
			name:         "label not found returns default",
			collection:   NewLabels(),
			label:        "l1",
			defaultValue: "default",
			want:         "default",
		},
		{
			name: "label found returns actual value",
			collection: NewLabels().SetMany(map[string]string{
				"l1": "v1",
			}),
			label:        "l1",
			defaultValue: "default",
			want:         "v1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.collection.ValueOrDefault(tt.label, tt.defaultValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLabelsCollection_ValueIs(t *testing.T) {
	tests := []struct {
		name       string
		collection *Labels
		label      string
		value      string
		want       bool
	}{
		{
			name:       "empty collection",
			collection: NewLabels(),
			label:      "l1",
			value:      "v1",
			want:       false,
		},
		{
			name: "label exists with matching value",
			collection: NewLabels().SetMany(map[string]string{
				"l1": "v1",
				"l2": "v2",
			}),
			label: "l1",
			value: "v1",
			want:  true,
		},
		{
			name: "label exists with different value",
			collection: NewLabels().SetMany(map[string]string{
				"l1": "v1",
				"l2": "v2",
			}),
			label: "l1",
			value: "v999",
			want:  false,
		},
		{
			name: "label does not exist",
			collection: NewLabels().SetMany(map[string]string{
				"l1": "v1",
			}),
			label: "l2",
			value: "v2",
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.collection.ValueIs(tt.label, tt.value))
		})
	}
}

func TestLabelsCollection_Every(t *testing.T) {
	tests := []struct {
		name       string
		collection *Labels
		predicate  Predicate
		want       bool
	}{
		{
			name:       "empty collection returns true (vacuous truth)",
			collection: NewLabels(),
			predicate:  func(k, v string) bool { return false },
			want:       true,
		},
		{
			name: "all labels satisfy predicate",
			collection: NewLabels().SetMany(map[string]string{
				"l1": "v1",
				"l2": "v2",
				"l3": "v3",
			}),
			predicate: func(k, v string) bool { return strings.HasPrefix(k, "l") },
			want:      true,
		},
		{
			name: "not all labels satisfy predicate",
			collection: NewLabels().SetMany(map[string]string{
				"l1":  "v1",
				"l2":  "v2",
				"xyz": "v3",
			}),
			predicate: func(k, v string) bool { return strings.HasPrefix(k, "l") },
			want:      false,
		},
		{
			name: "all values non-empty",
			collection: NewLabels().SetMany(map[string]string{
				"l1": "v1",
				"l2": "v2",
			}),
			predicate: func(k, v string) bool { return v != "" },
			want:      true,
		},
		{
			name: "some values empty",
			collection: NewLabels().SetMany(map[string]string{
				"l1": "v1",
				"l2": "",
			}),
			predicate: func(k, v string) bool { return v != "" },
			want:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.collection.Every(tt.predicate))
		})
	}
}

func TestLabelsCollection_Any(t *testing.T) {
	tests := []struct {
		name       string
		collection *Labels
		predicate  Predicate
		want       bool
	}{
		{
			name:       "empty collection returns false",
			collection: NewLabels(),
			predicate:  func(k, v string) bool { return true },
			want:       false,
		},
		{
			name: "at least one label satisfies predicate",
			collection: NewLabels().SetMany(map[string]string{
				"l1":  "v1",
				"l2":  "v2",
				"xyz": "v3",
			}),
			predicate: func(k, v string) bool { return strings.HasPrefix(k, "x") },
			want:      true,
		},
		{
			name: "no labels satisfy predicate",
			collection: NewLabels().SetMany(map[string]string{
				"l1": "v1",
				"l2": "v2",
				"l3": "v3",
			}),
			predicate: func(k, v string) bool { return strings.HasPrefix(k, "x") },
			want:      false,
		},
		{
			name: "at least one value matches condition",
			collection: NewLabels().SetMany(map[string]string{
				"env":  "production",
				"tier": "frontend",
			}),
			predicate: func(k, v string) bool { return v == "production" },
			want:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.collection.Any(tt.predicate))
		})
	}
}

func TestLabelsCollection_Len(t *testing.T) {
	tests := []struct {
		name       string
		collection *Labels
		want       int
	}{
		{
			name:       "empty collection",
			collection: NewLabels(),
			want:       0,
		},
		{
			name: "collection with labels",
			collection: NewLabels().SetMany(map[string]string{
				"l1": "v1",
				"l2": "v2",
				"l3": "v3",
			}),
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.collection.Len())
		})
	}
}

func TestLabelsCollection_Map(t *testing.T) {
	tests := []struct {
		name       string
		collection *Labels
		mapper     Mapper
		assertions func(t *testing.T, result *Labels)
	}{
		{
			name:       "empty collection returns empty",
			collection: NewLabels(),
			mapper:     func(k, v string) (string, string) { return strings.ToUpper(k), strings.ToUpper(v) },
			assertions: func(t *testing.T, result *Labels) {
				assert.Equal(t, 0, result.Len())
			},
		},
		{
			name: "transform keys to uppercase",
			collection: NewLabels().SetMany(map[string]string{
				"env":  "production",
				"tier": "backend",
			}),
			mapper: func(k, v string) (string, string) { return strings.ToUpper(k), v },
			assertions: func(t *testing.T, result *Labels) {
				assert.Equal(t, 2, result.Len())
				assert.True(t, result.Has("ENV"))
				assert.True(t, result.Has("TIER"))
				assert.True(t, result.ValueIs("ENV", "production"))
				assert.True(t, result.ValueIs("TIER", "backend"))
			},
		},
		{
			name: "transform values to uppercase",
			collection: NewLabels().SetMany(map[string]string{
				"env":  "production",
				"tier": "backend",
			}),
			mapper: func(k, v string) (string, string) { return k, strings.ToUpper(v) },
			assertions: func(t *testing.T, result *Labels) {
				assert.Equal(t, 2, result.Len())
				assert.True(t, result.ValueIs("env", "PRODUCTION"))
				assert.True(t, result.ValueIs("tier", "BACKEND"))
			},
		},
		{
			name: "add prefix to keys",
			collection: NewLabels().SetMany(map[string]string{
				"region": "us-east",
				"zone":   "1a",
			}),
			mapper: func(k, v string) (string, string) { return "app." + k, v },
			assertions: func(t *testing.T, result *Labels) {
				assert.Equal(t, 2, result.Len())
				assert.True(t, result.Has("app.region"))
				assert.True(t, result.Has("app.zone"))
				assert.True(t, result.ValueIs("app.region", "us-east"))
			},
		},
		{
			name: "transform both keys and values",
			collection: NewLabels().SetMany(map[string]string{
				"env":  "dev",
				"tier": "frontend",
			}),
			mapper: func(k, v string) (string, string) {
				return "system." + k, "[" + v + "]"
			},
			assertions: func(t *testing.T, result *Labels) {
				assert.Equal(t, 2, result.Len())
				assert.True(t, result.ValueIs("system.env", "[dev]"))
				assert.True(t, result.ValueIs("system.tier", "[frontend]"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.collection.Map(tt.mapper)
			if tt.assertions != nil {
				tt.assertions(t, result)
			}
		})
	}
}

func TestLabelsCollection_Filter(t *testing.T) {
	tests := []struct {
		name       string
		collection *Labels
		predicate  Predicate
		assertions func(t *testing.T, result *Labels)
	}{
		{
			name:       "empty collection returns empty",
			collection: NewLabels(),
			predicate:  func(k, v string) bool { return true },
			assertions: func(t *testing.T, result *Labels) {
				assert.Equal(t, 0, result.Len())
			},
		},
		{
			name: "filter by key prefix",
			collection: NewLabels().SetMany(map[string]string{
				"app.env":    "production",
				"app.tier":   "backend",
				"system.cpu": "high",
			}),
			predicate: func(k, v string) bool { return strings.HasPrefix(k, "app.") },
			assertions: func(t *testing.T, result *Labels) {
				assert.Equal(t, 2, result.Len())
				assert.True(t, result.Has("app.env"))
				assert.True(t, result.Has("app.tier"))
				assert.False(t, result.Has("system.cpu"))
			},
		},
		{
			name: "filter by value",
			collection: NewLabels().SetMany(map[string]string{
				"env1": "production",
				"env2": "dev",
				"env3": "production",
			}),
			predicate: func(k, v string) bool { return v == "production" },
			assertions: func(t *testing.T, result *Labels) {
				assert.Equal(t, 2, result.Len())
				assert.True(t, result.Has("env1"))
				assert.True(t, result.Has("env3"))
				assert.False(t, result.Has("env2"))
			},
		},
		{
			name: "no matches returns empty collection",
			collection: NewLabels().SetMany(map[string]string{
				"l1": "v1",
				"l2": "v2",
			}),
			predicate: func(k, v string) bool { return false },
			assertions: func(t *testing.T, result *Labels) {
				assert.Equal(t, 0, result.Len())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.collection.Filter(tt.predicate)
			if tt.assertions != nil {
				tt.assertions(t, result)
			}
		})
	}
}

func TestLabelsCollection_Chainable(t *testing.T) {
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

func TestLabelsCollection_ErrorHandling(t *testing.T) {
	t.Run("query methods return safe defaults on empty collection", func(t *testing.T) {
		lc := NewLabels()
		assert.False(t, lc.Has("test"))
		assert.Equal(t, 0, lc.Len())
		assert.True(t, lc.Every(func(k, v string) bool { return true })) // vacuous truth
		assert.False(t, lc.Any(func(k, v string) bool { return true }))  // empty

		val, err := lc.Value("test")
		require.Error(t, err)
		assert.Empty(t, val)
	})
}

func TestLabelsCollection_HasAllFrom(t *testing.T) {
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
			name: "a empty, b non-empty → false",
			a:    NewLabels(),
			b:    NewLabels().SetMany(map[string]string{"x": "1"}),
			want: false,
		},
		{
			name: "b empty → true",
			a:    NewLabels().SetMany(map[string]string{"x": "1"}),
			b:    NewLabels(),
			want: true,
		},
		{
			name: "a contains all labels from b",
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
			got := tt.a.HasAllFrom(tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLabelsCollection_HasAnyFrom(t *testing.T) {
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
			name: "a empty → false",
			a:    NewLabels(),
			b:    NewLabels().SetMany(map[string]string{"x": "1"}),
			want: false,
		},
		{
			name: "b empty → false",
			a:    NewLabels().SetMany(map[string]string{"x": "1"}),
			b:    NewLabels(),
			want: false,
		},
		{
			name: "at least one label matches",
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
		{
			name: "multiple labels, still one match",
			a:    NewLabels().SetMany(map[string]string{"l1": "v1", "l2": "v2", "l3": "v3"}),
			b:    NewLabels().SetMany(map[string]string{"l10": "v", "l2": "v"}),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a.HasAnyFrom(tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLabelsCollection_Count(t *testing.T) {
	tests := []struct {
		name       string
		collection *Labels
		pred       Predicate
		want       int
	}{
		{
			name:       "empty collection → 0",
			collection: NewLabels(),
			pred:       func(_, _ string) bool { return true },
			want:       0,
		},
		{
			name: "match none",
			collection: NewLabels().SetMany(map[string]string{
				"a": "1",
				"b": "2",
			}),
			pred: func(_, v string) bool { return v == "zzz" },
			want: 0,
		},
		{
			name: "match some",
			collection: NewLabels().SetMany(map[string]string{
				"a": "1",
				"b": "2",
				"c": "2",
			}),
			pred: func(_, v string) bool { return v == "2" },
			want: 2,
		},
		{
			name: "match all",
			collection: NewLabels().SetMany(map[string]string{
				"a": "x",
				"b": "y",
			}),
			pred: func(_, _ string) bool { return true },
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.collection.Count(tt.pred)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLabelsCollection_Keys(t *testing.T) {
	t.Run("returns sorted keys", func(t *testing.T) {
		c := NewLabels().SetMany(map[string]string{"b": "2", "a": "1", "c": "3"})
		assert.Equal(t, []string{"a", "b", "c"}, c.Keys())
	})

	t.Run("empty collection", func(t *testing.T) {
		assert.Empty(t, NewLabels().Keys())
	})
}

func TestLabelsCollection_Values(t *testing.T) {
	t.Run("returns values sorted by key", func(t *testing.T) {
		c := NewLabels().SetMany(map[string]string{"b": "beta", "a": "alpha", "c": "gamma"})
		assert.Equal(t, []string{"alpha", "beta", "gamma"}, c.Values())
	})

	t.Run("empty collection", func(t *testing.T) {
		assert.Empty(t, NewLabels().Values())
	})
}

func TestLabelsCollection_Merge(t *testing.T) {
	t.Run("merges two collections", func(t *testing.T) {
		a := NewLabels().SetMany(map[string]string{"x": "1", "y": "2"})
		b := NewLabels().SetMany(map[string]string{"y": "overridden", "z": "3"})
		merged := a.Merge(b)

		assert.Equal(t, 3, merged.Len())
		assert.True(t, merged.ValueIs("x", "1"))
		assert.True(t, merged.ValueIs("y", "overridden"), "other wins on conflict")
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
		a := NewLabels().Set("k", "v")
		merged := a.Merge(NewLabels())
		assert.Equal(t, 1, merged.Len())
		assert.True(t, merged.ValueIs("k", "v"))
	})

	t.Run("merge into empty", func(t *testing.T) {
		b := NewLabels().Set("k", "v")
		merged := NewLabels().Merge(b)
		assert.True(t, merged.ValueIs("k", "v"))
	})
}

func TestLabelsCollection_ValueIs_edge_cases(t *testing.T) {
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
