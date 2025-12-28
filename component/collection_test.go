package component

import (
	"fmt"
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
			name:       "component not found",
			components: NewCollection().Add(New("c1"), New("c2")),
			args: args{
				name: "c3",
			},
			want: New("n/a").WithChainableErr(fmt.Errorf("%w, component name: %s", errNotFound, "c3")),
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

func TestCollection_LeafMethodsDoNotPoisonCollection(t *testing.T) {
	t.Run("ByName does not poison collection on not found", func(t *testing.T) {
		collection := NewCollection().Add(New("c1"), New("c2"))

		// Query for non-existent component
		result := collection.ByName("nonexistent")

		// Result should have error
		assert.True(t, result.HasChainableErr())
		require.ErrorContains(t, result.ChainableErr(), "not found")

		// But collection should NOT be poisoned
		assert.False(t, collection.HasChainableErr())
		assert.Equal(t, 2, collection.Len())

		// Collection should still be usable
		c1 := collection.ByName("c1")
		assert.False(t, c1.HasChainableErr())
		assert.Equal(t, "c1", c1.Name())
	})

	t.Run("Any does not poison collection when empty", func(t *testing.T) {
		collection := NewCollection()

		// Query any on empty collection
		result := collection.Any()

		// Result should have error
		assert.True(t, result.HasChainableErr())
		require.ErrorIs(t, result.ChainableErr(), ErrNoComponentsInCollection)

		// But collection should NOT be poisoned
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

		// Result should have error
		assert.True(t, result.HasChainableErr())
		require.ErrorIs(t, result.ChainableErr(), ErrNoComponentMatchesPredicate)

		// But collection should NOT be poisoned
		assert.False(t, collection.HasChainableErr())
		assert.Equal(t, 2, collection.Len())

		// Subsequent FindAny should work
		found := collection.FindAny(func(c *Component) bool {
			return c.Name() == "c1"
		})
		assert.False(t, found.HasChainableErr())
		assert.Equal(t, "c1", found.Name())
	})
}
