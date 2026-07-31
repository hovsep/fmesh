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
