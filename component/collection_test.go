package component

import (
	"github.com/hovsep/fmesh/port"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCollection_ByName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name       string
		components Collection
		args       args
		want       *Component
	}{
		{
			name:       "component found",
			components: NewComponentCollection().Add(New("c1"), New("c2")),
			args: args{
				name: "c2",
			},
			want: &Component{
				name:        "c2",
				description: "",
				inputs:      port.Collection{},
				outputs:     port.Collection{},
				f:           nil,
			},
		},
		{
			name:       "component not found",
			components: NewComponentCollection().Add(New("c1"), New("c2")),
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
	type args struct {
		components []*Component
	}
	tests := []struct {
		name       string
		collection Collection
		args       args
		assertions func(t *testing.T, collection Collection)
	}{
		{
			name:       "adding nothing to empty collection",
			collection: NewComponentCollection(),
			args: args{
				components: nil,
			},
			assertions: func(t *testing.T, collection Collection) {
				assert.Len(t, collection, 0)
			},
		},
		{
			name:       "adding to empty collection",
			collection: NewComponentCollection(),
			args: args{
				components: []*Component{New("c1"), New("c2")},
			},
			assertions: func(t *testing.T, collection Collection) {
				assert.Len(t, collection, 2)
				assert.NotNil(t, collection.ByName("c1"))
				assert.NotNil(t, collection.ByName("c2"))
				assert.Nil(t, collection.ByName("c999"))
			},
		},
		{
			name:       "adding to existing collection",
			collection: NewComponentCollection().Add(New("c1"), New("c2")),
			args: args{
				components: []*Component{New("c3"), New("c4")},
			},
			assertions: func(t *testing.T, collection Collection) {
				assert.Len(t, collection, 4)
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
			tt.collection.Add(tt.args.components...)
			if tt.assertions != nil {
				tt.assertions(t, tt.collection)
			}
		})
	}
}
