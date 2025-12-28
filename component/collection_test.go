package component

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollection_ByName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name       string
		components *Collection
		args       args
		want       *Component
	}{
		{
			name:       "component found",
			components: NewCollection().Add(New("c1"), New("c2")),
			args: args{
				name: "c2",
			},
			want: New("c2"),
		},
		{
			name:       "component not found returns nil",
			components: NewCollection().Add(New("c1"), New("c2")),
			args: args{
				name: "c3",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.components.ByName(tt.args.name))
		})
	}
}

func TestCollection_Add(t *testing.T) {
	tests := []struct {
		name       string
		collection *Collection
		components []*Component
		assertions func(t *testing.T, collection *Collection)
	}{
		{
			name:       "adding to empty collection",
			collection: NewCollection(),
			components: []*Component{New("c1"), New("c2")},
			assertions: func(t *testing.T, collection *Collection) {
				assert.Equal(t, 2, collection.Len())
				assert.Equal(t, "c1", collection.ByName("c1").Name())
				assert.Equal(t, "c2", collection.ByName("c2").Name())
			},
		},
		{
			name:       "adding to non-empty collection",
			collection: NewCollection().Add(New("existing")),
			components: []*Component{New("c1"), New("c2")},
			assertions: func(t *testing.T, collection *Collection) {
				assert.Equal(t, 3, collection.Len())
				assert.Equal(t, "existing", collection.ByName("existing").Name())
				assert.Equal(t, "c1", collection.ByName("c1").Name())
				assert.Equal(t, "c2", collection.ByName("c2").Name())
			},
		},
		{
			name:       "adding 2 components with the same name",
			collection: NewCollection().Add(New("existing")),
			components: []*Component{New("existing")},
			assertions: func(t *testing.T, collection *Collection) {
				require.Error(t, collection.ChainableErr())
				require.ErrorContains(t, collection.ChainableErr(), "component with name 'existing' already exists")
				assert.True(t, collection.HasChainableErr())
				assert.Equal(t, 0, collection.Len())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.collection.Add(tt.components...)
			if tt.assertions != nil {
				tt.assertions(t, result)
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
			collection: NewCollection().Add(New("c1"), New("c2"), New("c3")),
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
			collection: NewCollection().Add(New("c1")),
			want:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.collection.IsEmpty())
		})
	}
}

func TestCollection_AllMatch(t *testing.T) {
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
			collection: NewCollection().Add(New("c1"), New("c2")),
			predicate:  func(c *Component) bool { return c.Name() != "" },
			want:       true,
		},
		{
			name:       "not all match",
			collection: NewCollection().Add(New("c1"), New("")),
			predicate:  func(c *Component) bool { return c.Name() != "" },
			want:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.collection.AllMatch(tt.predicate))
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
			collection: NewCollection().Add(New("c1"), New("")),
			predicate:  func(c *Component) bool { return c.Name() != "" },
			want:       true,
		},
		{
			name:       "none match",
			collection: NewCollection().Add(New(""), New("")),
			predicate:  func(c *Component) bool { return c.Name() != "" },
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
			collection: NewCollection().Add(New("c1"), New("c2"), New("c3")),
			predicate:  func(c *Component) bool { return c.Name() != "c2" },
			want:       2,
		},
		{
			name:       "filter all components",
			collection: NewCollection().Add(New("c1"), New("c2")),
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
		collection := NewCollection().Add(New("c1"))
		result := collection.Any()
		require.NotNil(t, result)
		assert.Equal(t, "c1", result.Name())
	})

	t.Run("returns nil from empty collection", func(t *testing.T) {
		collection := NewCollection()
		result := collection.Any()
		assert.Nil(t, result)
	})

	t.Run("returns nil when collection has error", func(t *testing.T) {
		collection := NewCollection().WithChainableErr(assert.AnError)
		result := collection.Any()
		assert.Nil(t, result)
	})
}

func TestCollection_FindAny(t *testing.T) {
	t.Run("finds matching component", func(t *testing.T) {
		collection := NewCollection().Add(New("c1"), New("c2"), New("target"))
		result := collection.FindAny(func(c *Component) bool {
			return c.Name() == "target"
		})
		require.NotNil(t, result)
		assert.Equal(t, "target", result.Name())
	})

	t.Run("returns nil when no match", func(t *testing.T) {
		collection := NewCollection().Add(New("c1"), New("c2"))
		result := collection.FindAny(func(c *Component) bool {
			return c.Name() == "nonexistent"
		})
		assert.Nil(t, result)
	})

	t.Run("returns nil when collection has error", func(t *testing.T) {
		collection := NewCollection().Add(New("c1")).WithChainableErr(assert.AnError)
		result := collection.FindAny(func(c *Component) bool {
			return true
		})
		assert.Nil(t, result)
	})
}

func TestCollection_CountMatch(t *testing.T) {
	t.Run("counts matching components", func(t *testing.T) {
		collection := NewCollection().Add(New("a1"), New("a2"), New("b1"))
		count := collection.CountMatch(func(c *Component) bool {
			return c.Name()[0] == 'a'
		})
		assert.Equal(t, 2, count)
	})

	t.Run("returns 0 for empty collection", func(t *testing.T) {
		collection := NewCollection()
		count := collection.CountMatch(func(c *Component) bool {
			return true
		})
		assert.Equal(t, 0, count)
	})

	t.Run("returns 0 when collection has error", func(t *testing.T) {
		collection := NewCollection().Add(New("c1")).WithChainableErr(assert.AnError)
		count := collection.CountMatch(func(c *Component) bool {
			return true
		})
		assert.Equal(t, 0, count)
	})
}

func TestCollection_Map(t *testing.T) {
	t.Run("transforms components", func(t *testing.T) {
		collection := NewCollection().Add(New("c1"), New("c2"))
		mapped := collection.Map(func(c *Component) *Component {
			return New("mapped_" + c.Name())
		})
		assert.Equal(t, 2, mapped.Len())
		assert.NotNil(t, mapped.ByName("mapped_c1"))
		assert.NotNil(t, mapped.ByName("mapped_c2"))
	})

	t.Run("filters out nil results", func(t *testing.T) {
		collection := NewCollection().Add(New("c1"), New("c2"), New("c3"))
		mapped := collection.Map(func(c *Component) *Component {
			if c.Name() == "c2" {
				return nil
			}
			return c
		})
		assert.Equal(t, 2, mapped.Len())
	})

	t.Run("propagates error from source collection", func(t *testing.T) {
		collection := NewCollection().WithChainableErr(assert.AnError)
		mapped := collection.Map(func(c *Component) *Component {
			return c
		})
		assert.True(t, mapped.HasChainableErr())
	})
}

func TestCollection_ForEach(t *testing.T) {
	t.Run("applies action to each component", func(t *testing.T) {
		collection := NewCollection().Add(New("c1"), New("c2"))
		visited := make([]string, 0)
		collection.ForEach(func(c *Component) error {
			visited = append(visited, c.Name())
			return nil
		})
		assert.Len(t, visited, 2)
	})

	t.Run("stops on error and sets chainable error", func(t *testing.T) {
		collection := NewCollection().Add(New("c1"), New("c2"), New("c3"))
		result := collection.ForEach(func(c *Component) error {
			return assert.AnError
		})
		assert.True(t, result.HasChainableErr())
	})

	t.Run("skips when collection has error", func(t *testing.T) {
		collection := NewCollection().WithChainableErr(assert.AnError)
		visited := false
		collection.ForEach(func(c *Component) error {
			visited = true
			return nil
		})
		assert.False(t, visited)
	})
}

func TestCollection_Clear(t *testing.T) {
	t.Run("removes all components", func(t *testing.T) {
		collection := NewCollection().Add(New("c1"), New("c2"))
		result := collection.Clear()
		assert.Equal(t, 0, result.Len())
		assert.True(t, result.IsEmpty())
	})

	t.Run("skips when collection has error", func(t *testing.T) {
		collection := NewCollection().Add(New("c1")).WithChainableErr(assert.AnError)
		result := collection.Clear()
		assert.True(t, result.HasChainableErr())
	})
}

func TestCollection_Without(t *testing.T) {
	t.Run("removes specified components", func(t *testing.T) {
		collection := NewCollection().Add(New("c1"), New("c2"), New("c3"))
		result := collection.Without("c1", "c3")
		assert.Equal(t, 1, result.Len())
		assert.NotNil(t, result.ByName("c2"))
		assert.Nil(t, result.ByName("c1"))
	})

	t.Run("handles non-existent names gracefully", func(t *testing.T) {
		collection := NewCollection().Add(New("c1"))
		result := collection.Without("nonexistent")
		assert.Equal(t, 1, result.Len())
	})

	t.Run("skips when collection has error", func(t *testing.T) {
		collection := NewCollection().Add(New("c1")).WithChainableErr(assert.AnError)
		result := collection.Without("c1")
		assert.True(t, result.HasChainableErr())
	})
}

func TestCollection_All(t *testing.T) {
	t.Run("returns all components", func(t *testing.T) {
		collection := NewCollection().Add(New("c1"), New("c2"))
		all, err := collection.All()
		require.NoError(t, err)
		assert.Len(t, all, 2)
	})

	t.Run("returns error when collection has error", func(t *testing.T) {
		collection := NewCollection().WithChainableErr(assert.AnError)
		_, err := collection.All()
		require.Error(t, err)
	})
}

func TestCollection_LeafMethodsDoNotPoisonCollection(t *testing.T) {
	t.Run("ByName does not poison collection on not found", func(t *testing.T) {
		collection := NewCollection().Add(New("c1"), New("c2"))

		// Query for non-existent component
		result := collection.ByName("nonexistent")

		// Result should be nil
		assert.Nil(t, result)

		// Collection should NOT be poisoned
		assert.False(t, collection.HasChainableErr())
		assert.Equal(t, 2, collection.Len())

		// Collection should still be usable
		c1 := collection.ByName("c1")
		require.NotNil(t, c1)
		assert.Equal(t, "c1", c1.Name())
	})

	t.Run("Any does not poison collection when empty", func(t *testing.T) {
		collection := NewCollection()

		// Query any on empty collection
		result := collection.Any()

		// Result should be nil
		assert.Nil(t, result)

		// Collection should NOT be poisoned
		assert.False(t, collection.HasChainableErr())

		// Collection should still be usable for adding
		collection.Add(New("c1"))
		assert.Equal(t, 1, collection.Len())
	})

	t.Run("FindAny does not poison collection when no match", func(t *testing.T) {
		collection := NewCollection().Add(New("c1"), New("c2"))

		// Query with predicate that matches nothing
		result := collection.FindAny(func(c *Component) bool {
			return c.Name() == "nonexistent"
		})

		// Result should be nil
		assert.Nil(t, result)

		// Collection should NOT be poisoned
		assert.False(t, collection.HasChainableErr())
		assert.Equal(t, 2, collection.Len())

		// Subsequent FindAny should work
		found := collection.FindAny(func(c *Component) bool {
			return c.Name() == "c1"
		})
		require.NotNil(t, found)
		assert.Equal(t, "c1", found.Name())
	})
}
