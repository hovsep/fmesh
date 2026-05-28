package component

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newCol is a test helper to build a collection from names, panicking on error.
func newCol(names ...string) *Collection {
	col := NewCollection()
	for _, name := range names {
		c, err := New(name)
		if err != nil {
			panic(err)
		}
		if err := col.Add(c); err != nil {
			panic(err)
		}
	}
	return col
}

func TestCollection_ByName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name       string
		components *Collection
		args       args
		wantName   string
		wantNil    bool
	}{
		{
			name:       "component found",
			components: newCol("c1", "c2"),
			args:       args{name: "c2"},
			wantName:   "c2",
		},
		{
			name:       "component not found returns nil",
			components: newCol("c1", "c2"),
			args:       args{name: "c3"},
			wantNil:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.components.ByName(tt.args.name)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.wantName, result.Name())
			}
		})
	}
}

func TestCollection_Add(t *testing.T) {
	tests := []struct {
		name       string
		collection *Collection
		toAdd      []string
		assertions func(t *testing.T, collection *Collection, addErr error)
	}{
		{
			name:       "adding to empty collection",
			collection: NewCollection(),
			toAdd:      []string{"c1", "c2"},
			assertions: func(t *testing.T, collection *Collection, addErr error) {
				require.NoError(t, addErr)
				assert.Equal(t, 2, collection.Len())
				assert.Equal(t, "c1", collection.ByName("c1").Name())
				assert.Equal(t, "c2", collection.ByName("c2").Name())
			},
		},
		{
			name:       "adding to non-empty collection",
			collection: newCol("existing"),
			toAdd:      []string{"c1", "c2"},
			assertions: func(t *testing.T, collection *Collection, addErr error) {
				require.NoError(t, addErr)
				assert.Equal(t, 3, collection.Len())
				assert.Equal(t, "existing", collection.ByName("existing").Name())
				assert.Equal(t, "c1", collection.ByName("c1").Name())
				assert.Equal(t, "c2", collection.ByName("c2").Name())
			},
		},
		{
			name:       "adding 2 components with the same name",
			collection: newCol("existing"),
			toAdd:      []string{"existing"},
			assertions: func(t *testing.T, collection *Collection, addErr error) {
				require.Error(t, addErr)
				require.ErrorContains(t, addErr, `component with name "existing" already exists`)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var addErr error
			for _, name := range tt.toAdd {
				c, err := New(name)
				require.NoError(t, err)
				if err := tt.collection.Add(c); err != nil {
					addErr = err
					break
				}
			}
			if tt.assertions != nil {
				tt.assertions(t, tt.collection, addErr)
			}
		})
	}
}

func TestCollection_Len(t *testing.T) {
	tests := []struct {
		name       string
		collection *Collection
		want       int
	}{
		{
			name:       "empty collection",
			collection: NewCollection(),
			want:       0,
		},
		{
			name:       "non-empty collection",
			collection: newCol("c1", "c2", "c3"),
			want:       3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.collection.Len())
		})
	}
}

func TestCollection_IsEmpty(t *testing.T) {
	tests := []struct {
		name       string
		collection *Collection
		want       bool
	}{
		{
			name:       "empty collection",
			collection: NewCollection(),
			want:       true,
		},
		{
			name:       "non-empty collection",
			collection: newCol("c1"),
			want:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.collection.IsEmpty())
		})
	}
}

func TestCollection_Every(t *testing.T) {
	tests := []struct {
		name       string
		collection *Collection
		predicate  Predicate
		want       bool
	}{
		{
			name:       "empty collection returns true",
			collection: NewCollection(),
			predicate:  func(c *Component) bool { return false },
			want:       true,
		},
		{
			name:       "all match",
			collection: newCol("c1", "c2"),
			predicate:  func(c *Component) bool { return c.Name() != "" },
			want:       true,
		},
		{
			name:       "not all match",
			collection: newCol("c1", "c2_noname_placeholder"),
			predicate:  func(c *Component) bool { return c.Name() == "c1" },
			want:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.collection.Every(tt.predicate))
		})
	}
}

func TestCollection_AnyMatch(t *testing.T) {
	tests := []struct {
		name       string
		collection *Collection
		predicate  Predicate
		want       bool
	}{
		{
			name:       "empty collection returns false",
			collection: NewCollection(),
			predicate:  func(c *Component) bool { return true },
			want:       false,
		},
		{
			name:       "at least one matches",
			collection: newCol("c1", "c2"),
			predicate:  func(c *Component) bool { return c.Name() == "c1" },
			want:       true,
		},
		{
			name:       "none match",
			collection: newCol("b1", "b2"),
			predicate:  func(c *Component) bool { return c.Name() == "" },
			want:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.collection.AnyMatch(tt.predicate))
		})
	}
}

func TestCollection_Filter(t *testing.T) {
	tests := []struct {
		name       string
		collection *Collection
		predicate  Predicate
		want       int // expected length after filtering
	}{
		{
			name:       "empty collection",
			collection: NewCollection(),
			predicate:  func(c *Component) bool { return true },
			want:       0,
		},
		{
			name:       "filter some components",
			collection: newCol("c1", "c2", "c3"),
			predicate:  func(c *Component) bool { return c.Name() != "c2" },
			want:       2,
		},
		{
			name:       "filter all components",
			collection: newCol("c1", "c2"),
			predicate:  func(c *Component) bool { return false },
			want:       0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.collection.Filter(tt.predicate)
			assert.Equal(t, tt.want, result.Len())
		})
	}
}

func TestCollection_Any(t *testing.T) {
	t.Run("returns component from non-empty collection", func(t *testing.T) {
		collection := newCol("c1")
		result := collection.Any()
		require.NotNil(t, result)
		assert.Equal(t, "c1", result.Name())
	})

	t.Run("returns nil from empty collection", func(t *testing.T) {
		collection := NewCollection()
		result := collection.Any()
		assert.Nil(t, result)
	})
}

func TestCollection_FindAny(t *testing.T) {
	t.Run("finds matching component", func(t *testing.T) {
		collection := newCol("c1", "c2", "target")
		result := collection.FindAny(func(c *Component) bool {
			return c.Name() == "target"
		})
		require.NotNil(t, result)
		assert.Equal(t, "target", result.Name())
	})

	t.Run("returns nil when no match", func(t *testing.T) {
		collection := newCol("c1", "c2")
		result := collection.FindAny(func(c *Component) bool {
			return c.Name() == "nonexistent"
		})
		assert.Nil(t, result)
	})
}

func TestCollection_Count(t *testing.T) {
	t.Run("counts matching components", func(t *testing.T) {
		collection := newCol("a1", "a2", "b1")
		count := collection.Count(func(c *Component) bool {
			return c.Name()[0] == 'a'
		})
		assert.Equal(t, 2, count)
	})

	t.Run("returns 0 for empty collection", func(t *testing.T) {
		collection := NewCollection()
		count := collection.Count(func(c *Component) bool {
			return true
		})
		assert.Equal(t, 0, count)
	})
}

func TestCollection_Map(t *testing.T) {
	t.Run("transforms components", func(t *testing.T) {
		collection := newCol("c1", "c2")
		mapped, err := collection.Map(func(c *Component) *Component {
			nc, nerr := New("mapped_" + c.Name())
			if nerr != nil {
				panic(nerr)
			}
			return nc
		})
		require.NoError(t, err)
		assert.Equal(t, 2, mapped.Len())
		assert.NotNil(t, mapped.ByName("mapped_c1"))
		assert.NotNil(t, mapped.ByName("mapped_c2"))
	})

	t.Run("filters out nil results", func(t *testing.T) {
		collection := newCol("c1", "c2", "c3")
		mapped, err := collection.Map(func(c *Component) *Component {
			if c.Name() == "c2" {
				return nil
			}
			return c
		})
		require.NoError(t, err)
		assert.Equal(t, 2, mapped.Len())
	})
}

func TestCollection_ForEach(t *testing.T) {
	t.Run("applies action to each component", func(t *testing.T) {
		collection := newCol("c1", "c2")
		visited := make([]string, 0)
		err := collection.ForEach(func(c *Component) error {
			visited = append(visited, c.Name())
			return nil
		})
		require.NoError(t, err)
		assert.Len(t, visited, 2)
	})

	t.Run("stops on error and returns error", func(t *testing.T) {
		collection := newCol("c1", "c2", "c3")
		err := collection.ForEach(func(c *Component) error {
			return assert.AnError
		})
		assert.Error(t, err)
	})
}

func TestCollection_Clear(t *testing.T) {
	t.Run("removes all components", func(t *testing.T) {
		collection := newCol("c1", "c2")
		result := collection.Clear()
		assert.Equal(t, 0, result.Len())
		assert.True(t, result.IsEmpty())
	})
}

func TestCollection_Without(t *testing.T) {
	t.Run("removes specified components", func(t *testing.T) {
		collection := newCol("c1", "c2", "c3")
		result := collection.Without("c1", "c3")
		assert.Equal(t, 1, result.Len())
		assert.NotNil(t, result.ByName("c2"))
		assert.Nil(t, result.ByName("c1"))
	})

	t.Run("handles non-existent names gracefully", func(t *testing.T) {
		collection := newCol("c1")
		result := collection.Without("nonexistent")
		assert.Equal(t, 1, result.Len())
	})
}

func TestCollection_All(t *testing.T) {
	t.Run("returns all components", func(t *testing.T) {
		collection := newCol("c1", "c2")
		all := collection.All()
		assert.Len(t, all, 2)
	})

	t.Run("returns empty map for empty collection", func(t *testing.T) {
		collection := NewCollection()
		all := collection.All()
		assert.Empty(t, all)
	})
}

func TestCollection_LeafMethodsDoNotPoisonCollection(t *testing.T) {
	t.Run("ByName does not poison collection on not found", func(t *testing.T) {
		collection := newCol("c1", "c2")

		// Query for non-existent component
		result := collection.ByName("nonexistent")

		// Result should be nil
		assert.Nil(t, result)

		// Collection should still have 2 components
		assert.Equal(t, 2, collection.Len())

		// Collection should still be usable
		c1 := collection.ByName("c1")
		require.NotNil(t, c1)
		assert.Equal(t, "c1", c1.Name())
	})

	t.Run("Any does not affect collection when empty", func(t *testing.T) {
		collection := NewCollection()

		// Query any on empty collection
		result := collection.Any()
		assert.Nil(t, result)

		// Collection should still be usable for adding
		c, err := New("c1")
		require.NoError(t, err)
		require.NoError(t, collection.Add(c))
		assert.Equal(t, 1, collection.Len())
	})

	t.Run("FindAny does not affect collection when no match", func(t *testing.T) {
		collection := newCol("c1", "c2")

		// Query with predicate that matches nothing
		result := collection.FindAny(func(c *Component) bool {
			return c.Name() == "nonexistent"
		})
		assert.Nil(t, result)

		// Collection should have 2 components
		assert.Equal(t, 2, collection.Len())

		// Subsequent FindAny should work
		found := collection.FindAny(func(c *Component) bool {
			return c.Name() == "c1"
		})
		require.NotNil(t, found)
		assert.Equal(t, "c1", found.Name())
	})
}
