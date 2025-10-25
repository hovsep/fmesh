package component

import (
	"fmt"
	"testing"

	"github.com/hovsep/fmesh/labels"
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
			want: New("").WithErr(fmt.Errorf("%w, component name: %s", errNotFound, "c3")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.components.ByName(tt.args.name))
		})
	}
}

func TestCollection_ByLabelValue(t *testing.T) {
	type args struct {
		labelName  string
		labelValue string
	}
	tests := []struct {
		name       string
		components *Collection
		args       args
		want       *Collection
	}{
		{
			name:       "no labels, nothing found",
			components: NewCollection().With(New("c1"), New("c2")),
			args: args{
				labelName:  "version",
				labelValue: "v100",
			},
			want: NewCollection(),
		},
		{
			name: "no relevant labels, nothing found",
			components: NewCollection().With(New("c1").
				WithLabels(labels.Map{
					"l1": "v1",
					"l2": "v2",
				}),
				New("c2").
					WithLabels(labels.Map{
						"l1": "v1",
						"l2": "v2",
					})),
			args: args{
				labelName:  "version",
				labelValue: "v100",
			},
			want: NewCollection(),
		},
		{
			name: "found one",
			components: NewCollection().With(New("c1").
				WithLabels(labels.Map{
					"version": "v1",
				}),
				New("c2").
					WithLabels(labels.Map{
						"version": "v2",
					}),
				New("c3").
					WithLabels(labels.Map{
						"version": "v3",
					}),
				New("c4").
					WithLabels(labels.Map{
						"version": "v4",
					})),

			args: args{
				labelName:  "version",
				labelValue: "v2",
			},
			want: NewCollection().With(New("c2").
				WithLabels(labels.Map{
					"version": "v2",
				})),
		},
		{
			name: "found several",
			components: NewCollection().With(New("c1").
				WithLabels(labels.Map{
					"env": "stage",
				}),
				New("c2").
					WithLabels(labels.Map{
						"env": "prod",
					}),
				New("c3").
					WithLabels(labels.Map{
						"env": "stage",
					}),
				New("c4").
					WithLabels(labels.Map{
						"env": "prod",
					})),

			args: args{
				labelName:  "env",
				labelValue: "prod",
			},
			want: NewCollection().With(New("c2").
				WithLabels(labels.Map{
					"env": "prod",
				}),
				New("c4").
					WithLabels(labels.Map{
						"env": "prod",
					})),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.components.ByLabelValue(tt.args.labelName, tt.args.labelValue))
		})
	}
}

func TestCollection_One(t *testing.T) {
	tests := []struct {
		name       string
		components *Collection
		assertions func(t *testing.T, component *Component)
	}{
		{
			name:       "empty collection",
			components: NewCollection(),
			assertions: func(t *testing.T, component *Component) {
				t.Helper()
				assert.True(t, component.HasErr())
				require.Error(t, component.Err())
				require.ErrorIs(t, component.Err(), errNotFound)
			},
		},
		{
			name:       "one component",
			components: NewCollection().With(New("c1")),
			assertions: func(t *testing.T, component *Component) {
				t.Helper()
				assert.False(t, component.HasErr())
				assert.Equal(t, "c1", component.Name())
			},
		},
		{
			name:       "multiple components",
			components: NewCollection().With(New("c1"), New("c2"), New("c3")),
			assertions: func(t *testing.T, component *Component) {
				t.Helper()
				assert.False(t, component.HasErr())
				require.NoError(t, component.Err())
				// As map iteration is not determined - any component can be returned,so we do not check for name
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.components.One()
			if tt.assertions != nil {
				tt.assertions(t, got)
			}
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
