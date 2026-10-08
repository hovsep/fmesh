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
				require.ErrorContains(t, addErr, `component "existing" already exists`)
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

func TestCollection_Remove(t *testing.T) {
	t.Run("removes specified components", func(t *testing.T) {
		collection := newCol("c1", "c2", "c3")
		result := collection.Remove("c1", "c3")
		assert.Equal(t, 1, result.Len())
		assert.NotNil(t, result.ByName("c2"))
		assert.Nil(t, result.ByName("c1"))
	})

	t.Run("handles non-existent names gracefully", func(t *testing.T) {
		collection := newCol("c1")
		result := collection.Remove("nonexistent")
		assert.Equal(t, 1, result.Len())
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
}

// TestCollection_PromotedReadSurface smoke-checks the read methods promoted
// from internal/collection.Keyed; that package's suite is their source of truth.
func TestCollection_PromotedReadSurface(t *testing.T) {
	t.Run("promoted methods work through the Collection facade", func(t *testing.T) {
		collection := newCol("c2", "c1")
		assert.Equal(t, 2, collection.Len())
		assert.False(t, collection.IsEmpty())
		assert.Equal(t, "c1", collection.AllOrdered()[0].Name(), "traversal is name-ordered")
		assert.True(t, collection.Any(func(c *Component) bool { return c.Name() == "c2" }))
	})
}
