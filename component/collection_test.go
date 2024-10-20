package component

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
			components: NewCollection().With(New("c1"), New("c2")),
			args: args{
				name: "c2",
			},
			want: New("c2"),
		},
		{
			name:       "component not found",
			components: NewCollection().With(New("c1"), New("c2")),
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

func TestCollection_With(t *testing.T) {
	type args struct {
		components []*Component
	}
	tests := []struct {
		name       string
		collection *Collection
		args       args
		assertions func(t *testing.T, collection *Collection)
	}{
		{
			name:       "adding nothing to empty collection",
			collection: NewCollection(),
			args: args{
				components: nil,
			},
			assertions: func(t *testing.T, collection *Collection) {
				assert.Zero(t, collection.Len())
			},
		},
		{
			name:       "adding to empty collection",
			collection: NewCollection(),
			args: args{
				components: []*Component{New("c1"), New("c2")},
			},
			assertions: func(t *testing.T, collection *Collection) {
				assert.Equal(t, 2, collection.Len())
				assert.NotNil(t, collection.ByName("c1"))
				assert.NotNil(t, collection.ByName("c2"))
				assert.Nil(t, collection.ByName("c999"))
			},
		},
		{
			name:       "adding to non-empty collection",
			collection: NewCollection().With(New("c1"), New("c2")),
			args: args{
				components: []*Component{New("c3"), New("c4")},
			},
			assertions: func(t *testing.T, collection *Collection) {
				assert.Equal(t, 4, collection.Len())
				assert.NotNil(t, collection.ByName("c1"))
				assert.NotNil(t, collection.ByName("c2"))
				assert.NotNil(t, collection.ByName("c3"))
				assert.NotNil(t, collection.ByName("c4"))
				assert.Nil(t, collection.ByName("c999"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collectionAfter := tt.collection.With(tt.args.components...)
			if tt.assertions != nil {
				tt.assertions(t, collectionAfter)
			}
		})
	}
}
