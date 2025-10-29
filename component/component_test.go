package component

import (
	"testing"

	"github.com/hovsep/fmesh/labels"
	"github.com/hovsep/fmesh/port"
	"github.com/stretchr/testify/assert"
)

func TestNewComponent(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want *Component
	}{
		{
			name: "empty name is valid",
			args: args{
				name: "",
			},
			want: New(""),
		},
		{
			name: "with name",
			args: args{
				name: "multiplier",
			},
			want: New("multiplier"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, New(tt.args.name))
		})
	}
}

func TestComponent_WithDescription(t *testing.T) {
	type args struct {
		description string
	}
	tests := []struct {
		name      string
		component *Component
		args      args
		want      *Component
	}{
		{
			name:      "happy path",
			component: New("c1"),
			args: args{
				description: "descr",
			},
			want: &Component{
				name:         "c1",
				description:  "descr",
				labels:       labels.NewCollection(nil),
				chainableErr: nil,
				inputs: port.NewCollection().WithDefaultLabels(labels.Map{
					port.DirectionLabel: port.DirectionIn,
				}),
				outputs: port.NewCollection().WithDefaultLabels(labels.Map{
					port.DirectionLabel: port.DirectionOut,
				}),
				f:     nil,
				state: NewState(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.component.WithDescription(tt.args.description))
		})
	}
}

func TestComponent_SetLabels(t *testing.T) {
	tests := []struct {
		name       string
		component  *Component
		labels     labels.Map
		assertions func(t *testing.T, component *Component)
	}{
		{
			name:      "set labels on new component",
			component: New("c1"),
			labels: labels.Map{
				"l1": "v1",
				"l2": "v2",
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, 2, component.Labels().Len())
				assert.True(t, component.labels.HasAll("l1", "l2"))
			},
		},
		{
			name:      "set labels replaces existing labels",
			component: New("c1").AddLabels(labels.Map{"old": "value"}),
			labels: labels.Map{
				"l1": "v1",
				"l2": "v2",
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, 2, component.Labels().Len())
				assert.True(t, component.labels.HasAll("l1", "l2"))
				assert.False(t, component.labels.Has("old"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			componentAfter := tt.component.SetLabels(tt.labels)
			if tt.assertions != nil {
				tt.assertions(t, componentAfter)
			}
		})
	}
}

func TestComponent_AddLabels(t *testing.T) {
	tests := []struct {
		name       string
		component  *Component
		labels     labels.Map
		assertions func(t *testing.T, component *Component)
	}{
		{
			name:      "add labels to new component",
			component: New("c1"),
			labels: labels.Map{
				"l1": "v1",
				"l2": "v2",
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, 2, component.Labels().Len())
				assert.True(t, component.labels.HasAll("l1", "l2"))
			},
		},
		{
			name:      "add labels merges with existing",
			component: New("c1").AddLabels(labels.Map{"existing": "label"}),
			labels: labels.Map{
				"l1": "v1",
				"l2": "v2",
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, 3, component.Labels().Len())
				assert.True(t, component.labels.HasAll("existing", "l1", "l2"))
			},
		},
		{
			name:      "add labels updates existing key",
			component: New("c1").AddLabels(labels.Map{"l1": "old"}),
			labels: labels.Map{
				"l1": "new",
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, 1, component.Labels().Len())
				assert.True(t, component.labels.ValueIs("l1", "new"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			componentAfter := tt.component.AddLabels(tt.labels)
			if tt.assertions != nil {
				tt.assertions(t, componentAfter)
			}
		})
	}
}
