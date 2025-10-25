package labels

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLabelsCollection(t *testing.T) {
	tests := []struct {
		name       string
		labels     Map
		assertions func(t *testing.T, lc *Collection)
	}{
		{
			name:   "nil map creates empty collection",
			labels: nil,
			assertions: func(t *testing.T, lc *Collection) {
				assert.NotNil(t, lc)
				assert.Equal(t, 0, lc.Len())
				assert.False(t, lc.HasErr())
			},
		},
		{
			name:   "empty map",
			labels: Map{},
			assertions: func(t *testing.T, lc *Collection) {
				assert.Equal(t, 0, lc.Len())
			},
		},
		{
			name: "with labels",
			labels: Map{
				"label1": "value1",
				"label2": "value2",
			},
			assertions: func(t *testing.T, lc *Collection) {
				assert.Equal(t, 2, lc.Len())
				assert.True(t, lc.Has("label1"))
				assert.True(t, lc.Has("label2"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCollection(tt.labels)
			if tt.assertions != nil {
				tt.assertions(t, got)
			}
		})
	}
}

func TestLabelsCollection_With(t *testing.T) {
	tests := []struct {
		name       string
		collection *Collection
		label      string
		value      string
		assertions func(t *testing.T, result *Collection)
	}{
		{
			name:       "adding to empty collection",
			collection: NewCollection(nil),
			label:      "l1",
			value:      "v1",
			assertions: func(t *testing.T, result *Collection) {
				assert.Equal(t, 1, result.Len())
				value, err := result.Value("l1")
				require.NoError(t, err)
				assert.Equal(t, "v1", value)
			},
		},
		{
			name: "adding to non-empty collection",
			collection: NewCollection(Map{
				"l1": "v1",
			}),
			label: "l2",
			value: "v2",
			assertions: func(t *testing.T, result *Collection) {
				assert.Equal(t, 2, result.Len())
				assert.True(t, result.Has("l1"))
				assert.True(t, result.Has("l2"))
			},
		},
		{
			name: "overwriting existing label",
			collection: NewCollection(Map{
				"l1": "v1",
				"l2": "v2",
			}),
			label: "l2",
			value: "v3",
			assertions: func(t *testing.T, result *Collection) {
				assert.Equal(t, 2, result.Len())
				assert.True(t, result.ValueIs("l2", "v3"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.collection.With(tt.label, tt.value)
			if tt.assertions != nil {
				tt.assertions(t, result)
			}
		})
	}
}

func TestLabelsCollection_WithMany(t *testing.T) {
	tests := []struct {
		name       string
		collection *Collection
		labels     Map
		assertions func(t *testing.T, result *Collection)
	}{
		{
			name:       "adding to empty collection",
			collection: NewCollection(nil),
			labels: Map{
				"l1": "v1",
				"l2": "v2",
			},
			assertions: func(t *testing.T, result *Collection) {
				assert.Equal(t, 2, result.Len())
			},
		},
		{
			name: "adding to non-empty collection with overwrite",
			collection: NewCollection(Map{
				"l1": "v1",
				"l2": "v2",
				"l3": "v3",
			}),
			labels: Map{
				"l3": "v100",
				"l4": "v4",
				"l5": "v5",
			},
			assertions: func(t *testing.T, result *Collection) {
				assert.Equal(t, 5, result.Len())
				assert.True(t, result.ValueIs("l3", "v100"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.collection.WithMany(tt.labels)
			if tt.assertions != nil {
				tt.assertions(t, result)
			}
		})
	}
}

func TestLabelsCollection_Without(t *testing.T) {
	tests := []struct {
		name       string
		collection *Collection
		labels     []string
		assertions func(t *testing.T, result *Collection)
	}{
		{
			name: "label found and deleted",
			collection: NewCollection(Map{
				"l1": "v1",
				"l2": "v2",
			}),
			labels: []string{"l1"},
			assertions: func(t *testing.T, result *Collection) {
				assert.Equal(t, 1, result.Len())
				assert.False(t, result.Has("l1"))
				assert.True(t, result.Has("l2"))
			},
		},
		{
			name: "label not found is no-op",
			collection: NewCollection(Map{
				"l1": "v1",
				"l2": "v2",
			}),
			labels: []string{"l3"},
			assertions: func(t *testing.T, result *Collection) {
				assert.Equal(t, 2, result.Len())
			},
		},
		{
			name: "delete multiple labels",
			collection: NewCollection(Map{
				"l1": "v1",
				"l2": "v2",
				"l3": "v3",
			}),
			labels: []string{"l1", "l3"},
			assertions: func(t *testing.T, result *Collection) {
				assert.Equal(t, 1, result.Len())
				assert.False(t, result.Has("l1"))
				assert.True(t, result.Has("l2"))
				assert.False(t, result.Has("l3"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.collection.Without(tt.labels...)
			if tt.assertions != nil {
				tt.assertions(t, result)
			}
		})
	}
}

func TestLabelsCollection_Has(t *testing.T) {
	tests := []struct {
		name       string
		collection *Collection
		label      string
		want       bool
	}{
		{
			name:       "empty collection",
			collection: NewCollection(nil),
			label:      "l1",
			want:       false,
		},
		{
			name: "label exists",
			collection: NewCollection(Map{
				"l1": "v1",
				"l2": "v2",
			}),
			label: "l1",
			want:  true,
		},
		{
			name: "label does not exist",
			collection: NewCollection(Map{
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
		collection *Collection
		labels     []string
		want       bool
	}{
		{
			name:       "empty collection",
			collection: NewCollection(nil),
			labels:     []string{"l1"},
			want:       false,
		},
		{
			name: "has all labels",
			collection: NewCollection(Map{
				"l1": "v1",
				"l2": "v2",
				"l3": "v3",
			}),
			labels: []string{"l1", "l2"},
			want:   true,
		},
		{
			name: "missing some labels",
			collection: NewCollection(Map{
				"l1": "v1",
				"l2": "v2",
				"l3": "v3",
			}),
			labels: []string{"l1", "l2", "l4"},
			want:   false,
		},
		{
			name: "empty labels list returns true",
			collection: NewCollection(Map{
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
		collection *Collection
		labels     []string
		want       bool
	}{
		{
			name:       "empty collection",
			collection: NewCollection(nil),
			labels:     []string{"l1"},
			want:       false,
		},
		{
			name: "has at least one label",
			collection: NewCollection(Map{
				"l1": "v1",
				"l2": "v2",
				"l3": "v3",
			}),
			labels: []string{"l1", "l10"},
			want:   true,
		},
		{
			name: "has none of the labels",
			collection: NewCollection(Map{
				"l1": "v1",
				"l2": "v2",
				"l3": "v3",
			}),
			labels: []string{"l10", "l20", "l4"},
			want:   false,
		},
		{
			name: "empty labels list returns false",
			collection: NewCollection(Map{
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
		collection *Collection
		label      string
		want       string
		wantErr    bool
	}{
		{
			name:       "label not found in empty collection",
			collection: NewCollection(nil),
			label:      "l1",
			want:       "",
			wantErr:    true,
		},
		{
			name: "label found",
			collection: NewCollection(Map{
				"l1": "v1",
				"l2": "v2",
			}),
			label:   "l2",
			want:    "v2",
			wantErr: false,
		},
		{
			name: "label not found in non-empty collection",
			collection: NewCollection(Map{
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
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLabelsCollection_ValueOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		collection   *Collection
		label        string
		defaultValue string
		want         string
	}{
		{
			name:         "label not found returns default",
			collection:   NewCollection(nil),
			label:        "l1",
			defaultValue: "default",
			want:         "default",
		},
		{
			name: "label found returns actual value",
			collection: NewCollection(Map{
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
		collection *Collection
		label      string
		value      string
		want       bool
	}{
		{
			name:       "empty collection",
			collection: NewCollection(nil),
			label:      "l1",
			value:      "v1",
			want:       false,
		},
		{
			name: "label exists with matching value",
			collection: NewCollection(Map{
				"l1": "v1",
				"l2": "v2",
			}),
			label: "l1",
			value: "v1",
			want:  true,
		},
		{
			name: "label exists with different value",
			collection: NewCollection(Map{
				"l1": "v1",
				"l2": "v2",
			}),
			label: "l1",
			value: "v999",
			want:  false,
		},
		{
			name: "label does not exist",
			collection: NewCollection(Map{
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

func TestLabelsCollection_MatchesAll(t *testing.T) {
	tests := []struct {
		name       string
		collection *Collection
		labels     Map
		want       bool
	}{
		{
			name:       "empty collection vs empty map",
			collection: NewCollection(nil),
			labels:     Map{},
			want:       true,
		},
		{
			name:       "empty collection vs non-empty map",
			collection: NewCollection(nil),
			labels:     Map{"l1": "v1"},
			want:       false,
		},
		{
			name: "collection matches all key-value pairs",
			collection: NewCollection(Map{
				"l1": "v1",
				"l2": "v2",
				"l3": "v3",
			}),
			labels: Map{
				"l1": "v1",
				"l2": "v2",
			},
			want: true,
		},
		{
			name: "collection has key but wrong value",
			collection: NewCollection(Map{
				"l1": "v1",
				"l2": "v2",
				"l3": "v3",
			}),
			labels: Map{
				"l1": "v1",
				"l2": "v999",
			},
			want: false,
		},
		{
			name: "collection missing required key",
			collection: NewCollection(Map{
				"l1": "v1",
			}),
			labels: Map{
				"l1": "v1",
				"l2": "v2",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.collection.MatchesAll(tt.labels))
		})
	}
}

func TestLabelsCollection_MatchesAny(t *testing.T) {
	tests := []struct {
		name       string
		collection *Collection
		labels     Map
		want       bool
	}{
		{
			name:       "empty collection",
			collection: NewCollection(nil),
			labels:     Map{"l1": "v1"},
			want:       false,
		},
		{
			name: "matches at least one pair",
			collection: NewCollection(Map{
				"l1": "v1",
				"l2": "v2",
				"l3": "v3",
			}),
			labels: Map{
				"l1": "v1",
				"l2": "v999",
			},
			want: true,
		},
		{
			name: "matches none",
			collection: NewCollection(Map{
				"l1": "v1",
				"l2": "v2",
			}),
			labels: Map{
				"l1": "v999",
				"l2": "v888",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.collection.MatchesAny(tt.labels))
		})
	}
}

func TestLabelsCollection_All(t *testing.T) {
	tests := []struct {
		name       string
		collection *Collection
		predicate  LabelPredicate
		want       bool
	}{
		{
			name:       "empty collection returns true (vacuous truth)",
			collection: NewCollection(nil),
			predicate:  func(k, v string) bool { return false },
			want:       true,
		},
		{
			name: "all labels satisfy predicate",
			collection: NewCollection(Map{
				"l1": "v1",
				"l2": "v2",
				"l3": "v3",
			}),
			predicate: func(k, v string) bool { return strings.HasPrefix(k, "l") },
			want:      true,
		},
		{
			name: "not all labels satisfy predicate",
			collection: NewCollection(Map{
				"l1":  "v1",
				"l2":  "v2",
				"xyz": "v3",
			}),
			predicate: func(k, v string) bool { return strings.HasPrefix(k, "l") },
			want:      false,
		},
		{
			name: "all values non-empty",
			collection: NewCollection(Map{
				"l1": "v1",
				"l2": "v2",
			}),
			predicate: func(k, v string) bool { return v != "" },
			want:      true,
		},
		{
			name: "some values empty",
			collection: NewCollection(Map{
				"l1": "v1",
				"l2": "",
			}),
			predicate: func(k, v string) bool { return v != "" },
			want:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.collection.All(tt.predicate))
		})
	}
}

func TestLabelsCollection_Any(t *testing.T) {
	tests := []struct {
		name       string
		collection *Collection
		predicate  LabelPredicate
		want       bool
	}{
		{
			name:       "empty collection returns false",
			collection: NewCollection(nil),
			predicate:  func(k, v string) bool { return true },
			want:       false,
		},
		{
			name: "at least one label satisfies predicate",
			collection: NewCollection(Map{
				"l1":  "v1",
				"l2":  "v2",
				"xyz": "v3",
			}),
			predicate: func(k, v string) bool { return strings.HasPrefix(k, "x") },
			want:      true,
		},
		{
			name: "no labels satisfy predicate",
			collection: NewCollection(Map{
				"l1": "v1",
				"l2": "v2",
				"l3": "v3",
			}),
			predicate: func(k, v string) bool { return strings.HasPrefix(k, "x") },
			want:      false,
		},
		{
			name: "at least one value matches condition",
			collection: NewCollection(Map{
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
		collection *Collection
		want       int
	}{
		{
			name:       "empty collection",
			collection: NewCollection(nil),
			want:       0,
		},
		{
			name: "collection with labels",
			collection: NewCollection(Map{
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

func TestLabelsCollection_Chainable(t *testing.T) {
	t.Run("chaining multiple operations", func(t *testing.T) {
		lc := NewCollection(nil).
			With("env", "dev").
			With("tier", "backend").
			WithMany(Map{
				"region": "us-east",
				"zone":   "1a",
			}).
			Without("zone")

		assert.Equal(t, 3, lc.Len())
		assert.True(t, lc.Has("env"))
		assert.True(t, lc.Has("tier"))
		assert.True(t, lc.Has("region"))
		assert.False(t, lc.Has("zone"))
	})
}

func TestLabelsCollection_ErrorHandling(t *testing.T) {
	tests := []struct {
		name       string
		collection *Collection
		assertions func(t *testing.T, lc *Collection)
	}{
		{
			name:       "mutating methods return self without changes when error present",
			collection: NewCollection(nil).WithErr(errors.New("test error")),
			assertions: func(t *testing.T, lc *Collection) {
				result := lc.With("test", "value")
				assert.True(t, result.HasErr())
				assert.Equal(t, 0, result.Len())
			},
		},
		{
			name:       "query methods return safe defaults when error present",
			collection: NewCollection(nil).WithErr(errors.New("test error")),
			assertions: func(t *testing.T, lc *Collection) {
				assert.False(t, lc.Has("test"))
				assert.Equal(t, 0, lc.Len())
				assert.False(t, lc.All(func(k, v string) bool { return true }))
				assert.False(t, lc.Any(func(k, v string) bool { return true }))

				val, err := lc.Value("test")
				assert.Error(t, err)
				assert.Equal(t, "", val)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.assertions != nil {
				tt.assertions(t, tt.collection)
			}
		})
	}
}
