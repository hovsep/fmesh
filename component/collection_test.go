package component

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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
			want: New("").WithChainableErr(fmt.Errorf("%w, component name: %s", errNotFound, "c3")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.components.ByName(tt.args.name))
		})
	}
}

func TestCollection_With(t *testing.T) {
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
